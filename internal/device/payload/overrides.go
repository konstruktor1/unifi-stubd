// Package payload applies operator-specified port state, speed, media, counters,
// identity, and role metadata after profile generation. Setter order is explicit
// because speed, media, and up/down state interact in controller-visible ways.
package payload

import (
	"fmt"
	"net"
	"strings"
)

// ApplyPortOverrides applies per-port overrides to ports.
func ApplyPortOverrides(ports []Port, overrides []PortOverride) []Port {
	if len(overrides) == 0 || len(ports) == 0 {
		return ports
	}
	for _, override := range overrides {
		if override.Port < 1 || override.Port > len(ports) {
			continue
		}
		port := &ports[override.Port-1]
		for _, setter := range portOverrideSetters {
			setter(port, override)
		}
	}
	return ports
}

// ClonePortOverrides returns a detached copy of per-port runtime overrides.
func ClonePortOverrides(overrides []PortOverride) []PortOverride {
	if len(overrides) == 0 {
		return nil
	}
	out := make([]PortOverride, len(overrides))
	for index, override := range overrides {
		out[index] = override
		if override.Up != nil {
			up := *override.Up
			out[index].Up = &up
		}
	}
	return out
}

// ValidatePortOverride validates one per-port runtime override.
func ValidatePortOverride(override PortOverride, portCount int) error {
	if override.Port < 1 || override.Port > portCount {
		return fmt.Errorf("invalid port override %d; use 1..%d", override.Port, portCount)
	}
	if override.Speed < 0 {
		return fmt.Errorf("invalid speed override %d on port %d; use 0 or a positive Mbps value", override.Speed, override.Port)
	}
	for _, field := range portOverrideStringFields {
		if err := field.validate(override); err != nil {
			return err
		}
	}
	if PortOverrideEmpty(override) {
		return fmt.Errorf("empty port override on port %d", override.Port)
	}
	return nil
}

// PortOverrideEmpty reports whether override has no effective field besides port.
func PortOverrideEmpty(override PortOverride) bool {
	if !portOverrideStringsEmpty(override) {
		return false
	}
	return override.Speed == 0 &&
		override.Up == nil &&
		!override.Disabled &&
		override.RXBytes == 0 &&
		override.TXBytes == 0 &&
		override.RXPackets == 0 &&
		override.TXPackets == 0 &&
		override.RXErrors == 0 &&
		override.TXErrors == 0 &&
		!override.TrafficRatesSet
}

// NormalizePortOverride returns a copy suitable for status and plan output.
func NormalizePortOverride(override PortOverride) PortOverride {
	for _, field := range portOverrideStringFields {
		field.setOverride(&override, field.normalize(field.get(override)))
	}
	if override.Up != nil {
		up := *override.Up
		override.Up = &up
	}
	return override
}

func validGatewayRole(role string) bool {
	switch role {
	case gatewayPortRoleWAN, gatewayPortRoleLAN, gatewayPortRoleWAN2, gatewayPortRoleLAN2:
		return true
	default:
		return false
	}
}

type portOverrideSetter func(*Port, PortOverride)

var portOverrideSetters = []portOverrideSetter{
	setPortOverrideStrings,
	setPortOverrideSpeed,
	setPortOverrideCounters,
	setPortOverrideRates,
	setPortOverrideMedia,
	setPortOverrideLinkState,
	setPortOverrideDisabled,
}

func setPortOverrideStrings(port *Port, override PortOverride) {
	for _, field := range portOverrideStringFields {
		if field.applyAfterSpeed {
			continue
		}
		field.setPort(port, field.normalize(field.get(override)))
	}
}

func setPortOverrideSpeed(port *Port, override PortOverride) {
	if override.Speed <= 0 {
		return
	}
	port.Speed = override.Speed
	if strings.TrimSpace(override.Media) == "" {
		port.Media = mediaForSpeed(override.Speed)
	}
}

func setPortOverrideCounters(port *Port, override PortOverride) {
	for _, binding := range portCounterOverrides {
		if value := binding.get(override); value != 0 {
			binding.set(port, value)
		}
	}
}

func setPortOverrideRates(port *Port, override PortOverride) {
	if !override.TrafficRatesSet {
		return
	}
	port.RXBytesRate = override.RXBytesRate
	port.TXBytesRate = override.TXBytesRate
	port.TrafficRatesEnabled = true
	port.TrafficRatesSet = true
}

type portCounterOverride struct {
	get func(PortOverride) int64
	set func(*Port, int64)
}

var portCounterOverrides = []portCounterOverride{
	{func(override PortOverride) int64 { return override.RXBytes }, func(port *Port, value int64) { port.RXBytes = value }},
	{func(override PortOverride) int64 { return override.TXBytes }, func(port *Port, value int64) { port.TXBytes = value }},
	{func(override PortOverride) int64 { return override.RXPackets }, func(port *Port, value int64) { port.RXPackets = value }},
	{func(override PortOverride) int64 { return override.TXPackets }, func(port *Port, value int64) { port.TXPackets = value }},
	{func(override PortOverride) int64 { return override.RXErrors }, func(port *Port, value int64) { port.RXErrors = value }},
	{func(override PortOverride) int64 { return override.TXErrors }, func(port *Port, value int64) { port.TXErrors = value }},
}

func setPortOverrideMedia(port *Port, override PortOverride) {
	for _, field := range portOverrideStringFields {
		if field.applyAfterSpeed {
			field.setPort(port, field.normalize(field.get(override)))
		}
	}
}

func setPortOverrideLinkState(port *Port, override PortOverride) {
	if override.Up == nil {
		return
	}
	port.Up = *override.Up
	if !*override.Up && override.Speed <= 0 {
		port.Speed = 0
	}
}

func setPortOverrideDisabled(port *Port, override PortOverride) {
	if !override.Disabled {
		return
	}
	port.Disabled = true
	port.Up = false
	port.Speed = 0
	port.MACs = nil
}

type portOverrideStringField struct {
	get             func(PortOverride) string
	setOverride     func(*PortOverride, string)
	setPort         func(*Port, string)
	normalize       func(string) string
	validateValue   func(PortOverride, string) error
	applyAfterSpeed bool
}

var portOverrideStringFields = []portOverrideStringField{
	{
		get:         func(override PortOverride) string { return override.Name },
		setOverride: func(override *PortOverride, value string) { override.Name = value },
		setPort:     func(port *Port, value string) { setNonEmptyString(value, func() { port.Name = value }) },
		normalize:   strings.TrimSpace,
	},
	{
		get:         func(override PortOverride) string { return override.Interface },
		setOverride: func(override *PortOverride, value string) { override.Interface = value },
		setPort:     func(port *Port, value string) { setNonEmptyString(value, func() { port.Interface = value }) },
		normalize:   strings.TrimSpace,
		validateValue: func(override PortOverride, value string) error {
			if strings.Contains(value, "/") {
				return fmt.Errorf("invalid interface override %q on port %d", value, override.Port)
			}
			return nil
		},
	},
	{
		get:         func(override PortOverride) string { return override.MAC },
		setOverride: func(override *PortOverride, value string) { override.MAC = value },
		setPort:     func(port *Port, value string) { setNonEmptyString(value, func() { port.MAC = value }) },
		normalize:   lowerTrimmed,
		validateValue: func(override PortOverride, value string) error {
			if _, err := net.ParseMAC(value); err != nil {
				return fmt.Errorf("invalid port override mac %q on port %d: %w", value, override.Port, err)
			}
			return nil
		},
	},
	{
		get:         func(override PortOverride) string { return override.IP },
		setOverride: func(override *PortOverride, value string) { override.IP = value },
		setPort:     func(port *Port, value string) { setNonEmptyString(value, func() { port.IP = value }) },
		normalize:   strings.TrimSpace,
		validateValue: func(override PortOverride, value string) error {
			if net.ParseIP(value).To4() == nil {
				return fmt.Errorf("invalid port override ip %q on port %d", value, override.Port)
			}
			return nil
		},
	},
	{
		get:         func(override PortOverride) string { return override.Netmask },
		setOverride: func(override *PortOverride, value string) { override.Netmask = value },
		setPort:     func(port *Port, value string) { setNonEmptyString(value, func() { port.Netmask = value }) },
		normalize:   strings.TrimSpace,
		validateValue: func(override PortOverride, value string) error {
			if net.ParseIP(value).To4() == nil {
				return fmt.Errorf("invalid port override netmask %q on port %d", value, override.Port)
			}
			return nil
		},
	},
	{
		get:         func(override PortOverride) string { return override.Role },
		setOverride: func(override *PortOverride, value string) { override.Role = value },
		setPort:     func(port *Port, value string) { setNonEmptyString(value, func() { port.Role = value }) },
		normalize:   normalizeGatewayRole,
		validateValue: func(override PortOverride, value string) error {
			if !validGatewayRole(value) {
				return fmt.Errorf("invalid port override role %q on port %d; use wan, lan, wan2, or lan2", override.Role, override.Port)
			}
			return nil
		},
	},
	{
		get:         func(override PortOverride) string { return override.NetworkGroup },
		setOverride: func(override *PortOverride, value string) { override.NetworkGroup = value },
		setPort:     func(port *Port, value string) { setNonEmptyString(value, func() { port.NetworkGroup = value }) },
		normalize:   normalizeGatewayNetworkGroup,
		validateValue: func(override PortOverride, value string) error {
			if strings.ContainsAny(value, "\r\n\t") {
				return fmt.Errorf("invalid port override network_group %q on port %d", value, override.Port)
			}
			return nil
		},
	},
	{
		get:             func(override PortOverride) string { return override.Media },
		setOverride:     func(override *PortOverride, value string) { override.Media = value },
		setPort:         func(port *Port, value string) { setNonEmptyString(value, func() { port.Media = value }) },
		normalize:       strings.TrimSpace,
		applyAfterSpeed: true,
	},
}

func (field portOverrideStringField) validate(override PortOverride) error {
	value := field.normalize(field.get(override))
	if value == "" || field.validateValue == nil {
		return nil
	}
	return field.validateValue(override, value)
}

func portOverrideStringsEmpty(override PortOverride) bool {
	for _, field := range portOverrideStringFields {
		if field.normalize(field.get(override)) != "" {
			return false
		}
	}
	return true
}

func lowerTrimmed(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func setNonEmptyString(value string, set func()) {
	if value != "" {
		set()
	}
}

package device

import (
	"fmt"
	"net"
	"strings"
)

type portOverrideNormalizer int

const (
	portOverrideTrimSpace portOverrideNormalizer = iota
	portOverrideLowerTrimmed
	portOverrideNormalizeRole
	portOverrideNormalizeNetworkGroup
)

type portOverrideTextKey int

const (
	portOverrideTextName portOverrideTextKey = iota
	portOverrideTextInterface
	portOverrideTextMAC
	portOverrideTextIP
	portOverrideTextNetmask
	portOverrideTextRole
	portOverrideTextNetworkGroup
	portOverrideTextMedia
	portOverrideTextPortConfID
	portOverrideTextNetworkConfID
	portOverrideTextNativeNetworkConfID
	portOverrideTextNetworkName
)

type portOverrideValidator int

const (
	portOverrideValidateNone portOverrideValidator = iota
	portOverrideValidateInterface
	portOverrideValidateMAC
	portOverrideValidateIPv4
	portOverrideValidateNetmask
	portOverrideValidateRole
	portOverrideValidateNetworkGroup
)

// portOverrideStringField centralizes normalization, validation, and application
// rules for text-like override fields.
type portOverrideStringField struct {
	key             portOverrideTextKey
	normalizer      portOverrideNormalizer
	validator       portOverrideValidator
	applyAfterSpeed bool
}

func (field portOverrideStringField) get(override PortOverride) string {
	switch field.key {
	case portOverrideTextName:
		return override.Name
	case portOverrideTextInterface:
		return override.Interface
	case portOverrideTextMAC:
		return override.MAC
	case portOverrideTextIP:
		return override.IP
	case portOverrideTextNetmask:
		return override.Netmask
	case portOverrideTextRole:
		return override.Role
	case portOverrideTextNetworkGroup:
		return override.NetworkGroup
	case portOverrideTextMedia:
		return override.Media
	case portOverrideTextPortConfID:
		return override.PortConfID
	case portOverrideTextNetworkConfID:
		return override.NetworkConfID
	case portOverrideTextNativeNetworkConfID:
		return override.NativeNetworkConfID
	case portOverrideTextNetworkName:
		return override.NetworkName
	default:
		return ""
	}
}

func (field portOverrideStringField) setOverride(override *PortOverride, value string) {
	switch field.key {
	case portOverrideTextName:
		override.Name = value
	case portOverrideTextInterface:
		override.Interface = value
	case portOverrideTextMAC:
		override.MAC = value
	case portOverrideTextIP:
		override.IP = value
	case portOverrideTextNetmask:
		override.Netmask = value
	case portOverrideTextRole:
		override.Role = value
	case portOverrideTextNetworkGroup:
		override.NetworkGroup = value
	case portOverrideTextMedia:
		override.Media = value
	case portOverrideTextPortConfID:
		override.PortConfID = value
	case portOverrideTextNetworkConfID:
		override.NetworkConfID = value
	case portOverrideTextNativeNetworkConfID:
		override.NativeNetworkConfID = value
	case portOverrideTextNetworkName:
		override.NetworkName = value
	}
}

func (field portOverrideStringField) setPort(port *Port, value string) {
	setNonEmptyString(value, func() {
		switch field.key {
		case portOverrideTextName:
			port.Name = value
		case portOverrideTextInterface:
			port.Interface = value
		case portOverrideTextMAC:
			port.MAC = value
		case portOverrideTextIP:
			port.IP = value
		case portOverrideTextNetmask:
			port.Netmask = value
		case portOverrideTextRole:
			port.Role = value
		case portOverrideTextNetworkGroup:
			port.NetworkGroup = value
		case portOverrideTextMedia:
			port.Media = value
		case portOverrideTextPortConfID:
			port.PortConfID = value
		case portOverrideTextNetworkConfID:
			port.NetworkConfID = value
		case portOverrideTextNativeNetworkConfID:
			port.NativeNetworkConfID = value
		case portOverrideTextNetworkName:
			port.NetworkName = value
		}
	})
}

func (field portOverrideStringField) normalize(value string) string {
	switch field.normalizer {
	case portOverrideLowerTrimmed:
		return strings.ToLower(strings.TrimSpace(value))
	case portOverrideNormalizeRole:
		return normalizeGatewayRole(value)
	case portOverrideNormalizeNetworkGroup:
		return normalizeGatewayNetworkGroup(value)
	default:
		return strings.TrimSpace(value)
	}
}

// validate checks one normalized text override against the field-specific
// payload policy.
func (field portOverrideStringField) validate(override PortOverride) error {
	value := field.normalize(field.get(override))
	if value == "" || field.validator == portOverrideValidateNone {
		return nil
	}
	switch field.validator {
	case portOverrideValidateNone:
		return nil
	case portOverrideValidateInterface:
		if strings.Contains(value, "/") {
			return fmt.Errorf("invalid interface override %q on port %d", value, override.Port)
		}
	case portOverrideValidateMAC:
		if _, err := net.ParseMAC(value); err != nil {
			return fmt.Errorf("invalid port override mac %q on port %d: %w", value, override.Port, err)
		}
	case portOverrideValidateIPv4:
		if net.ParseIP(value).To4() == nil {
			return fmt.Errorf("invalid port override ip %q on port %d", value, override.Port)
		}
	case portOverrideValidateNetmask:
		if net.ParseIP(value).To4() == nil {
			return fmt.Errorf("invalid port override netmask %q on port %d", value, override.Port)
		}
	case portOverrideValidateRole:
		if !validGatewayRole(value) {
			return fmt.Errorf("invalid port override role %q on port %d; use wan, lan, wan2, lan2, or unassigned", override.Role, override.Port)
		}
	case portOverrideValidateNetworkGroup:
		if strings.ContainsAny(value, "\r\n\t") {
			return fmt.Errorf("invalid port override network_group %q on port %d", value, override.Port)
		}
	}
	return nil
}

// portOverrideStringsEmpty checks normalized text fields before validation
// decides whether an override has any effect.
func portOverrideStringsEmpty(override PortOverride) bool {
	for _, field := range portOverrideStringFields {
		if field.normalize(field.get(override)) != "" {
			return false
		}
	}
	return true
}

// validGatewayRole keeps override roles aligned with the gateway renderer's
// known WAN/LAN role set.
func validGatewayRole(role string) bool {
	switch role {
	case gatewayPortRoleWAN, gatewayPortRoleLAN, gatewayPortRoleWAN2, gatewayPortRoleLAN2, gatewayPortRoleNone:
		return true
	default:
		return false
	}
}

// setNonEmptyString applies optional text overrides only when a normalized value
// is present.
func setNonEmptyString(value string, set func()) {
	if value != "" {
		set()
	}
}

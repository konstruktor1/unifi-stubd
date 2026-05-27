// Package device applies operator-specified port state, speed, media, counters,
// identity, and role metadata after profile generation.
package device

import "fmt"

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
		out[index].Up = cloneBoolRef(override.Up)
		out[index].WANUptimePercent = cloneFloat64Ref(override.WANUptimePercent)
		out[index].WANConnected = cloneBoolRef(override.WANConnected)
	}
	return out
}

// PortOverridesFromWANHealthResults converts active WAN health samples into
// health-only port overrides. It intentionally leaves role, assignment,
// addressing, VLAN, and link-state fields unset.
func PortOverridesFromWANHealthResults(results []WANHealthResult) []PortOverride {
	if len(results) == 0 {
		return nil
	}
	out := make([]PortOverride, 0, len(results))
	for _, result := range results {
		if result.Port < 1 {
			continue
		}
		connected := result.Connected
		uptime := clampFloat64(result.UptimePercent, 0, 100)
		out = append(out, PortOverride{
			Port:               result.Port,
			WANConnected:       &connected,
			WANLatencyMS:       nonNegativeInt(result.LatencyMS),
			WANDowntimeSeconds: nonNegativeInt(result.DowntimeSeconds),
			WANUptimePercent:   &uptime,
		})
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
	if override.VLAN < 0 {
		return fmt.Errorf("invalid vlan override %d on port %d; use 0 or a positive VLAN ID", override.VLAN, override.Port)
	}
	if override.WANUptimePercent != nil && (*override.WANUptimePercent < 0 || *override.WANUptimePercent > 100) {
		return fmt.Errorf("invalid wan_uptime_percent override %.2f on port %d; use 0..100", *override.WANUptimePercent, override.Port)
	}
	if override.WANLatencyMS < 0 {
		return fmt.Errorf("invalid wan_latency_ms override %d on port %d; use 0 or a positive millisecond value", override.WANLatencyMS, override.Port)
	}
	if override.WANDowntimeSeconds < 0 {
		return fmt.Errorf("invalid wan_downtime_seconds override %d on port %d; use 0 or a positive seconds value", override.WANDowntimeSeconds, override.Port)
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
		override.VLAN == 0 &&
		override.WANUptimePercent == nil &&
		override.WANLatencyMS == 0 &&
		override.WANDowntimeSeconds == 0 &&
		override.WANConnected == nil &&
		!override.TrafficRatesSet
}

// NormalizePortOverride returns a copy suitable for status and plan output.
func NormalizePortOverride(override PortOverride) PortOverride {
	for _, field := range portOverrideStringFields {
		field.setOverride(&override, field.normalize(field.get(override)))
	}
	override.Up = cloneBoolRef(override.Up)
	override.WANUptimePercent = cloneFloat64Ref(override.WANUptimePercent)
	override.WANConnected = cloneBoolRef(override.WANConnected)
	return override
}

package device

import "strings"

// portOverrideSetter applies one ordered group of override fields.
type portOverrideSetter func(*Port, PortOverride)

// setPortOverrideAssignment mirrors explicit controller assignment metadata
// into the runtime port model. These fields describe controller state; they do
// not apply host networking or mutate the controller.
func setPortOverrideAssignment(port *Port, override PortOverride) {
	port.VLAN = override.VLAN
}

// setPortOverrideStrings applies identity and role text before speed/media
// dependent fields are resolved.
func setPortOverrideStrings(port *Port, override PortOverride) {
	for _, field := range portOverrideStringFields {
		if field.applyAfterSpeed {
			continue
		}
		field.setPort(port, field.normalize(field.get(override)))
	}
}

// setPortOverrideSpeed applies explicit speed before media so default media can
// still be replaced by a later explicit media override.
func setPortOverrideSpeed(port *Port, override PortOverride) {
	if override.Speed <= 0 {
		return
	}
	port.Speed = override.Speed
	// A speed override also implies the controller media label unless the
	// operator supplied an explicit media override later in the setter order.
	if strings.TrimSpace(override.Media) == "" {
		port.Media = mediaForSpeed(override.Speed)
	}
}

// setPortOverrideWANHealth applies lab-only WAN reachability hints. These are
// telemetry fields and do not change host routes or controller provisioning.
func setPortOverrideWANHealth(port *Port, override PortOverride) {
	port.WANUptimePercent = cloneFloat64Ref(override.WANUptimePercent)
	port.WANLatencyMS = override.WANLatencyMS
	port.WANDowntimeSeconds = override.WANDowntimeSeconds
	port.WANConnected = cloneBoolRef(override.WANConnected)
}

// setPortOverrideCounters overlays non-zero operator or observation counters
// onto the generated port.
func setPortOverrideCounters(port *Port, override PortOverride) {
	for _, binding := range portCounterOverrides {
		if value := binding.get(override); value != 0 {
			binding.set(port, value)
		}
	}
}

// setPortOverrideRates marks operator-provided or observed byte rates as
// explicit so the renderer does not synthesize heartbeat rates for that port.
func setPortOverrideRates(port *Port, override PortOverride) {
	if !override.TrafficRatesSet {
		return
	}
	port.RXBytesRate = override.RXBytesRate
	port.TXBytesRate = override.TXBytesRate
	port.TrafficRatesEnabled = true
	port.TrafficRatesSet = true
}

// setPortOverrideMedia applies explicit media after speed so it can override
// the speed-derived label.
func setPortOverrideMedia(port *Port, override PortOverride) {
	for _, field := range portOverrideStringFields {
		if field.applyAfterSpeed {
			field.setPort(port, field.normalize(field.get(override)))
		}
	}
}

// setPortOverrideLinkState applies explicit up/down state after counters and
// media so disconnected ports render consistently.
func setPortOverrideLinkState(port *Port, override PortOverride) {
	if override.Up == nil {
		return
	}
	port.Up = *override.Up
	if !*override.Up && override.Speed <= 0 {
		// Link-down without an explicit speed should render as disconnected, not
		// as a forced-speed port that happens to be down.
		port.Speed = 0
	}
}

// setPortOverrideDisabled is the final override step because disabling a port
// must clear link, speed, and learned MAC state after all other metadata merges.
func setPortOverrideDisabled(port *Port, override PortOverride) {
	if !override.Disabled {
		return
	}
	port.Disabled = true
	port.Up = false
	port.Speed = 0
	port.MACs = nil
}

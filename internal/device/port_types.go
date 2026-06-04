package device

// MacTableEntry represents a learned MAC entry for a switch port.
type MacTableEntry struct {
	// MAC is the learned client or neighbor MAC address.
	MAC string `json:"mac"`
	// Hostname is the optional controller-facing client name.
	Hostname string `json:"hostname,omitempty"`
	// IP is the optional client IPv4 address.
	IP string `json:"ip,omitempty"`
	// Age is the controller-facing age counter for this entry.
	Age int `json:"age"`
	// Uptime is the number of seconds the entry has been visible.
	Uptime int `json:"uptime"`
	// VLAN is the optional VLAN associated with the entry.
	VLAN int `json:"vlan,omitempty"`
	// Static reports that the entry is a configured lab hint, not a learned row.
	Static bool `json:"static,omitempty"`
	// Type describes the learned device type when known.
	Type string `json:"type,omitempty"`
}

// Port describes one fake switch port in the UniFi payload.
type Port struct {
	// Index is the one-based UniFi port index.
	Index int
	// Name is the display name reported for the port.
	Name string
	// Interface is the optional local host interface that supplied this port's
	// data. Gateway renderers expose it as source_interface, never as the
	// controller-facing ifname.
	Interface string
	// MAC is the optional interface MAC address reported for this port.
	MAC string
	// IP is the optional IPv4 address reported for this port.
	IP string
	// Netmask is the optional IPv4 netmask reported for this port.
	Netmask string
	// Role is the effective gateway function, such as wan, lan, wan2, or lan2.
	// It can differ from ProfileRole when a lab uses a different physical port.
	Role string
	// ProfileRole is the immutable gateway role assigned by the selected
	// profile. It records the hardware default even when Role is overridden.
	ProfileRole string
	// NetworkGroup is the UniFi network group, such as WAN, WAN2, or LAN.
	NetworkGroup string
	// PortConfID is the optional UniFi port profile ID mirrored from controller config.
	PortConfID string
	// NetworkConfID is the optional UniFi network ID mirrored from controller config.
	NetworkConfID string
	// NativeNetworkConfID is the optional native network ID mirrored from controller config.
	NativeNetworkConfID string
	// NetworkName is the optional controller-facing network name.
	NetworkName string
	// VLAN is the optional controller-facing VLAN ID.
	VLAN int
	// WANUptimePercent is the controller-facing WAN availability percentage.
	// It is telemetry only and must not drive host networking.
	WANUptimePercent *float64
	// WANLatencyMS is the controller-facing WAN latency in milliseconds.
	WANLatencyMS int
	// WANDowntimeSeconds is the controller-facing WAN downtime counter.
	WANDowntimeSeconds int
	// WANConnected overrides WAN reachability independently from link state.
	WANConnected *bool
	// Media is the UniFi media label, such as GE or SFP+.
	Media string
	// Uplink marks the currently active upstream connection.
	Uplink bool
	// ProfileUplink marks a profile-defined uplink-capable port group, such as
	// dedicated SFP/SFP+ cages. It does not imply current link direction.
	ProfileUplink bool
	// Disabled reports that the port should be administratively disabled.
	Disabled bool
	// Up reports whether link is up.
	Up bool
	// Speed is the negotiated speed in Mbps.
	Speed int
	// RXBytes is the receive byte counter.
	RXBytes int64
	// TXBytes is the transmit byte counter.
	TXBytes int64
	// RXPackets is the receive packet counter.
	RXPackets int64
	// TXPackets is the transmit packet counter.
	TXPackets int64
	// RXErrors is the receive error counter.
	RXErrors int64
	// TXErrors is the transmit error counter.
	TXErrors int64
	// RXBytesRate is the receive byte rate in bytes per second.
	RXBytesRate int64
	// TXBytesRate is the transmit byte rate in bytes per second.
	TXBytesRate int64
	// TrafficRatesEnabled reports that the runtime traffic-rate switch is on.
	TrafficRatesEnabled bool
	// TrafficRatesSet reports that RXBytesRate and TXBytesRate are real samples.
	TrafficRatesSet bool
	// MACs contains learned MAC entries for this port.
	MACs []MacTableEntry
}

// PortGroup describes one contiguous block in a switch port layout.
type PortGroup struct {
	// Count is the number of ports in this block.
	Count int `yaml:"count"`
	// Speed is the negotiated speed in Mbps for ports in this block.
	Speed int `yaml:"speed"`
	// Media is the UniFi media label for ports in this block.
	Media string `yaml:"media"`
	// Uplink marks this block as the profile-defined uplink-capable group; the
	// first port is used as the active upstream default.
	Uplink bool `yaml:"uplink"`
}

// PortBuildOptions configures runtime changes to profile-derived ports.
type PortBuildOptions struct {
	// Count overrides the profile port count when positive.
	Count int
	// LinkSpeed overrides all profile port speeds when positive.
	LinkSpeed int
	// UplinkSpeed overrides only the active uplink speed when positive.
	UplinkSpeed int
	// UplinkPort overrides the generated uplink port when positive.
	UplinkPort int
}

// PortNeighbor describes one configured MAC-table entry on a specific port.
type PortNeighbor struct {
	// Port is the one-based switch port index.
	Port int
	// Entry is the controller-facing MAC table entry to expose.
	Entry MacTableEntry
}

// Package payload defines the neutral data model shared by profiles,
// observation, and JSON renderers. These types describe what should be reported,
// not how a specific controller table is encoded.
package payload

// Profile contains profile-driven inform payload rendering metadata.
type Profile struct {
	// Kind selects the generic payload renderer: switch or gateway.
	Kind string
	// RequiredVersion is reported in the inform payload.
	RequiredVersion string
	// ManagementInterface is the controller-facing management interface name.
	ManagementInterface string
	// GatewayInterfacePrefix prefixes generated gateway interface names.
	GatewayInterfacePrefix string
	// HasDPI reports whether gateway DPI capability should be advertised.
	HasDPI bool
}

// Identity contains the device attributes reported in inform payloads.
type Identity struct {
	// MAC is the fake device MAC address in controller-facing text form.
	MAC string
	// IP is the device management IP address reported to UniFi.
	IP string
	// Hostname is the device name reported to UniFi.
	Hostname string
	// Model is the UniFi model identifier.
	Model string
	// ModelDisplay is the human-readable UniFi model name.
	ModelDisplay string
	// DeviceType is the controller-facing UniFi device family.
	DeviceType string
	// Version is the firmware version reported by the stub.
	Version string
	// Serial is the serial number reported by the stub.
	Serial string
	// InformURL is the controller inform URL currently known by the device.
	InformURL string
	// InformIP is the numeric controller inform endpoint address reported to UniFi.
	InformIP string
	// CFGVersion is the controller configuration version applied to the device.
	CFGVersion string
	// ManagementVLAN is the optional controller-facing management VLAN ID.
	ManagementVLAN int
	// UptimeSeconds is the monotonic runtime uptime reported in inform payloads.
	UptimeSeconds int
	// Adopted reports whether the stub should present itself as adopted.
	Adopted bool
}

// MacTableEntry represents a learned MAC entry for a switch port.
type MacTableEntry struct {
	// MAC is the learned client or neighbor MAC address.
	MAC string `json:"mac"`
	// Age is the controller-facing age counter for this entry.
	Age int `json:"age"`
	// Uptime is the number of seconds the entry has been visible.
	Uptime int `json:"uptime"`
	// VLAN is the optional VLAN associated with the entry.
	VLAN int `json:"vlan,omitempty"`
	// Type describes the learned device type when known.
	Type string `json:"type,omitempty"`
}

// Port describes one fake switch port in the UniFi payload.
type Port struct {
	// Index is the one-based UniFi port index.
	Index int
	// Name is the display name reported for the port.
	Name string
	// Interface is the optional host interface that supplied this port's data.
	Interface string
	// MAC is the optional interface MAC address reported for this port.
	MAC string
	// IP is the optional IPv4 address reported for this port.
	IP string
	// Netmask is the optional IPv4 netmask reported for this port.
	Netmask string
	// Role is the gateway-facing role, such as wan, lan, wan2, or lan2.
	Role string
	// NetworkGroup is the UniFi network group, such as WAN, WAN2, or LAN.
	NetworkGroup string
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
	// MACs contains learned MAC entries for this port.
	MACs []MacTableEntry
}

// PortGroup describes one contiguous block in a switch port layout.
type PortGroup struct {
	// Count is the number of ports in this block.
	Count int
	// Speed is the negotiated speed in Mbps for ports in this block.
	Speed int
	// Media is the UniFi media label for ports in this block.
	Media string
	// Uplink marks this block as the profile-defined uplink-capable group; the
	// first port is used as the active upstream default.
	Uplink bool
}

// PortOptions configures generated switch port defaults.
type PortOptions struct {
	// Speed is the default access port speed in Mbps.
	Speed int
	// UplinkSpeed is the uplink port speed in Mbps.
	UplinkSpeed int
	// Media is the default access port media label.
	Media string
	// UplinkMedia is the uplink port media label.
	UplinkMedia string
	// UplinkPort overrides the generated uplink port when positive.
	UplinkPort int
	// PortGroups optionally describe a non-uniform physical port layout.
	PortGroups []PortGroup
	// PortNames optionally override one-based port display labels.
	PortNames []string
	// PortRoles optionally assign one-based gateway roles.
	PortRoles []string
	// PortNetworkGroups optionally assign one-based UniFi network groups.
	PortNetworkGroups []string
}

// PortOverride describes one per-port runtime override.
type PortOverride struct {
	// Port is the one-based switch port index.
	Port int `yaml:"port" json:"port"`
	// Name overrides the controller-facing port label when set.
	Name string `yaml:"name" json:"name,omitempty"`
	// Interface names the optional host interface used as a passive source.
	Interface string `yaml:"interface" json:"interface,omitempty"`
	// MAC overrides the controller-facing interface MAC when set.
	MAC string `yaml:"mac" json:"mac,omitempty"`
	// IP overrides the controller-facing interface IPv4 address when set.
	IP string `yaml:"ip" json:"ip,omitempty"`
	// Netmask overrides the controller-facing interface IPv4 netmask when set.
	Netmask string `yaml:"netmask" json:"netmask,omitempty"`
	// Role overrides the gateway-facing role when set.
	Role string `yaml:"role" json:"role,omitempty"`
	// NetworkGroup overrides the UniFi network group when set.
	NetworkGroup string `yaml:"network_group" json:"network_group,omitempty"`
	// Speed overrides the negotiated speed in Mbps when positive.
	Speed int `yaml:"speed" json:"speed,omitempty"`
	// Media overrides the controller-facing media label when set.
	Media string `yaml:"media" json:"media,omitempty"`
	// Up overrides link state when set.
	Up *bool `yaml:"up" json:"up,omitempty"`
	// Disabled administratively disables the rendered port. It is set by
	// port-map disabled entries, not by controller provisioning.
	Disabled bool `yaml:"-" json:"disabled,omitempty"`
	// RXBytes overrides the receive byte counter when non-zero.
	RXBytes int64 `yaml:"-" json:"-"`
	// TXBytes overrides the transmit byte counter when non-zero.
	TXBytes int64 `yaml:"-" json:"-"`
	// RXPackets overrides the receive packet counter when non-zero.
	RXPackets int64 `yaml:"-" json:"-"`
	// TXPackets overrides the transmit packet counter when non-zero.
	TXPackets int64 `yaml:"-" json:"-"`
	// RXErrors overrides the receive error counter when non-zero.
	RXErrors int64 `yaml:"-" json:"-"`
	// TXErrors overrides the transmit error counter when non-zero.
	TXErrors int64 `yaml:"-" json:"-"`
}

// PortNeighbor describes one configured MAC-table entry on a specific port.
type PortNeighbor struct {
	// Port is the one-based switch port index.
	Port int
	// Entry is the controller-facing MAC table entry to expose.
	Entry MacTableEntry
}

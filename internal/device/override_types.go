package device

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
	// IPv6 overrides the controller-facing interface IPv6 CIDR addresses when set.
	IPv6 []string `yaml:"ipv6" json:"ipv6,omitempty"`
	// Role overrides the gateway-facing role when set.
	Role string `yaml:"role" json:"role,omitempty"`
	// NetworkGroup overrides the UniFi network group when set.
	NetworkGroup string `yaml:"network_group" json:"network_group,omitempty"`
	// PortConfID mirrors the UniFi port profile ID into gateway port tables when set.
	PortConfID string `yaml:"portconf_id" json:"portconf_id,omitempty"`
	// NetworkConfID mirrors the UniFi network ID into gateway port tables when set.
	NetworkConfID string `yaml:"networkconf_id" json:"networkconf_id,omitempty"`
	// NativeNetworkConfID mirrors the native network ID into gateway port tables when set.
	NativeNetworkConfID string `yaml:"native_networkconf_id" json:"native_networkconf_id,omitempty"`
	// NetworkName mirrors the controller-facing network name into gateway port tables when set.
	NetworkName string `yaml:"network_name" json:"network_name,omitempty"`
	// VLAN mirrors the controller-facing VLAN ID into gateway port tables when positive.
	VLAN int `yaml:"vlan" json:"vlan,omitempty"`
	// WANUptimePercent mirrors gateway WAN availability into status telemetry.
	WANUptimePercent *float64 `yaml:"wan_uptime_percent" json:"wan_uptime_percent,omitempty"`
	// WANLatencyMS mirrors gateway WAN latency into status telemetry.
	WANLatencyMS int `yaml:"wan_latency_ms" json:"wan_latency_ms,omitempty"`
	// WANDowntimeSeconds mirrors gateway WAN downtime into status telemetry.
	WANDowntimeSeconds int `yaml:"wan_downtime_seconds" json:"wan_downtime_seconds,omitempty"`
	// WANConnected overrides gateway WAN reachability independently from link state.
	WANConnected *bool `yaml:"wan_connected" json:"wan_connected,omitempty"`
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
	// RXBytesRate overrides the receive byte rate when TrafficRatesSet is true.
	RXBytesRate int64 `yaml:"-" json:"-"`
	// TXBytesRate overrides the transmit byte rate when TrafficRatesSet is true.
	TXBytesRate int64 `yaml:"-" json:"-"`
	// TrafficRatesSet reports that RXBytesRate and TXBytesRate are real samples.
	TrafficRatesSet bool `yaml:"-" json:"-"`
}

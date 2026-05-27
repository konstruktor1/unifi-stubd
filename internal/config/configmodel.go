package config

import "github.com/konstruktor1/unifi-stubd/internal/device"

// ManagementLAN describes controller-facing switch management VLAN behavior.
type ManagementLAN struct {
	// Enabled opts into the structured management LAN model.
	Enabled bool `yaml:"enabled"`
	// VLAN is the management VLAN ID. 0 leaves it unset.
	VLAN int `yaml:"vlan"`
	// NetworkName is an optional controller-facing label for status and docs.
	NetworkName string `yaml:"network_name"`
	// Mode selects metadata-only, preexisting-interface, or planned-host-vlan.
	Mode string `yaml:"mode"`
	// Interface is the existing host VLAN interface used by preexisting-interface.
	Interface string `yaml:"interface"`
	// IP optionally pins the management IP instead of reading the interface IPv4.
	IP string `yaml:"ip"`
	// ControllerReachable controls optional live reachability validation.
	ControllerReachable string `yaml:"controller_reachable"`
	// AdoptionStrategy documents whether adoption starts untagged or tagged.
	AdoptionStrategy string `yaml:"adoption_strategy"`
}

// UplinkNeighbor describes a configured fake upstream neighbor.
type UplinkNeighbor struct {
	// MAC is the neighbor MAC address to expose on the uplink port.
	MAC string `yaml:"mac"`
	// Name is an optional lab label reported as the client hostname.
	Name string `yaml:"name"`
	// Hostname is the optional controller-facing client hostname.
	Hostname string `yaml:"hostname"`
	// IP is the optional client IPv4 address.
	IP string `yaml:"ip"`
	// VLAN is the optional VLAN associated with the neighbor.
	VLAN int `yaml:"vlan"`
	// Static reports the neighbor as a configured lab hint.
	Static bool `yaml:"static"`
	// Type is the controller-facing neighbor type.
	Type string `yaml:"type"`
	// Age is the controller-facing MAC-table age counter.
	Age int `yaml:"age"`
	// Uptime is the number of seconds the neighbor has been visible.
	Uptime int `yaml:"uptime"`
}

// PortNeighbor describes a configured fake neighbor on a specific port.
type PortNeighbor struct {
	// Port is the one-based switch port index.
	Port int `yaml:"port"`
	// MAC is the neighbor MAC address to expose on the port.
	MAC string `yaml:"mac"`
	// Name is an optional lab label reported as the client hostname.
	Name string `yaml:"name"`
	// Hostname is the optional controller-facing client hostname.
	Hostname string `yaml:"hostname"`
	// IP is the optional client IPv4 address.
	IP string `yaml:"ip"`
	// VLAN is the optional VLAN associated with the neighbor.
	VLAN int `yaml:"vlan"`
	// Static reports the neighbor as a configured lab hint.
	Static bool `yaml:"static"`
	// Type is the controller-facing neighbor type.
	Type string `yaml:"type"`
	// Age is the controller-facing MAC-table age counter.
	Age int `yaml:"age"`
	// Uptime is the number of seconds the neighbor has been visible.
	Uptime int `yaml:"uptime"`
}

// PortOverride describes one per-port YAML override.
type PortOverride = device.PortOverride

// WANHealthConfig configures optional read-only gateway WAN health reporting.
// The fields are operator YAML only; controller provisioning does not enable
// probes or write these settings.
type WANHealthConfig struct {
	// Source selects off, static, or ping.
	Source string `yaml:"source"`
	// IntervalSeconds is the sampling interval used for active probes.
	IntervalSeconds int `yaml:"interval_seconds"`
	// TimeoutMS is the per-target active probe timeout.
	TimeoutMS int `yaml:"timeout_ms"`
	// Targets maps active probe hosts to represented WAN ports.
	Targets []WANHealthTarget `yaml:"targets"`
}

// WANHealthTarget maps one active health probe to one represented WAN port.
type WANHealthTarget struct {
	// Port is the one-based UniFi port index.
	Port int `yaml:"port"`
	// Host is the IP or hostname passed to the local ping command.
	Host string `yaml:"host"`
}

// BridgeObserve describes a bridge represented as one virtual switch.
type BridgeObserve struct {
	// Bridge is the host bridge whose learned MAC table is observed.
	Bridge string `yaml:"bridge"`
	// UplinkInterface is the bridge member that points upstream.
	UplinkInterface string `yaml:"uplink_interface"`
	// IgnoredMembers are bridge members excluded from UniFi port mapping.
	IgnoredMembers []string `yaml:"ignored_members"`
	// MemberPortMap pins bridge members to one-based UniFi ports.
	MemberPortMap []BridgeMemberPortMap `yaml:"member_port_map"`
}

// BridgeMemberPortMap pins one bridge member interface to one UniFi port.
type BridgeMemberPortMap struct {
	// Member is the bridge member interface name, such as tap101i0.
	Member string `yaml:"member"`
	// Port is the one-based UniFi port index.
	Port int `yaml:"port"`
}

// PortMapping maps one UniFi port to a physical interface or explicit state.
type PortMapping struct {
	// Port is the one-based UniFi port index.
	Port int `yaml:"port"`
	// Interface is the physical OS interface used as the source for this port.
	Interface string `yaml:"interface"`
	// Disabled reports this port as link-down with speed 0.
	Disabled bool `yaml:"disabled"`
	// Unmapped leaves this port on profile defaults without an observation source.
	Unmapped bool `yaml:"unmapped"`
}

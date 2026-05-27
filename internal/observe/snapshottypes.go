package observe

import "github.com/konstruktor1/unifi-stubd/internal/device"

// Config selects passive Linux observation sources.
type Config struct {
	// Interface is the host interface used for counters and link speed.
	Interface string
	// Bridge is the Linux bridge used for FDB MAC table data.
	Bridge string
	// IgnoredMembers excludes bridge member interfaces from UniFi port mapping.
	IgnoredMembers []string
	// MemberPortMap pins bridge member interfaces to one-based UniFi ports.
	MemberPortMap map[string]int
	// SysfsRoot is the sysfs root, usually /sys.
	SysfsRoot string
}

// InterfaceStats contains passive counters and link speed for one interface.
type InterfaceStats struct {
	// RXBytes is the received byte counter.
	RXBytes int64 `json:"rx_bytes,omitempty"`
	// TXBytes is the transmitted byte counter.
	TXBytes int64 `json:"tx_bytes,omitempty"`
	// RXPackets is the received packet counter.
	RXPackets int64 `json:"rx_packets,omitempty"`
	// TXPackets is the transmitted packet counter.
	TXPackets int64 `json:"tx_packets,omitempty"`
	// RXErrors is the receive error counter.
	RXErrors int64 `json:"rx_errors,omitempty"`
	// TXErrors is the transmit error counter.
	TXErrors int64 `json:"tx_errors,omitempty"`
	// SpeedMbps is the reported link speed in Mbps.
	SpeedMbps int `json:"speed_mbps,omitempty"`
}

// Snapshot contains passive data that can be merged into generated switch ports.
type Snapshot struct {
	// UplinkPortIndex is the one-based target port for uplink observations.
	UplinkPortIndex int
	// Interface is the observed host interface name.
	Interface string
	// Bridge is the observed Linux bridge name.
	Bridge string
	// Stats contains counters and link speed from the observed interface.
	Stats InterfaceStats
	// MACs contains learned MAC entries flattened for the uplink fallback.
	MACs []device.MacTableEntry
	// DeviceMACs contains learned MAC entries grouped by bridge member.
	DeviceMACs map[string][]device.MacTableEntry
	// RemoteMACs contains MACs learned behind the physical uplink neighbor.
	RemoteMACs map[string]bool
	// MemberPorts contains observed interface state grouped by bridge member.
	MemberPorts map[string]PortObservation
	// MemberPortMap pins bridge member interfaces to one-based UniFi ports.
	MemberPortMap map[string]int
	// MemberRoles classifies bridge members before they are mapped to ports.
	MemberRoles map[string]BridgeMemberRole
}

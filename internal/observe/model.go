// Package observe normalizes Linux bridge data, FreeBSD bridge data, and
// explicit port-map sources before payload merge. It describes read-only facts
// about the host, not host-network changes to perform.
package observe

import (
	"context"

	"github.com/konstruktor1/unifi-stubd/internal/device"
)

// ObservationSource reads host network state without mutating the host.
type ObservationSource interface {
	Bridge(context.Context, BridgeConfig) (BridgeObservation, []error)
	Ports(context.Context, PortMapConfig) (PortMapObservation, []error)
}

// BridgeConfig selects a host bridge represented as a virtual switch.
type BridgeConfig struct {
	Bridge          string
	UplinkInterface string
	MemberPortMap   map[string]int
}

// BridgeMemberRole describes how a bridge member should map to UniFi ports.
type BridgeMemberRole string

const (
	// BridgeMemberRoleUnknown keeps a member eligible for normal port mapping
	// when the platform cannot classify it safely.
	BridgeMemberRoleUnknown BridgeMemberRole = "unknown"
	// BridgeMemberRoleUplink maps the member to the selected UniFi uplink port.
	BridgeMemberRoleUplink BridgeMemberRole = "uplink"
	// BridgeMemberRoleAccess maps the member to an access/downstream port.
	BridgeMemberRoleAccess BridgeMemberRole = "access"
	// BridgeMemberRoleBridge marks the bridge device itself; it is metadata and
	// must not consume a UniFi port.
	BridgeMemberRoleBridge BridgeMemberRole = "bridge"
	// BridgeMemberRoleIgnored excludes a member from payload port mapping.
	BridgeMemberRoleIgnored BridgeMemberRole = "ignored"
)

// PortMapConfig selects explicit host-interface sources for UniFi ports.
type PortMapConfig struct {
	Mappings []PortMapping
}

// PortMapping maps one UniFi port to a host interface or explicit state.
type PortMapping struct {
	Port      int
	Interface string
	Disabled  bool
	Unmapped  bool
}

// BridgeObservation contains bridge member MACs and optional uplink stats.
type BridgeObservation struct {
	Bridge          string
	UplinkInterface string
	Uplink          PortObservation
	MemberMACs      map[string][]device.MacTableEntry
	RemoteMACs      map[string]bool
	MemberPorts     map[string]PortObservation
	MemberPortMap   map[string]int
	MemberRoles     map[string]BridgeMemberRole
}

// PortMapObservation contains explicit per-port physical observations.
type PortMapObservation struct {
	Ports map[int]PortObservation
}

// PortObservation contains the physical state observed for one source port.
type PortObservation struct {
	Port      int
	Interface string
	MAC       string
	IP        string
	Netmask   string
	Up        *bool
	SpeedMbps int
	Media     string
	Stats     InterfaceStats
	MACs      []device.MacTableEntry
	Disabled  bool
	Unmapped  bool
}

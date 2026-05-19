package device

import (
	"fmt"

	payloadpkg "github.com/konstruktor1/unifi-stubd/internal/device/payload"
)

// Identity contains the device attributes reported in inform payloads.
type Identity = payloadpkg.Identity

// MacTableEntry represents a learned MAC entry for a switch port.
type MacTableEntry = payloadpkg.MacTableEntry

// Port describes one fake switch port in the UniFi payload.
type Port = payloadpkg.Port

// PortGroup describes one contiguous block in a switch port layout.
type PortGroup = payloadpkg.PortGroup

// PortOptions configures generated switch port defaults.
type PortOptions = payloadpkg.PortOptions

// PortOverride describes one per-port runtime override.
type PortOverride = payloadpkg.PortOverride

// PortNeighbor describes one configured MAC-table entry on a specific port.
type PortNeighbor = payloadpkg.PortNeighbor

// BuildPayload returns a JSON inform payload using profile-driven renderer metadata.
func BuildPayload(profile Profile, id Identity, ports []Port) ([]byte, error) {
	data, err := payloadpkg.BuildPayload(payloadProfile(profile.PayloadOptions()), id, ports)
	if err != nil {
		return nil, fmt.Errorf("build device payload: %w", err)
	}
	return data, nil
}

// MinimalSwitchPayload returns a JSON inform payload with a switch-shaped port table.
func MinimalSwitchPayload(id Identity, ports []Port) ([]byte, error) {
	data, err := payloadpkg.MinimalSwitchPayload(id, ports)
	if err != nil {
		return nil, fmt.Errorf("build minimal switch payload: %w", err)
	}
	return data, nil
}

// SwitchPorts returns count generated switch ports with profile-neutral defaults.
func SwitchPorts(count int) []Port {
	return payloadpkg.SwitchPorts(count)
}

// SwitchPortsWithOptions returns count generated switch ports using options.
func SwitchPortsWithOptions(count int, options PortOptions) []Port {
	return payloadpkg.SwitchPortsWithOptions(count, options)
}

// ApplyPortOverrides applies per-port overrides to ports.
func ApplyPortOverrides(ports []Port, overrides []PortOverride) []Port {
	return payloadpkg.ApplyPortOverrides(ports, overrides)
}

// ApplyUplinkNeighbor adds a configured neighbor entry to the uplink port.
func ApplyUplinkNeighbor(ports []Port, neighbor *MacTableEntry) []Port {
	return payloadpkg.ApplyUplinkNeighbor(ports, neighbor)
}

// ApplyPortNeighbors adds configured MAC-table entries to their target ports.
func ApplyPortNeighbors(ports []Port, neighbors []PortNeighbor) []Port {
	return payloadpkg.ApplyPortNeighbors(ports, neighbors)
}

func payloadProfile(profile PayloadProfile) payloadpkg.Profile {
	return payloadpkg.Profile{
		Kind:                   profile.Kind,
		RequiredVersion:        profile.RequiredVersion,
		ManagementInterface:    profile.ManagementInterface,
		GatewayInterfacePrefix: profile.GatewayInterfacePrefix,
		HasDPI:                 profile.HasDPI,
	}
}

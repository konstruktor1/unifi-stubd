package main

import (
	"fmt"
	"net"
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/device"
)

// validatePortOverrides checks configured payload metadata before it can reach
// generated ports, keeping invalid MAC/IP/speed data out of inform payloads.
func validatePortOverrides(flags runtimeFlags, profile device.Profile) error {
	if flags.uplinkPort < 0 || flags.uplinkPort > flags.portCount {
		return fmt.Errorf("invalid -uplink-port %d; use 0 or 1..%d", flags.uplinkPort, flags.portCount)
	}
	if flags.uplinkNeighbor != nil {
		if _, err := net.ParseMAC(flags.uplinkNeighbor.MAC); err != nil {
			return fmt.Errorf("invalid uplink_neighbor mac %q: %w", flags.uplinkNeighbor.MAC, err)
		}
		if flags.uplinkNeighbor.VLAN < 0 {
			return fmt.Errorf("invalid uplink_neighbor vlan %d; use 0 or a positive VLAN ID", flags.uplinkNeighbor.VLAN)
		}
		if ip := strings.TrimSpace(flags.uplinkNeighbor.IP); ip != "" && net.ParseIP(ip).To4() == nil {
			return fmt.Errorf("invalid uplink_neighbor ip %q; use an IPv4 address", flags.uplinkNeighbor.IP)
		}
		if flags.uplinkNeighbor.RemotePort < 0 || flags.uplinkNeighbor.RemotePort > 65535 {
			return fmt.Errorf("invalid uplink_neighbor remote_port %d; use 0 or 1..65535", flags.uplinkNeighbor.RemotePort)
		}
	}
	for _, neighbor := range flags.portNeighbors {
		if neighbor.Port < 1 || neighbor.Port > flags.portCount {
			return fmt.Errorf("invalid port neighbor %d; use 1..%d", neighbor.Port, flags.portCount)
		}
		if _, err := net.ParseMAC(neighbor.Entry.MAC); err != nil {
			return fmt.Errorf("invalid port neighbor mac %q on port %d: %w", neighbor.Entry.MAC, neighbor.Port, err)
		}
		if neighbor.Entry.VLAN < 0 {
			return fmt.Errorf("invalid port neighbor vlan %d on port %d; use 0 or a positive VLAN ID", neighbor.Entry.VLAN, neighbor.Port)
		}
		if ip := strings.TrimSpace(neighbor.Entry.IP); ip != "" && net.ParseIP(ip).To4() == nil {
			return fmt.Errorf("invalid port neighbor ip %q on port %d; use an IPv4 address", neighbor.Entry.IP, neighbor.Port)
		}
		if neighbor.RemotePort < 0 || neighbor.RemotePort > 65535 {
			return fmt.Errorf("invalid port neighbor remote_port %d on port %d; use 0 or 1..65535", neighbor.RemotePort, neighbor.Port)
		}
	}
	for _, override := range flags.portOverrides {
		if err := device.ValidatePortOverride(override, flags.portCount); err != nil {
			return fmt.Errorf("validate port overrides: %w", err)
		}
	}
	ports := device.BuildPorts(profile, device.PortBuildOptions{
		Count:      flags.portCount,
		LinkSpeed:  flags.linkSpeed,
		UplinkPort: effectiveUplinkPort(profile, flags),
	})
	ports = device.ApplyPortNeighbors(ports, flags.portNeighbors)
	ports = device.ApplyUplinkNeighbor(ports, flags.uplinkNeighbor)
	if _, err := device.ResolveTopology(ports); err != nil {
		return fmt.Errorf("validate topology neighbors: %w", err)
	}
	return nil
}

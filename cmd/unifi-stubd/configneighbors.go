package main

import (
	"strings"

	appconfig "github.com/konstruktor1/unifi-stubd/internal/config"
	"github.com/konstruktor1/unifi-stubd/internal/device"
)

// configUplinkNeighbor converts YAML neighbor metadata into the payload-facing
// MAC-table entry used on the represented uplink port.
func configUplinkNeighbor(neighbor *appconfig.UplinkNeighbor) *device.MacTableEntry {
	if neighbor == nil || strings.TrimSpace(neighbor.MAC) == "" {
		return nil
	}
	return &device.MacTableEntry{
		MAC:      strings.TrimSpace(neighbor.MAC),
		Hostname: defaultNeighborHostname(neighbor.Hostname, neighbor.Name),
		IP:       strings.TrimSpace(neighbor.IP),
		Age:      defaultNeighborAge(neighbor.Age),
		Uptime:   defaultNeighborUptime(neighbor.Uptime),
		VLAN:     neighbor.VLAN,
		Static:   neighbor.Static,
		Type:     defaultNeighborType(neighbor.Type),
	}
}

// configPortNeighbors converts YAML per-port neighbor metadata into payload
// MAC-table entries without requiring live observation.
func configPortNeighbors(neighbors []appconfig.PortNeighbor) []device.PortNeighbor {
	out := make([]device.PortNeighbor, 0, len(neighbors))
	for _, neighbor := range neighbors {
		if strings.TrimSpace(neighbor.MAC) == "" {
			continue
		}
		out = append(out, device.PortNeighbor{
			Port: neighbor.Port,
			Entry: device.MacTableEntry{
				MAC:      strings.TrimSpace(neighbor.MAC),
				Hostname: defaultNeighborHostname(neighbor.Hostname, neighbor.Name),
				IP:       strings.TrimSpace(neighbor.IP),
				Age:      defaultNeighborAge(neighbor.Age),
				Uptime:   defaultNeighborUptime(neighbor.Uptime),
				VLAN:     neighbor.VLAN,
				Static:   neighbor.Static,
				Type:     defaultPortNeighborType(neighbor.Type),
			},
		})
	}
	return out
}

// defaultNeighborAge supplies a fresh-looking MAC-table age for synthetic
// configured neighbors.
func defaultNeighborAge(age int) int {
	if age == 0 {
		return 4
	}
	return age
}

// defaultNeighborUptime supplies a stable non-zero uptime for configured
// neighbors shown in controller topology views.
func defaultNeighborUptime(uptime int) int {
	if uptime == 0 {
		return 1200
	}
	return uptime
}

// defaultNeighborType marks the configured uplink neighbor as switch-like when
// the operator did not provide a more specific topology type.
func defaultNeighborType(neighborType string) string {
	neighborType = strings.TrimSpace(neighborType)
	if neighborType == "" {
		return "usw"
	}
	return neighborType
}

// defaultPortNeighborType marks per-port configured neighbors as clients unless
// the operator supplied a specific topology type.
func defaultPortNeighborType(neighborType string) string {
	neighborType = strings.TrimSpace(neighborType)
	if neighborType == "" {
		return "client"
	}
	return neighborType
}

// defaultNeighborHostname accepts both the new hostname field and the legacy
// name alias used by older config examples.
func defaultNeighborHostname(hostname, name string) string {
	if hostname = strings.TrimSpace(hostname); hostname != "" {
		return hostname
	}
	return strings.TrimSpace(name)
}

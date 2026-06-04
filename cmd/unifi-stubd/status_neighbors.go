package main

import "github.com/konstruktor1/unifi-stubd/internal/device"

// statusUplinkNeighborEntry copies configured uplink neighbor metadata into the
// stable status schema.
func statusUplinkNeighborEntry(neighbor *device.MacTableEntry) *statusUplinkNeighbor {
	if neighbor == nil {
		return nil
	}
	return &statusUplinkNeighbor{
		MAC:      neighbor.MAC,
		Hostname: neighbor.Hostname,
		IP:       neighbor.IP,
		VLAN:     neighbor.VLAN,
		Static:   neighbor.Static,
		Type:     neighbor.Type,
		Age:      neighbor.Age,
		Uptime:   neighbor.Uptime,
	}
}

// statusPortNeighbors converts configured per-port neighbors into status rows
// without exposing renderer-only fields.
func statusPortNeighbors(neighbors []device.PortNeighbor) []statusPortNeighbor {
	out := make([]statusPortNeighbor, 0, len(neighbors))
	for _, neighbor := range neighbors {
		out = append(out, statusPortNeighbor{
			Port:     neighbor.Port,
			MAC:      neighbor.Entry.MAC,
			Hostname: neighbor.Entry.Hostname,
			IP:       neighbor.Entry.IP,
			VLAN:     neighbor.Entry.VLAN,
			Static:   neighbor.Entry.Static,
			Type:     neighbor.Entry.Type,
			Age:      neighbor.Entry.Age,
			Uptime:   neighbor.Entry.Uptime,
		})
	}
	return out
}

// statusPortOverrides normalizes override text for display while keeping the
// runtime override slice detached.
func statusPortOverrides(overrides []device.PortOverride) []device.PortOverride {
	out := device.ClonePortOverrides(overrides)
	for index := range out {
		out[index] = device.NormalizePortOverride(out[index])
	}
	return out
}

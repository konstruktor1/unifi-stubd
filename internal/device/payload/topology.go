package payload

import (
	"strconv"
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/device"
)

type lldpRow struct {
	ChassisID     string `json:"chassis_id"`
	IsWired       bool   `json:"is_wired"`
	LocalPortIdx  int    `json:"local_port_idx"`
	LocalPortName string `json:"local_port_name"`
	PortID        string `json:"port_id,omitempty"`
}

func lldpTable(ports []PortView) []lldpRow {
	out := make([]lldpRow, 0)
	for _, port := range ports {
		if port.TopologyNeighbor == nil {
			continue
		}
		selectedMAC := strings.ToLower(strings.TrimSpace(port.TopologyNeighbor.MAC))
		for _, neighbor := range port.Neighbors {
			if neighbor.Source != device.NeighborSourceLLDP ||
				strings.ToLower(strings.TrimSpace(neighbor.MAC)) != selectedMAC {
				continue
			}
			portID := strings.TrimSpace(neighbor.RemotePortID)
			if portID == "" && neighbor.RemotePort > 0 {
				portID = "Port " + strconv.Itoa(neighbor.RemotePort)
			}
			out = append(out, lldpRow{
				ChassisID:     strings.ToLower(strings.TrimSpace(neighbor.MAC)),
				IsWired:       true,
				LocalPortIdx:  port.Index,
				LocalPortName: port.Name,
				PortID:        portID,
			})
		}
	}
	return out
}

func topologyMACEntry(neighbor device.TopologyNeighbor) device.MacTableEntry {
	return device.MacTableEntry{
		MAC:      strings.ToLower(strings.TrimSpace(neighbor.MAC)),
		Hostname: strings.TrimSpace(neighbor.Hostname),
		IP:       strings.TrimSpace(neighbor.IP),
		Age:      firstNonZero(neighbor.Age, 4),
		Uptime:   firstNonZero(neighbor.Uptime, 1200),
		VLAN:     neighbor.VLAN,
		Static:   neighbor.Static,
		Type:     strings.TrimSpace(neighbor.DeviceType),
	}
}

func downstreamMACTable(view PortView) []device.MacTableEntry {
	out := make([]device.MacTableEntry, 0, len(view.MACs)+1)
	if view.TopologyNeighbor != nil && view.TopologyDirection == device.DirectionDownstream {
		out = append(out, topologyMACEntry(*view.TopologyNeighbor))
	}
	out = append(out, view.MACs...)
	return out
}

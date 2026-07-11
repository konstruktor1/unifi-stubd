package payload

import (
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/device"
)

func gatewayConnectionFields(view PortView) connectionFields {
	out := connectionFields{Connected: view.Up}
	if !view.Up || !view.Uplink || view.TopologyNeighbor == nil || view.TopologyDirection != device.DirectionUpstream {
		return out
	}
	entry := *view.TopologyNeighbor
	connection := gatewayLastConnection{
		MAC:    strings.ToLower(strings.TrimSpace(entry.MAC)),
		Source: jsonKeyMACTable,
	}
	if ip := strings.TrimSpace(entry.IP); ip != "" {
		connection.IP = ip
	}
	if hostname := strings.TrimSpace(entry.Hostname); hostname != "" {
		connection.Hostname = hostname
	}
	if entryType := strings.TrimSpace(entry.DeviceType); entryType != "" {
		connection.Type = entryType
	}
	out.LastConnection = &connection
	return out
}

func physicalConnectionFields(view PortView) connectionFields {
	out := connectionFields{Connected: view.Up}
	if !view.Up {
		return out
	}
	if view.TopologyNeighbor != nil && view.TopologyDirection == device.DirectionDownstream {
		entry := *view.TopologyNeighbor
		connection := gatewayLastConnection{
			MAC:    strings.ToLower(strings.TrimSpace(entry.MAC)),
			Source: jsonKeyMACTable,
		}
		if ip := strings.TrimSpace(entry.IP); ip != "" {
			connection.IP = ip
		}
		if hostname := strings.TrimSpace(entry.Hostname); hostname != "" {
			connection.Hostname = hostname
		}
		if entryType := strings.TrimSpace(entry.DeviceType); entryType != "" {
			connection.Type = entryType
		}
		out.LastConnection = &connection
		return out
	}
	entry, ok := firstClientMAC(view.MACs)
	if !ok {
		return out
	}
	connection := gatewayLastConnection{
		MAC:    strings.ToLower(strings.TrimSpace(entry.MAC)),
		Source: jsonKeyMACTable,
	}
	if ip := strings.TrimSpace(entry.IP); ip != "" {
		connection.IP = ip
	}
	if hostname := strings.TrimSpace(entry.Hostname); hostname != "" {
		connection.Hostname = hostname
	}
	if entryType := strings.TrimSpace(entry.Type); entryType != "" {
		connection.Type = entryType
	}
	out.LastConnection = &connection
	return out
}

func firstClientMAC(entries []device.MacTableEntry) (device.MacTableEntry, bool) {
	for _, entry := range entries {
		if strings.TrimSpace(entry.MAC) != "" {
			return entry, true
		}
	}
	return device.MacTableEntry{}, false
}

func portMACTable(view PortView) []device.MacTableEntry {
	out := downstreamMACTable(view)
	if len(out) == 0 {
		return nil
	}
	return out
}

func gatewayHostTable(view PortView) []gatewayHostRow {
	port := view.Port
	out := make([]gatewayHostRow, 0, len(port.MACs))
	for _, entry := range port.MACs {
		entryType := strings.TrimSpace(entry.Type)
		if !gatewayHostEntryVisible(entryType) {
			continue
		}
		row := gatewayHostRow{
			MAC:        strings.ToLower(strings.TrimSpace(entry.MAC)),
			Age:        entry.Age,
			Authorized: true,
			RXBytes:    port.RXBytes,
			TXBytes:    port.TXBytes,
			RXPackets:  packetCounter(port.RXPackets, port.CountersSet),
			TXPackets:  packetCounter(port.TXPackets, port.CountersSet),
			Uptime:     firstNonZero(entry.Uptime, 1200),
		}
		if hostname := strings.TrimSpace(entry.Hostname); hostname != "" {
			row.Hostname = hostname
		}
		if ip := strings.TrimSpace(entry.IP); ip != "" {
			row.IP = ip
		}
		if entryType != "" {
			row.Type = entryType
		}
		if entry.VLAN > 0 {
			row.VLAN = entry.VLAN
		}
		if entry.Static {
			row.Static = true
		}
		out = append(out, row)
	}
	return out
}

func gatewayHostEntryVisible(entryType string) bool {
	entryType = strings.ToLower(strings.TrimSpace(entryType))
	return entryType == "" || entryType == deviceTypeClient
}

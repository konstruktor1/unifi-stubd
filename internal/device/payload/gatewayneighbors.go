package payload

import (
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/device"
)

// gatewayConnectionFields uses the resolved port view as the single source for
// gateway connection and topology state.
func gatewayConnectionFields(view PortView) connectionFields {
	out := connectionFields{Connected: view.Up}
	if !view.Up || !view.Uplink {
		return out
	}
	// Controllers use last_connection as a topology hint. The first MAC entry
	// is therefore treated as metadata about the visible neighbor, not as host
	// configuration to apply.
	entry, ok := gatewayFirstTopologyMAC(view)
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

func gatewayPhysicalPortConnectionFields(view PortView) connectionFields {
	out := connectionFields{Connected: view.Up}
	if !view.Up {
		return out
	}
	entry, ok := gatewayFirstPhysicalPortTopologyMAC(view)
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

func gatewayFirstPhysicalPortTopologyMAC(view PortView) (device.MacTableEntry, bool) {
	for _, entry := range view.MACs {
		if gatewayMACEntryVisibleOnPhysicalGatewayPort(view, entry) {
			return entry, true
		}
	}
	return device.MacTableEntry{}, false
}

func gatewayFirstTopologyMAC(view PortView) (device.MacTableEntry, bool) {
	for _, entry := range view.MACs {
		if gatewayMACEntryVisibleOnGatewayPort(view, entry) {
			return entry, true
		}
	}
	return device.MacTableEntry{}, false
}

func gatewayPortMACTable(view PortView) []device.MacTableEntry {
	out := make([]device.MacTableEntry, 0, len(view.MACs))
	for _, entry := range view.MACs {
		if gatewayMACEntryVisibleOnPhysicalGatewayPort(view, entry) {
			out = append(out, entry)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func gatewayMACEntryVisibleOnGatewayPort(view PortView, entry device.MacTableEntry) bool {
	if view.Uplink {
		return true
	}
	entryType := strings.ToLower(strings.TrimSpace(entry.Type))
	return entryType == "" || entryType == deviceTypeClient
}

func gatewayMACEntryVisibleOnPhysicalGatewayPort(view PortView, entry device.MacTableEntry) bool {
	if strings.TrimSpace(entry.MAC) == "" {
		return false
	}
	if view.Uplink || gatewayPortActsLikeSwitchPort(view) {
		return true
	}
	entryType := strings.ToLower(strings.TrimSpace(entry.Type))
	return entryType == "" || entryType == deviceTypeClient
}

func gatewayPortActsLikeSwitchPort(view PortView) bool {
	switch gatewayPortRole(view.Port) {
	case gatewayPortRoleLAN, gatewayPortRoleLAN2:
		return true
	default:
		return false
	}
}

// gatewayHostTable renders learned downstream MACs for one gateway port.
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
			RXPackets:  firstNonZeroInt64(port.RXPackets, 1),
			TXPackets:  firstNonZeroInt64(port.TXPackets, 1),
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

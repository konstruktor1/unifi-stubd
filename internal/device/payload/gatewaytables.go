package payload

import (
	"strconv"
	"strings"
	"time"

	"github.com/konstruktor1/unifi-stubd/internal/device"
)

// gatewayPortTable renders read-only physical port inventory for controller UI
// surfaces. Operator-provided assignment IDs and VLAN bindings are mirrored as
// controller display hints only.
func gatewayPortTable(ports []PortView, uptime int) []gatewayPortRow {
	out := make([]gatewayPortRow, 0, len(ports))
	lanIP := gatewayFirstLANIP(ports)
	for _, view := range ports {
		iface := view.GatewayInterface
		ifname := gatewayPhysicalPortIfName(view)
		row := gatewayPortRow{
			PortIdx:                 view.Index,
			IfName:                  ifname,
			Name:                    ifname,
			MAC:                     strings.ToLower(strings.TrimSpace(iface.MAC)),
			IP:                      gatewayPortTableIP(view, lanIP),
			NetworkGroup:            iface.NetworkGroup,
			Role:                    view.Role,
			Type:                    "ethernet",
			NumPort:                 1,
			Enable:                  view.Enabled,
			Up:                      view.Up,
			Connected:               view.Up,
			IsUplink:                view.Uplink,
			OpMode:                  payloadModeSwitch,
			FullDuplex:              true,
			Autoneg:                 true,
			FlowctrlRX:              false,
			FlowctrlTX:              false,
			PortPOE:                 false,
			POEEnable:               false,
			POECaps:                 0,
			POEClass:                "Class 0",
			POEPower:                "0.00",
			MACTable:                gatewayPortMACTable(view),
			RXBroadcast:             0,
			RXMulticast:             0,
			RXDropped:               0,
			TXBroadcast:             0,
			TXMulticast:             0,
			TXDropped:               0,
			gatewayAssignmentFields: gatewayPortAssignmentFields(view),
			gatewayPortLinkFields:   gatewayPortLinkFieldsFor(view.Speed, view.Media),
			counterFields:           portCounterFields(view.Port),
			optionalRateFields:      explicitPortRateFields(view.Port),
			gatewayWANInlineHealth:  gatewayWANInlineHealthFor(view, uptime),
			SourceInterface:         view.SourceInterface,
			connectionFields:        gatewayPhysicalPortConnectionFields(view),
		}
		out = append(out, row)
	}
	return out
}

func gatewayFirstLANIP(ports []PortView) string {
	for _, role := range []string{gatewayPortRoleLAN, gatewayPortRoleLAN2} {
		for _, view := range ports {
			if gatewayPortRole(view.Port) != role {
				continue
			}
			if ip := strings.TrimSpace(view.GatewayInterface.IP); ip != "" && ip != gatewayNoIP {
				return ip
			}
		}
	}
	return ""
}

func gatewayPortTableIP(view PortView, lanIP string) string {
	ip := strings.TrimSpace(view.GatewayInterface.IP)
	if ip == "" || ip == gatewayNoIP {
		if strings.EqualFold(strings.TrimSpace(view.NetworkGroup), gatewayNetworkGroupNone) &&
			strings.TrimSpace(lanIP) != "" {
			return lanIP
		}
	}
	return ip
}

func gatewayPortStatsTable(ports []PortView) []gatewayPortStatsRow {
	out := make([]gatewayPortStatsRow, 0, len(ports))
	for _, view := range ports {
		iface := view.GatewayInterface
		ifname := gatewayPhysicalPortIfName(view)
		out = append(out, gatewayPortStatsRow{
			PortIdx:            view.Index,
			Name:               ifname,
			IfName:             ifname,
			MAC:                strings.ToLower(strings.TrimSpace(iface.MAC)),
			Type:               "ethernet",
			NumPort:            1,
			Media:              view.Media,
			Speed:              view.Speed,
			SpeedCaps:          gatewaySpeedCapsCode(view.Speed, view.Media),
			Enable:             view.Enabled,
			Up:                 view.Up,
			IsUplink:           view.Uplink,
			FullDuplex:         true,
			Autoneg:            true,
			FlowctrlRX:         false,
			FlowctrlTX:         false,
			PortPOE:            false,
			POEEnable:          false,
			POEClass:           "Class 0",
			POEPower:           "0.00",
			MACTable:           gatewayPortMACTable(view),
			RXBroadcast:        0,
			RXMulticast:        0,
			RXDropped:          0,
			TXBroadcast:        0,
			TXMulticast:        0,
			TXDropped:          0,
			counterFields:      portCounterFields(view.Port),
			optionalRateFields: explicitPortRateFields(view.Port),
		})
	}
	return out
}

// gatewayEthernetTable renders the stable physical port inventory expected by
// Network's gateway port surfaces.
func gatewayEthernetTable(ports []PortView) []gatewayEthernetTableRow {
	out := make([]gatewayEthernetTableRow, 0, len(ports))
	for _, view := range ports {
		iface := view.GatewayInterface
		ifname := gatewayPhysicalPortIfName(view)
		out = append(out, gatewayEthernetTableRow{
			MAC:       strings.ToLower(strings.TrimSpace(iface.MAC)),
			PortIdx:   view.Index,
			NumPort:   1,
			Name:      ifname,
			IfName:    ifname,
			Media:     view.Media,
			PortPOE:   false,
			SpeedCaps: gatewaySpeedCapsCode(view.Speed, view.Media),
		})
	}
	return out
}

// gatewayConfigPortTable renders controller-visible port assignment hints,
// including explicit controller assignment metadata when configured.
func gatewayConfigPortTable(ports []PortView, uptime int) []gatewayConfigPortRow {
	out := make([]gatewayConfigPortRow, 0, len(ports))
	for _, view := range ports {
		ifname := gatewayPhysicalPortIfName(view)
		out = append(out, gatewayConfigPortRow{
			Name:                    view.Name,
			IfName:                  ifname,
			PortIdx:                 view.Index,
			NetworkGroup:            view.NetworkGroup,
			Role:                    view.Role,
			Up:                      view.Up,
			Enable:                  view.Enabled,
			IsUplink:                view.Uplink,
			gatewayAssignmentFields: gatewayPortAssignmentFields(view),
			linkFields:              portLinkFields(view.Speed, view.Media),
			gatewayWANInlineHealth:  gatewayWANInlineHealthFor(view, uptime),
			SourceInterface:         view.SourceInterface,
			connectionFields:        gatewayPhysicalPortConnectionFields(view),
		})
	}
	return out
}

// gatewayEthernetOverrides renders interface binding hints that older tested
// gateway payloads exposed, including explicit assignment metadata.
func gatewayEthernetOverrides(ports []PortView, uptime int) []gatewayEthernetOverrideRow {
	out := make([]gatewayEthernetOverrideRow, 0, len(ports))
	complete := gatewayNeedsCompleteEthernetOverrides(ports)
	for _, view := range ports {
		if !gatewayIsLogicalRole(view.Role) {
			if complete && strings.EqualFold(strings.TrimSpace(view.NetworkGroup), gatewayNetworkGroupNone) {
				out = append(out, gatewayDisabledEthernetOverride(view))
			}
			continue
		}
		if !complete {
			switch gatewayPortRole(view.Port) {
			case gatewayPortRoleLAN, gatewayPortRoleLAN2:
			default:
				continue
			}
		}
		if strings.TrimSpace(gatewayPhysicalPortIfName(view)) == "" {
			continue
		}
		iface := view.GatewayInterface
		ifname := gatewayPhysicalPortIfName(view)
		out = append(out, gatewayEthernetOverrideRow{
			Name:                    ifname,
			IfName:                  ifname,
			PortIdx:                 view.Index,
			MAC:                     strings.ToLower(strings.TrimSpace(iface.MAC)),
			NetworkGroup:            view.NetworkGroup,
			Role:                    view.Role,
			Up:                      view.Up,
			Enable:                  view.Enabled,
			gatewayAssignmentFields: gatewayPortAssignmentFields(view),
			linkFields:              portLinkFields(view.Speed, view.Media),
			gatewayWANInlineHealth:  gatewayWANInlineHealthFor(view, uptime),
			SourceInterface:         view.SourceInterface,
			connectionFields:        gatewayPhysicalPortConnectionFields(view),
		})
	}
	return out
}

func gatewayNeedsCompleteEthernetOverrides(ports []PortView) bool {
	for _, view := range ports {
		port := view.Port
		if !gatewayIsLogicalRole(view.Role) &&
			strings.EqualFold(strings.TrimSpace(view.NetworkGroup), gatewayNetworkGroupNone) {
			return true
		}
		if strings.TrimSpace(port.PortConfID) != "" ||
			strings.TrimSpace(port.NetworkConfID) != "" ||
			strings.TrimSpace(port.NativeNetworkConfID) != "" {
			return true
		}
	}
	return false
}

func gatewayPhysicalPortIfName(view PortView) string {
	if ifname := strings.TrimSpace(view.PhysicalIfName); ifname != "" {
		return ifname
	}
	return strings.TrimSpace(view.GatewayInterface.IfName)
}

func gatewayDisabledEthernetOverride(view PortView) gatewayEthernetOverrideRow {
	ifname := gatewayPhysicalPortIfName(view)
	iface := view.GatewayInterface
	return gatewayEthernetOverrideRow{
		Name:            ifname,
		IfName:          ifname,
		PortIdx:         view.Index,
		MAC:             strings.ToLower(strings.TrimSpace(iface.MAC)),
		NetworkGroup:    gatewayNetworkGroupNone,
		Role:            view.Role,
		Up:              false,
		Enable:          false,
		Disabled:        true,
		SourceInterface: view.SourceInterface,
	}
}

func gatewayPortAssignmentFields(view PortView) gatewayAssignmentFields {
	port := view.Port
	return gatewayAssignmentFields{
		PortConfID:          strings.TrimSpace(port.PortConfID),
		NetworkConfID:       strings.TrimSpace(port.NetworkConfID),
		NativeNetworkConfID: strings.TrimSpace(port.NativeNetworkConfID),
		NetworkName:         gatewayNetworkName(view),
		VLAN:                port.VLAN,
	}
}

func gatewayNetworkName(view PortView) string {
	if value := strings.TrimSpace(view.Port.NetworkName); value != "" {
		return value
	}
	if !gatewayIsLogicalRole(view.Role) {
		return ""
	}
	if value := strings.TrimSpace(view.NetworkGroup); value != "" {
		return strings.ToLower(value)
	}
	return strings.ToLower(normalizeGatewayRole(view.Role))
}

// gatewayUplinkPortIndex returns the one-based uplink port index.
func gatewayUplinkPortIndex(ports []PortView) int {
	for _, port := range ports {
		if port.Uplink {
			return port.Index
		}
	}
	return 1
}

// gatewayIfTable renders physical interfaces for gateway inform payloads.
func gatewayIfTable(_ device.Profile, id device.Identity, ports []PortView, now time.Time, uptime int) []gatewayIfRow {
	ports = gatewayLogicalInterfacePorts(ports)
	out := make([]gatewayIfRow, 0, len(ports))
	uplinkIndex := gatewayUplinkPortIndex(ports)
	for _, view := range ports {
		iface := view.GatewayInterface
		row := gatewayIfRow{
			Name:               iface.Name,
			IfName:             iface.IfName,
			Comment:            iface.Comment,
			PortIdx:            view.Index,
			MAC:                iface.MAC,
			IP:                 iface.IP,
			Netmask:            iface.Netmask,
			NumPort:            1,
			Up:                 view.Up,
			Enable:             view.Enabled,
			NetworkGroup:       iface.NetworkGroup,
			FullDuplex:         true,
			PhysicalPorts:      []int{view.Index},
			linkFields:         portLinkFields(view.Speed, view.Media),
			counterFields:      portCounterFields(view.Port),
			optionalRateFields: explicitPortRateFields(view.Port),
			gatewayWANInlineHealth: gatewayWANInlineHealthFor(
				view,
				uptime,
			),
			SourceInterface:  view.SourceInterface,
			connectionFields: gatewayConnectionFields(view),
		}
		if view.Index == uplinkIndex {
			health := gatewayWANHealthFor(view, uptime)
			row.VLAN = id.ManagementVLAN
			row.ManagementVLAN = id.ManagementVLAN
			row.gatewayWANUplinkHealthFields = gatewayWANUplinkHealthFieldsFor(health, now, uptime)
		}
		out = append(out, row)
	}
	return out
}

func gatewayLogicalInterfacePorts(ports []PortView) []PortView {
	out := make([]PortView, 0, len(ports))
	seen := make(map[string]int)
	for _, view := range ports {
		if !gatewayIsLogicalRole(view.Role) {
			continue
		}
		ifname := strings.TrimSpace(view.GatewayInterface.IfName)
		if ifname == "" {
			continue
		}
		index, exists := seen[ifname]
		if !exists {
			seen[ifname] = len(out)
			out = append(out, view)
			continue
		}
		if gatewayPreferInterfaceView(view, out[index]) {
			out[index] = view
		}
	}
	return out
}

func gatewayIsLogicalRole(role string) bool {
	switch normalizeGatewayRole(role) {
	case gatewayPortRoleWAN, gatewayPortRoleWAN2, gatewayPortRoleLAN, gatewayPortRoleLAN2:
		return true
	default:
		return false
	}
}

func gatewayPreferInterfaceView(candidate, current PortView) bool {
	candidateScore := gatewayInterfaceViewScore(candidate)
	currentScore := gatewayInterfaceViewScore(current)
	if candidateScore != currentScore {
		return candidateScore > currentScore
	}
	return candidate.Index < current.Index
}

func gatewayInterfaceViewScore(view PortView) int {
	score := 0
	if view.Up {
		score += 4
	}
	if strings.TrimSpace(view.SourceInterface) != "" {
		score += 2
	}
	if view.Uplink {
		score++
	}
	return score
}

// gatewayNetworkTable renders the routed network view for each gateway port.
func gatewayNetworkTable(_ device.Profile, _ device.Identity, ports []PortView, uptime int) []gatewayNetworkRow {
	ports = gatewayLogicalInterfacePorts(ports)
	out := make([]gatewayNetworkRow, 0, len(ports))
	for _, view := range ports {
		iface := view.GatewayInterface
		ifname := gatewayNetworkInterfaceName(view)
		entry := gatewayNetworkRow{
			Name:                 iface.Name,
			IfName:               ifname,
			GatewayInterfaceName: ifname,
			PortIdx:              view.Index,
			MAC:                  iface.MAC,
			NetworkGroup:         iface.NetworkGroup,
			IP:                   iface.IP,
			Netmask:              iface.Netmask,
			Address:              iface.Address,
			Addresses:            []string{iface.Address},
			Up:                   boolText(view.Up),
			L1Up:                 boolText(view.Up),
			Autoneg:              "true",
			Duplex:               "full",
			Speed:                strconv.Itoa(view.Speed),
			MaxSpeed:             strconv.Itoa(view.Speed),
			MTU:                  "1500",
			Stats:                gatewayNetworkStats{counterFields: portCounterFields(view.Port), optionalRateFields: explicitPortRateFields(view.Port)},
			gatewayWANInlineHealth: gatewayWANInlineHealthFor(
				view,
				uptime,
			),
			SourceInterface:  view.SourceInterface,
			connectionFields: connectionFields{Connected: view.Up},
		}
		if hosts := gatewayHostTable(view); len(hosts) > 0 {
			entry.HostTable = hosts
		}
		out = append(out, entry)
	}
	return out
}

func gatewayNetworkInterfaceName(view PortView) string {
	switch gatewayPortRole(view.Port) {
	case gatewayPortRoleLAN, gatewayPortRoleLAN2:
		if ifname := gatewayPhysicalPortIfName(view); ifname != "" {
			return ifname
		}
	}
	return view.GatewayInterface.IfName
}

// gatewayReportedNetworks renders a read-only network summary per gateway
// port. It mirrors network_table values without inventing host configuration.
func gatewayReportedNetworks(ports []PortView, uptime int) []gatewayReportedNetworkRow {
	ports = gatewayLogicalInterfacePorts(ports)
	out := make([]gatewayReportedNetworkRow, 0, len(ports))
	for _, view := range ports {
		iface := view.GatewayInterface
		ifname := gatewayNetworkInterfaceName(view)
		row := gatewayReportedNetworkRow{
			Name:             iface.NetworkGroup,
			IfName:           ifname,
			PortIdx:          view.Index,
			NetworkGroup:     iface.NetworkGroup,
			Type:             view.Role,
			IP:               iface.IP,
			Netmask:          iface.Netmask,
			Address:          iface.Address,
			Addresses:        []string{iface.Address},
			Up:               view.Up,
			SourceInterface:  view.SourceInterface,
			connectionFields: connectionFields{Connected: view.Up},
		}
		if gatewayPortRole(view.Port) == gatewayPortRoleWAN || gatewayPortRole(view.Port) == gatewayPortRoleWAN2 {
			health := gatewayWANHealthFor(view, uptime)
			row.Availability = health.uptimePercent
			row.Latency = health.latencyMS
			row.Downtime = health.downtime
			row.IsWANUp = health.up
			row.IsWANConnected = health.connected
		}
		out = append(out, row)
	}
	return out
}

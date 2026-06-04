package payload

import (
	"strconv"
	"strings"
)

// gatewayConfigNetwork renders the controller-facing network assignment view
// for the first port matching the requested WAN role.
func gatewayConfigNetwork(ports []PortView, role string) (gatewayConfigNetworkRow, bool) {
	for _, view := range gatewayRoleCandidates(ports, role) {
		if gatewayPortRole(view.Port) != role {
			continue
		}
		iface := view.GatewayInterface
		speed := "auto"
		if view.Speed > 0 && !view.Up {
			speed = strconv.Itoa(view.Speed)
		}
		return gatewayConfigNetworkRow{
			Type:         payloadTypeDHCP,
			Name:         iface.NetworkGroup,
			IfName:       iface.IfName,
			PortIdx:      view.Index,
			NetworkGroup: iface.NetworkGroup,
			UplinkIfName: iface.IfName,
			IP:           stringRef(iface.IP),
			Netmask:      stringRef(iface.Netmask),
			Speed:        stringRef(speed),
			Autoneg:      boolRef(true),
			FullDuplex:   boolRef(true),
			DHCPOptions:  []emptyObject{},
		}, true
	}
	return gatewayConfigNetworkRow{Type: payloadTypeDHCP, DHCPOptions: []emptyObject{}}, false
}

// gatewayConfigLAN renders the LAN configuration summary that Network stores
// separately from live interface rows for UXG-class gateways.
func gatewayConfigLAN(ports []PortView) (gatewayConfigLANRow, bool) {
	view, ok := gatewayLANConfigPort(ports)
	if !ok {
		return gatewayConfigLANRow{}, false
	}
	iface := view.GatewayInterface
	lanIfName := physicalIfName(view)
	if lanIfName == "" {
		lanIfName = iface.IfName
	}
	vlan := view.Port.VLAN
	if vlan <= 0 {
		vlan = 1
	}
	return gatewayConfigLANRow{
		Name:                iface.NetworkGroup,
		IfName:              lanIfName,
		PortIdx:             view.Index,
		NetworkGroup:        iface.NetworkGroup,
		IP:                  stringRef(iface.IP),
		Netmask:             stringRef(iface.Netmask),
		UplinkIfName:        lanIfName,
		NetworkConfID:       strings.TrimSpace(view.Port.NetworkConfID),
		NativeNetworkConfID: strings.TrimSpace(view.Port.NativeNetworkConfID),
		NetworkName:         networkName(view),
		DHCPEnabled:         false,
		DHCPRangeStart:      "",
		DHCPRangeStop:       "",
		CIDR:                iface.Address,
		VLAN:                vlan,
	}, true
}

func gatewayLANConfigPort(ports []PortView) (PortView, bool) {
	for _, activeOnly := range []bool{true, false} {
		for _, role := range []string{gatewayPortRoleLAN, gatewayPortRoleLAN2} {
			for _, view := range ports {
				if gatewayPortRole(view.Port) != role {
					continue
				}
				if activeOnly && !view.Up && strings.TrimSpace(view.SourceInterface) == "" {
					continue
				}
				return view, true
			}
		}
	}
	return PortView{}, false
}

// gatewayWANStatus renders live WAN-like state from the same resolved port
// view used by the gateway interface tables.
func gatewayWANStatus(ports []PortView, role string, uptime int) *gatewayWANStatusRow {
	for _, view := range gatewayRoleCandidates(ports, role) {
		if gatewayPortRole(view.Port) != role {
			continue
		}
		iface := view.GatewayInterface
		health := wanHealth(view, uptime)
		row := gatewayWANStatusRow{
			Type:               payloadTypeDHCP,
			Name:               iface.NetworkGroup,
			IfName:             iface.IfName,
			PortIdx:            view.Index,
			NetworkGroup:       iface.NetworkGroup,
			Role:               view.Role,
			MAC:                iface.MAC,
			IP:                 iface.IP,
			Netmask:            iface.Netmask,
			Address:            iface.Address,
			Up:                 view.Up,
			Enable:             view.Enabled,
			Uptime:             uptime,
			Latency:            health.latencyMS,
			UplinkIfName:       iface.IfName,
			linkFields:         portLinkFields(view.Speed, view.Media),
			counterFields:      portCounterFields(view.Port),
			optionalRateFields: explicitPortRateFields(view.Port),
			gatewayRateFields:  gatewayPortRateFields(view.Port),
			SourceInterface:    view.SourceInterface,
		}
		return &row
	}
	return nil
}

func gatewayTrafficSummaryFor(ports []PortView, role string) gatewayTrafficSummary {
	for _, view := range gatewayRoleCandidates(ports, role) {
		if gatewayPortRole(view.Port) != role {
			continue
		}
		port := view.Port
		// Gateway root counters follow the controller summary convention.
		// Per-port rows keep interface-local RX/TX direction; the root summary
		// is WAN-facing and therefore uses the opposite direction.
		out := gatewayTrafficSummary{
			Bytes:   port.RXBytes + port.TXBytes,
			RXBytes: port.TXBytes,
			TXBytes: port.RXBytes,
		}
		if port.TrafficRatesSet || port.TrafficRatesEnabled {
			rxByteRate := port.TXBytesRate
			txByteRate := port.RXBytesRate
			out.BytesRate = int64Ref(rxByteRate + txByteRate)
			out.RXBytesRate = int64Ref(rxByteRate)
			out.TXBytesRate = int64Ref(txByteRate)
			out.RXRate = int64Ref(bitsPerSecond(rxByteRate))
			out.TXRate = int64Ref(bitsPerSecond(txByteRate))
		}
		return out
	}
	return gatewayTrafficSummary{}
}

func gatewayWans(ports []PortView) []gatewayWANInventoryRow {
	rows := make([]gatewayWANInventoryRow, 0, 2)
	for _, role := range []string{gatewayPortRoleWAN, gatewayPortRoleWAN2} {
		for _, view := range gatewayRoleCandidates(ports, role) {
			if gatewayPortRole(view.Port) != role {
				continue
			}
			iface := view.GatewayInterface
			rows = append(rows, gatewayWANInventoryRow{
				Enabled:   view.Enabled,
				Interface: iface.IfName,
				IPv4:      iface.IP,
				MAC:       iface.MAC,
				Plugged:   view.Up,
				Port:      view.Index,
				Type:      strings.ToUpper(iface.NetworkGroup),
			})
			break
		}
	}
	if len(rows) == 0 {
		return nil
	}
	return rows
}

func gatewayRoleCandidates(ports []PortView, role string) []PortView {
	matches := make([]PortView, 0, len(ports))
	active := make([]PortView, 0, len(ports))
	for _, view := range ports {
		if gatewayPortRole(view.Port) != role {
			continue
		}
		matches = append(matches, view)
		if view.Up && (view.Uplink || strings.TrimSpace(view.SourceInterface) != "") {
			active = append(active, view)
		}
	}
	if len(active) > 0 {
		return active
	}
	if len(matches) > 0 {
		return matches
	}
	return ports
}

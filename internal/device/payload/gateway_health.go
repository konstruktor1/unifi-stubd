package payload

import "github.com/konstruktor1/unifi-stubd/internal/device"

func gatewayUplinkTable(id device.Identity, ports []PortView) []gatewayUplinkRow {
	uplinkIndex := uplinkPortIndex(ports)
	for _, view := range ports {
		if view.Index != uplinkIndex {
			continue
		}
		iface := view.GatewayInterface
		row := gatewayUplinkRow{
			Name:               iface.Name,
			IfName:             iface.IfName,
			PortIdx:            view.Index,
			MAC:                iface.MAC,
			Type:               payloadTypeWire,
			NetworkGroup:       iface.NetworkGroup,
			UplinkIfName:       iface.IfName,
			IP:                 iface.IP,
			Up:                 view.Up,
			Enable:             view.Enabled,
			FullDuplex:         true,
			VLAN:               id.ManagementVLAN,
			ManagementVLAN:     id.ManagementVLAN,
			linkFields:         portLinkFields(view.Speed, view.Media),
			counterFields:      portCounterFields(view.Port),
			optionalRateFields: explicitPortRateFields(view.Port),
			gatewayRateFields:  gatewayPortRateFields(view.Port),
			SourceInterface:    view.SourceInterface,
			connectionFields:   gatewayConnectionFields(view),
		}
		return []gatewayUplinkRow{row}
	}
	return nil
}

func gatewayLastWANIP(ports []PortView) string {
	for _, view := range gatewayRoleCandidates(ports, gatewayPortRoleWAN) {
		if gatewayPortRole(view.Port) != gatewayPortRoleWAN {
			continue
		}
		return view.GatewayInterface.IP
	}
	return ""
}

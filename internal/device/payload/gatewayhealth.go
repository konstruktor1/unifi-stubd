package payload

import (
	"time"

	"github.com/konstruktor1/unifi-stubd/internal/device"
)

// gatewayUplinkTable renders the controller-facing uplink entry.
func gatewayUplinkTable(_ device.Profile, id device.Identity, ports []PortView, now time.Time, uptime int) []gatewayUplinkRow {
	uplinkIndex := gatewayUplinkPortIndex(ports)
	for _, view := range ports {
		if view.Index != uplinkIndex {
			continue
		}
		iface := view.GatewayInterface
		health := gatewayWANHealthFor(view, uptime)
		row := gatewayUplinkRow{
			Name:         iface.Name,
			IfName:       iface.IfName,
			PortIdx:      view.Index,
			MAC:          iface.MAC,
			Type:         "wire",
			NetworkGroup: iface.NetworkGroup,
			UplinkIfName: iface.IfName,
			IP:           iface.IP,
			Up:           view.Up,
			Enable:       view.Enabled,
			FullDuplex:   true,
			Availability: health.uptimePercent,
			Latency:      health.latencyMS,
			Downtime:     health.downtime,
			gatewayWANUplinkHealthFields: gatewayWANUplinkHealthFieldsFor(
				health,
				now,
				uptime,
			),
			IsWANConnected:     health.connected,
			IsWANUp:            health.up,
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

func gatewayWANUplinkHealthFieldsFor(health gatewayWANHealth, now time.Time, uptime int) gatewayWANUplinkHealthFields {
	if !health.connected {
		return gatewayWANUplinkHealthFields{}
	}
	return gatewayWANUplinkHealthFields{
		Uptime:           gatewayWANUplinkUptime(health, uptime),
		SpeedtestStatus:  gatewaySpeedtestStatusText(health.connected),
		SpeedtestLastRun: gatewaySpeedtestLastRun(health.connected, now),
		SpeedtestPing:    health.latencyMS,
	}
}

func gatewayWANUplinkUptime(health gatewayWANHealth, uptime int) int {
	if !health.connected {
		return 0
	}
	return uptime
}

func gatewaySpeedtestStatusText(connected bool) string {
	if connected {
		return "Success"
	}
	return "Idle"
}

func gatewaySpeedtestLastRun(connected bool, now time.Time) int {
	if !connected {
		return 0
	}
	return int(now.Unix())
}

func gatewayWANUptimeStats(ports []PortView, uptime int) map[string]gatewayWANHealthRow {
	stats := make(map[string]gatewayWANHealthRow, 2)
	for _, role := range []string{gatewayPortRoleWAN, gatewayPortRoleWAN2} {
		for _, view := range gatewayRoleCandidates(ports, role) {
			if gatewayPortRole(view.Port) != role {
				continue
			}
			iface := view.GatewayInterface
			health := gatewayWANHealthFor(view, uptime)
			stats[iface.NetworkGroup] = gatewayWANHealthRow{
				NetworkGroup:   iface.NetworkGroup,
				IfName:         iface.IfName,
				UplinkIfName:   iface.IfName,
				IP:             iface.IP,
				PortIdx:        view.Index,
				Uptime:         health.uptimePercent,
				Availability:   health.uptimePercent,
				Latency:        health.latencyMS,
				Downtime:       health.downtime,
				IsWANUp:        health.up,
				IsWANConnected: health.connected,
			}
			break
		}
	}
	if len(stats) == 0 {
		return nil
	}
	return stats
}

func gatewayInternetHealth(ports []PortView, uptime int) *gatewayInternetHealthRow {
	for _, view := range gatewayRoleCandidates(ports, gatewayPortRoleWAN) {
		if gatewayPortRole(view.Port) != gatewayPortRoleWAN {
			continue
		}
		iface := view.GatewayInterface
		health := gatewayWANHealthFor(view, uptime)
		status := "offline"
		if health.connected {
			status = "ok"
		}
		return &gatewayInternetHealthRow{
			Status: status,
			WANStatus: map[string]string{
				iface.NetworkGroup: gatewayWANStatusText(health.connected),
			},
			WANIP:          iface.IP,
			Netmask:        iface.Netmask,
			IfName:         iface.IfName,
			UplinkIfName:   iface.IfName,
			PortIdx:        view.Index,
			Latency:        health.latencyMS,
			Uptime:         health.uptimePercent,
			Availability:   health.uptimePercent,
			Downtime:       health.downtime,
			Drops:          0,
			IsWANUp:        health.up,
			IsWANConnected: health.connected,
		}
	}
	return nil
}

func gatewayLastWANStatus(ports []PortView, uptime int) map[string]string {
	status := make(map[string]string, 2)
	for _, role := range []string{gatewayPortRoleWAN, gatewayPortRoleWAN2} {
		for _, view := range gatewayRoleCandidates(ports, role) {
			if gatewayPortRole(view.Port) != role {
				continue
			}
			iface := view.GatewayInterface
			health := gatewayWANHealthFor(view, uptime)
			status[iface.NetworkGroup] = gatewayWANStatusText(health.connected)
			break
		}
	}
	if len(status) == 0 {
		return nil
	}
	return status
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

func gatewayWANStatusText(connected bool) string {
	if connected {
		return "online"
	}
	return "offline"
}

func gatewayWANHealthFor(view PortView, uptime int) gatewayWANHealth {
	up := view.Up && view.Enabled
	connected := up
	// WANConnected may come from static YAML hints or active wan_health ping
	// samples. It overlays reachability telemetry only; link state remains in up.
	if view.Port.WANConnected != nil {
		connected = *view.Port.WANConnected
	}
	uptimePercent := 0.0
	if connected {
		uptimePercent = 100.0
	}
	if view.Port.WANUptimePercent != nil {
		uptimePercent = *view.Port.WANUptimePercent
	}
	downtime := view.Port.WANDowntimeSeconds
	if downtime == 0 && !connected && uptime > 0 {
		downtime = uptime
	}
	return gatewayWANHealth{
		uptimePercent: uptimePercent,
		latencyMS:     view.Port.WANLatencyMS,
		downtime:      downtime,
		up:            up,
		connected:     connected,
	}
}

func gatewayWANInlineHealthFor(view PortView, uptime int) gatewayWANInlineHealth {
	switch gatewayPortRole(view.Port) {
	case gatewayPortRoleWAN, gatewayPortRoleWAN2:
	default:
		return gatewayWANInlineHealth{}
	}
	health := gatewayWANHealthFor(view, uptime)
	return gatewayWANInlineHealth{
		Availability:   float64Ref(health.uptimePercent),
		Latency:        intRef(health.latencyMS),
		Downtime:       intRef(health.downtime),
		IsWANUp:        boolRef(health.up),
		IsWANConnected: boolRef(health.connected),
	}
}

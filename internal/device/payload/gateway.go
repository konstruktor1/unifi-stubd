// Package payload derives gateway interface and link-state tables from profile
// hardware metadata plus passive observations. Explicitly configured WAN/LAN,
// VLAN, and port-profile assignment metadata may be mirrored to the controller;
// it is never inferred from controller provisioning or applied to the host.
package payload

import (
	"strings"
	"time"

	"github.com/konstruktor1/unifi-stubd/internal/device"
)

// buildGatewayPayload fills gateway status tables with device-originated facts.
func buildGatewayPayload(base basePayload, profile device.Profile, id device.Identity, ports []PortView, now time.Time, uptime int) gatewayPayload {
	configWAN, _ := gatewayConfigNetwork(ports, gatewayPortRoleWAN)
	configWAN2, hasWAN2 := gatewayConfigNetwork(ports, gatewayPortRoleWAN2)
	configLAN, hasLAN := gatewayConfigLAN(ports)
	payload := gatewayPayload{
		basePayload:      base,
		gatewayTelemetry: newGatewayTelemetry(id, now, uptime, base.CFGVersion),
		gatewayTrafficSummary: gatewayTrafficSummaryFor(
			ports,
			gatewayPortRoleWAN,
		),
		IfTable:           gatewayIfTable(profile, id, ports, now, uptime),
		NetworkTable:      gatewayNetworkTable(profile, id, ports, uptime),
		ConfigPortTable:   gatewayConfigPortTable(ports, uptime),
		EthernetTable:     gatewayEthernetTable(ports),
		EthernetOverrides: gatewayEthernetOverrides(ports, uptime),
		PortTable:         gatewayPortTable(ports, uptime),
		PortStats:         gatewayPortStatsTable(ports),
		ReportedNetworks:  gatewayReportedNetworks(ports, uptime),
		Wans:              gatewayWans(ports),
		Uplink:            gatewayUplinkInterfaceName(profile, ports),
		UplinkTable:       gatewayUplinkTable(profile, id, ports, now, uptime),
		UptimeStats:       gatewayWANUptimeStats(ports, uptime),
		InternetHealth:    gatewayInternetHealth(ports, uptime),
		LastWANStatus:     gatewayLastWANStatus(ports, uptime),
		LastWANIP:         gatewayLastWANIP(ports),
		LANIP:             gatewayLANIP(configLAN, hasLAN),
		HasEth1:           gatewayHasEth1(ports),
		HasDPI:            profile.Payload.HasDPI,
		ConfigNetworkWAN:  configWAN,
		WAN1:              gatewayWANStatus(ports, gatewayPortRoleWAN, uptime),
		WAN2:              gatewayWANStatus(ports, gatewayPortRoleWAN2, uptime),
	}
	payload.SpeedtestStatus = gatewaySpeedtestStatusFor(ports, now, uptime)
	if hasLAN {
		payload.ConfigNetworkLAN = &configLAN
	}
	if hasWAN2 {
		payload.ConfigNetworkWAN2 = &configWAN2
	}
	return payload
}

func gatewayLANIP(configLAN gatewayConfigLANRow, ok bool) string {
	if !ok || configLAN.IP == nil {
		return ""
	}
	return *configLAN.IP
}

func gatewayHasEth1(ports []PortView) bool {
	for _, view := range ports {
		if strings.TrimSpace(view.GatewayInterface.IfName) == "eth1" {
			return true
		}
	}
	return false
}

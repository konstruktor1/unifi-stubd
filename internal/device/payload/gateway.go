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
	telemetry := newGatewayTelemetry(profile, id, now, uptime, base.CFGVersion)
	telemetry.LLDPTable = lldpTable(ports)
	payload := gatewayPayload{
		basePayload:      base,
		gatewayTelemetry: telemetry,
		gatewayTrafficSummary: gatewayTrafficSummaryFor(
			ports,
			gatewayPortRoleWAN,
		),
		IfTable:           gatewayIfTable(id, ports),
		NetworkTable:      gatewayNetworkTable(ports),
		ConfigPortTable:   gatewayConfigPortTable(ports),
		EthernetTable:     gatewayEthernetTable(ports),
		EthernetOverrides: gatewayEthernetOverrides(ports),
		PortTable:         gatewayPortTable(ports, id.GatewayMACIPv6Requested),
		PortStats:         gatewayPortStatsTable(ports),
		ReportedNetworks:  gatewayReportedNetworks(ports),
		Wans:              gatewayWans(ports),
		Uplink:            gatewayUplinkInterfaceName(profile, ports),
		UplinkTable:       gatewayUplinkTable(id, ports),
		LastWANIP:         gatewayLastWANIP(ports),
		LANIP:             gatewayLANIP(configLAN, hasLAN),
		HasEth1:           gatewayHasEth1(ports),
		HasDPI:            profile.Payload.HasDPI,
		ConfigNetworkWAN:  configWAN,
	}
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

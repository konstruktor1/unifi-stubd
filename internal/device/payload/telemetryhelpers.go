package payload

import (
	"fmt"
	"net"
	"time"
)

// formatUptime mirrors compact firmware-style uptime text for gateway status
// blocks.
func formatUptime(seconds int) string {
	if seconds < 1 {
		seconds = 1
	}
	days := seconds / 86400
	seconds %= 86400
	hours := seconds / 3600
	seconds %= 3600
	minutes := seconds / 60
	seconds %= 60
	switch {
	case days > 0:
		return fmt.Sprintf("%dd%dh%dm%ds", days, hours, minutes, seconds)
	case hours > 0:
		return fmt.Sprintf("%dh%dm%ds", hours, minutes, seconds)
	case minutes > 0:
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	default:
		return fmt.Sprintf("%ds", seconds)
	}
}

// gatewaySpeedtestStatus returns an idle speed-test block for gateway payloads.
func gatewaySpeedtestStatus() speedtestStatusPayload {
	return speedtestStatusPayload{
		Latency: 0,
		RunDate: 0,
		Runtime: 0,
		Server: speedtestServerPayload{
			CountryCode: "",
			City:        "",
			Country:     "",
			Latitude:    0.0,
			Longitude:   0.0,
			Provider:    "",
			ProviderURL: "",
		},
		SourceIf:       "",
		StatusDownload: 0,
		StatusPing:     0,
		StatusSummary:  0,
		StatusUpload:   0,
		XputDownload:   0.0,
		XputUpload:     0.0,
	}
}

func gatewaySpeedtestStatusFor(ports []PortView, now time.Time, uptime int) speedtestStatusPayload {
	status := gatewaySpeedtestStatus()
	for _, view := range gatewayRoleCandidates(ports, gatewayPortRoleWAN) {
		if gatewayPortRole(view.Port) != gatewayPortRoleWAN {
			continue
		}
		health := gatewayWANHealthFor(view, uptime)
		if !health.connected {
			return status
		}
		status.Latency = health.latencyMS
		status.RunDate = int(now.Unix())
		status.Runtime = 1
		status.SourceIf = view.GatewayInterface.IfName
		status.StatusDownload = 2
		status.StatusPing = 2
		status.StatusSummary = 2
		status.StatusUpload = 2
		return status
	}
	return status
}

// gatewaySysID derives a stable numeric system ID from the device MAC.
func gatewaySysID(macText string) int {
	mac, err := net.ParseMAC(macText)
	if err != nil || len(mac) < 2 {
		return 42615
	}
	return int(mac[len(mac)-2])<<8 | int(mac[len(mac)-1])
}

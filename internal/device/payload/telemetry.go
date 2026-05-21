// Package payload fills low-risk deterministic gateway telemetry blocks that
// UniFi Network expects around gateway payloads. Values are placeholders and
// must not be copied from private controller state or firmware secrets.
package payload

import (
	"fmt"
	"net"
	"time"
)

// applyGatewayTelemetry adds deterministic low-risk gateway metadata fields.
func applyGatewayTelemetry(payload map[string]any, id Identity, now time.Time, uptime int) {
	cfgVersion, _ := payload["cfgversion"].(string)
	if cfgVersion == "" {
		cfgVersion = "?"
	}
	payload["anon_id"] = ""
	payload["architecture"] = "aarch64"
	payload["ble_caps"] = 0
	payload["board_rev"] = 1
	payload["bomrev"] = "unknown"
	payload["bomrev_id"] = "00000000"
	payload["boot"] = map[string]any{}
	payload["bootid"] = -1
	payload["bootrom_version"] = "unknown"
	payload["cfgversion_effective"] = cfgVersion
	payload["connections"] = []map[string]any{}
	payload["content_filtering_status"] = map[string]any{"feature_status": "UNAVAILABLE_NO_SUBSCRIPTION"}
	payload["dns_shield"] = map[string]any{"hash": ""}
	payload["dpi_stats"] = []map[string]any{}
	payload["dualboot"] = false
	payload["ever_crash"] = false
	payload["fingerprint"] = "00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00"
	payload["fingerprints"] = []map[string]any{}
	payload["fw2_caps"] = 0
	payload["fw_caps"] = 0
	payload["guest_kicks"] = 0
	payload["guest_token"] = ""
	payload["gw_caps"] = map[string]any{}
	payload["hardware_uuid"] = "00000000-0000-4000-8000-000000000000"
	payload["has_default_route_distance"] = true
	payload["has_speaker"] = false
	payload["has_ssh_disable"] = true
	payload["has_vti"] = true
	payload["hw_caps"] = 0
	payload["ids_ips_rule"] = map[string]any{"rule_count": 0, "sha256": "", "signature_type": "", "update_time": ""}
	payload["inform_min_interval"] = 1
	payload["ipv4_active_leases"] = []map[string]any{}
	payload["isolated"] = false
	payload["kernel_version"] = "6.12.0-stubd"
	payload["last_error_conns"] = []map[string]any{}
	payload["led_state"] = map[string]any{"pattern": "0", "tempo": 120}
	payload["lldp_table"] = []map[string]any{}
	payload["locating"] = false
	payload["manufacturer_id"] = 61
	payload["netmask"] = "255.255.255.0"
	payload["outlet_enabled"] = false
	payload["outlet_overrides"] = []map[string]any{}
	payload["outlet_table"] = []map[string]any{}
	payload["pingtest-status"] = []map[string]any{}
	payload["qrid"] = ""
	payload["reboot_duration"] = 30
	payload["selfrun_beacon"] = true
	payload["speedtest-status"] = gatewaySpeedtestStatus()
	payload["speedtest-status-udapi"] = []map[string]any{}
	payload["ssh_session_table"] = []map[string]any{}
	payload["stats_inform_interval"] = 0
	payload["switch_caps"] = map[string]any{"feature_caps": 1048576, "max_aggregate_sessions": 0, "max_mirror_sessions": 1}
	payload["sys_error_caps"] = 0
	payload["sysid"] = gatewaySysID(id.MAC)
	payload["teleport_version"] = 1
	payload["time_ms"] = now.UnixMilli()
	payload["timestamp"] = now.UTC().Format("2006-01-02T15:04:05")
	payload["tm_ready"] = false
	payload["triggers"] = []map[string]any{}
	payload["triggers_dns_filter"] = []map[string]any{}
	payload["triggers_geo"] = []map[string]any{}
	payload["udapi_caps"] = 0
	payload["udapi_version"] = map[string]any{}
	payload["upgrade_duration"] = 150
	payload["uptime_str"] = formatUptime(uptime)
	payload["usg2_caps"] = 0
	payload["usg_caps"] = 0
	payload["wifi_caps"] = 0
}

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
func gatewaySpeedtestStatus() map[string]any {
	return map[string]any{
		"latency":         0,
		"rundate":         0,
		"runtime":         0,
		"server":          map[string]any{"cc": "", "city": "", "country": "", "lat": 0.0, "lon": 0.0, "provider": "", "provider_url": ""},
		jsonKeySourceIf:   "",
		"status_download": 0,
		"status_ping":     0,
		"status_summary":  0,
		"status_upload":   0,
		"xput_download":   0.0,
		"xput_upload":     0.0,
	}
}

// gatewaySysID derives a stable numeric system ID from the device MAC.
func gatewaySysID(macText string) int {
	mac, err := net.ParseMAC(macText)
	if err != nil || len(mac) < 2 {
		return 42615
	}
	return int(mac[len(mac)-2])<<8 | int(mac[len(mac)-1])
}

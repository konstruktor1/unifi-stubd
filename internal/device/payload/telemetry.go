// Package payload fills low-risk deterministic gateway telemetry blocks that
// UniFi Network expects around gateway payloads. Values are placeholders and
// must not be copied from private controller state or firmware secrets.
package payload

import (
	"fmt"
	"net"
	"time"

	"github.com/konstruktor1/unifi-stubd/internal/device"
)

type emptyObject struct{}

type featureStatusPayload struct {
	FeatureStatus string `json:"feature_status"`
}

type hashPayload struct {
	Hash string `json:"hash"`
}

type gatewayRulePayload struct {
	RuleCount     int    `json:"rule_count"`
	SHA256        string `json:"sha256"`
	SignatureType string `json:"signature_type"`
	UpdateTime    string `json:"update_time"`
}

type ledStatePayload struct {
	Pattern string `json:"pattern"`
	Tempo   int    `json:"tempo"`
}

type speedtestServerPayload struct {
	CountryCode string  `json:"cc"`
	City        string  `json:"city"`
	Country     string  `json:"country"`
	Latitude    float64 `json:"lat"`
	Longitude   float64 `json:"lon"`
	Provider    string  `json:"provider"`
	ProviderURL string  `json:"provider_url"`
}

type speedtestStatusPayload struct {
	Latency        int                    `json:"latency"`
	RunDate        int                    `json:"rundate"`
	Runtime        int                    `json:"runtime"`
	Server         speedtestServerPayload `json:"server"`
	SourceIf       string                 `json:"source_interface"`
	StatusDownload int                    `json:"status_download"`
	StatusPing     int                    `json:"status_ping"`
	StatusSummary  int                    `json:"status_summary"`
	StatusUpload   int                    `json:"status_upload"`
	XputDownload   float64                `json:"xput_download"`
	XputUpload     float64                `json:"xput_upload"`
}

type switchCapsPayload struct {
	FeatureCaps          int `json:"feature_caps"`
	MaxAggregateSessions int `json:"max_aggregate_sessions"`
	MaxMirrorSessions    int `json:"max_mirror_sessions"`
}

type gatewayTelemetry struct {
	AnonID                 string                 `json:"anon_id"`
	Architecture           string                 `json:"architecture"`
	BLECaps                int                    `json:"ble_caps"`
	BoardRevision          int                    `json:"board_rev"`
	BOMRevision            string                 `json:"bomrev"`
	BOMRevisionID          string                 `json:"bomrev_id"`
	Boot                   emptyObject            `json:"boot"`
	BootID                 int                    `json:"bootid"`
	BootROMVersion         string                 `json:"bootrom_version"`
	CFGVersionEffective    string                 `json:"cfgversion_effective"`
	Connections            []emptyObject          `json:"connections"`
	ContentFilteringStatus featureStatusPayload   `json:"content_filtering_status"`
	DNSShield              hashPayload            `json:"dns_shield"`
	DPIStats               []emptyObject          `json:"dpi_stats"`
	Dualboot               bool                   `json:"dualboot"`
	EverCrash              bool                   `json:"ever_crash"`
	Fingerprint            string                 `json:"fingerprint"`
	Fingerprints           []emptyObject          `json:"fingerprints"`
	FW2Caps                int                    `json:"fw2_caps"`
	FWCaps                 int                    `json:"fw_caps"`
	GuestKicks             int                    `json:"guest_kicks"`
	GuestToken             string                 `json:"guest_token"`
	GWCapabilities         emptyObject            `json:"gw_caps"`
	HardwareUUID           string                 `json:"hardware_uuid"`
	HasDefaultRouteDist    bool                   `json:"has_default_route_distance"`
	HasSpeaker             bool                   `json:"has_speaker"`
	HasSSHDisable          bool                   `json:"has_ssh_disable"`
	HasVTI                 bool                   `json:"has_vti"`
	HWCapabilities         int                    `json:"hw_caps"`
	IDsIPSRule             gatewayRulePayload     `json:"ids_ips_rule"`
	InformMinInterval      int                    `json:"inform_min_interval"`
	IPv4ActiveLeases       []emptyObject          `json:"ipv4_active_leases"`
	Isolated               bool                   `json:"isolated"`
	KernelVersion          string                 `json:"kernel_version"`
	LastErrorConns         []emptyObject          `json:"last_error_conns"`
	LEDState               ledStatePayload        `json:"led_state"`
	LLDPTable              []emptyObject          `json:"lldp_table"`
	Locating               bool                   `json:"locating"`
	ManufacturerID         int                    `json:"manufacturer_id"`
	Netmask                string                 `json:"netmask"`
	OutletEnabled          bool                   `json:"outlet_enabled"`
	OutletOverrides        []emptyObject          `json:"outlet_overrides"`
	OutletTable            []emptyObject          `json:"outlet_table"`
	PingtestStatus         []emptyObject          `json:"pingtest-status"`
	QRID                   string                 `json:"qrid"`
	RebootDuration         int                    `json:"reboot_duration"`
	SelfrunBeacon          bool                   `json:"selfrun_beacon"`
	SpeedtestStatus        speedtestStatusPayload `json:"speedtest-status"`
	SpeedtestStatusUDAPI   []emptyObject          `json:"speedtest-status-udapi"`
	SSHSessionTable        []emptyObject          `json:"ssh_session_table"`
	StatsInformInterval    int                    `json:"stats_inform_interval"`
	SwitchCaps             switchCapsPayload      `json:"switch_caps"`
	SysErrorCaps           int                    `json:"sys_error_caps"`
	SysID                  int                    `json:"sysid"`
	TeleportVersion        int                    `json:"teleport_version"`
	TimeMS                 int64                  `json:"time_ms"`
	Timestamp              string                 `json:"timestamp"`
	TMReady                bool                   `json:"tm_ready"`
	Triggers               []emptyObject          `json:"triggers"`
	TriggersDNSFilter      []emptyObject          `json:"triggers_dns_filter"`
	TriggersGeo            []emptyObject          `json:"triggers_geo"`
	UDAPICaps              int                    `json:"udapi_caps"`
	UDAPIVersion           emptyObject            `json:"udapi_version"`
	UpgradeDuration        int                    `json:"upgrade_duration"`
	UptimeText             string                 `json:"uptime_str"`
	USG2Caps               int                    `json:"usg2_caps"`
	USGCaps                int                    `json:"usg_caps"`
	WiFiCaps               int                    `json:"wifi_caps"`
}

// newGatewayTelemetry returns deterministic low-risk gateway metadata fields.
func newGatewayTelemetry(id device.Identity, now time.Time, uptime int, cfgVersion string) gatewayTelemetry {
	if cfgVersion == "" {
		cfgVersion = "?"
	}
	emptyList := []emptyObject{}
	return gatewayTelemetry{
		AnonID:              "",
		Architecture:        "aarch64",
		BLECaps:             0,
		BoardRevision:       1,
		BOMRevision:         "unknown",
		BOMRevisionID:       "00000000",
		Boot:                emptyObject{},
		BootID:              -1,
		BootROMVersion:      "unknown",
		CFGVersionEffective: cfgVersion,
		Connections:         emptyList,
		ContentFilteringStatus: featureStatusPayload{
			FeatureStatus: "UNAVAILABLE_NO_SUBSCRIPTION",
		},
		DNSShield:           hashPayload{Hash: ""},
		DPIStats:            emptyList,
		Dualboot:            false,
		EverCrash:           false,
		Fingerprint:         "00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00",
		Fingerprints:        emptyList,
		FW2Caps:             0,
		FWCaps:              0,
		GuestKicks:          0,
		GuestToken:          "",
		GWCapabilities:      emptyObject{},
		HardwareUUID:        "00000000-0000-4000-8000-000000000000",
		HasDefaultRouteDist: true,
		HasSpeaker:          false,
		HasSSHDisable:       true,
		HasVTI:              true,
		HWCapabilities:      0,
		IDsIPSRule: gatewayRulePayload{
			RuleCount:     0,
			SHA256:        "",
			SignatureType: "",
			UpdateTime:    "",
		},
		InformMinInterval:    1,
		IPv4ActiveLeases:     emptyList,
		Isolated:             false,
		KernelVersion:        "6.12.0-stubd",
		LastErrorConns:       emptyList,
		LEDState:             ledStatePayload{Pattern: "0", Tempo: 120},
		LLDPTable:            emptyList,
		Locating:             false,
		ManufacturerID:       61,
		Netmask:              "255.255.255.0",
		OutletEnabled:        false,
		OutletOverrides:      emptyList,
		OutletTable:          emptyList,
		PingtestStatus:       emptyList,
		QRID:                 "",
		RebootDuration:       30,
		SelfrunBeacon:        true,
		SpeedtestStatus:      gatewaySpeedtestStatus(),
		SpeedtestStatusUDAPI: emptyList,
		SSHSessionTable:      emptyList,
		StatsInformInterval:  0,
		SwitchCaps: switchCapsPayload{
			FeatureCaps:          1048576,
			MaxAggregateSessions: 0,
			MaxMirrorSessions:    1,
		},
		SysErrorCaps:      0,
		SysID:             gatewaySysID(id.MAC),
		TeleportVersion:   1,
		TimeMS:            now.UnixMilli(),
		Timestamp:         now.UTC().Format("2006-01-02T15:04:05"),
		TMReady:           false,
		Triggers:          emptyList,
		TriggersDNSFilter: emptyList,
		TriggersGeo:       emptyList,
		UDAPICaps:         0,
		UDAPIVersion:      emptyObject{},
		UpgradeDuration:   150,
		UptimeText:        formatUptime(uptime),
		USG2Caps:          0,
		USGCaps:           0,
		WiFiCaps:          0,
	}
}

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

// gatewaySysID derives a stable numeric system ID from the device MAC.
func gatewaySysID(macText string) int {
	mac, err := net.ParseMAC(macText)
	if err != nil || len(mac) < 2 {
		return 42615
	}
	return int(mac[len(mac)-2])<<8 | int(mac[len(mac)-1])
}

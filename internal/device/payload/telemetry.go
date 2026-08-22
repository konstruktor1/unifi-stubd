package payload

import (
	"math"
	"time"

	"github.com/konstruktor1/unifi-stubd/internal/device"
)

// newGatewayTelemetry returns deterministic low-risk gateway metadata fields.
func newGatewayTelemetry(profile device.Profile, id device.Identity, now time.Time, uptime int, cfgVersion string) gatewayTelemetry {
	if cfgVersion == "" {
		cfgVersion = "?"
	}
	emptyList := []emptyObject{}
	telemetry := gatewayTelemetry{
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
		DPIStats:            &emptyList,
		Dualboot:            false,
		EverCrash:           false,
		Fingerprint:         "00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00",
		Fingerprints:        &emptyList,
		FW2Caps:             669375863,
		FWCaps:              1659698733,
		GuestKicks:          0,
		GuestToken:          "",
		HardwareUUID:        "00000000-0000-4000-8000-000000000000",
		HasDefaultRouteDist: true,
		HasSpeaker:          false,
		HasSSHDisable:       true,
		HasVTI:              true,
		HWCapabilities:      152,
		IDsIPSRule: gatewayRulePayload{
			RuleCount:     0,
			SHA256:        "",
			SignatureType: "",
			UpdateTime:    "",
		},
		InformMinInterval:   1,
		Isolated:            false,
		KernelVersion:       "6.12.0-stubd",
		LastErrorConns:      emptyList,
		LEDState:            ledStatePayload{Pattern: "0", Tempo: 120},
		LLDPTable:           []lldpRow{},
		Locating:            false,
		ManufacturerID:      61,
		Netmask:             "255.255.255.0",
		OutletEnabled:       false,
		OutletOverrides:     emptyList,
		OutletTable:         emptyList,
		PingtestStatus:      emptyList,
		QRID:                "",
		RebootDuration:      30,
		SelfrunBeacon:       true,
		SSHSessionTable:     emptyList,
		StatsInformInterval: 0,
		SwitchCaps: switchCapsPayload{
			FeatureCaps:          1048576,
			MaxAggregateSessions: 0,
			MaxMirrorSessions:    1,
		},
		SysErrorCaps:      0,
		SysID:             gatewaySysID(id.MAC),
		TeleportVersion:   1,
		TimeMS:            saturatedUnixMilli(now),
		Timestamp:         now.UTC().Format("2006-01-02T15:04:05"),
		TMReady:           false,
		Triggers:          emptyList,
		TriggersDNSFilter: emptyList,
		TriggersGeo:       emptyList,
		UDAPICaps:         1610591215,
		UDAPIVersion:      gatewayUDAPIVersionRef(),
		UpgradeDuration:   150,
		UptimeText:        formatUptime(uptime),
		USG2Caps:          0,
		USGCaps:           0,
		WiFiCaps:          0,
	}
	reportingPolicy := device.ResolveControllerReportingPolicy(profile, id.Model, id.Version)
	if reportingPolicy.IdleSpeedtest {
		idle := gatewaySpeedtestStatus()
		telemetry.SpeedtestStatus = &idle
		telemetry.SpeedtestStatusUDAPI = &emptyList
	}
	if id.GWCapabilitiesRequested {
		telemetry.GWCapabilities = &gatewayCapabilities{
			GTI:                true,
			HB46PPIPIP:         true,
			HB46PPMapEHubSpoke: true,
			JPIXMapE:           true,
			MDNSTable:          true,
			NTTMapE:            true,
			WANMagic:           true,
		}
	}
	if id.IPv4ActiveLeasesRequested {
		telemetry.IPv4ActiveLeases = &emptyList
	}
	if id.UDAPIVersionExcluded {
		telemetry.UDAPIVersion = nil
	}
	if id.DPIStatsExcluded {
		telemetry.DPIStats = nil
	}
	if id.FingerprintsExcluded {
		telemetry.Fingerprints = nil
	}
	return telemetry
}

func saturatedUnixMilli(value time.Time) int64 {
	seconds := value.Unix()
	milliseconds := int64(value.Nanosecond() / int(time.Millisecond))
	if seconds > math.MaxInt64/1000 ||
		(seconds == math.MaxInt64/1000 && milliseconds > math.MaxInt64%1000) {
		return math.MaxInt64
	}
	if seconds < math.MinInt64/1000 {
		return math.MinInt64
	}
	return seconds*1000 + milliseconds
}

func gatewayUDAPIVersionRef() *udapiVersionPayload {
	return &udapiVersionPayload{
		Path:    "/system/ubios/udm/configuration",
		Version: 48,
		VersionDetail: map[string]int{
			"configuration": 48,
		},
	}
}

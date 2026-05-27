// Package payload fills low-risk deterministic gateway telemetry blocks that
// UniFi Network expects around gateway payloads. Values are placeholders and
// must not be copied from private controller state or firmware secrets.
package payload

import (
	"time"

	"github.com/konstruktor1/unifi-stubd/internal/device"
)

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
		DNSShield:    hashPayload{Hash: ""},
		DPIStats:     emptyList,
		Dualboot:     false,
		EverCrash:    false,
		Fingerprint:  "00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00",
		Fingerprints: emptyList,
		FW2Caps:      669375863,
		FWCaps:       1659698733,
		GuestKicks:   0,
		GuestToken:   "",
		GWCapabilities: gatewayCapabilities{
			GTI:                true,
			HB46PPIPIP:         true,
			HB46PPMapEHubSpoke: true,
			JPIXMapE:           true,
			MDNSTable:          true,
			NTTMapE:            true,
			WANMagic:           true,
		},
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
		SpeedtestStatusSaved: true,
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
		UDAPICaps:         1610591215,
		UDAPIVersion:      gatewayUDAPIVersion(),
		UpgradeDuration:   150,
		UptimeText:        formatUptime(uptime),
		USG2Caps:          0,
		USGCaps:           0,
		WiFiCaps:          0,
	}
}

func gatewayUDAPIVersion() udapiVersionPayload {
	return udapiVersionPayload{
		Path:    "/system/ubios/udm/configuration",
		Version: 48,
		VersionDetail: map[string]int{
			"configuration": 48,
		},
	}
}

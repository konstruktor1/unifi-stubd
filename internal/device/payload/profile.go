package payload

import (
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/device"
)

// isGatewayDeviceType reports whether a device type needs gateway-shaped tables.
func isGatewayDeviceType(deviceType string) bool {
	switch strings.TrimSpace(deviceType) {
	case deviceTypeUGW, deviceTypeUXG, deviceTypeUDM:
		return true
	default:
		return false
	}
}

// normalizePayloadProfile turns sparse profile metadata into the renderer
// defaults used by both legacy switch payloads and gateway-shaped payloads.
func normalizePayloadProfile(profile device.Profile, id device.Identity) device.Profile {
	profile.Payload.Kind = strings.ToLower(strings.TrimSpace(profile.Payload.Kind))
	if profile.Payload.Kind == "" {
		if isGatewayDeviceType(deviceTypeOrDefault(id.DeviceType)) {
			profile.Payload.Kind = payloadKindGateway
		} else {
			profile.Payload.Kind = payloadKindSwitch
		}
	}
	if profile.Payload.Kind != payloadKindGateway {
		profile.Payload.Kind = payloadKindSwitch
	}
	if strings.TrimSpace(profile.Payload.RequiredVersion) == "" {
		profile.Payload.RequiredVersion = defaultRequiredVersion
	}
	if strings.TrimSpace(profile.Payload.ManagementInterface) == "" {
		profile.Payload.ManagementInterface = "eth0"
	}
	if strings.TrimSpace(profile.Payload.GatewayInterfacePrefix) == "" {
		profile.Payload.GatewayInterfacePrefix = "eth"
	}
	return profile
}

// deviceTypeOrDefault keeps older switch payloads usable when no type is configured.
func deviceTypeOrDefault(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return deviceTypeUSW
	}
	return value
}

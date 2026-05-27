package device

import "strings"

// normalizeProfile delegates model-level normalization shared by built-in and
// external profiles.
func normalizeProfile(profile *Profile) {
	NormalizeProfile(profile)
}

// applyProfileDefaults fills renderer metadata that older or minimal profiles
// may omit, without changing fields explicitly set in YAML.
func applyProfileDefaults(profile *Profile) {
	setDefaultInt(&profile.SchemaVersion, schemaVersion)
	for _, field := range []struct {
		target *string
		value  string
	}{
		{target: &profile.Stability, value: "tested"},
		{target: &profile.Payload.Kind, value: defaultPayloadKind(profile.DeviceType)},
		{target: &profile.Payload.RequiredVersion, value: defaultRequiredVersion},
		{target: &profile.Payload.ManagementInterface, value: defaultMgmtInterface},
		{target: &profile.Payload.GatewayInterfacePrefix, value: defaultGatewayPrefix},
	} {
		setDefaultString(field.target, field.value)
	}
}

// defaultPayloadKind selects gateway-shaped payloads only for gateway device
// families; switches remain the conservative default.
func defaultPayloadKind(deviceType string) string {
	switch strings.TrimSpace(deviceType) {
	case "ugw", "uxg", "udm":
		return payloadKindGateway
	default:
		return payloadKindSwitch
	}
}

// setDefaultString fills profile defaults after decode without overwriting YAML
// values.
func setDefaultString(target *string, value string) {
	if *target == "" {
		*target = value
	}
}

// setDefaultInt fills numeric profile defaults after decode without overwriting
// YAML values.
func setDefaultInt(target *int, value int) {
	if *target == 0 {
		*target = value
	}
}

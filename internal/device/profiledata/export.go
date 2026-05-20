// Package profiledata renders validated records as canonical YAML for template
// and export CLI actions. The registry remains the source of truth for defaults
// and built-in override markers.
package profiledata

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// ExportYAML returns a profile as canonical YAML.
func (r Registry) ExportYAML(name string) ([]byte, error) {
	profile, ok := r.Lookup(name)
	if !ok {
		return nil, fmt.Errorf("profile %q not found", strings.TrimSpace(name))
	}
	if profile.SourceType == sourceTypeBuiltIn {
		profile.AllowBuiltinOverride = true
	}
	return profileYAML(profile)
}

// TemplateYAML returns a starter profile template for kind.
func TemplateYAML(kind string) ([]byte, error) {
	kind = strings.ToLower(strings.TrimSpace(kind))
	switch kind {
	case payloadKindSwitch:
		return profileYAML(Profile{
			SchemaVersion: schemaVersion,
			Name:          "custom-switch",
			Model:         "CUSTOMSW",
			ModelDisplay:  "Custom Lab Switch",
			DeviceType:    "usw",
			Version:       "7.4.1.16850",
			Ports:         8,
			PortSpeed:     1000,
			UplinkSpeed:   1000,
			PortMedia:     "GE",
			UplinkMedia:   "GE",
			Stability:     stabilityExternal,
			Recommended:   false,
			Payload: PayloadProfile{
				Kind:                payloadKindSwitch,
				RequiredVersion:     defaultRequiredVersion,
				ManagementInterface: defaultMgmtInterface,
			},
			Description: "external lab switch profile",
		})
	case payloadKindGateway:
		return profileYAML(Profile{
			SchemaVersion: schemaVersion,
			Name:          "custom-gateway",
			Model:         "CUSTOMGW",
			ModelDisplay:  "Custom Lab Gateway",
			DeviceType:    "uxg",
			Version:       "5.0.16.30689",
			Ports:         2,
			PortNames:     []string{"WAN", "LAN"},
			PortRoles:     []string{"wan", "lan"},
			PortNetworkGroups: []string{
				"WAN",
				"LAN",
			},
			PortSpeed:   1000,
			UplinkSpeed: 1000,
			PortMedia:   "GE",
			UplinkMedia: "GE",
			Stability:   stabilityExternal,
			Recommended: false,
			Payload: PayloadProfile{
				Kind:                   payloadKindGateway,
				RequiredVersion:        defaultRequiredVersion,
				ManagementInterface:    defaultMgmtInterface,
				GatewayInterfacePrefix: defaultGatewayPrefix,
				HasDPI:                 false,
			},
			Description: "external lab gateway profile",
		})
	default:
		return nil, fmt.Errorf("invalid profile template kind %q; use switch or gateway", kind)
	}
}

func profileYAML(profile Profile) ([]byte, error) {
	profile = cloneProfile(profile)
	profile.Source = ""
	profile.SourceType = ""
	profile.SchemaVersion = firstNonZero(profile.SchemaVersion, schemaVersion)
	profile.Order = 0
	data, err := yaml.Marshal(profile)
	if err != nil {
		return nil, fmt.Errorf("marshal profile YAML: %w", err)
	}
	return data, nil
}

func classifyProfileError(err error) string {
	if err == nil {
		return ""
	}
	if strings.Contains(err.Error(), "decode profile YAML") {
		return ErrorKindParse
	}
	return ErrorKindValidation
}

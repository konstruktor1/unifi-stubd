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
	recommended := profile.Recommended
	hasDPI := profile.Payload.HasDPI
	cfg := profileConfig{
		SchemaVersion:               firstNonZero(profile.SchemaVersion, schemaVersion),
		Extends:                     profile.Extends,
		AllowBuiltinOverride:        profile.AllowBuiltinOverride,
		Name:                        profile.Name,
		Model:                       profile.Model,
		ModelDisplay:                profile.ModelDisplay,
		DeviceType:                  profile.DeviceType,
		Version:                     profile.Version,
		Ports:                       profile.Ports,
		PortGroups:                  clonePortGroups(profile.PortGroups),
		PortNames:                   cloneStrings(profile.PortNames),
		PortRoles:                   cloneStrings(profile.PortRoles),
		PortNetworkGroups:           cloneStrings(profile.PortNetworkGroups),
		PortSpeed:                   profile.PortSpeed,
		UplinkSpeed:                 profile.UplinkSpeed,
		PortMedia:                   profile.PortMedia,
		UplinkMedia:                 profile.UplinkMedia,
		Stability:                   profile.Stability,
		Recommended:                 &recommended,
		ValidatedControllerVersions: cloneStrings(profile.ValidatedControllerVersions),
		Payload: payloadConfig{
			Kind:                   profile.Payload.Kind,
			RequiredVersion:        profile.Payload.RequiredVersion,
			ManagementInterface:    profile.Payload.ManagementInterface,
			GatewayInterfacePrefix: profile.Payload.GatewayInterfacePrefix,
			HasDPI:                 &hasDPI,
		},
		Description: profile.Description,
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("marshal profile YAML: %w", err)
	}
	return data, nil
}

func classifyProfileError(err error) string {
	if err == nil {
		return ""
	}
	if strings.Contains(err.Error(), "decode profile config") {
		return ErrorKindParse
	}
	return ErrorKindValidation
}

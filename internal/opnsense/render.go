package opnsense

import (
	"fmt"

	appconfig "github.com/konstruktor1/unifi-stubd/internal/config"
	"gopkg.in/yaml.v3"
)

// GenerateConfig overlays OPNsense-derived port data onto a loaded base config.
func GenerateConfig(base appconfig.Config, source SourceConfig, interfaces map[string]InterfaceStatus, gateways map[string]GatewayStatus) appconfig.Config {
	out := base
	generatedOverrides := OverridesFromState(source.Interfaces, interfaces, gateways)
	out.PortOverrides = MergeOverrides(generatedOverrides, base.PortOverrides)
	if source.UplinkPort > 0 && out.UplinkPort == 0 {
		out.UplinkPort = source.UplinkPort
	}
	if source.WANHealth.Source != "" {
		out.WANHealth = source.WANHealth
	}
	return out
}

// MarshalConfig renders the generated config as reviewable YAML.
func MarshalConfig(cfg appconfig.Config) ([]byte, error) {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("marshal generated unifi-stubd config: %w", err)
	}
	return data, nil
}

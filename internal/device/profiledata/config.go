// Package profiledata loads and stores embedded device profile data.
package profiledata

// This file decodes embedded YAML profile records into validated profiles.

import (
	"bytes"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// profileConfig is the embedded YAML schema before validation and registration.
type profileConfig struct {
	// Order controls profile display order.
	Order int `yaml:"order"`
	// Name is the short CLI and config name.
	Name string `yaml:"name"`
	// Model is the UniFi model identifier.
	Model string `yaml:"model"`
	// ModelDisplay is the human-readable UniFi model name.
	ModelDisplay string `yaml:"model_display"`
	// DeviceType is the controller-facing UniFi device family.
	DeviceType string `yaml:"device_type"`
	// Version is the firmware version reported by this profile.
	Version string `yaml:"version"`
	// Ports is the number of reported Ethernet ports.
	Ports int `yaml:"ports"`
	// PortGroups describe non-uniform physical port layouts.
	PortGroups []PortGroup `yaml:"port_groups"`
	// PortNames optionally override one-based port display labels.
	PortNames []string `yaml:"port_names"`
	// PortRoles optionally define one-based gateway roles for the profile.
	PortRoles []string `yaml:"port_roles"`
	// PortNetworkGroups optionally define one-based UniFi network groups.
	PortNetworkGroups []string `yaml:"port_network_groups"`
	// PortSpeed is the default access port speed in Mbps.
	PortSpeed int `yaml:"port_speed"`
	// UplinkSpeed is the uplink port speed in Mbps.
	UplinkSpeed int `yaml:"uplink_speed"`
	// PortMedia is the default access port media label.
	PortMedia string `yaml:"port_media"`
	// UplinkMedia is the uplink port media label.
	UplinkMedia string `yaml:"uplink_media"`
	// Description is the short label shown in profile listings.
	Description string `yaml:"description"`
}

// RegisterConfig decodes and registers one embedded profile YAML document.
func RegisterConfig(source string, data []byte) {
	profile, order, err := decodeProfileConfig(data)
	if err != nil {
		panic(fmt.Sprintf("load profile %s: %v", source, err))
	}
	Register(source, order, profile)
}

func decodeProfileConfig(data []byte) (Profile, int, error) {
	var cfg profileConfig
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return Profile{}, 0, fmt.Errorf("decode profile config: %w", err)
	}
	profile := Profile{
		Name:              strings.TrimSpace(cfg.Name),
		Model:             strings.TrimSpace(cfg.Model),
		ModelDisplay:      strings.TrimSpace(cfg.ModelDisplay),
		DeviceType:        strings.TrimSpace(cfg.DeviceType),
		Version:           strings.TrimSpace(cfg.Version),
		Ports:             cfg.Ports,
		PortGroups:        clonePortGroups(cfg.PortGroups),
		PortNames:         cloneStrings(cfg.PortNames),
		PortRoles:         cloneStrings(cfg.PortRoles),
		PortNetworkGroups: cloneStrings(cfg.PortNetworkGroups),
		PortSpeed:         cfg.PortSpeed,
		UplinkSpeed:       cfg.UplinkSpeed,
		PortMedia:         strings.TrimSpace(cfg.PortMedia),
		UplinkMedia:       strings.TrimSpace(cfg.UplinkMedia),
		Description:       strings.TrimSpace(cfg.Description),
	}
	if err := validateProfile(profile); err != nil {
		return Profile{}, 0, err
	}
	return profile, cfg.Order, nil
}

func validateProfile(profile Profile) error {
	if profile.Name == "" {
		return fmt.Errorf("name is required")
	}
	if profile.Model == "" {
		return fmt.Errorf("model is required for %q", profile.Name)
	}
	if profile.Ports < 1 {
		return fmt.Errorf("ports must be positive for %q", profile.Name)
	}
	if profile.PortSpeed < 0 {
		return fmt.Errorf("port_speed must not be negative for %q", profile.Name)
	}
	if profile.UplinkSpeed < 0 {
		return fmt.Errorf("uplink_speed must not be negative for %q", profile.Name)
	}
	return nil
}

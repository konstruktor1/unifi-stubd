// Package profiledata loads and stores embedded device profile data.
package profiledata

// This file decodes embedded YAML profile records into validated profiles.

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	schemaVersion          = 1
	payloadKindSwitch      = "switch"
	payloadKindGateway     = "gateway"
	defaultRequiredVersion = "5.0.0"
	defaultMgmtInterface   = "eth0"
	defaultGatewayPrefix   = "eth"
)

// profileConfig is the embedded YAML schema before validation and registration.
type profileConfig struct {
	// SchemaVersion is the profile YAML schema version.
	SchemaVersion int `yaml:"schema_version"`
	// Extends names a base profile for external derived profiles.
	Extends string `yaml:"extends"`
	// AllowBuiltinOverride permits an external profile to replace a built-in profile.
	AllowBuiltinOverride bool `yaml:"allow_builtin_override"`
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
	// Stability describes profile maturity, such as tested or experimental.
	Stability string `yaml:"stability"`
	// Recommended marks profiles that should be preferred for ordinary labs.
	Recommended *bool `yaml:"recommended"`
	// ValidatedControllerVersions lists controller versions validated in the lab.
	ValidatedControllerVersions []string `yaml:"validated_controller_versions"`
	// Payload controls renderer behavior for this profile.
	Payload payloadConfig `yaml:"payload"`
	// Description is the short label shown in profile listings.
	Description string `yaml:"description"`
}

type payloadConfig struct {
	// Kind selects the generic payload renderer: switch or gateway.
	Kind string `yaml:"kind"`
	// RequiredVersion is reported in the inform payload.
	RequiredVersion string `yaml:"required_version"`
	// ManagementInterface is the controller-facing management interface name.
	ManagementInterface string `yaml:"management_interface"`
	// GatewayInterfacePrefix prefixes generated gateway interface names.
	GatewayInterfacePrefix string `yaml:"gateway_interface_prefix"`
	// HasDPI reports whether gateway DPI capability should be advertised.
	HasDPI *bool `yaml:"has_dpi"`
}

// RegisterConfig decodes and registers one embedded profile YAML document.
func RegisterConfig(source string, data []byte) {
	profile, order, err := DecodeConfig(data)
	if err != nil {
		panic(fmt.Sprintf("load profile %s: %v", source, err))
	}
	Register(source, order, profile)
}

// LoadPath loads one profile YAML file or all profile YAML files in a directory.
func (r *Registry) LoadPath(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return &PathError{Path: path, Kind: ErrorKindIO, Err: err}
	}
	if !info.IsDir() {
		return r.LoadFile(path)
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return &PathError{Path: path, Kind: ErrorKindIO, Err: err}
	}
	var paths []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if ext == ".yaml" || ext == ".yml" {
			paths = append(paths, filepath.Join(path, name))
		}
	}
	sort.Strings(paths)
	for _, profilePath := range paths {
		if err := r.LoadFile(profilePath); err != nil {
			return err
		}
	}
	return nil
}

// LoadFile loads one external profile YAML file into r.
func (r *Registry) LoadFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return &PathError{Path: path, Kind: ErrorKindIO, Err: err}
	}
	profile, order, err := r.DecodeExternalConfig(data)
	if err != nil {
		return &PathError{Path: path, Kind: classifyProfileError(err), Err: err}
	}
	if err := r.register(path, order, profile, false); err != nil {
		return &PathError{Path: path, Kind: ErrorKindValidation, Err: err}
	}
	return nil
}

// DecodeConfig decodes and validates one standalone profile YAML document.
func DecodeConfig(data []byte) (Profile, int, error) {
	cfg, err := decodeProfileConfig(data)
	if err != nil {
		return Profile{}, 0, err
	}
	profile := profileFromConfig(cfg)
	applyProfileDefaults(&profile)
	if err := validateProfile(profile); err != nil {
		return Profile{}, 0, err
	}
	return profile, cfg.Order, nil
}

// DecodeExternalConfig decodes a profile YAML document with optional inheritance.
func (r *Registry) DecodeExternalConfig(data []byte) (Profile, int, error) {
	cfg, err := decodeProfileConfig(data)
	if err != nil {
		return Profile{}, 0, err
	}
	profile := profileFromConfig(cfg)
	if strings.TrimSpace(profile.Extends) != "" {
		base, ok := r.lookup(profile.Extends)
		if !ok {
			return Profile{}, 0, fmt.Errorf("extends %q not found", profile.Extends)
		}
		profile = mergeProfileConfig(base, cfg)
	}
	applyProfileDefaults(&profile)
	if err := validateProfile(profile); err != nil {
		return Profile{}, 0, err
	}
	return profile, cfg.Order, nil
}

func decodeProfileConfig(data []byte) (profileConfig, error) {
	var cfg profileConfig
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return profileConfig{}, fmt.Errorf("decode profile config: %w", err)
	}
	return cfg, nil
}

func profileFromConfig(cfg profileConfig) Profile {
	recommended := false
	if cfg.Recommended != nil {
		recommended = *cfg.Recommended
	}
	hasDPI := false
	if cfg.Payload.HasDPI != nil {
		hasDPI = *cfg.Payload.HasDPI
	}
	return Profile{
		SchemaVersion:               cfg.SchemaVersion,
		Extends:                     strings.TrimSpace(cfg.Extends),
		AllowBuiltinOverride:        cfg.AllowBuiltinOverride,
		Name:                        strings.TrimSpace(cfg.Name),
		Model:                       strings.TrimSpace(cfg.Model),
		ModelDisplay:                strings.TrimSpace(cfg.ModelDisplay),
		DeviceType:                  strings.TrimSpace(cfg.DeviceType),
		Version:                     strings.TrimSpace(cfg.Version),
		Ports:                       cfg.Ports,
		PortGroups:                  clonePortGroups(cfg.PortGroups),
		PortNames:                   cloneStrings(cfg.PortNames),
		PortRoles:                   cloneStrings(cfg.PortRoles),
		PortNetworkGroups:           cloneStrings(cfg.PortNetworkGroups),
		PortSpeed:                   cfg.PortSpeed,
		UplinkSpeed:                 cfg.UplinkSpeed,
		PortMedia:                   strings.TrimSpace(cfg.PortMedia),
		UplinkMedia:                 strings.TrimSpace(cfg.UplinkMedia),
		Stability:                   strings.TrimSpace(cfg.Stability),
		Recommended:                 recommended,
		ValidatedControllerVersions: cloneStrings(cfg.ValidatedControllerVersions),
		Payload: PayloadProfile{
			Kind:                   strings.TrimSpace(cfg.Payload.Kind),
			RequiredVersion:        strings.TrimSpace(cfg.Payload.RequiredVersion),
			ManagementInterface:    strings.TrimSpace(cfg.Payload.ManagementInterface),
			GatewayInterfacePrefix: strings.TrimSpace(cfg.Payload.GatewayInterfacePrefix),
			HasDPI:                 hasDPI,
		},
		Description: strings.TrimSpace(cfg.Description),
	}
}

func mergeProfileConfig(base Profile, cfg profileConfig) Profile {
	profile := cloneProfile(base)
	overlay := profileFromConfig(cfg)
	profile.SchemaVersion = overlay.SchemaVersion
	profile.Extends = overlay.Extends
	profile.AllowBuiltinOverride = overlay.AllowBuiltinOverride
	if overlay.Name != "" {
		profile.Name = overlay.Name
	}
	if overlay.Model != "" {
		profile.Model = overlay.Model
	}
	if overlay.ModelDisplay != "" {
		profile.ModelDisplay = overlay.ModelDisplay
	}
	if overlay.DeviceType != "" {
		profile.DeviceType = overlay.DeviceType
	}
	if overlay.Version != "" {
		profile.Version = overlay.Version
	}
	if overlay.Ports != 0 {
		profile.Ports = overlay.Ports
	}
	if len(overlay.PortGroups) > 0 {
		profile.PortGroups = clonePortGroups(overlay.PortGroups)
	}
	if len(overlay.PortNames) > 0 {
		profile.PortNames = cloneStrings(overlay.PortNames)
	}
	if len(overlay.PortRoles) > 0 {
		profile.PortRoles = cloneStrings(overlay.PortRoles)
	}
	if len(overlay.PortNetworkGroups) > 0 {
		profile.PortNetworkGroups = cloneStrings(overlay.PortNetworkGroups)
	}
	if overlay.PortSpeed != 0 {
		profile.PortSpeed = overlay.PortSpeed
	}
	if overlay.UplinkSpeed != 0 {
		profile.UplinkSpeed = overlay.UplinkSpeed
	}
	if overlay.PortMedia != "" {
		profile.PortMedia = overlay.PortMedia
	}
	if overlay.UplinkMedia != "" {
		profile.UplinkMedia = overlay.UplinkMedia
	}
	if overlay.Stability != "" {
		profile.Stability = overlay.Stability
	}
	if cfg.Recommended != nil {
		profile.Recommended = overlay.Recommended
	}
	if len(overlay.ValidatedControllerVersions) > 0 {
		profile.ValidatedControllerVersions = cloneStrings(overlay.ValidatedControllerVersions)
	}
	if overlay.Payload.Kind != "" {
		profile.Payload.Kind = overlay.Payload.Kind
	}
	if overlay.Payload.RequiredVersion != "" {
		profile.Payload.RequiredVersion = overlay.Payload.RequiredVersion
	}
	if overlay.Payload.ManagementInterface != "" {
		profile.Payload.ManagementInterface = overlay.Payload.ManagementInterface
	}
	if overlay.Payload.GatewayInterfacePrefix != "" {
		profile.Payload.GatewayInterfacePrefix = overlay.Payload.GatewayInterfacePrefix
	}
	if cfg.Payload.HasDPI != nil {
		profile.Payload.HasDPI = overlay.Payload.HasDPI
	}
	if overlay.Description != "" {
		profile.Description = overlay.Description
	}
	return profile
}

func applyProfileDefaults(profile *Profile) {
	if profile.SchemaVersion == 0 {
		profile.SchemaVersion = schemaVersion
	}
	if profile.Stability == "" {
		profile.Stability = "tested"
	}
	if profile.Payload.Kind == "" {
		profile.Payload.Kind = defaultPayloadKind(profile.DeviceType)
	}
	if profile.Payload.RequiredVersion == "" {
		profile.Payload.RequiredVersion = defaultRequiredVersion
	}
	if profile.Payload.ManagementInterface == "" {
		profile.Payload.ManagementInterface = defaultMgmtInterface
	}
	if profile.Payload.GatewayInterfacePrefix == "" {
		profile.Payload.GatewayInterfacePrefix = defaultGatewayPrefix
	}
}

func validateProfile(profile Profile) error {
	if profile.SchemaVersion != schemaVersion {
		return fmt.Errorf("schema_version must be %d", schemaVersion)
	}
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
	if err := validatePortGroups(profile); err != nil {
		return err
	}
	if err := validateOneBasedStrings("port_names", profile.Name, profile.Ports, profile.PortNames); err != nil {
		return err
	}
	if err := validatePortRoles(profile); err != nil {
		return err
	}
	if err := validateOneBasedStrings("port_network_groups", profile.Name, profile.Ports, profile.PortNetworkGroups); err != nil {
		return err
	}
	if err := validatePayload(profile); err != nil {
		return err
	}
	return nil
}

func validatePortGroups(profile Profile) error {
	total := 0
	uplinkGroups := 0
	for index, group := range profile.PortGroups {
		if group.Count < 1 {
			return fmt.Errorf("port_groups[%d].count must be positive for %q", index, profile.Name)
		}
		if group.Speed < 0 {
			return fmt.Errorf("port_groups[%d].speed must not be negative for %q", index, profile.Name)
		}
		if group.Uplink {
			uplinkGroups++
		}
		total += group.Count
	}
	if len(profile.PortGroups) > 0 && total != profile.Ports {
		return fmt.Errorf("port_groups total %d != ports %d", total, profile.Ports)
	}
	if uplinkGroups > 1 {
		return fmt.Errorf("only one port_groups entry may set uplink for %q", profile.Name)
	}
	return nil
}

func validateOneBasedStrings(field, name string, ports int, values []string) error {
	if len(values) > ports {
		return fmt.Errorf("%s length %d exceeds ports %d for %q", field, len(values), ports, name)
	}
	for index, value := range values {
		if strings.ContainsAny(value, "\r\n\t") {
			return fmt.Errorf("%s[%d] contains unsupported whitespace for %q", field, index, name)
		}
	}
	return nil
}

func validatePortRoles(profile Profile) error {
	if err := validateOneBasedStrings("port_roles", profile.Name, profile.Ports, profile.PortRoles); err != nil {
		return err
	}
	for index, role := range profile.PortRoles {
		role = strings.ToLower(strings.TrimSpace(role))
		if role == "" {
			continue
		}
		switch role {
		case "wan", "lan", "wan2", "lan2":
		default:
			return fmt.Errorf("port_roles[%d] has invalid role %q; use wan, lan, wan2, or lan2", index, role)
		}
	}
	return nil
}

func validatePayload(profile Profile) error {
	switch strings.ToLower(strings.TrimSpace(profile.Payload.Kind)) {
	case payloadKindSwitch, payloadKindGateway:
	default:
		return fmt.Errorf("payload.kind %q is invalid; use switch or gateway", profile.Payload.Kind)
	}
	if strings.TrimSpace(profile.Payload.ManagementInterface) == "" {
		return fmt.Errorf("payload.management_interface is required for %q", profile.Name)
	}
	if strings.Contains(profile.Payload.ManagementInterface, "/") {
		return fmt.Errorf("payload.management_interface %q is invalid for %q", profile.Payload.ManagementInterface, profile.Name)
	}
	if strings.TrimSpace(profile.Payload.GatewayInterfacePrefix) == "" {
		return fmt.Errorf("payload.gateway_interface_prefix is required for %q", profile.Name)
	}
	if strings.ContainsAny(profile.Payload.GatewayInterfacePrefix, "/ \t\r\n") {
		return fmt.Errorf("payload.gateway_interface_prefix %q is invalid for %q", profile.Payload.GatewayInterfacePrefix, profile.Name)
	}
	return nil
}

func defaultPayloadKind(deviceType string) string {
	switch strings.TrimSpace(deviceType) {
	case "ugw", "uxg", "udm":
		return payloadKindGateway
	default:
		return payloadKindSwitch
	}
}

// Package profilemodel contains the canonical typed device profile model.
package profilemodel

import (
	"strings"

	payloadpkg "github.com/konstruktor1/unifi-stubd/internal/device/payload"
)

// Profile defines a UniFi device profile.
type Profile struct {
	// Source is the built-in source path or external profile path.
	Source string `yaml:"-"`
	// SourceType is built-in or external.
	SourceType string `yaml:"-"`
	// SchemaVersion is the profile YAML schema version.
	SchemaVersion int `yaml:"schema_version"`
	// Extends names an existing profile used as the base for this profile.
	Extends string `yaml:"extends"`
	// AllowBuiltinOverride permits an external profile to replace a built-in record.
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
	Recommended bool `yaml:"recommended"`
	// ValidatedControllerVersions lists controller versions validated in the lab.
	ValidatedControllerVersions []string `yaml:"validated_controller_versions"`
	// Payload controls renderer behavior for this profile.
	Payload PayloadProfile `yaml:"payload"`
	// Description is the short label shown in profile listings.
	Description string `yaml:"description"`
}

// PortGroup describes one contiguous block in a profile port layout.
type PortGroup = payloadpkg.PortGroup

// PayloadProfile contains profile-driven inform payload rendering metadata.
type PayloadProfile struct {
	// Kind selects the generic payload renderer: switch or gateway.
	Kind string `yaml:"kind"`
	// RequiredVersion is reported in the inform payload.
	RequiredVersion string `yaml:"required_version"`
	// ManagementInterface is the controller-facing management interface name.
	ManagementInterface string `yaml:"management_interface"`
	// GatewayInterfacePrefix prefixes generated gateway interface names.
	GatewayInterfacePrefix string `yaml:"gateway_interface_prefix"`
	// HasDPI reports whether gateway DPI capability should be advertised.
	HasDPI bool `yaml:"has_dpi"`
}

// PortOptions converts p to generated switch port options.
func (p Profile) PortOptions() payloadpkg.PortOptions {
	return payloadpkg.PortOptions{
		Speed:             p.PortSpeed,
		UplinkSpeed:       p.UplinkSpeed,
		Media:             p.PortMedia,
		UplinkMedia:       p.UplinkMedia,
		UplinkPort:        0,
		PortGroups:        cloneNonEmptySlice(p.PortGroups),
		PortNames:         cloneNonEmptySlice(p.PortNames),
		PortRoles:         cloneNonEmptySlice(p.PortRoles),
		PortNetworkGroups: cloneNonEmptySlice(p.PortNetworkGroups),
	}
}

// PayloadOptions converts p to payload renderer options.
func (p Profile) PayloadOptions() payloadpkg.Profile {
	return payloadpkg.Profile{
		Kind:                   p.Payload.Kind,
		RequiredVersion:        p.Payload.RequiredVersion,
		ManagementInterface:    p.Payload.ManagementInterface,
		GatewayInterfacePrefix: p.Payload.GatewayInterfacePrefix,
		HasDPI:                 p.Payload.HasDPI,
	}
}

// Normalize trims user-facing string fields in profile.
func Normalize(profile *Profile) {
	trimStrings(
		&profile.Extends,
		&profile.Name,
		&profile.Model,
		&profile.ModelDisplay,
		&profile.DeviceType,
		&profile.Version,
		&profile.PortMedia,
		&profile.UplinkMedia,
		&profile.Stability,
		&profile.Payload.Kind,
		&profile.Payload.RequiredVersion,
		&profile.Payload.ManagementInterface,
		&profile.Payload.GatewayInterfacePrefix,
		&profile.Description,
	)
}

func trimStrings(values ...*string) {
	for _, value := range values {
		*value = strings.TrimSpace(*value)
	}
}

func cloneNonEmptySlice[T any](values []T) []T {
	if len(values) == 0 {
		return nil
	}
	out := make([]T, len(values))
	copy(out, values)
	return out
}

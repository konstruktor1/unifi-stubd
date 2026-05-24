// Package device defines the canonical profile model. AutoMAC provides
// deterministic local identity generation for lab defaults.
package device

import (
	"crypto/sha256"
	"net"
	"strings"
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

// NormalizeProfile trims user-facing string fields in profile.
func NormalizeProfile(profile *Profile) {
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

// AutoMAC derives a stable locally administered MAC address from seed.
func AutoMAC(seed string) net.HardwareAddr {
	sum := sha256.Sum256([]byte(strings.TrimSpace(seed)))
	mac := net.HardwareAddr{sum[0], sum[1], sum[2], sum[3], sum[4], sum[5]}
	mac[0] = (mac[0] | 0x02) & 0xfe
	return mac
}

// trimStrings normalizes profile text before validation and registry storage.
func trimStrings(values ...*string) {
	for _, value := range values {
		*value = strings.TrimSpace(*value)
	}
}

package device

import (
	"crypto/sha256"
	"net"
	"strings"
)

// Profile defines a built-in UniFi device profile.
type Profile struct {
	// Source is the built-in source path or external profile path.
	Source string
	// SourceType is built-in or external.
	SourceType string
	// SchemaVersion is the profile YAML schema version.
	SchemaVersion int
	// Extends names an existing profile used as the base for this profile.
	Extends string
	// AllowBuiltinOverride permits an external profile to replace a built-in record.
	AllowBuiltinOverride bool
	// Name is the short CLI and config name.
	Name string
	// Model is the UniFi model identifier.
	Model string
	// ModelDisplay is the human-readable UniFi model name.
	ModelDisplay string
	// DeviceType is the controller-facing UniFi device family.
	DeviceType string
	// Version is the firmware version reported by this profile.
	Version string
	// Ports is the number of reported Ethernet ports.
	Ports int
	// PortGroups describe non-uniform physical port layouts.
	PortGroups []PortGroup
	// PortNames optionally override one-based port display labels.
	PortNames []string
	// PortRoles optionally define one-based gateway roles for the profile.
	PortRoles []string
	// PortNetworkGroups optionally define one-based UniFi network groups.
	PortNetworkGroups []string
	// PortSpeed is the default access port speed in Mbps.
	PortSpeed int
	// UplinkSpeed is the uplink port speed in Mbps.
	UplinkSpeed int
	// PortMedia is the default access port media label.
	PortMedia string
	// UplinkMedia is the uplink port media label.
	UplinkMedia string
	// Stability describes profile maturity, such as tested or experimental.
	Stability string
	// Recommended marks profiles that should be preferred for ordinary labs.
	Recommended bool
	// ValidatedControllerVersions lists controller versions validated in the lab.
	ValidatedControllerVersions []string
	// Payload controls renderer behavior for this profile.
	Payload PayloadProfile
	// Description is the short label shown in profile listings.
	Description string
}

// PayloadProfile contains profile-driven inform payload rendering metadata.
type PayloadProfile struct {
	// Kind selects the generic payload renderer: switch or gateway.
	Kind string
	// RequiredVersion is reported in the inform payload.
	RequiredVersion string
	// ManagementInterface is the controller-facing management interface name.
	ManagementInterface string
	// GatewayInterfacePrefix prefixes generated gateway interface names.
	GatewayInterfacePrefix string
	// HasDPI reports whether gateway DPI capability should be advertised.
	HasDPI bool
}

// PortOptions converts p to generated switch port options.
func (p Profile) PortOptions() PortOptions {
	return PortOptions{
		Speed:             p.PortSpeed,
		UplinkSpeed:       p.UplinkSpeed,
		Media:             p.PortMedia,
		UplinkMedia:       p.UplinkMedia,
		UplinkPort:        0,
		PortGroups:        clonePortGroups(p.PortGroups),
		PortNames:         cloneStrings(p.PortNames),
		PortRoles:         cloneStrings(p.PortRoles),
		PortNetworkGroups: cloneStrings(p.PortNetworkGroups),
	}
}

// PayloadOptions converts p to payload renderer options.
func (p Profile) PayloadOptions() PayloadProfile {
	return PayloadProfile{
		Kind:                   p.Payload.Kind,
		RequiredVersion:        p.Payload.RequiredVersion,
		ManagementInterface:    p.Payload.ManagementInterface,
		GatewayInterfacePrefix: p.Payload.GatewayInterfacePrefix,
		HasDPI:                 p.Payload.HasDPI,
	}
}

func clonePortGroups(groups []PortGroup) []PortGroup {
	if len(groups) == 0 {
		return nil
	}
	out := make([]PortGroup, len(groups))
	copy(out, groups)
	return out
}

func cloneStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, len(values))
	copy(out, values)
	return out
}

// AutoMAC derives a stable locally administered MAC address from seed.
func AutoMAC(seed string) net.HardwareAddr {
	sum := sha256.Sum256([]byte(strings.TrimSpace(seed)))
	mac := net.HardwareAddr{sum[0], sum[1], sum[2], sum[3], sum[4], sum[5]}
	mac[0] = (mac[0] | 0x02) & 0xfe
	return mac
}

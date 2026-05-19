package profiledata

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

// PortGroup describes one contiguous block in a profile port layout.
type PortGroup struct {
	// Count is the number of ports in this block.
	Count int
	// Speed is the negotiated speed in Mbps for ports in this block.
	Speed int
	// Media is the UniFi media label for ports in this block.
	Media string
	// Uplink marks the first port in this block as the upstream connection.
	Uplink bool
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

// Package device turns profile layout data into deterministic UniFi ports
// before observations and overrides are merged. The generator preserves profile
// media, speed groups, names, roles, and uplink selection.
package device

const (
	deviceTypeUSW       = "usw"
	gatewayPortRoleLAN  = "lan"
	gatewayPortRoleLAN2 = "lan2"
	gatewayPortRoleNone = "unassigned"
	gatewayPortRoleWAN  = "wan"
	gatewayPortRoleWAN2 = "wan2"
	mediaSFPPlus        = "SFP+"
)

// BuildPorts returns generated switch ports from profile plus runtime options.
func BuildPorts(profile Profile, options PortBuildOptions) []Port {
	count := profile.Ports
	if options.Count > 0 {
		count = options.Count
	}
	return switchPortsWithLayout(count, profilePortLayout(profile, options))
}

// SwitchPorts returns count generated switch ports with profile-neutral defaults.
func SwitchPorts(count int) []Port {
	return BuildPorts(Profile{Ports: count}, PortBuildOptions{})
}

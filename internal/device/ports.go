package device

const (
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

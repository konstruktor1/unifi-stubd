package device

// ControllerReportingPolicy contains the controller-driven behavior enabled
// for one exact built-in product identity.
type ControllerReportingPolicy struct {
	BlockSelection  bool
	ControllerClock bool
	IdleSpeedtest   bool
	Reason          string
}

// ResolveControllerReportingPolicy resolves reporting controls without a
// family, model-prefix, version-prefix, or external-profile fallback.
func ResolveControllerReportingPolicy(profile Profile, effectiveModel, effectiveVersion string) ControllerReportingPolicy {
	productPolicy, ok := resolveUserProfileProtocol(profile, effectiveModel, effectiveVersion)
	if !ok {
		return ControllerReportingPolicy{Reason: profileProtocolNoBinding}
	}
	return ControllerReportingPolicy{
		BlockSelection:  productPolicy.blockSelection,
		ControllerClock: productPolicy.controllerClock,
		IdleSpeedtest:   productPolicy.idleSpeedtest,
	}
}

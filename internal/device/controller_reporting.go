package device

const (
	controllerReportingUnsupported = "exact profile evidence has no supported protocol behavior"
	controllerReportingNoControls  = "exact protocol behavior has no bound controller reporting controls"
)

// ControllerReportingPolicy binds controller-selected payload producers and
// controller time to one exact supported profile/build. Unsupported profiles
// may still exchange safe lifecycle messages, but these reporting controls
// remain inert.
type ControllerReportingPolicy struct {
	EvidenceID      string
	BehaviorID      string
	BlockSelection  bool
	ControllerClock bool
	Reason          string
}

// ResolveControllerReportingPolicy resolves reporting controls without a
// family, model-prefix, version-prefix, or external-profile fallback.
func ResolveControllerReportingPolicy(profile Profile, effectiveModel, effectiveVersion string) ControllerReportingPolicy {
	binding, ok := resolveUserProfileProtocol(profile, effectiveModel, effectiveVersion)
	if !ok {
		return ControllerReportingPolicy{
			BehaviorID: protocolBehaviorUnsupported,
			Reason:     profileProtocolNoBinding,
		}
	}
	policy := ControllerReportingPolicy{
		EvidenceID: binding.evidenceID,
		BehaviorID: binding.behavior.id(),
	}
	if !binding.behavior.supported() {
		policy.Reason = controllerReportingUnsupported
		return policy
	}
	policy.BlockSelection = binding.blockSelection
	policy.ControllerClock = binding.controllerClock
	if !policy.BlockSelection && !policy.ControllerClock {
		policy.Reason = controllerReportingNoControls
	}
	return policy
}

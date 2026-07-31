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
	BehaviorID      ProtocolBehaviorID
	BlockSelection  bool
	ControllerClock bool
	Reason          string
}

// ResolveControllerReportingPolicy resolves reporting controls without a
// family, model-prefix, version-prefix, or external-profile fallback.
func ResolveControllerReportingPolicy(profile Profile, effectiveModel, effectiveVersion string) ControllerReportingPolicy {
	protocolProfile, _, ok := resolveUserProfileProtocol(profile, effectiveModel, effectiveVersion)
	if !ok {
		return ControllerReportingPolicy{
			BehaviorID: ProtocolBehaviorUnsupported,
			Reason:     profileProtocolNoBinding,
		}
	}
	policy := ControllerReportingPolicy{
		EvidenceID: protocolProfile.EvidenceID,
		BehaviorID: protocolProfile.BehaviorID,
	}
	if !protocolProfile.IsSupported() {
		policy.Reason = controllerReportingUnsupported
		return policy
	}
	capabilities := protocolProfile.Policies().Capabilities
	policy.BlockSelection = capabilities.Has(CapabilityGatewayProducerTelemetry)
	policy.ControllerClock = capabilities.Has(CapabilityControllerClock)
	if !policy.BlockSelection && !policy.ControllerClock {
		policy.Reason = controllerReportingNoControls
	}
	return policy
}

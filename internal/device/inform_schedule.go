package device

import "time"

const (
	informScheduleNoProfileBinding = profileProtocolNoBinding
	informScheduleNoMinimumBinding = "firmware minimum is not cataloged for the exact profile and evidence binding"
)

// InformSchedulePolicy is the exact firmware-evidence decision for controller
// interval hints. A missing minimum disables hints without preventing the
// operator-selected base schedule from running.
type InformSchedulePolicy struct {
	EvidenceID      string
	ProtocolBuild   ProtocolBuild
	BehaviorID      ProtocolBehaviorID
	FirmwareMinimum time.Duration
	Supported       bool
	Reason          string
}

// EffectiveMinimum returns the stricter firmware or operator lower bound.
// Unsupported policies return zero because no controller hint may be applied.
func (policy InformSchedulePolicy) EffectiveMinimum(operatorMinimum time.Duration) time.Duration {
	if !policy.Supported || policy.FirmwareMinimum <= 0 {
		return 0
	}
	if operatorMinimum > policy.FirmwareMinimum {
		return operatorMinimum
	}
	return policy.FirmwareMinimum
}

// ResolveInformSchedulePolicy resolves one effective product identity. It
// never falls back by family, model similarity, version prefix, or a nearby
// protocol build.
func ResolveInformSchedulePolicy(profile Profile, effectiveModel, effectiveVersion string) InformSchedulePolicy {
	protocolProfile, binding, ok := resolveUserProfileProtocol(profile, effectiveModel, effectiveVersion)
	if !ok {
		return InformSchedulePolicy{
			BehaviorID: ProtocolBehaviorUnsupported,
			Reason:     informScheduleNoProfileBinding,
		}
	}
	policy := InformSchedulePolicy{
		EvidenceID:    protocolProfile.EvidenceID,
		ProtocolBuild: protocolProfile.Build,
		BehaviorID:    protocolProfile.BehaviorID,
	}
	if binding.minimumSeconds <= 0 {
		policy.Reason = informScheduleNoMinimumBinding
		return policy
	}
	if !protocolProfile.IsSupported() {
		policy.Reason = informScheduleNoProfileBinding
		return policy
	}
	policy.FirmwareMinimum = time.Duration(binding.minimumSeconds) * time.Second
	policy.Supported = true
	return policy
}

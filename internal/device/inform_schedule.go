package device

import "time"

// InformSchedulePolicy is the product decision for controller interval hints.
// An unsupported identity still uses the operator-selected base schedule.
type InformSchedulePolicy struct {
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
	productPolicy, ok := resolveUserProfileProtocol(profile, effectiveModel, effectiveVersion)
	if !ok || productPolicy.firmwareMinimumSeconds <= 0 {
		return InformSchedulePolicy{Reason: profileProtocolNoBinding}
	}
	return InformSchedulePolicy{
		FirmwareMinimum: time.Duration(productPolicy.firmwareMinimumSeconds) * time.Second,
		Supported:       true,
	}
}

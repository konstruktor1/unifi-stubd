package main

import (
	"fmt"

	"github.com/konstruktor1/unifi-stubd/internal/device"
	"github.com/konstruktor1/unifi-stubd/internal/inform"
)

func informScheduleConfig(flags runtimeFlags, policy device.InformSchedulePolicy) inform.ScheduleConfig {
	return inform.ScheduleConfig{
		BaseInterval:          flags.interval,
		ControllerIntervalMin: flags.controllerIntervalMin,
		ControllerIntervalMax: flags.controllerIntervalMax,
		FirmwareIntervalMin:   policy.FirmwareMinimum,
		ControllerHints:       policy.Supported,
		BackoffMax:            flags.informBackoffMax,
		JitterPercent:         flags.informBackoffJitterPercent,
	}
}

func validateInformSchedule(flags runtimeFlags, policy device.InformSchedulePolicy) error {
	if err := inform.ValidateScheduleConfig(informScheduleConfig(flags, policy)); err != nil {
		return fmt.Errorf("invalid inform schedule: %w", err)
	}
	return nil
}

func controllerIntervalPolicyStatus(flags runtimeFlags, policy device.InformSchedulePolicy) statusControllerIntervalPolicy {
	status := statusControllerIntervalPolicy{
		Supported:  policy.Supported,
		EvidenceID: policy.EvidenceID,
		Reason:     policy.Reason,
	}
	if policy.FirmwareMinimum > 0 {
		status.FirmwareMinimum = policy.FirmwareMinimum.String()
	}
	if effective := policy.EffectiveMinimum(flags.controllerIntervalMin); effective > 0 {
		status.EffectiveMinimum = effective.String()
	}
	return status
}

func controllerReportingPolicyStatus(policy device.ControllerReportingPolicy) statusControllerReportingPolicy {
	return statusControllerReportingPolicy{
		EvidenceID:      policy.EvidenceID,
		BehaviorID:      string(policy.BehaviorID),
		BlockSelection:  policy.BlockSelection,
		ControllerClock: policy.ControllerClock,
		Reason:          policy.Reason,
	}
}

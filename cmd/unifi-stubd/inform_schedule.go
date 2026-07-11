package main

import (
	"fmt"

	"github.com/konstruktor1/unifi-stubd/internal/inform"
)

func informScheduleConfig(flags runtimeFlags) inform.ScheduleConfig {
	return inform.ScheduleConfig{
		BaseInterval:          flags.interval,
		ControllerIntervalMin: flags.controllerIntervalMin,
		ControllerIntervalMax: flags.controllerIntervalMax,
		BackoffMax:            flags.informBackoffMax,
		JitterPercent:         flags.informBackoffJitterPercent,
	}
}

func validateInformSchedule(flags runtimeFlags) error {
	if err := inform.ValidateScheduleConfig(informScheduleConfig(flags)); err != nil {
		return fmt.Errorf("invalid inform schedule: %w", err)
	}
	return nil
}

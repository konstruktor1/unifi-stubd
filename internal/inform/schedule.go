package inform

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	immediateInformDelay            = 100 * time.Millisecond
	maximumControllerInformInterval = 300 * time.Second
)

// ScheduleConfig defines the operator-controlled heartbeat and retry bounds.
type ScheduleConfig struct {
	BaseInterval          time.Duration
	ControllerIntervalMin time.Duration
	ControllerIntervalMax time.Duration
	FirmwareIntervalMin   time.Duration
	ControllerHints       bool
	BackoffMax            time.Duration
	JitterPercent         int
}

// ScheduleDecision describes the next heartbeat delay without exposing
// controller payload data.
type ScheduleDecision struct {
	EffectiveInterval     time.Duration
	NextDelay             time.Duration
	ConsecutiveFailures   int
	Backoff               bool
	ControllerHintApplied bool
	ControllerHintIgnored bool
	ControllerHints       bool
	ImmediateApplied      bool
	ImmediateSuppressed   bool
}

// Schedule applies controller interval hints and bounded heartbeat-failure
// backoff for one daemon instance.
type Schedule struct {
	config           ScheduleConfig
	effective        time.Duration
	failures         int
	immediateLatched bool
	random           *rand.Rand
}

// NewSchedule validates config and starts at the operator-selected base
// interval. The random source is local to this schedule and is not used for
// security-sensitive values.
func NewSchedule(config ScheduleConfig) (*Schedule, error) {
	if err := ValidateScheduleConfig(config); err != nil {
		return nil, err
	}
	return &Schedule{
		config:    config,
		effective: config.BaseInterval,
		random:    rand.New(rand.NewSource(time.Now().UnixNano())), //nolint:gosec // Retry jitter is not security-sensitive.
	}, nil
}

// ValidateScheduleConfig rejects invalid bounds before the daemon starts.
func ValidateScheduleConfig(config ScheduleConfig) error {
	if config.BaseInterval <= 0 {
		return fmt.Errorf("inform interval must be positive, got %s", config.BaseInterval)
	}
	if config.ControllerIntervalMin <= 0 {
		return fmt.Errorf("controller interval minimum must be positive, got %s", config.ControllerIntervalMin)
	}
	if config.ControllerIntervalMin > maximumControllerInformInterval {
		return fmt.Errorf("controller interval minimum %s exceeds firmware maximum %s", config.ControllerIntervalMin, maximumControllerInformInterval)
	}
	if config.ControllerIntervalMax < config.ControllerIntervalMin {
		return fmt.Errorf("controller interval maximum %s is below minimum %s", config.ControllerIntervalMax, config.ControllerIntervalMin)
	}
	if !config.ControllerHints && config.FirmwareIntervalMin != 0 {
		return fmt.Errorf("firmware interval minimum must be zero when controller hints are disabled, got %s", config.FirmwareIntervalMin)
	}
	if config.ControllerHints && config.FirmwareIntervalMin <= 0 {
		return fmt.Errorf("firmware interval minimum must be positive when controller hints are enabled, got %s", config.FirmwareIntervalMin)
	}
	if config.ControllerHints && config.FirmwareIntervalMin > maximumControllerInformInterval {
		return fmt.Errorf("firmware interval minimum %s exceeds firmware maximum %s", config.FirmwareIntervalMin, maximumControllerInformInterval)
	}
	effectiveMinimum := config.FirmwareIntervalMin
	if config.ControllerIntervalMin > effectiveMinimum {
		effectiveMinimum = config.ControllerIntervalMin
	}
	if config.ControllerHints && config.ControllerIntervalMax < effectiveMinimum {
		return fmt.Errorf("controller interval maximum %s is below effective firmware minimum %s", config.ControllerIntervalMax, effectiveMinimum)
	}
	if config.BackoffMax <= 0 {
		return fmt.Errorf("inform backoff maximum must be positive, got %s", config.BackoffMax)
	}
	if config.JitterPercent < 0 || config.JitterPercent > 100 {
		return fmt.Errorf("inform backoff jitter percent must be between 0 and 100, got %d", config.JitterPercent)
	}
	return nil
}

// Current returns the normal delay without changing controller hints or retry
// state. It is used when no inform URL is configured.
func (schedule *Schedule) Current() ScheduleDecision {
	return schedule.decision(schedule.effective, false, false, false)
}

// Reset discards controller interval and transport-failure state while keeping
// the operator-selected schedule configuration.
func (schedule *Schedule) Reset() ScheduleDecision {
	schedule.effective = schedule.config.BaseInterval
	schedule.failures = 0
	schedule.immediateLatched = false
	return schedule.Current()
}

// Success applies a positive controller hint within configured bounds and
// resets all heartbeat-failure backoff.
func (schedule *Schedule) Success(controllerHint time.Duration) ScheduleDecision {
	return schedule.SuccessWithOptions(controllerHint, false)
}

// SuccessWithOptions applies a controller hint and optionally schedules one
// bounded near-immediate inform without changing the retained interval.
func (schedule *Schedule) SuccessWithOptions(controllerHint time.Duration, immediate bool) ScheduleDecision {
	hintApplied := schedule.config.ControllerHints && controllerHint > 0
	hintIgnored := !schedule.config.ControllerHints && controllerHint > 0
	if hintApplied {
		maximum := schedule.config.ControllerIntervalMax
		if maximum > maximumControllerInformInterval {
			maximum = maximumControllerInformInterval
		}
		minimum := schedule.config.FirmwareIntervalMin
		if schedule.config.ControllerIntervalMin > minimum {
			minimum = schedule.config.ControllerIntervalMin
		}
		schedule.effective = clampDuration(
			controllerHint,
			minimum,
			maximum,
		)
	}
	schedule.failures = 0
	delay := schedule.effective
	grantImmediate := immediate && !schedule.immediateLatched
	schedule.immediateLatched = immediate
	if grantImmediate {
		delay = immediateInformDelay
	}
	decision := schedule.decision(delay, false, hintApplied, hintIgnored)
	decision.ImmediateApplied = grantImmediate
	decision.ImmediateSuppressed = immediate && !grantImmediate
	return decision
}

// Failure advances bounded exponential backoff and applies symmetric jitter.
// The first failed heartbeat waits twice the effective normal interval.
func (schedule *Schedule) Failure() ScheduleDecision {
	schedule.failures++
	schedule.immediateLatched = false
	maximum := schedule.config.BackoffMax
	if maximum < schedule.effective {
		maximum = schedule.effective
	}
	delay := schedule.effective
	for count := 0; count < schedule.failures; count++ {
		if delay >= maximum || delay > maximum/2 {
			delay = maximum
			break
		}
		delay *= 2
	}
	delay = schedule.jitter(delay, schedule.effective, maximum)
	return schedule.decision(delay, true, false, false)
}

func (schedule *Schedule) decision(delay time.Duration, backoff, hintApplied, hintIgnored bool) ScheduleDecision {
	return ScheduleDecision{
		EffectiveInterval:     schedule.effective,
		NextDelay:             delay,
		ConsecutiveFailures:   schedule.failures,
		Backoff:               backoff,
		ControllerHintApplied: hintApplied,
		ControllerHintIgnored: hintIgnored,
		ControllerHints:       schedule.config.ControllerHints,
	}
}

func (schedule *Schedule) jitter(delay, minimum, maximum time.Duration) time.Duration {
	if schedule.config.JitterPercent == 0 || delay <= 0 {
		return delay
	}
	window := delay * time.Duration(schedule.config.JitterPercent) / 100
	if window <= 0 {
		return delay
	}
	offset := time.Duration(schedule.random.Int63n(int64(window) + 1))
	if schedule.random.Intn(2) == 0 {
		offset = -offset
	}
	return clampDuration(delay+offset, minimum, maximum)
}

func clampDuration(value, minimum, maximum time.Duration) time.Duration {
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}

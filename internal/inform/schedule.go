package inform

import (
	"fmt"
	"math/rand"
	"time"
)

// ScheduleConfig defines the operator-controlled heartbeat and retry bounds.
type ScheduleConfig struct {
	BaseInterval          time.Duration
	ControllerIntervalMin time.Duration
	ControllerIntervalMax time.Duration
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
}

// Schedule applies controller interval hints and bounded transport-failure
// backoff for one daemon instance.
type Schedule struct {
	config    ScheduleConfig
	effective time.Duration
	failures  int
	random    *rand.Rand
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
	if config.ControllerIntervalMax < config.ControllerIntervalMin {
		return fmt.Errorf("controller interval maximum %s is below minimum %s", config.ControllerIntervalMax, config.ControllerIntervalMin)
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
	return schedule.decision(schedule.effective, false, false)
}

// Success applies a positive controller hint within configured bounds and
// resets all transport-failure backoff.
func (schedule *Schedule) Success(controllerHint time.Duration) ScheduleDecision {
	hintApplied := controllerHint > 0
	if hintApplied {
		schedule.effective = clampDuration(
			controllerHint,
			schedule.config.ControllerIntervalMin,
			schedule.config.ControllerIntervalMax,
		)
	}
	schedule.failures = 0
	return schedule.decision(schedule.effective, false, hintApplied)
}

// Failure advances bounded exponential backoff and applies symmetric jitter.
// The first transport failure waits twice the effective normal interval.
func (schedule *Schedule) Failure() ScheduleDecision {
	schedule.failures++
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
	return schedule.decision(delay, true, false)
}

func (schedule *Schedule) decision(delay time.Duration, backoff, hintApplied bool) ScheduleDecision {
	return ScheduleDecision{
		EffectiveInterval:     schedule.effective,
		NextDelay:             delay,
		ConsecutiveFailures:   schedule.failures,
		Backoff:               backoff,
		ControllerHintApplied: hintApplied,
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

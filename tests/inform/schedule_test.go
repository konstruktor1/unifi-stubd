// Inform schedule tests verify controller hint bounds and retry behavior
// independently from packet framing.
package inform_test

import (
	"strings"
	"testing"
	"time"

	"github.com/konstruktor1/unifi-stubd/internal/inform"
)

func TestScheduleClampsControllerHintsAndKeepsLastHint(t *testing.T) {
	schedule := newTestSchedule(t, inform.ScheduleConfig{
		BaseInterval:          10 * time.Second,
		ControllerIntervalMin: 5 * time.Second,
		ControllerIntervalMax: 60 * time.Second,
		BackoffMax:            5 * time.Minute,
	})

	decision := schedule.Success(time.Second)
	if decision.EffectiveInterval != 5*time.Second || decision.NextDelay != 5*time.Second || !decision.ControllerHintApplied {
		t.Fatalf("low controller hint decision = %+v", decision)
	}
	decision = schedule.Success(2 * time.Minute)
	if decision.EffectiveInterval != time.Minute || decision.NextDelay != time.Minute || !decision.ControllerHintApplied {
		t.Fatalf("high controller hint decision = %+v", decision)
	}
	decision = schedule.Success(0)
	if decision.EffectiveInterval != time.Minute || decision.NextDelay != time.Minute || decision.ControllerHintApplied {
		t.Fatalf("missing controller hint decision = %+v", decision)
	}
}

func TestScheduleBackoffIsExponentialBoundedAndResetBySuccess(t *testing.T) {
	schedule := newTestSchedule(t, inform.ScheduleConfig{
		BaseInterval:          10 * time.Second,
		ControllerIntervalMin: 5 * time.Second,
		ControllerIntervalMax: time.Minute,
		BackoffMax:            45 * time.Second,
	})

	for index, want := range []time.Duration{20 * time.Second, 40 * time.Second, 45 * time.Second, 45 * time.Second} {
		decision := schedule.Failure()
		if decision.NextDelay != want || decision.ConsecutiveFailures != index+1 || !decision.Backoff {
			t.Fatalf("failure %d decision = %+v, want delay %s", index+1, decision, want)
		}
	}
	decision := schedule.Success(15 * time.Second)
	if decision.NextDelay != 15*time.Second || decision.EffectiveInterval != 15*time.Second || decision.ConsecutiveFailures != 0 || decision.Backoff {
		t.Fatalf("success reset decision = %+v", decision)
	}
	decision = schedule.Failure()
	if decision.NextDelay != 30*time.Second || decision.ConsecutiveFailures != 1 {
		t.Fatalf("post-reset failure decision = %+v", decision)
	}
}

func TestScheduleJitterStaysWithinNormalAndMaximumBounds(t *testing.T) {
	schedule := newTestSchedule(t, inform.ScheduleConfig{
		BaseInterval:          10 * time.Second,
		ControllerIntervalMin: 5 * time.Second,
		ControllerIntervalMax: time.Minute,
		BackoffMax:            25 * time.Second,
		JitterPercent:         20,
	})
	first := schedule.Failure()
	if first.NextDelay < 16*time.Second || first.NextDelay > 24*time.Second {
		t.Fatalf("first jittered delay = %s, want 16s..24s", first.NextDelay)
	}
	for range 20 {
		decision := schedule.Failure()
		if decision.NextDelay < 10*time.Second || decision.NextDelay > 25*time.Second {
			t.Fatalf("bounded jitter delay = %s, want 10s..25s", decision.NextDelay)
		}
	}
}

func TestScheduleRejectsInvalidConfiguration(t *testing.T) {
	valid := inform.ScheduleConfig{
		BaseInterval:          10 * time.Second,
		ControllerIntervalMin: 5 * time.Second,
		ControllerIntervalMax: time.Minute,
		BackoffMax:            5 * time.Minute,
		JitterPercent:         20,
	}
	tests := []struct {
		name   string
		mutate func(*inform.ScheduleConfig)
		want   string
	}{
		{name: "base", mutate: func(config *inform.ScheduleConfig) { config.BaseInterval = 0 }, want: "interval must be positive"},
		{name: "minimum", mutate: func(config *inform.ScheduleConfig) { config.ControllerIntervalMin = 0 }, want: "minimum must be positive"},
		{name: "bounds", mutate: func(config *inform.ScheduleConfig) { config.ControllerIntervalMax = time.Second }, want: "below minimum"},
		{name: "backoff", mutate: func(config *inform.ScheduleConfig) { config.BackoffMax = 0 }, want: "backoff maximum must be positive"},
		{name: "jitter", mutate: func(config *inform.ScheduleConfig) { config.JitterPercent = 101 }, want: "between 0 and 100"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := valid
			test.mutate(&config)
			if err := inform.ValidateScheduleConfig(config); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("ValidateScheduleConfig() error = %v, want %q", err, test.want)
			}
		})
	}
}

func newTestSchedule(t *testing.T, config inform.ScheduleConfig) *inform.Schedule {
	t.Helper()
	schedule, err := inform.NewSchedule(config)
	if err != nil {
		t.Fatal(err)
	}
	return schedule
}

package inform

import (
	"math"
	"time"
)

// ControllerClock retains a process-local reporting offset learned from
// server_time_in_utc. It never mutates the host clock or system time.
type ControllerClock struct {
	offsetSeconds int64
	set           bool
}

// Update applies the observed server_ms/1000-local_seconds formula.
func (clock *ControllerClock) Update(serverTimeMillis int64, local time.Time) {
	clock.offsetSeconds = saturatingSubtract(serverTimeMillis/1000, local.Unix())
	clock.set = true
}

// Apply shifts one supplied time by the retained whole-second offset.
func (clock *ControllerClock) Apply(local time.Time) time.Time {
	if clock == nil || !clock.set {
		return local
	}
	return time.Unix(saturatingAdd(local.Unix(), clock.offsetSeconds), int64(local.Nanosecond())).In(local.Location())
}

// OffsetSeconds returns the retained controller-minus-device offset.
func (clock *ControllerClock) OffsetSeconds() int64 {
	if clock == nil || !clock.set {
		return 0
	}
	return clock.offsetSeconds
}

// Set reports whether controller time metadata has established an offset.
func (clock *ControllerClock) Set() bool {
	return clock != nil && clock.set
}

// Reset clears controller-derived time state after local de-adoption.
func (clock *ControllerClock) Reset() {
	if clock == nil {
		return
	}
	clock.offsetSeconds = 0
	clock.set = false
}

func saturatingAdd(left, right int64) int64 {
	switch {
	case right > 0 && left > math.MaxInt64-right:
		return math.MaxInt64
	case right < 0 && left < math.MinInt64-right:
		return math.MinInt64
	default:
		return left + right
	}
}

func saturatingSubtract(left, right int64) int64 {
	switch {
	case right < 0 && left > math.MaxInt64+right:
		return math.MaxInt64
	case right > 0 && left < math.MinInt64+right:
		return math.MinInt64
	default:
		return left - right
	}
}

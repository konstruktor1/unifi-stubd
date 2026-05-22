package observe

import (
	"strings"
	"time"
)

// InterfaceRates contains byte-per-second rates derived from monotonic counters.
type InterfaceRates struct {
	RXBytesRate int64
	TXBytesRate int64
}

// TrafficRateTracker computes interface rates from successive counter samples.
type TrafficRateTracker struct {
	previous map[string]trafficRateSample
}

// trafficRateSample stores the previous monotonic counters for one source
// interface.
type trafficRateSample struct {
	time    time.Time
	rxBytes int64
	txBytes int64
}

// NewTrafficRateTracker returns an empty traffic rate tracker.
func NewTrafficRateTracker() *TrafficRateTracker {
	return &TrafficRateTracker{previous: map[string]trafficRateSample{}}
}

// Rates returns RX/TX byte rates for key and stores the current sample.
func (t *TrafficRateTracker) Rates(now time.Time, key string, stats InterfaceStats, up bool) InterfaceRates {
	if t == nil {
		return InterfaceRates{}
	}
	key = strings.ToLower(strings.TrimSpace(key))
	if key == "" || !hasByteCounters(stats) {
		return InterfaceRates{}
	}
	if t.previous == nil {
		t.previous = map[string]trafficRateSample{}
	}
	current := trafficRateSample{time: now, rxBytes: stats.RXBytes, txBytes: stats.TXBytes}
	previous, ok := t.previous[key]
	t.previous[key] = current
	if !ok || !up {
		return InterfaceRates{}
	}
	elapsed := now.Sub(previous.time).Seconds()
	if elapsed <= 0 || stats.RXBytes < previous.rxBytes || stats.TXBytes < previous.txBytes {
		return InterfaceRates{}
	}
	return InterfaceRates{
		RXBytesRate: int64(float64(stats.RXBytes-previous.rxBytes) / elapsed),
		TXBytesRate: int64(float64(stats.TXBytes-previous.txBytes) / elapsed),
	}
}

// hasByteCounters keeps rate tracking disabled until a source provides real
// monotonic counters.
func hasByteCounters(stats InterfaceStats) bool {
	return stats.RXBytes > 0 || stats.TXBytes > 0
}

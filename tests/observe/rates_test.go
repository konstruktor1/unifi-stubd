package observe_test

import (
	"testing"
	"time"

	"github.com/konstruktor1/unifi-stubd/internal/observe"
)

func TestTrafficRatesRequireTwoMonotonicUpSamples(t *testing.T) {
	const interfaceName = "eno1"

	tracker := observe.NewTrafficRateTracker()
	start := time.Unix(100, 0)
	first := tracker.Rates(start, "ENO1", observe.InterfaceStats{RXBytes: 1000, TXBytes: 2000}, true)
	if first.Set {
		t.Fatalf("first sample produced rates: %+v", first)
	}
	second := tracker.Rates(start.Add(10*time.Second), interfaceName, observe.InterfaceStats{RXBytes: 1600, TXBytes: 2900}, true)
	if !second.Set || second.RXBytesRate != 60 || second.TXBytesRate != 90 {
		t.Fatalf("second sample rates = %+v, want 60/90", second)
	}
	reset := tracker.Rates(start.Add(20*time.Second), interfaceName, observe.InterfaceStats{RXBytes: 10, TXBytes: 20}, true)
	if reset.Set {
		t.Fatalf("counter reset produced rates: %+v", reset)
	}
}

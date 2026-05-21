package observe_test

import (
	"testing"
	"time"

	"github.com/konstruktor1/unifi-stubd/internal/observe"
)

func TestTrafficRateTrackerCalculatesDeltas(t *testing.T) {
	tracker := observe.NewTrafficRateTracker()
	start := time.Unix(100, 0)

	first := tracker.Rates(start, "eth0", observe.InterfaceStats{RXBytes: 1000, TXBytes: 2000}, true)
	if first.RXBytesRate != 0 || first.TXBytesRate != 0 {
		t.Fatalf("first rates = %+v, want zero", first)
	}

	second := tracker.Rates(start.Add(10*time.Second), "eth0", observe.InterfaceStats{RXBytes: 1600, TXBytes: 2600}, true)
	if second.RXBytesRate != 60 || second.TXBytesRate != 60 {
		t.Fatalf("second rates = %+v, want 60/60", second)
	}
}

func TestTrafficRateTrackerHandlesResetAndDownLinks(t *testing.T) {
	tracker := observe.NewTrafficRateTracker()
	start := time.Unix(100, 0)
	_ = tracker.Rates(start, "eth0", observe.InterfaceStats{RXBytes: 1000, TXBytes: 2000}, true)

	reset := tracker.Rates(start.Add(10*time.Second), "eth0", observe.InterfaceStats{RXBytes: 10, TXBytes: 20}, true)
	if reset.RXBytesRate != 0 || reset.TXBytesRate != 0 {
		t.Fatalf("reset rates = %+v, want zero", reset)
	}

	down := tracker.Rates(start.Add(20*time.Second), "eth0", observe.InterfaceStats{RXBytes: 1010, TXBytes: 2020}, false)
	if down.RXBytesRate != 0 || down.TXBytesRate != 0 {
		t.Fatalf("down rates = %+v, want zero", down)
	}
}

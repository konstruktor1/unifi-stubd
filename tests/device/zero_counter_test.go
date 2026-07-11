package device_test

import (
	"encoding/json"
	"testing"

	"github.com/konstruktor1/unifi-stubd/internal/device"
	"github.com/konstruktor1/unifi-stubd/internal/device/payload"
)

// TestCompleteZeroCounterOverrideClearsSyntheticCounters verifies an observed
// interface reset replaces generated non-zero port counters with zero.
func TestCompleteZeroCounterOverrideClearsSyntheticCounters(t *testing.T) {
	t.Parallel()

	ports := []device.Port{{
		Index:     1,
		RXBytes:   1000,
		TXBytes:   900,
		RXPackets: 1,
		TXPackets: 1,
		RXErrors:  2,
		TXErrors:  3,
	}}
	out := device.ApplyPortOverrides(ports, []device.PortOverride{{Port: 1, CountersSet: true}})
	got := out[0]
	if !got.CountersSet {
		t.Fatalf("complete counter marker was not propagated: %+v", got)
	}
	if got.RXBytes != 0 || got.TXBytes != 0 || got.RXPackets != 0 ||
		got.TXPackets != 0 || got.RXErrors != 0 || got.TXErrors != 0 {
		t.Fatalf("zero counter snapshot did not replace generated counters: %+v", got)
	}

	data, err := payload.Build(device.Profile{}, device.Identity{
		MAC:      "02:00:5e:10:53:99",
		IP:       "192.0.2.199",
		Hostname: "counter-reset",
		Model:    "TEST8",
		Version:  "1.0.0",
	}, out)
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		PortTable []struct {
			RXPackets int64 `json:"rx_packets"`
			TXPackets int64 `json:"tx_packets"`
		} `json:"port_table"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	if len(doc.PortTable) != 1 || doc.PortTable[0].RXPackets != 0 || doc.PortTable[0].TXPackets != 0 {
		t.Fatalf("payload did not preserve zero packet counters: %+v", doc.PortTable)
	}
}

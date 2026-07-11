package observe_test

import (
	"context"
	"testing"

	"github.com/konstruktor1/unifi-stubd/internal/device"
	"github.com/konstruktor1/unifi-stubd/internal/observe"
	"github.com/konstruktor1/unifi-stubd/internal/observe/ifsource"
)

type interfaceReaderFunc func(context.Context, string) (observe.PortObservation, []error)

const (
	refreshInterface       = "wan0"
	refreshMedia           = "SFP+"
	refreshObservedNetmask = "255.255.255.192"
	refreshObservedIPv6    = "2001:db8:22::20/64"
)

func (read interfaceReaderFunc) Interface(ctx context.Context, iface string) (observe.PortObservation, []error) {
	return read(ctx, iface)
}

// TestRefreshOverridesUsesNewSnapshot verifies successive refreshes start from
// immutable configuration instead of treating the first observation as input.
func TestRefreshOverridesUsesNewSnapshot(t *testing.T) {
	t.Parallel()

	up := true
	down := false
	snapshots := []observe.PortObservation{
		{
			Interface: refreshInterface,
			MAC:       "02:00:5e:00:53:01",
			IP:        "192.0.2.111",
			Netmask:   "255.255.254.0",
			IPv6:      []string{"2001:db8:1::10/64"},
			Up:        &up,
			SpeedMbps: 1000,
			Media:     "GE",
			Stats: observe.InterfaceStats{
				CountersSet: true,
				RXBytes:     100,
				TXBytes:     200,
				RXPackets:   10,
				TXPackets:   20,
				RXErrors:    1,
				TXErrors:    2,
			},
		},
		{
			Interface: refreshInterface,
			MAC:       "02:00:5e:00:53:02",
			IP:        "198.51.100.20",
			Netmask:   refreshObservedNetmask,
			IPv6:      []string{refreshObservedIPv6},
			Up:        &down,
			SpeedMbps: 10000,
			Media:     refreshMedia,
			Stats:     observe.InterfaceStats{CountersSet: true},
		},
	}
	reads := 0
	reader := interfaceReaderFunc(func(_ context.Context, iface string) (observe.PortObservation, []error) {
		if iface != refreshInterface {
			t.Fatalf("interface = %q, want wan0", iface)
		}
		observation := snapshots[reads]
		reads++
		return observation, nil
	})
	configured := []device.PortOverride{{Port: 3, Interface: " " + refreshInterface + " ", Role: "wan"}}

	first, warnings := ifsource.RefreshOverrides(context.Background(), reader, configured)
	if len(warnings) != 0 {
		t.Fatalf("first warnings = %v", warnings)
	}
	second, warnings := ifsource.RefreshOverrides(context.Background(), reader, configured)
	if len(warnings) != 0 {
		t.Fatalf("second warnings = %v", warnings)
	}
	if reads != 2 {
		t.Fatalf("interface reads = %d, want 2", reads)
	}
	if first[0].MAC != "02:00:5e:00:53:01" || first[0].IP != "192.0.2.111" || first[0].Speed != 1000 {
		t.Fatalf("first snapshot = %+v", first[0])
	}
	if second[0].MAC != "02:00:5e:00:53:02" ||
		second[0].IP != "198.51.100.20" ||
		second[0].Netmask != refreshObservedNetmask ||
		len(second[0].IPv6) != 1 || second[0].IPv6[0] != refreshObservedIPv6 ||
		second[0].Up == nil || *second[0].Up ||
		second[0].Speed != 10000 || second[0].Media != refreshMedia {
		t.Fatalf("second snapshot = %+v", second[0])
	}
	if !second[0].CountersSet || second[0].RXBytes != 0 || second[0].TXBytes != 0 ||
		second[0].RXPackets != 0 || second[0].TXPackets != 0 ||
		second[0].RXErrors != 0 || second[0].TXErrors != 0 {
		t.Fatalf("second counters = %+v", second[0])
	}
	if configured[0].Interface != " "+refreshInterface+" " || configured[0].MAC != "" || configured[0].IP != "" || configured[0].Up != nil {
		t.Fatalf("configured override was mutated: %+v", configured[0])
	}
}

// TestRefreshOverridesPreservesOperatorValues verifies explicit configuration
// remains authoritative over a fresh interface observation.
func TestRefreshOverridesPreservesOperatorValues(t *testing.T) {
	t.Parallel()

	configuredUp := true
	observedUp := false
	configured := []device.PortOverride{{
		Port:      3,
		Interface: refreshInterface,
		MAC:       "02:00:5e:00:53:aa",
		IP:        "203.0.113.10",
		Netmask:   "255.255.252.0",
		IPv6:      []string{"2001:db8:ff::10/64"},
		Up:        &configuredUp,
		Speed:     2500,
		Media:     refreshMedia,
	}}
	reader := interfaceReaderFunc(func(_ context.Context, _ string) (observe.PortObservation, []error) {
		return observe.PortObservation{
			MAC:       "02:00:5e:00:53:bb",
			IP:        "192.0.2.20",
			Netmask:   refreshObservedNetmask,
			IPv6:      []string{refreshObservedIPv6},
			Up:        &observedUp,
			SpeedMbps: 10000,
			Media:     "GE",
			Stats:     observe.InterfaceStats{CountersSet: true, RXBytes: 42},
		}, nil
	})

	refreshed, warnings := ifsource.RefreshOverrides(context.Background(), reader, configured)
	if len(warnings) != 0 {
		t.Fatalf("warnings = %v", warnings)
	}
	got := refreshed[0]
	if got.MAC != configured[0].MAC || got.IP != configured[0].IP || got.Netmask != configured[0].Netmask ||
		len(got.IPv6) != 1 || got.IPv6[0] != configured[0].IPv6[0] ||
		got.Up == nil || !*got.Up || got.Speed != 2500 || got.Media != refreshMedia {
		t.Fatalf("operator values were replaced: %+v", got)
	}
	if !got.CountersSet || got.RXBytes != 42 {
		t.Fatalf("observed counters missing: %+v", got)
	}
}

// TestReadInterfaceStatsMarksZeroCounters verifies a complete all-zero sample
// remains distinguishable from an unavailable counter source.
func TestReadInterfaceStatsMarksZeroCounters(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	for _, path := range []string{
		"statistics/rx_bytes",
		"statistics/tx_bytes",
		"statistics/rx_packets",
		"statistics/tx_packets",
		"statistics/rx_errors",
		"statistics/tx_errors",
		"speed",
	} {
		writeSysfsCounter(t, root, "eth0", path, "0")
	}
	stats, err := observe.ReadInterfaceStats(root, "eth0")
	if err != nil {
		t.Fatal(err)
	}
	if !stats.CountersSet {
		t.Fatalf("zero counter sample was not marked complete: %+v", stats)
	}
}

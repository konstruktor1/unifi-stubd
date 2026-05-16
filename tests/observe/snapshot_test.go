package observe_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/konstruktor1/unifi-stubd/internal/adapters/linuxbridge"
	"github.com/konstruktor1/unifi-stubd/internal/device"
	"github.com/konstruktor1/unifi-stubd/internal/observe"
)

func TestReadInterfaceStatsFromSysfsFixture(t *testing.T) {
	root := t.TempDir()
	writeSysfsCounter(t, root, "eth0", "statistics/rx_bytes", "1234")
	writeSysfsCounter(t, root, "eth0", "statistics/tx_bytes", "5678")
	writeSysfsCounter(t, root, "eth0", "statistics/rx_packets", "12")
	writeSysfsCounter(t, root, "eth0", "statistics/tx_packets", "34")
	writeSysfsCounter(t, root, "eth0", "statistics/rx_errors", "1")
	writeSysfsCounter(t, root, "eth0", "statistics/tx_errors", "2")
	writeSysfsCounter(t, root, "eth0", "speed", "25000")

	stats, err := observe.ReadInterfaceStats(root, "eth0")
	if err != nil {
		t.Fatal(err)
	}
	if stats.RXBytes != 1234 || stats.TXBytes != 5678 {
		t.Fatalf("byte counters = %+v", stats)
	}
	if stats.RXPackets != 12 || stats.TXPackets != 34 {
		t.Fatalf("packet counters = %+v", stats)
	}
	if stats.RXErrors != 1 || stats.TXErrors != 2 {
		t.Fatalf("error counters = %+v", stats)
	}
	if stats.SpeedMbps != 25000 {
		t.Fatalf("SpeedMbps = %d", stats.SpeedMbps)
	}
}

func TestMACEntriesFiltersLocalAndMulticastFDBRows(t *testing.T) {
	entries := []linuxbridge.FDBEntry{
		{MAC: "00:11:22:33:44:55", Device: "tap101i0", VLAN: 20, Dynamic: true},
		{MAC: "02:aa:bb:cc:dd:ee", Device: "eth0", Permanent: true, Self: true},
		{MAC: "33:33:00:00:00:01", Device: "eth0", Dynamic: true},
		{MAC: "00:11:22:33:44:55", Device: "tap101i0", Dynamic: true},
	}
	macs := observe.MACEntries(entries)
	if len(macs) != 1 {
		t.Fatalf("len(macs) = %d, want 1: %+v", len(macs), macs)
	}
	if macs[0].MAC != "00:11:22:33:44:55" {
		t.Fatalf("MAC = %q", macs[0].MAC)
	}
	if macs[0].VLAN != 20 {
		t.Fatalf("VLAN = %d", macs[0].VLAN)
	}
}

func TestApplySnapshotUpdatesUplinkPort(t *testing.T) {
	ports := device.SwitchPortsWithOptions(4, device.PortOptions{
		Speed:       10000,
		UplinkSpeed: 25000,
		Media:       "SFP+",
		UplinkMedia: "SFP28",
		PortGroups: []device.PortGroup{
			{Count: 3, Speed: 10000, Media: "SFP+"},
			{Count: 1, Speed: 25000, Media: "SFP28", Uplink: true},
		},
	})
	out := observe.Apply(ports, observe.Snapshot{
		UplinkPortIndex: 4,
		Stats: observe.InterfaceStats{
			RXBytes:   100,
			TXBytes:   200,
			RXPackets: 10,
			TXPackets: 20,
			RXErrors:  1,
			TXErrors:  2,
			SpeedMbps: 10000,
		},
		MACs: []device.MacTableEntry{{MAC: "00:11:22:33:44:55", Age: 4, Uptime: 1200}},
	})
	port := out[3]
	if port.Speed != 10000 {
		t.Fatalf("Speed = %d", port.Speed)
	}
	if port.RXBytes != 100 || port.TXBytes != 200 {
		t.Fatalf("bytes = %+v", port)
	}
	if len(port.MACs) != 1 || port.MACs[0].MAC != "00:11:22:33:44:55" {
		t.Fatalf("MACs = %+v", port.MACs)
	}
	if ports[3].RXBytes == out[3].RXBytes {
		t.Fatal("Apply mutated input ports")
	}
}

func writeSysfsCounter(t *testing.T, root, iface, name, value string) {
	t.Helper()
	path := filepath.Join(root, "class", "net", iface, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(value+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

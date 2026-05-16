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

func TestMACEntriesByDeviceGroupsBridgeMembers(t *testing.T) {
	entries := []linuxbridge.FDBEntry{
		{MAC: "00:11:22:33:44:55", Device: "tap101i0", VLAN: 20, Dynamic: true},
		{MAC: "00:11:22:33:44:66", Device: "tap101i0", Dynamic: true},
		{MAC: "00:11:22:33:44:77", Device: "veth200i0", Dynamic: true},
		{MAC: "02:aa:bb:cc:dd:ee", Device: "eth0", Permanent: true, Self: true},
	}
	byDevice := observe.MACEntriesByDevice(entries)
	if len(byDevice["tap101i0"]) != 2 {
		t.Fatalf("tap101i0 entries = %+v", byDevice["tap101i0"])
	}
	if byDevice["tap101i0"][0].VLAN != 20 {
		t.Fatalf("tap101i0 VLAN = %d", byDevice["tap101i0"][0].VLAN)
	}
	if len(byDevice["veth200i0"]) != 1 {
		t.Fatalf("veth200i0 entries = %+v", byDevice["veth200i0"])
	}
	if _, ok := byDevice["eth0"]; ok {
		t.Fatalf("local uplink entry was not filtered: %+v", byDevice["eth0"])
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

func TestApplySnapshotDistributesBridgeFDBDevices(t *testing.T) {
	ports := device.SwitchPortsWithOptions(5, device.PortOptions{
		Speed:       10000,
		UplinkSpeed: 25000,
		Media:       "SFP+",
		UplinkMedia: "SFP28",
		PortGroups: []device.PortGroup{
			{Count: 4, Speed: 10000, Media: "SFP+"},
			{Count: 1, Speed: 25000, Media: "SFP28", Uplink: true},
		},
	})
	out := observe.Apply(ports, observe.Snapshot{
		UplinkPortIndex: 5,
		Interface:       "eth0",
		Bridge:          "vmbr0",
		DeviceMACs: map[string][]device.MacTableEntry{
			"veth200i0": {{MAC: "00:11:22:33:44:77", Age: 4, Uptime: 1200}},
			"tap101i0":  {{MAC: "00:11:22:33:44:55", Age: 4, Uptime: 1200, VLAN: 20}},
			"eth0":      {{MAC: "00:11:22:33:44:99", Age: 4, Uptime: 1200}},
		},
	})

	if out[0].Name != "tap101i0" {
		t.Fatalf("port 1 name = %q", out[0].Name)
	}
	if len(out[0].MACs) != 1 || out[0].MACs[0].MAC != "00:11:22:33:44:55" {
		t.Fatalf("port 1 MACs = %+v", out[0].MACs)
	}
	if out[1].Name != "veth200i0" {
		t.Fatalf("port 2 name = %q", out[1].Name)
	}
	if len(out[1].MACs) != 1 || out[1].MACs[0].MAC != "00:11:22:33:44:77" {
		t.Fatalf("port 2 MACs = %+v", out[1].MACs)
	}
	if out[4].Name != "eth0" {
		t.Fatalf("uplink name = %q", out[4].Name)
	}
	if len(out[4].MACs) != 1 || out[4].MACs[0].MAC != "00:11:22:33:44:99" {
		t.Fatalf("uplink MACs = %+v", out[4].MACs)
	}
	if ports[0].Name == out[0].Name {
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

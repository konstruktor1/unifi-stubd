//nolint:goconst // Repeated sysfs and port fixture literals keep observation cases explicit.
package observe_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/konstruktor1/unifi-stubd/internal/adapters/freebsdifconfig"
	"github.com/konstruktor1/unifi-stubd/internal/adapters/linuxbridge"
	"github.com/konstruktor1/unifi-stubd/internal/device"
	"github.com/konstruktor1/unifi-stubd/internal/observe"
)

func groupedSnapshotPorts(count int) []device.Port {
	return device.BuildPorts(device.Profile{
		Ports:       count,
		PortSpeed:   10000,
		UplinkSpeed: 25000,
		PortMedia:   "SFP+",
		UplinkMedia: "SFP28",
		PortGroups: []device.PortGroup{
			{Count: count - 1, Speed: 10000, Media: "SFP+"},
			{Count: 1, Speed: 25000, Media: "SFP28", Uplink: true},
		},
	}, device.PortBuildOptions{})
}

// TestReadInterfaceStatsFromSysfsFixture verifies sysfs counter and speed reads
// from fixture data.
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

// TestMACEntriesFiltersLocalAndMulticastFDBRows verifies bridge observation
// ignores rows that cannot represent downstream clients.
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

// TestMACEntriesByDeviceGroupsBridgeMembers verifies FDB entries are grouped by
// bridge member for port assignment.
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

// TestParseARPTableFiltersAndNormalizesRows verifies ARP enrichment accepts
// only normalized unicast IPv4 rows.
func TestParseARPTableFiltersAndNormalizesRows(t *testing.T) {
	rows := `IP address       HW type     Flags       HW address            Mask     Device
192.0.2.52       0x1         0x2         02:00:5e:00:53:03     *        vmbr0
192.0.2.54       0x1         0x0         00:00:00:00:00:00     *        vmbr0
2001:db8::1      0x1         0x2         02:00:5e:00:53:04     *        vmbr0
192.0.2.53       0x1         0x2         33:33:00:00:00:01     *        vmbr0
`
	entries := observe.ParseARPTable(strings.NewReader(rows))
	if len(entries) != 1 {
		t.Fatalf("ARP entries = %+v", entries)
	}
	if entries[0].IP != "192.0.2.52" ||
		entries[0].MAC != "02:00:5e:00:53:03" ||
		entries[0].Device != "vmbr0" {
		t.Fatalf("entry = %+v", entries[0])
	}
}

// TestEnrichMACsFromARPAddsClientIP verifies ARP data fills missing
// client IP metadata without replacing existing fields.
func TestEnrichMACsFromARPAddsClientIP(t *testing.T) {
	memberMACs := map[string][]device.MacTableEntry{
		"tap101i0": {
			{MAC: "02:00:5e:00:53:03", Age: 4, Uptime: 1200},
		},
	}

	observe.EnrichMACsFromARP(memberMACs, []observe.ARPEntry{
		{IP: "192.0.2.52", MAC: "02:00:5e:00:53:03", Device: "vmbr0"},
	})

	if got := memberMACs["tap101i0"][0].IP; got != "192.0.2.52" {
		t.Fatalf("IP = %q, want 192.0.2.52", got)
	}
}

// TestFreeBSDMACsByInterfaceFiltersLocalAndMulticast verifies FreeBSD
// bridge rows follow the same downstream-client filter.
func TestFreeBSDMACsByInterfaceFiltersLocalAndMulticast(t *testing.T) {
	entries := []freebsdifconfig.BridgeAddress{
		{MAC: "00:11:22:33:44:55", VLAN: 20, Interface: "tap101", Age: 99},
		{MAC: "02:aa:bb:cc:dd:ee", VLAN: 1, Interface: "em0", Local: true},
		{MAC: "33:33:00:00:00:01", VLAN: 1, Interface: "em0"},
	}
	byInterface := observe.FreeBSDMACsByInterface(entries)
	if len(byInterface["tap101"]) != 1 {
		t.Fatalf("tap101 entries = %+v", byInterface["tap101"])
	}
	entry := byInterface["tap101"][0]
	if entry.MAC != "00:11:22:33:44:55" || entry.VLAN != 20 || entry.Age != 99 {
		t.Fatalf("entry = %+v", entry)
	}
	if _, ok := byInterface["em0"]; ok {
		t.Fatalf("filtered entries leaked through: %+v", byInterface["em0"])
	}
}

// TestClassifyMembersDistinguishesUplinkAccessAndBridge verifies bridge
// member roles before port mapping.
func TestClassifyMembersDistinguishesUplinkAccessAndBridge(t *testing.T) {
	memberMACs := map[string][]device.MacTableEntry{
		"vmbr0":    {{MAC: "00:11:22:33:44:00", Age: 4, Uptime: 1200}},
		"enp100s0": {{MAC: "00:11:22:33:44:01", Age: 4, Uptime: 1200}},
		"tap101i0": {{MAC: "00:11:22:33:44:02", Age: 4, Uptime: 1200}},
		"veth200i0": {
			{MAC: "00:11:22:33:44:03", Age: 4, Uptime: 1200},
		},
		"fwpr104p0": {{MAC: "00:11:22:33:44:04", Age: 4, Uptime: 1200}},
	}

	roles := observe.ClassifyMembers(memberMACs, "vmbr0", "enp100s0")
	if roles["vmbr0"] != observe.BridgeMemberRoleBridge {
		t.Fatalf("vmbr0 role = %q", roles["vmbr0"])
	}
	if roles["enp100s0"] != observe.BridgeMemberRoleUplink {
		t.Fatalf("enp100s0 role = %q", roles["enp100s0"])
	}
	for _, member := range []string{"tap101i0", "veth200i0", "fwpr104p0"} {
		if roles[member] != observe.BridgeMemberRoleAccess {
			t.Fatalf("%s role = %q", member, roles[member])
		}
	}
}

// TestClassifyMembersPromotesSinglePhysicalCandidate verifies the
// conservative single-physical-uplink heuristic.
func TestClassifyMembersPromotesSinglePhysicalCandidate(t *testing.T) {
	memberMACs := map[string][]device.MacTableEntry{
		"bridge0": {{MAC: "00:11:22:33:44:00", Age: 4, Uptime: 1200}},
		"igb0":    {{MAC: "00:11:22:33:44:01", Age: 4, Uptime: 1200}},
		"epair0a": {{MAC: "00:11:22:33:44:02", Age: 4, Uptime: 1200}},
	}

	roles := observe.ClassifyMembers(memberMACs, "bridge0", "")
	if roles["bridge0"] != observe.BridgeMemberRoleBridge {
		t.Fatalf("bridge0 role = %q", roles["bridge0"])
	}
	if roles["igb0"] != observe.BridgeMemberRoleUplink {
		t.Fatalf("igb0 role = %q", roles["igb0"])
	}
	if roles["epair0a"] != observe.BridgeMemberRoleAccess {
		t.Fatalf("epair0a role = %q", roles["epair0a"])
	}
}

// TestClassifyMembersHonorsIgnoredMembers verifies ignored members cannot
// consume represented UniFi ports.
func TestClassifyMembersHonorsIgnoredMembers(t *testing.T) {
	memberMACs := map[string][]device.MacTableEntry{
		"vmbr0":      {{MAC: "00:11:22:33:44:00", Age: 4, Uptime: 1200}},
		"eno1":       {{MAC: "00:11:22:33:44:01", Age: 4, Uptime: 1200}},
		"tap10000i0": {{MAC: "00:11:22:33:44:02", Age: 4, Uptime: 1200}},
	}

	roles := observe.ClassifyMembersWithIgnores(memberMACs, "vmbr0", "eno1", []string{"TAP10000I0"})
	if roles["tap10000i0"] != observe.BridgeMemberRoleIgnored {
		t.Fatalf("tap10000i0 role = %q", roles["tap10000i0"])
	}
	if roles["eno1"] != observe.BridgeMemberRoleUplink {
		t.Fatalf("eno1 role = %q", roles["eno1"])
	}
}

// TestRemoteMACsByBridgeMemberDetectsUplinkNeighbor verifies upstream MACs are
// separated from local bridge participants.
func TestRemoteMACsByBridgeMemberDetectsUplinkNeighbor(t *testing.T) {
	memberMACs := map[string][]device.MacTableEntry{
		"enp100s0": {
			{MAC: "02:00:5e:00:53:80", Age: 4, Uptime: 1200},
			{MAC: "02:00:5e:00:53:81", Age: 4, Uptime: 1200},
		},
		"tap101i0": {{MAC: "00:11:22:33:44:55", Age: 4, Uptime: 1200}},
	}
	roles := map[string]observe.BridgeMemberRole{
		"enp100s0": observe.BridgeMemberRoleUplink,
		"tap101i0": observe.BridgeMemberRoleAccess,
	}

	remote := observe.RemoteMACsByBridgeMember(memberMACs, roles, "enp100s0", "vmbr0")
	if !remote["02:00:5e:00:53:80"] || !remote["02:00:5e:00:53:81"] {
		t.Fatalf("remote uplink MACs = %+v", remote)
	}
	if remote["00:11:22:33:44:55"] {
		t.Fatalf("local access MAC was marked remote: %+v", remote)
	}
}

// TestApplySnapshotIgnoresBridgeMembers verifies bridge metadata and ignored
// members are not rendered as access ports.
func TestApplySnapshotIgnoresBridgeMembers(t *testing.T) {
	ports := groupedSnapshotPorts(4)
	out := observe.Apply(ports, observe.Snapshot{
		UplinkPortIndex: 4,
		Interface:       "eno1",
		Bridge:          "vmbr0",
		DeviceMACs: map[string][]device.MacTableEntry{
			"tap10000i0": {{MAC: "00:11:22:33:44:55", Age: 4, Uptime: 1200}},
			"veth200i0":  {{MAC: "00:11:22:33:44:77", Age: 4, Uptime: 1200}},
			"eno1":       {{MAC: "00:11:22:33:44:99", Age: 4, Uptime: 1200}},
		},
		MemberRoles: map[string]observe.BridgeMemberRole{
			"tap10000i0": observe.BridgeMemberRoleIgnored,
			"veth200i0":  observe.BridgeMemberRoleAccess,
			"eno1":       observe.BridgeMemberRoleUplink,
		},
	})

	if out[0].Name != "veth200i0" {
		t.Fatalf("port 1 name = %q", out[0].Name)
	}
	if len(out[0].MACs) != 1 || out[0].MACs[0].MAC != "00:11:22:33:44:77" {
		t.Fatalf("port 1 MACs = %+v", out[0].MACs)
	}
	for _, port := range out {
		for _, entry := range port.MACs {
			if entry.MAC == "00:11:22:33:44:55" {
				t.Fatalf("ignored member MAC leaked into payload: port=%+v", port)
			}
		}
	}
}

// TestApplySnapshotUpdatesUplinkPort verifies uplink observation updates speed,
// counters, and MAC metadata.
func TestApplySnapshotUpdatesUplinkPort(t *testing.T) {
	ports := groupedSnapshotPorts(4)
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

// TestApplySnapshotDistributesBridgeFDBDevices verifies deterministic bridge
// member assignment across generated ports.
func TestApplySnapshotDistributesBridgeFDBDevices(t *testing.T) {
	ports := groupedSnapshotPorts(5)
	out := observe.Apply(ports, observe.Snapshot{
		UplinkPortIndex: 5,
		Interface:       "eth0",
		Bridge:          "vmbr0",
		DeviceMACs: map[string][]device.MacTableEntry{
			"vmbr0":     {{MAC: "00:11:22:33:44:11", Age: 4, Uptime: 1200}},
			"veth200i0": {{MAC: "00:11:22:33:44:77", Age: 4, Uptime: 1200}},
			"tap101i0": {
				{MAC: "00:11:22:33:44:55", Age: 4, Uptime: 1200, VLAN: 20},
				{MAC: "02:00:5e:00:53:80", Age: 4, Uptime: 1200, VLAN: 20},
			},
			"eth0": {
				{MAC: "00:11:22:33:44:99", Age: 4, Uptime: 1200},
				{MAC: "00:11:22:33:44:aa", Age: 4, Uptime: 1200},
				{MAC: "02:00:5e:00:53:80", Age: 4, Uptime: 1200},
			},
		},
		MemberPorts: map[string]observe.PortObservation{
			"tap101i0": {
				Interface: "tap101i0",
				SpeedMbps: 10000,
				Stats:     observe.InterfaceStats{RXBytes: 101, TXBytes: 202},
			},
			"veth200i0": {
				Interface: "veth200i0",
				SpeedMbps: 10000,
			},
			"eth0": {
				Interface: "eth0",
				SpeedMbps: 2500,
			},
		},
		MemberRoles: map[string]observe.BridgeMemberRole{
			"vmbr0":     observe.BridgeMemberRoleBridge,
			"eth0":      observe.BridgeMemberRoleUplink,
			"tap101i0":  observe.BridgeMemberRoleAccess,
			"veth200i0": observe.BridgeMemberRoleAccess,
		},
	})

	if out[0].Name != "tap101i0" {
		t.Fatalf("port 1 name = %q", out[0].Name)
	}
	if len(out[0].MACs) != 1 || out[0].MACs[0].MAC != "00:11:22:33:44:55" {
		t.Fatalf("port 1 MACs = %+v", out[0].MACs)
	}
	if out[0].Speed != 10000 || out[0].Interface != "tap101i0" || out[0].RXBytes != 101 || out[0].TXBytes != 202 {
		t.Fatalf("port 1 observation = %+v", out[0])
	}
	if out[1].Name != "veth200i0" {
		t.Fatalf("port 2 name = %q", out[1].Name)
	}
	if len(out[1].MACs) != 1 || out[1].MACs[0].MAC != "00:11:22:33:44:77" {
		t.Fatalf("port 2 MACs = %+v", out[1].MACs)
	}
	if out[1].Speed != 10000 || out[1].Interface != "veth200i0" {
		t.Fatalf("port 2 observation = %+v", out[1])
	}
	for _, index := range []int{3, 4} {
		if out[index-1].Up || out[index-1].Speed != 0 || len(out[index-1].MACs) != 0 {
			t.Fatalf("unused bridge port %d = %+v, want disconnected", index, out[index-1])
		}
	}
	if out[4].Name != "eth0" {
		t.Fatalf("uplink name = %q", out[4].Name)
	}
	if len(out[4].MACs) != 0 {
		t.Fatalf("uplink remote MACs leaked into local MAC table: %+v", out[4].MACs)
	}
	if out[4].Speed != 2500 || out[4].Interface != "eth0" {
		t.Fatalf("uplink observation = %+v", out[4])
	}
	for _, port := range out {
		for _, entry := range port.MACs {
			if entry.MAC == "00:11:22:33:44:11" {
				t.Fatalf("bridge device MAC leaked into payload: port=%+v", port)
			}
			if entry.MAC == "00:11:22:33:44:99" || entry.MAC == "00:11:22:33:44:aa" {
				t.Fatalf("remote uplink MAC leaked into payload: port=%+v", port)
			}
			if entry.MAC == "02:00:5e:00:53:80" {
				t.Fatalf("neighbor switch MAC leaked into local MAC table: port=%+v", port)
			}
		}
	}
	if ports[0].Name == out[0].Name {
		t.Fatal("Apply mutated input ports")
	}
}

// TestApplySnapshotHonorsBridgeMemberPortMap verifies explicit member pinning
// overrides automatic port assignment.
func TestApplySnapshotHonorsBridgeMemberPortMap(t *testing.T) {
	ports := groupedSnapshotPorts(5)
	out := observe.Apply(ports, observe.Snapshot{
		UplinkPortIndex: 5,
		Interface:       "eno1",
		Bridge:          "vmbr0",
		MemberPortMap: map[string]int{
			"tap101i0": 3,
		},
		DeviceMACs: map[string][]device.MacTableEntry{
			"tap101i0":  {{MAC: "00:11:22:33:44:55", Age: 4, Uptime: 1200}},
			"veth200i0": {{MAC: "00:11:22:33:44:77", Age: 4, Uptime: 1200}},
			"eno1":      {{MAC: "00:11:22:33:44:99", Age: 4, Uptime: 1200}},
		},
	})

	if out[2].Name != "tap101i0" {
		t.Fatalf("pinned port 3 name = %q", out[2].Name)
	}
	if len(out[2].MACs) != 1 || out[2].MACs[0].MAC != "00:11:22:33:44:55" {
		t.Fatalf("pinned port 3 MACs = %+v", out[2].MACs)
	}
	if out[0].Name != "veth200i0" {
		t.Fatalf("auto port 1 name = %q", out[0].Name)
	}
	if out[4].Name != "eno1" {
		t.Fatalf("uplink name = %q", out[4].Name)
	}
}

// writeSysfsCounter writes one fixture sysfs counter for observation tests.
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

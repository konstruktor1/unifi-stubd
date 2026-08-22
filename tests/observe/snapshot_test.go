package observe_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/konstruktor1/unifi-stubd/internal/device"
	"github.com/konstruktor1/unifi-stubd/internal/observe"
	"github.com/konstruktor1/unifi-stubd/internal/observe/ifsource"
)

func TestObservedLinksFailClosed(t *testing.T) {
	root := t.TempDir()
	values := map[string]string{
		"statistics/rx_bytes":   "1200",
		"statistics/tx_bytes":   "3400",
		"statistics/rx_packets": "12",
		"statistics/tx_packets": "34",
		"statistics/rx_errors":  "1",
		"statistics/tx_errors":  "2",
		"speed":                 "2500",
		"carrier":               "0",
	}
	for name, value := range values {
		path := filepath.Join(root, "class", "net", "eno1", name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(value), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	stats, err := observe.ReadInterfaceStats(root, "eno1")
	if err != nil {
		t.Fatal(err)
	}
	if !stats.CountersSet || stats.RXBytes != 1200 || stats.TXBytes != 3400 || stats.SpeedMbps != 2500 {
		t.Fatalf("interface snapshot = %+v", stats)
	}
	if stats.RXPackets != 12 || stats.TXPackets != 34 || stats.RXErrors != 1 || stats.TXErrors != 2 {
		t.Fatalf("packet/error snapshot = %+v", stats)
	}
	if stats.Up == nil || *stats.Up {
		t.Fatalf("carrier state = %v, want down", stats.Up)
	}
	ports := observe.Apply(
		[]device.Port{{Index: 1, Uplink: true, Up: true, Speed: 1000}},
		observe.Snapshot{Up: stats.Up, Stats: stats},
	)
	if ports[0].Up || ports[0].Speed != 0 {
		t.Fatalf("carrier-down uplink = up %t speed %d", ports[0].Up, ports[0].Speed)
	}

	const memberInterface = "eth1"
	override := device.PortOverride{Port: 1, Interface: memberInterface}
	ifsource.ApplyObservation(&override, observe.PortObservation{Interface: memberInterface})
	mapped := device.ApplyPortOverrides(
		[]device.Port{{Index: 1, Up: true, Speed: 1000}},
		[]device.PortOverride{override},
	)
	if mapped[0].Up || mapped[0].Speed != 0 {
		t.Fatalf("unknown interface mapping = up %t speed %d", mapped[0].Up, mapped[0].Speed)
	}

	uplinkUp := true
	bridged := observe.Apply(
		[]device.Port{{Index: 1, Uplink: true}, {Index: 2, Up: true, Speed: 1000}},
		observe.Snapshot{
			UplinkPortIndex: 1,
			Up:              &uplinkUp,
			DeviceMACs:      map[string][]device.MacTableEntry{memberInterface: {{MAC: "02:11:22:33:44:55"}}},
			MemberRoles:     map[string]observe.BridgeMemberRole{memberInterface: observe.BridgeMemberRoleAccess},
			MemberPorts:     map[string]observe.PortObservation{memberInterface: {Interface: memberInterface}},
		},
	)
	if bridged[1].Up || bridged[1].Speed != 0 {
		t.Fatalf("unknown bridge member = up %t speed %d", bridged[1].Up, bridged[1].Speed)
	}
}

package opnsense_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/konstruktor1/unifi-stubd/internal/device"
	"github.com/konstruktor1/unifi-stubd/internal/opnsense"
)

func TestOverridesFromStateMapsInterfacesAndGatewayHealth(t *testing.T) {
	t.Parallel()

	interfaces := loadInterfacesFixture(t)
	gateways := loadGatewayFixture(t)
	overrides := opnsense.OverridesFromState([]opnsense.InterfaceMapping{
		{
			Port:         3,
			Interface:    testInterfaceIXL0,
			Name:         "WAN SFP+",
			Role:         testRoleWAN,
			NetworkGroup: testNetworkWAN,
			VLAN:         3,
		},
		{
			Port:         4,
			Interface:    testInterfaceVTNET0,
			Name:         "LAN",
			Role:         testRoleLAN,
			NetworkGroup: testNetworkLAN,
		},
	}, interfaces, gateways)
	if len(overrides) != 2 {
		t.Fatalf("len(overrides) = %d, want 2", len(overrides))
	}
	wan := overrides[0]
	if wan.Interface != testInterfaceIXL0 || wan.MAC != testWANMAC || wan.IP != testWANIP {
		t.Fatalf("WAN override identity = %+v", wan)
	}
	if wan.Netmask != testWANNetmask || len(wan.IPv6) != 1 || wan.IPv6[0] != "2001:db8:100::9/64" {
		t.Fatalf("WAN addresses = %+v", wan)
	}
	if wan.Speed != 10000 || wan.Media != "SFP+" {
		t.Fatalf("WAN link = speed %d media %q", wan.Speed, wan.Media)
	}
	if wan.WANConnected == nil || !*wan.WANConnected || wan.WANLatencyMS != 7 {
		t.Fatalf("WAN health = %+v", wan)
	}
	lan := overrides[1]
	if lan.Interface != testInterfaceVTNET0 || lan.Role != testRoleLAN || lan.IP != "192.0.2.1" {
		t.Fatalf("LAN override = %+v", lan)
	}
}

func TestMergeOverridesKeepsBaseValues(t *testing.T) {
	t.Parallel()

	up := false
	merged := opnsense.MergeOverrides(
		[]device.PortOverride{
			{Port: 3, Interface: testInterfaceIXL0, IP: testWANIP, Role: testRoleWAN, NetworkGroup: testNetworkWAN, Speed: 10000},
		},
		[]device.PortOverride{
			{Port: 3, IP: testManualWANIP, Up: &up, Name: "manual WAN"},
		},
	)
	if len(merged) != 1 {
		t.Fatalf("len(merged) = %d, want 1", len(merged))
	}
	if merged[0].IP != testManualWANIP || merged[0].Interface != testInterfaceIXL0 || merged[0].Name != "manual WAN" {
		t.Fatalf("merged override = %+v", merged[0])
	}
	if merged[0].Up == nil || *merged[0].Up {
		t.Fatalf("merged Up = %v, want false", merged[0].Up)
	}
}

func TestDecodeInterfaceMessageWrapperUsesRequestedName(t *testing.T) {
	t.Parallel()

	status := opnsense.DecodeInterface(map[string]any{
		"message": map[string]any{
			"mac":    testWANMAC,
			"ipv4":   "203.0.113.10/24",
			"status": "up",
		},
	}, testInterfaceIXL0)
	if status.Interface != testInterfaceIXL0 || status.MAC != testWANMAC || status.Netmask != testWANNetmask {
		t.Fatalf("decoded interface = %+v", status)
	}
}

func loadInterfacesFixture(t *testing.T) map[string]opnsense.InterfaceStatus {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "fixtures", "opnsense", "interfaces_info.json"))
	if err != nil {
		t.Fatal(err)
	}
	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	return opnsense.DecodeInterfaces(raw)
}

func loadGatewayFixture(t *testing.T) map[string]opnsense.GatewayStatus {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "fixtures", "opnsense", "gateway_status.json"))
	if err != nil {
		t.Fatal(err)
	}
	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	return opnsense.DecodeGatewayStatuses(raw)
}

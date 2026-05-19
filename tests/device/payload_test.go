//nolint:goconst // Repeated payload fixture literals document expected UniFi shapes.
package device_test

import (
	"encoding/json"
	"testing"

	"github.com/konstruktor1/unifi-stubd/internal/device"
)

func TestMinimalSwitchPayloadReportsPortCount(t *testing.T) {
	payload, err := device.MinimalSwitchPayload(device.Identity{
		MAC:          "02:11:22:33:44:55",
		IP:           "192.0.2.50",
		Hostname:     "unifi-stubd-lab",
		Model:        "US16P150",
		ModelDisplay: "UniFi Switch 16 POE-150W",
		Version:      "7.4.1.16850",
		Serial:       "021122334455",
		InformURL:    "http://192.0.2.10:8080/inform",
	}, device.SwitchPorts(16))
	if err != nil {
		t.Fatal(err)
	}

	var doc struct {
		Adopted       bool             `json:"adopted"`
		Default       bool             `json:"default"`
		NumPort       int              `json:"num_port"`
		State         int              `json:"state"`
		EthernetTable []map[string]any `json:"ethernet_table"`
		IfTable       []map[string]any `json:"if_table"`
		PortTable     []map[string]any `json:"port_table"`
	}
	if err := json.Unmarshal(payload, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.NumPort != 16 {
		t.Fatalf("top-level num_port = %d, want 16", doc.NumPort)
	}
	if doc.State != 1 || !doc.Default || doc.Adopted {
		t.Fatalf("factory adoption fields = state %d default %t adopted %t", doc.State, doc.Default, doc.Adopted)
	}
	if got := int(doc.EthernetTable[0]["num_port"].(float64)); got != 16 {
		t.Fatalf("ethernet_table num_port = %d, want 16", got)
	}
	if got := int(doc.IfTable[0]["num_port"].(float64)); got != 16 {
		t.Fatalf("if_table num_port = %d, want 16", got)
	}
	if len(doc.PortTable) != 16 {
		t.Fatalf("port_table length = %d, want 16", len(doc.PortTable))
	}
}

func TestMinimalSwitchPayloadReportsAdoptedState(t *testing.T) {
	payload, err := device.MinimalSwitchPayload(device.Identity{
		MAC:          "02:11:22:33:44:56",
		IP:           "192.0.2.50",
		Hostname:     "unifi-stubd-lab",
		Model:        "US8",
		ModelDisplay: "UniFi Switch 8",
		Version:      "7.4.1.16850",
		Serial:       "021122334456",
		InformURL:    "http://192.0.2.10:8080/inform",
		CFGVersion:   "abc123",
		Adopted:      true,
	}, device.SwitchPorts(8))
	if err != nil {
		t.Fatal(err)
	}

	var doc struct {
		Adopted bool `json:"adopted"`
		Default bool `json:"default"`
		State   int  `json:"state"`
	}
	if err := json.Unmarshal(payload, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.State != 2 || doc.Default || !doc.Adopted {
		t.Fatalf("adopted fields = state %d default %t adopted %t", doc.State, doc.Default, doc.Adopted)
	}
}

func TestMinimalSwitchPayloadReportsTenGigUplink(t *testing.T) {
	payload, err := device.MinimalSwitchPayload(device.Identity{
		MAC:          "02:11:22:33:44:57",
		IP:           "192.0.2.50",
		Hostname:     "unifi-stubd-lab",
		Model:        "US16XG",
		ModelDisplay: "UniFi Switch 16 XG",
		Version:      "7.4.1.16850",
		Serial:       "021122334457",
		InformURL:    "http://192.0.2.10:8080/inform",
	}, device.SwitchPortsWithOptions(16, device.PortOptions{
		Speed:       10000,
		UplinkSpeed: 10000,
		Media:       "SFP+",
		UplinkMedia: "SFP+",
	}))
	if err != nil {
		t.Fatal(err)
	}

	var doc struct {
		IfTable   []map[string]any `json:"if_table"`
		PortTable []map[string]any `json:"port_table"`
	}
	if err := json.Unmarshal(payload, &doc); err != nil {
		t.Fatal(err)
	}
	if got := int(doc.IfTable[0]["speed"].(float64)); got != 10000 {
		t.Fatalf("if_table speed = %d, want 10000", got)
	}
	if got := int(doc.PortTable[0]["speed"].(float64)); got != 10000 {
		t.Fatalf("uplink port speed = %d, want 10000", got)
	}
	if got := doc.PortTable[0]["media"].(string); got != "SFP+" {
		t.Fatalf("uplink media = %q, want SFP+", got)
	}
}

func TestMinimalSwitchPayloadReportsManagementVLAN(t *testing.T) {
	payload, err := device.MinimalSwitchPayload(device.Identity{
		MAC:            "02:11:22:33:44:65",
		IP:             "192.0.2.50",
		Hostname:       "unifi-stubd-lab",
		Model:          "US8",
		ModelDisplay:   "UniFi Switch 8",
		Version:        "7.4.1.16850",
		Serial:         "021122334465",
		InformURL:      "http://192.0.2.10:8080/inform",
		ManagementVLAN: 42,
	}, device.SwitchPorts(8))
	if err != nil {
		t.Fatal(err)
	}

	var doc struct {
		ManagementVLAN int              `json:"management_vlan"`
		IfTable        []map[string]any `json:"if_table"`
	}
	if err := json.Unmarshal(payload, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.ManagementVLAN != 42 {
		t.Fatalf("management_vlan = %d, want 42", doc.ManagementVLAN)
	}
	if got := int(doc.IfTable[0]["management_vlan"].(float64)); got != 42 {
		t.Fatalf("if_table management_vlan = %d, want 42", got)
	}
	if got := int(doc.IfTable[0]["vlan"].(float64)); got != 42 {
		t.Fatalf("if_table vlan = %d, want 42", got)
	}
}

func TestSwitchPortsWithProfilePortGroups(t *testing.T) {
	profile, ok := device.LookupProfile("usw-pro-xg-48")
	if !ok {
		t.Fatal("profile not found")
	}
	ports := device.SwitchPortsWithOptions(profile.Ports, profile.PortOptions())

	if len(ports) != 52 {
		t.Fatalf("len(ports) = %d, want 52", len(ports))
	}
	assertPort := func(index, speed int, media string, uplink bool) {
		t.Helper()
		port := ports[index-1]
		if port.Speed != speed {
			t.Fatalf("port %d speed = %d, want %d", index, port.Speed, speed)
		}
		if port.Media != media {
			t.Fatalf("port %d media = %q, want %q", index, port.Media, media)
		}
		if port.Uplink != uplink {
			t.Fatalf("port %d uplink = %v, want %v", index, port.Uplink, uplink)
		}
	}
	assertPort(1, 2500, "GE", false)
	assertPort(16, 2500, "GE", false)
	assertPort(17, 10000, "GE", false)
	assertPort(48, 10000, "GE", false)
	assertPort(49, 25000, "SFP28", true)
	assertPort(52, 25000, "SFP28", false)
}

func TestSwitchPortsWithAggregationProPortGroups(t *testing.T) {
	profile, ok := device.LookupProfile("usaggpro")
	if !ok {
		t.Fatal("profile not found")
	}
	ports := device.SwitchPortsWithOptions(profile.Ports, profile.PortOptions())

	if len(ports) != 32 {
		t.Fatalf("len(ports) = %d, want 32", len(ports))
	}
	assertPort := func(index, speed int, media string, uplink bool) {
		t.Helper()
		port := ports[index-1]
		if port.Speed != speed {
			t.Fatalf("port %d speed = %d, want %d", index, port.Speed, speed)
		}
		if port.Media != media {
			t.Fatalf("port %d media = %q, want %q", index, port.Media, media)
		}
		if port.Uplink != uplink {
			t.Fatalf("port %d uplink = %v, want %v", index, port.Uplink, uplink)
		}
	}
	assertPort(1, 10000, "SFP+", false)
	assertPort(28, 10000, "SFP+", false)
	assertPort(29, 25000, "SFP28", true)
	assertPort(32, 25000, "SFP28", false)
}

func TestGatewayProfileReportsDeviceTypeAndPortNames(t *testing.T) {
	profile, ok := device.LookupProfile("ugw3")
	if !ok {
		t.Fatal("profile not found")
	}
	ports := device.SwitchPortsWithOptions(profile.Ports, profile.PortOptions())
	payload, err := device.MinimalSwitchPayload(device.Identity{
		MAC:          "02:11:22:33:44:61",
		IP:           "192.0.2.50",
		Hostname:     "unifi-stubd-router",
		Model:        profile.Model,
		ModelDisplay: profile.ModelDisplay,
		DeviceType:   profile.DeviceType,
		Version:      profile.Version,
		Serial:       "021122334461",
		InformURL:    "http://192.0.2.10:8080/inform",
	}, ports)
	if err != nil {
		t.Fatal(err)
	}

	var doc struct {
		DeviceType string           `json:"type"`
		NumPort    int              `json:"num_port"`
		IfTable    []map[string]any `json:"if_table"`
	}
	if err := json.Unmarshal(payload, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.DeviceType != "ugw" {
		t.Fatalf("type = %q, want ugw", doc.DeviceType)
	}
	if doc.NumPort != 3 {
		t.Fatalf("num_port = %d, want 3", doc.NumPort)
	}
	names := []string{"WAN 1", "LAN 1", "WAN 2 / LAN 2"}
	for index, name := range names {
		if got := doc.IfTable[index]["comment"].(string); got != name {
			t.Fatalf("port %d name = %q, want %q", index+1, got, name)
		}
	}
}

func TestTenGigGatewayProfileReportsPortLayout(t *testing.T) {
	profile, ok := device.LookupProfile("uxgpro")
	if !ok {
		t.Fatal("profile not found")
	}
	ports := device.SwitchPortsWithOptions(profile.Ports, profile.PortOptions())
	payload, err := device.MinimalSwitchPayload(device.Identity{
		MAC:          "02:11:22:33:44:62",
		IP:           "192.0.2.50",
		Hostname:     "unifi-stubd-uxg",
		Model:        profile.Model,
		ModelDisplay: profile.ModelDisplay,
		DeviceType:   profile.DeviceType,
		Version:      profile.Version,
		Serial:       "021122334462",
		InformURL:    "http://192.0.2.10:8080/inform",
		InformIP:     "192.0.2.10",
	}, ports)
	if err != nil {
		t.Fatal(err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(payload, &raw); err != nil {
		t.Fatal(err)
	}
	var doc struct {
		DeviceType   string           `json:"type"`
		InformIP     string           `json:"inform_ip"`
		NumPort      int              `json:"num_port"`
		Uplink       string           `json:"uplink"`
		IfTable      []map[string]any `json:"if_table"`
		NetworkTable []map[string]any `json:"network_table"`
		UplinkTable  []map[string]any `json:"uplink_table"`
	}
	if err := json.Unmarshal(payload, &doc); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"config_port_table", "ethernet_overrides", "ethernet_table", "port_overrides", "port_table", "reported_networks"} {
		if _, ok := raw[key]; ok {
			t.Fatalf("gateway payload contains switch table %q", key)
		}
	}
	if doc.DeviceType != "uxg" {
		t.Fatalf("type = %q, want uxg", doc.DeviceType)
	}
	if doc.InformIP != "192.0.2.10" {
		t.Fatalf("inform_ip = %q, want 192.0.2.10", doc.InformIP)
	}
	if doc.NumPort != 4 {
		t.Fatalf("num_port = %d, want 4", doc.NumPort)
	}
	if len(doc.IfTable) != 4 {
		t.Fatalf("if_table length = %d, want 4", len(doc.IfTable))
	}
	if got := int(doc.IfTable[0]["speed"].(float64)); got != 1000 {
		t.Fatalf("if_table eth0 speed = %d, want 1000", got)
	}
	if got := int(doc.IfTable[2]["speed"].(float64)); got != 10000 {
		t.Fatalf("if_table eth2 speed = %d, want 10000", got)
	}
	if len(doc.UplinkTable) != 1 {
		t.Fatalf("uplink_table length = %d, want 1", len(doc.UplinkTable))
	}
	if doc.Uplink != "eth0" {
		t.Fatalf("uplink = %q, want eth0", doc.Uplink)
	}
	if len(doc.NetworkTable) != 4 {
		t.Fatalf("network_table length = %d, want 4", len(doc.NetworkTable))
	}
	if got := doc.NetworkTable[2]["networkgroup"].(string); got != "WAN2" {
		t.Fatalf("network_table port 3 networkgroup = %q, want WAN2", got)
	}
}

func TestGatewayPayloadReportsManagementVLANOnUplink(t *testing.T) {
	profile, ok := device.LookupProfile("uxgpro")
	if !ok {
		t.Fatal("profile not found")
	}
	ports := device.SwitchPortsWithOptions(profile.Ports, profile.PortOptions())
	payload, err := device.MinimalSwitchPayload(device.Identity{
		MAC:            "02:11:22:33:44:66",
		IP:             "192.0.2.50",
		Hostname:       "unifi-stubd-uxg",
		Model:          profile.Model,
		ModelDisplay:   profile.ModelDisplay,
		DeviceType:     profile.DeviceType,
		Version:        profile.Version,
		Serial:         "021122334466",
		InformURL:      "http://192.0.2.10:8080/inform",
		ManagementVLAN: 99,
	}, ports)
	if err != nil {
		t.Fatal(err)
	}

	var doc struct {
		ManagementVLAN int              `json:"management_vlan"`
		IfTable        []map[string]any `json:"if_table"`
		UplinkTable    []map[string]any `json:"uplink_table"`
	}
	if err := json.Unmarshal(payload, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.ManagementVLAN != 99 {
		t.Fatalf("management_vlan = %d, want 99", doc.ManagementVLAN)
	}
	if got := int(doc.IfTable[0]["management_vlan"].(float64)); got != 99 {
		t.Fatalf("gateway uplink if_table management_vlan = %d, want 99", got)
	}
	if _, ok := doc.IfTable[1]["management_vlan"]; ok {
		t.Fatalf("non-uplink if_table has management_vlan: %+v", doc.IfTable[1])
	}
	if got := int(doc.UplinkTable[0]["management_vlan"].(float64)); got != 99 {
		t.Fatalf("uplink_table management_vlan = %d, want 99", got)
	}
}

func TestCloudGatewayFiberProfileReportsGatewayPayload(t *testing.T) {
	profile, ok := device.LookupProfile("ucg-fiber")
	if !ok {
		t.Fatal("profile not found")
	}
	ports := device.SwitchPortsWithOptions(profile.Ports, profile.PortOptions())
	payload, err := device.MinimalSwitchPayload(device.Identity{
		MAC:          "02:11:22:33:44:64",
		IP:           "192.0.2.50",
		Hostname:     "unifi-stubd-ucg-fiber",
		Model:        profile.Model,
		ModelDisplay: profile.ModelDisplay,
		DeviceType:   profile.DeviceType,
		Version:      profile.Version,
		Serial:       "021122334464",
		InformURL:    "http://192.0.2.10:8080/inform",
		InformIP:     "192.0.2.10",
	}, ports)
	if err != nil {
		t.Fatal(err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(payload, &raw); err != nil {
		t.Fatal(err)
	}
	var doc struct {
		DeviceType   string           `json:"type"`
		NumPort      int              `json:"num_port"`
		Uplink       string           `json:"uplink"`
		IfTable      []map[string]any `json:"if_table"`
		NetworkTable []map[string]any `json:"network_table"`
		UplinkTable  []map[string]any `json:"uplink_table"`
	}
	if err := json.Unmarshal(payload, &doc); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"config_port_table", "ethernet_overrides", "port_table", "reported_networks"} {
		if _, ok := raw[key]; ok {
			t.Fatalf("gateway payload contains switch table %q", key)
		}
	}
	if doc.DeviceType != "udm" {
		t.Fatalf("type = %q, want udm", doc.DeviceType)
	}
	if doc.NumPort != 7 {
		t.Fatalf("num_port = %d, want 7", doc.NumPort)
	}
	if len(doc.IfTable) != 7 {
		t.Fatalf("if_table length = %d, want 7", len(doc.IfTable))
	}
	if len(doc.UplinkTable) != 1 {
		t.Fatalf("uplink_table length = %d, want 1", len(doc.UplinkTable))
	}
	if doc.Uplink != "eth5" {
		t.Fatalf("uplink = %q, want eth5", doc.Uplink)
	}
	if len(doc.NetworkTable) != 7 {
		t.Fatalf("network_table length = %d, want 7", len(doc.NetworkTable))
	}
	if got := doc.NetworkTable[4]["networkgroup"].(string); got != "WAN2" {
		t.Fatalf("network_table port 5 networkgroup = %q, want WAN2", got)
	}
	if got := doc.NetworkTable[5]["networkgroup"].(string); got != "WAN" {
		t.Fatalf("network_table port 6 networkgroup = %q, want WAN", got)
	}
	if got := doc.NetworkTable[6]["networkgroup"].(string); got != "LAN" {
		t.Fatalf("network_table port 7 networkgroup = %q, want LAN", got)
	}
}

func TestGatewayPortAssignmentsCanBeOverriddenFromConfigModel(t *testing.T) {
	profile, ok := device.LookupProfile("uxgpro")
	if !ok {
		t.Fatal("profile not found")
	}
	ports := device.ApplyPortOverrides(device.SwitchPortsWithOptions(profile.Ports, profile.PortOptions()), []device.PortOverride{
		{Port: 1, Name: "WAN uplink", Role: "wan", NetworkGroup: "WAN", Interface: "ixl0"},
		{Port: 2, Name: "LAN bridge", Role: "lan", NetworkGroup: "LAN", Interface: "vtnet0"},
		{Port: 3, Name: "backup_wan", Role: "wan2", NetworkGroup: "WAN2", Interface: "vlan09"},
		{Port: 4, Name: "unused_lab_lan", Role: "lan2", NetworkGroup: "LAN", Interface: "vlan10"},
	})
	payload, err := device.MinimalSwitchPayload(device.Identity{
		MAC:          "02:11:22:33:44:63",
		IP:           "192.0.2.50",
		Hostname:     "unifi-stubd-uxg",
		Model:        profile.Model,
		ModelDisplay: profile.ModelDisplay,
		DeviceType:   profile.DeviceType,
		Version:      profile.Version,
		Serial:       "021122334463",
		InformURL:    "http://192.0.2.10:8080/inform",
	}, ports)
	if err != nil {
		t.Fatal(err)
	}

	var doc struct {
		NetworkTable []map[string]any `json:"network_table"`
	}
	if err := json.Unmarshal(payload, &doc); err != nil {
		t.Fatal(err)
	}
	if got := doc.NetworkTable[2]["source_interface"].(string); got != "vlan09" {
		t.Fatalf("network_table port 3 source_interface = %q, want vlan09", got)
	}
	if got := doc.NetworkTable[3]["networkgroup"].(string); got != "LAN" {
		t.Fatalf("network_table port 4 networkgroup = %q, want LAN", got)
	}
}

func TestSwitchPortsCanOverrideAggregationUplinkToTenGigPort(t *testing.T) {
	profile, ok := device.LookupProfile("usaggpro")
	if !ok {
		t.Fatal("profile not found")
	}
	options := profile.PortOptions()
	options.UplinkPort = 1
	ports := device.SwitchPortsWithOptions(profile.Ports, options)

	if !ports[0].Uplink {
		t.Fatal("port 1 is not uplink")
	}
	if ports[0].Speed != 10000 {
		t.Fatalf("port 1 speed = %d, want 10000", ports[0].Speed)
	}
	if ports[0].Media != "SFP+" {
		t.Fatalf("port 1 media = %q, want SFP+", ports[0].Media)
	}
	if len(ports[0].MACs) == 0 {
		t.Fatal("port 1 did not receive uplink MAC table")
	}
	if ports[28].Uplink {
		t.Fatal("port 29 is still uplink")
	}
	if ports[28].Speed != 25000 || ports[28].Media != "SFP28" {
		t.Fatalf("port 29 = speed %d media %q, want 25000 SFP28", ports[28].Speed, ports[28].Media)
	}
}

func TestApplyPortOverridesChangesSpeedAndLinkState(t *testing.T) {
	profile, ok := device.LookupProfile("usaggpro")
	if !ok {
		t.Fatal("profile not found")
	}
	linkDown := false
	ports := device.ApplyPortOverrides(device.SwitchPortsWithOptions(profile.Ports, profile.PortOptions()), []device.PortOverride{
		{Port: 2, Speed: 1000},
		{Port: 3, Speed: 2500},
		{Port: 4, Speed: 100},
		{Port: 5, Up: &linkDown},
	})

	assertPort := func(index, speed int, media string, up bool) {
		t.Helper()
		port := ports[index-1]
		if port.Speed != speed {
			t.Fatalf("port %d speed = %d, want %d", index, port.Speed, speed)
		}
		if port.Media != media {
			t.Fatalf("port %d media = %q, want %q", index, port.Media, media)
		}
		if port.Up != up {
			t.Fatalf("port %d up = %v, want %v", index, port.Up, up)
		}
	}
	assertPort(2, 1000, "GE", true)
	assertPort(3, 2500, "GE", true)
	assertPort(4, 100, "GE", true)
	assertPort(5, 0, "SFP+", false)
}

func TestGatewayPayloadReportsPortOverrideMACs(t *testing.T) {
	profile, ok := device.LookupProfile("ugw3")
	if !ok {
		t.Fatal("profile not found")
	}
	ports := device.ApplyPortOverrides(device.SwitchPortsWithOptions(2, profile.PortOptions()), []device.PortOverride{
		{Port: 1, Name: "WAN", MAC: "02:00:5e:00:53:01", IP: "192.0.2.2", Netmask: "255.255.255.0"},
		{Port: 2, Name: "LAN", MAC: "02:00:5e:00:53:02", IP: "192.0.2.1", Netmask: "255.255.255.0"},
	})
	payload, err := device.MinimalSwitchPayload(device.Identity{
		MAC:          "02:00:5e:00:53:01",
		IP:           "192.0.2.1",
		Hostname:     "opnsense",
		Model:        profile.Model,
		ModelDisplay: profile.ModelDisplay,
		DeviceType:   profile.DeviceType,
		Version:      profile.Version,
		Serial:       "02005E005301",
		InformURL:    "http://192.0.2.10:8080/inform",
	}, ports)
	if err != nil {
		t.Fatal(err)
	}

	var doc struct {
		IfTable      []map[string]any `json:"if_table"`
		NetworkTable []map[string]any `json:"network_table"`
	}
	if err := json.Unmarshal(payload, &doc); err != nil {
		t.Fatal(err)
	}
	if got := doc.IfTable[0]["mac"].(string); got != "02:00:5e:00:53:01" {
		t.Fatalf("WAN if_table mac = %q", got)
	}
	if got := doc.IfTable[1]["mac"].(string); got != "02:00:5e:00:53:02" {
		t.Fatalf("LAN if_table mac = %q", got)
	}
	if got := doc.NetworkTable[1]["mac"].(string); got != "02:00:5e:00:53:02" {
		t.Fatalf("LAN network_table mac = %q", got)
	}
	if got := doc.NetworkTable[1]["address"].(string); got != "192.0.2.1/24" {
		t.Fatalf("LAN network_table address = %q", got)
	}
}

func TestUXGGatewayPayloadUsesInterfaceOverrideData(t *testing.T) {
	profile, ok := device.LookupProfile("uxg-lite")
	if !ok {
		t.Fatal("profile not found")
	}
	ports := device.ApplyPortOverrides(device.SwitchPortsWithOptions(profile.Ports, profile.PortOptions()), []device.PortOverride{
		{
			Port:      1,
			Name:      "LAN",
			Interface: "vtnet0",
			MAC:       "02:00:5e:00:53:02",
			IP:        "192.0.2.1",
			Netmask:   "255.255.255.0",
			Speed:     10000,
			Media:     "GE",
			RXBytes:   1234,
			TXBytes:   5678,
		},
		{
			Port:      2,
			Name:      "WAN",
			Interface: "ixl0",
			MAC:       "02:00:5e:00:53:01",
			IP:        "198.51.100.9",
			Netmask:   "255.255.255.0",
			Speed:     10000,
			Media:     "SFP+",
		},
	})
	payload, err := device.MinimalSwitchPayload(device.Identity{
		MAC:          "02:00:5e:00:53:01",
		IP:           "192.0.2.1",
		Hostname:     "opnsense",
		Model:        profile.Model,
		ModelDisplay: profile.ModelDisplay,
		DeviceType:   profile.DeviceType,
		Version:      profile.Version,
		Serial:       "02005E005301",
		InformURL:    "http://192.0.2.10:8080/inform",
	}, ports)
	if err != nil {
		t.Fatal(err)
	}

	var doc struct {
		IfTable      []map[string]any `json:"if_table"`
		NetworkTable []map[string]any `json:"network_table"`
		UplinkTable  []map[string]any `json:"uplink_table"`
	}
	if err := json.Unmarshal(payload, &doc); err != nil {
		t.Fatal(err)
	}
	if got := doc.IfTable[0]["ip"].(string); got != "192.0.2.1" {
		t.Fatalf("LAN if_table ip = %q", got)
	}
	if got := doc.IfTable[1]["ip"].(string); got != "198.51.100.9" {
		t.Fatalf("WAN if_table ip = %q", got)
	}
	if got := doc.NetworkTable[0]["networkgroup"].(string); got != "LAN" {
		t.Fatalf("LAN networkgroup = %q", got)
	}
	if got := doc.NetworkTable[1]["networkgroup"].(string); got != "WAN" {
		t.Fatalf("WAN networkgroup = %q", got)
	}
	if got := doc.NetworkTable[0]["max_speed"].(string); got != "10000" {
		t.Fatalf("LAN max_speed = %q", got)
	}
	if got := doc.NetworkTable[0]["source_interface"].(string); got != "vtnet0" {
		t.Fatalf("LAN source_interface = %q", got)
	}
	if got := int(doc.UplinkTable[0]["max_speed"].(float64)); got != 10000 {
		t.Fatalf("uplink max_speed = %d", got)
	}
}

func TestCustomGatewayPayloadUsesProfileRolesWithoutModelSpecialCase(t *testing.T) {
	profile := device.Profile{
		Name:         "custom-gateway",
		Model:        "CUSTOMGW",
		ModelDisplay: "Custom Gateway",
		DeviceType:   "uxg",
		Version:      "5.0.16.30689",
		Ports:        2,
		PortNames:    []string{"LAN", "WAN"},
		PortRoles:    []string{"lan", "wan"},
		PortNetworkGroups: []string{
			"LAN",
			"WAN",
		},
		PortSpeed:   1000,
		UplinkSpeed: 1000,
		PortMedia:   "GE",
		UplinkMedia: "GE",
		Payload: device.PayloadProfile{
			Kind:                   "gateway",
			RequiredVersion:        "5.0.0",
			ManagementInterface:    "eth0",
			GatewayInterfacePrefix: "eth",
		},
	}
	ports := device.SwitchPortsWithOptions(profile.Ports, profile.PortOptions())
	payload, err := device.BuildPayload(profile, device.Identity{
		MAC:          "02:00:5e:00:53:70",
		IP:           "192.0.2.70",
		Hostname:     "custom-gateway",
		Model:        profile.Model,
		ModelDisplay: profile.ModelDisplay,
		DeviceType:   profile.DeviceType,
		Version:      profile.Version,
		Serial:       "02005E005370",
	}, ports)
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		IfTable      []map[string]any `json:"if_table"`
		NetworkTable []map[string]any `json:"network_table"`
		Uplink       string           `json:"uplink"`
	}
	if err := json.Unmarshal(payload, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.Uplink != "eth0" {
		t.Fatalf("uplink = %q, want eth0", doc.Uplink)
	}
	if got := doc.IfTable[0]["networkgroup"]; got != "LAN" {
		t.Fatalf("port 1 networkgroup = %v, want LAN", got)
	}
	if got := doc.IfTable[1]["networkgroup"]; got != "WAN" {
		t.Fatalf("port 2 networkgroup = %v, want WAN", got)
	}
	if got := doc.NetworkTable[0]["ip"]; got != "192.0.2.70" {
		t.Fatalf("LAN ip = %v, want management IP", got)
	}
	if got := doc.NetworkTable[1]["ip"]; got != "192.0.2.2" {
		t.Fatalf("WAN ip = %v, want documentation WAN IP", got)
	}
}

func TestApplyUplinkNeighborAddsConfiguredNeighbor(t *testing.T) {
	profile, ok := device.LookupProfile("usaggpro")
	if !ok {
		t.Fatal("profile not found")
	}
	options := profile.PortOptions()
	options.UplinkPort = 1
	ports := device.ApplyUplinkNeighbor(device.SwitchPortsWithOptions(profile.Ports, options), &device.MacTableEntry{
		MAC:  "02:aa:bb:cc:dd:01",
		VLAN: 1,
		Type: "usw",
	})

	if len(ports[0].MACs) == 0 {
		t.Fatal("uplink neighbor was not added")
	}
	entry := ports[0].MACs[0]
	if entry.MAC != "02:aa:bb:cc:dd:01" || entry.VLAN != 1 || entry.Type != "usw" {
		t.Fatalf("uplink neighbor = %+v", entry)
	}
	if entry.Age == 0 || entry.Uptime == 0 {
		t.Fatalf("uplink neighbor missing defaults: %+v", entry)
	}
}

func TestApplyPortNeighborsAddsConfiguredMacTableEntry(t *testing.T) {
	profile, ok := device.LookupProfile("ugw3")
	if !ok {
		t.Fatal("profile not found")
	}
	ports := device.ApplyPortNeighbors(device.SwitchPortsWithOptions(profile.Ports, profile.PortOptions()), []device.PortNeighbor{
		{
			Port: 2,
			Entry: device.MacTableEntry{
				MAC:  "02:00:5e:00:53:03",
				VLAN: 1,
				Type: "usw",
			},
		},
	})

	if len(ports[1].MACs) != 1 {
		t.Fatalf("port 2 MAC table length = %d, want 1", len(ports[1].MACs))
	}
	entry := ports[1].MACs[0]
	if entry.MAC != "02:00:5e:00:53:03" || entry.VLAN != 1 || entry.Type != "usw" {
		t.Fatalf("port 2 neighbor = %+v", entry)
	}
	if entry.Age == 0 || entry.Uptime == 0 {
		t.Fatalf("port 2 neighbor missing defaults: %+v", entry)
	}
	if len(ports[0].MACs) == 0 {
		t.Fatal("port 1 lost its generated uplink MAC table")
	}
}

func TestMinimalSwitchPayloadReportsPortOverrideLinkDown(t *testing.T) {
	linkDown := false
	payload, err := device.MinimalSwitchPayload(device.Identity{
		MAC:          "02:11:22:33:44:60",
		IP:           "192.0.2.50",
		Hostname:     "unifi-stubd-lab",
		Model:        "US8",
		ModelDisplay: "UniFi Switch 8",
		Version:      "7.4.1.16850",
		Serial:       "021122334460",
		InformURL:    "http://192.0.2.10:8080/inform",
	}, device.ApplyPortOverrides(device.SwitchPorts(8), []device.PortOverride{
		{Port: 5, Up: &linkDown},
	}))
	if err != nil {
		t.Fatal(err)
	}

	var doc struct {
		PortTable []map[string]any `json:"port_table"`
	}
	if err := json.Unmarshal(payload, &doc); err != nil {
		t.Fatal(err)
	}
	port := doc.PortTable[4]
	if up := port["up"].(bool); up {
		t.Fatal("port 5 is up, want down")
	}
	if got := int(port["speed"].(float64)); got != 0 {
		t.Fatalf("port 5 speed = %d, want 0", got)
	}
}

func TestMinimalSwitchPayloadReportsGroupedUplinkSpeed(t *testing.T) {
	profile, ok := device.LookupProfile("usw-pro-xg-48")
	if !ok {
		t.Fatal("profile not found")
	}
	payload, err := device.MinimalSwitchPayload(device.Identity{
		MAC:          "02:11:22:33:44:58",
		IP:           "192.0.2.50",
		Hostname:     "unifi-stubd-lab",
		Model:        profile.Model,
		ModelDisplay: profile.ModelDisplay,
		Version:      profile.Version,
		Serial:       "021122334458",
		InformURL:    "http://192.0.2.10:8080/inform",
	}, device.SwitchPortsWithOptions(profile.Ports, profile.PortOptions()))
	if err != nil {
		t.Fatal(err)
	}

	var doc struct {
		IfTable   []map[string]any `json:"if_table"`
		PortTable []map[string]any `json:"port_table"`
	}
	if err := json.Unmarshal(payload, &doc); err != nil {
		t.Fatal(err)
	}
	if got := int(doc.IfTable[0]["speed"].(float64)); got != 25000 {
		t.Fatalf("if_table speed = %d, want 25000", got)
	}
	if got := int(doc.PortTable[48]["speed"].(float64)); got != 25000 {
		t.Fatalf("uplink port speed = %d, want 25000", got)
	}
	if got := doc.PortTable[48]["media"].(string); got != "SFP28" {
		t.Fatalf("uplink media = %q, want SFP28", got)
	}
}

func TestMinimalSwitchPayloadReportsObservedCounters(t *testing.T) {
	payload, err := device.MinimalSwitchPayload(device.Identity{
		MAC:          "02:11:22:33:44:59",
		IP:           "192.0.2.50",
		Hostname:     "unifi-stubd-lab",
		Model:        "US8",
		ModelDisplay: "UniFi Switch 8",
		Version:      "7.4.1.16850",
		Serial:       "021122334459",
		InformURL:    "http://192.0.2.10:8080/inform",
	}, []device.Port{
		{
			Index:     1,
			Name:      "Port 1",
			Media:     "GE",
			Uplink:    true,
			Up:        true,
			Speed:     1000,
			RXBytes:   1234,
			TXBytes:   5678,
			RXPackets: 12,
			TXPackets: 34,
			RXErrors:  1,
			TXErrors:  2,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	var doc struct {
		PortTable []map[string]any `json:"port_table"`
	}
	if err := json.Unmarshal(payload, &doc); err != nil {
		t.Fatal(err)
	}
	port := doc.PortTable[0]
	if got := int64(port["rx_bytes"].(float64)); got != 1234 {
		t.Fatalf("rx_bytes = %d, want 1234", got)
	}
	if got := int64(port["tx_packets"].(float64)); got != 34 {
		t.Fatalf("tx_packets = %d, want 34", got)
	}
	if got := int64(port["rx_errors"].(float64)); got != 1 {
		t.Fatalf("rx_errors = %d, want 1", got)
	}
}

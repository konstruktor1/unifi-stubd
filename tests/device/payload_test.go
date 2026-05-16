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
		InformURL:    "http://10.10.0.30:8080/inform",
	}, device.SwitchPorts(16))
	if err != nil {
		t.Fatal(err)
	}

	var doc struct {
		NumPort       int              `json:"num_port"`
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

func TestMinimalSwitchPayloadReportsTenGigUplink(t *testing.T) {
	payload, err := device.MinimalSwitchPayload(device.Identity{
		MAC:          "02:11:22:33:44:57",
		IP:           "192.0.2.50",
		Hostname:     "unifi-stubd-lab",
		Model:        "US16XG",
		ModelDisplay: "UniFi Switch 16 XG",
		Version:      "7.4.1.16850",
		Serial:       "021122334457",
		InformURL:    "http://10.10.0.30:8080/inform",
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
		InformURL:    "http://10.10.0.30:8080/inform",
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

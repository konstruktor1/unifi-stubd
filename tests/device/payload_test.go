//nolint:goconst // Repeated wire keys keep the compact US8 contract explicit.
package device_test

import (
	"encoding/json"
	"testing"

	"github.com/konstruktor1/unifi-stubd/internal/device"
	"github.com/konstruktor1/unifi-stubd/internal/device/payload"
)

func buildUS8Payload(t *testing.T, update func([]device.Port)) map[string]any {
	t.Helper()
	profile, ok := device.LookupProfile("us8")
	if !ok {
		t.Fatal("US8 profile not found")
	}
	ports := device.BuildPorts(profile, device.PortBuildOptions{})
	if update != nil {
		update(ports)
	}
	data, err := payload.Build(profile, device.Identity{
		MAC:          "02:11:22:33:44:55",
		IP:           "192.0.2.50",
		Hostname:     "us8-lab",
		Model:        profile.Model,
		ModelDisplay: profile.ModelDisplay,
		DeviceType:   profile.DeviceType,
		Version:      profile.Version,
		Serial:       "021122334455",
		InformURL:    "http://192.0.2.10:8080/inform",
	}, ports)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	return doc
}

func TestUS8PayloadContract(t *testing.T) {
	doc := buildUS8Payload(t, nil)
	if doc["model"] != "US8" || doc["num_port"] != float64(8) || doc["uplink"] != "eth0" {
		t.Fatalf("US8 identity = model %v ports %v uplink %v", doc["model"], doc["num_port"], doc["uplink"])
	}
	ports := doc["port_table"].([]any)
	if len(ports) != 8 {
		t.Fatalf("port_table length = %d, want 8", len(ports))
	}
	uplinks := 0
	for _, raw := range ports {
		isUplink, _ := raw.(map[string]any)["is_uplink"].(bool)
		if isUplink {
			uplinks++
		}
	}
	if uplinks != 1 {
		t.Fatalf("active uplinks = %d, want 1", uplinks)
	}
	uplink := ports[0].(map[string]any)
	if uplink["port_idx"] != float64(1) || uplink["ifname"] != "eth0" || uplink["speed"] != float64(1000) {
		t.Fatalf("US8 uplink row = index %v ifname %v speed %v", uplink["port_idx"], uplink["ifname"], uplink["speed"])
	}
}

func TestUS8OmitsMeasurementsWithoutSource(t *testing.T) {
	doc := buildUS8Payload(t, func(ports []device.Port) {
		ports[0].TrafficRatesEnabled = true
	})
	measurementKeys := []string{
		"bytes", "rx_bytes", "tx_bytes", "rx_packets", "tx_packets",
		"rx_errors", "tx_errors", "bytes-r", "rx_bytes-r", "tx_bytes-r",
		"rx_rate", "tx_rate", "source_interface",
	}
	assertAbsent := func(label string, row map[string]any) {
		t.Helper()
		for _, key := range measurementKeys {
			if value, ok := row[key]; ok {
				t.Fatalf("source-less %s contains %q=%v", label, key, value)
			}
		}
	}
	assertAbsent("root", doc)
	for _, table := range []string{"if_table", "port_table"} {
		for _, raw := range doc[table].([]any) {
			assertAbsent(table, raw.(map[string]any))
		}
	}
}

func TestUS8ReportsObservedUplinkCountersAndRates(t *testing.T) {
	doc := buildUS8Payload(t, func(ports []device.Port) {
		port := &ports[0]
		port.RXBytes, port.TXBytes = 1000, 2000
		port.RXPackets, port.TXPackets = 10, 20
		port.RXErrors, port.TXErrors = 1, 2
		port.CountersSet = true
		port.RXBytesRate, port.TXBytesRate = 123, 456
		port.TrafficRatesSet = true
	})
	if doc["bytes"] != float64(3000) || doc["rx_bytes"] != float64(2000) || doc["tx_bytes"] != float64(1000) {
		t.Fatalf("root counters = %v/%v/%v", doc["bytes"], doc["rx_bytes"], doc["tx_bytes"])
	}
	if doc["bytes-r"] != float64(579) || doc["rx_bytes-r"] != float64(456) || doc["tx_bytes-r"] != float64(123) {
		t.Fatalf("root rates = %v/%v/%v", doc["bytes-r"], doc["rx_bytes-r"], doc["tx_bytes-r"])
	}
	if doc["rx_rate"] != float64(3648) || doc["tx_rate"] != float64(984) {
		t.Fatalf("root bit rates = %v/%v", doc["rx_rate"], doc["tx_rate"])
	}
	management := doc["if_table"].([]any)[0].(map[string]any)
	if management["rx_bytes"] != float64(1000) || management["tx_bytes"] != float64(2000) {
		t.Fatalf("management counters = %v/%v", management["rx_bytes"], management["tx_bytes"])
	}
	if management["rx_bytes-r"] != float64(123) || management["tx_bytes-r"] != float64(456) {
		t.Fatalf("management rates = %v/%v", management["rx_bytes-r"], management["tx_bytes-r"])
	}
	port := doc["port_table"].([]any)[0].(map[string]any)
	if port["rx_packets"] != float64(10) || port["tx_packets"] != float64(20) || port["rx_errors"] != float64(1) {
		t.Fatalf("uplink packet/error counters = %#v", port)
	}
	if port["rx_bytes-r"] != float64(123) || port["tx_bytes-r"] != float64(456) {
		t.Fatalf("uplink port-local rates = %v/%v", port["rx_bytes-r"], port["tx_bytes-r"])
	}
	downstream := doc["port_table"].([]any)[1].(map[string]any)
	for _, key := range []string{"bytes-r", "rx_bytes-r", "tx_bytes-r", "rx_rate", "tx_rate"} {
		if value, ok := downstream[key]; ok {
			t.Fatalf("downstream port contains uplink measurement %q=%v", key, value)
		}
	}
}

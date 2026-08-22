package device_test

import (
	"encoding/json"
	"testing"

	"github.com/konstruktor1/unifi-stubd/internal/device"
	"github.com/konstruktor1/unifi-stubd/internal/device/payload"
)

func buildUS8Payload(t *testing.T, update func(*device.Port)) map[string]any {
	t.Helper()
	profile, ok := device.NewProfileRegistry().LookupProfile("us8")
	if !ok {
		t.Fatal("US8 profile not found")
	}
	ports := device.BuildPorts(profile, device.PortBuildOptions{})
	if update != nil {
		update(&ports[0])
	}
	data, err := payload.Build(profile, device.Identity{
		MAC: "02:11:22:33:44:55", IP: "192.0.2.50", Hostname: "us8-lab",
		Model: profile.Model, ModelDisplay: profile.ModelDisplay,
		DeviceType: profile.DeviceType, Version: profile.Version,
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

func TestUS8ProductContract(t *testing.T) {
	doc := buildUS8Payload(t, nil)
	rows := doc["port_table"].([]any)
	if doc["model"] != "US8" || doc["num_port"] != float64(8) || len(rows) != 8 {
		t.Fatalf("US8 identity = model %v ports %v/%d", doc["model"], doc["num_port"], len(rows))
	}
	uplinks := 0
	for _, raw := range rows {
		if raw.(map[string]any)["is_uplink"] == true {
			uplinks++
		}
	}
	uplink := rows[0].(map[string]any)
	if uplinks != 1 || uplink["ifname"] != "eth0" || uplink["speed"] != float64(1000) {
		t.Fatalf("US8 uplink = count %d row %#v", uplinks, uplink)
	}
}

func TestUS8OmitsUnobservedMetrics(t *testing.T) {
	doc := buildUS8Payload(t, func(port *device.Port) {
		port.TrafficRatesEnabled = true
	})
	management := doc["if_table"].([]any)[0].(map[string]any)
	for _, field := range []string{"rx_bytes", "rx_bytes-r", "rx_rate"} {
		for surface, row := range map[string]map[string]any{"root": doc, "management": management} {
			if value, ok := row[field]; ok {
				t.Fatalf("unobserved %s contains %s=%v", surface, field, value)
			}
		}
	}
}

func TestUS8ProjectsObservedUplinkMetrics(t *testing.T) {
	doc := buildUS8Payload(t, func(port *device.Port) {
		port.RXBytes, port.TXBytes, port.CountersSet = 1000, 2000, true
		port.RXBytesRate, port.TXBytesRate, port.TrafficRatesSet = 123, 456, true
	})
	if doc["rx_bytes"] != float64(2000) || doc["tx_bytes"] != float64(1000) ||
		doc["rx_bytes-r"] != float64(456) || doc["tx_bytes-r"] != float64(123) {
		t.Fatalf("device-facing metrics = rx %v/%v tx %v/%v",
			doc["rx_bytes"], doc["rx_bytes-r"], doc["tx_bytes"], doc["tx_bytes-r"])
	}
	management := doc["if_table"].([]any)[0].(map[string]any)
	if management["rx_bytes"] != float64(1000) || management["tx_bytes"] != float64(2000) ||
		management["rx_rate"] != float64(984) || management["tx_rate"] != float64(3648) {
		t.Fatalf("current-uplink metrics = rx %v/%v tx %v/%v",
			management["rx_bytes"], management["rx_rate"], management["tx_bytes"], management["tx_rate"])
	}
}

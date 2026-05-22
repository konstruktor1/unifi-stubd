//nolint:goconst // Repeated payload fixture literals document expected UniFi shapes.
package device_test

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/konstruktor1/unifi-stubd/internal/device"
)

// TestMinimalSwitchPayloadReportsPortCount verifies that switch payloads expose
// generated port counts consistently across controller tables.
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

// TestPayloadReportsMonotonicFreshnessFields verifies that synthetic switch
// payloads still look fresh to controller health checks.
func TestPayloadReportsMonotonicFreshnessFields(t *testing.T) {
	payload, err := device.MinimalSwitchPayload(device.Identity{
		MAC:           "02:11:22:33:44:75",
		IP:            "192.0.2.50",
		Hostname:      "freshness-lab",
		Model:         "US8",
		ModelDisplay:  "UniFi Switch 8",
		Version:       "7.4.1.16850",
		Serial:        "021122334475",
		InformURL:     "http://192.0.2.10:8080/inform",
		UptimeSeconds: 125,
	}, device.SwitchPorts(8))
	if err != nil {
		t.Fatal(err)
	}

	var doc struct {
		Uptime      int              `json:"uptime"`
		SysStats    map[string]any   `json:"sys_stats"`
		SystemStats map[string]any   `json:"system-stats"`
		PortTable   []map[string]any `json:"port_table"`
	}
	if err := json.Unmarshal(payload, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.Uptime != 125 {
		t.Fatalf("uptime = %d, want 125", doc.Uptime)
	}
	if got := int(doc.SysStats["uptime"].(float64)); got != 125 {
		t.Fatalf("sys_stats uptime = %d, want 125", got)
	}
	if got := int(doc.SystemStats["uptime"].(float64)); got != 125 {
		t.Fatalf("system-stats uptime = %d, want 125", got)
	}
	if got := int64(doc.PortTable[0]["rx_bytes-r"].(float64)); got <= 0 {
		t.Fatalf("port rx_bytes-r = %d, want positive rate", got)
	}
	if got := int64(doc.PortTable[0]["tx_bytes-r"].(float64)); got <= 0 {
		t.Fatalf("port tx_bytes-r = %d, want positive rate", got)
	}
}

// TestGatewayPayloadReportsFreshnessFields verifies gateway-specific freshness
// and telemetry fields derived from deterministic lab data.
func TestGatewayPayloadReportsFreshnessFields(t *testing.T) {
	profile, ok := device.LookupProfile("uxg-lite")
	if !ok {
		t.Fatal("profile not found")
	}
	payload, err := device.BuildPayload(profile, device.Identity{
		MAC:           "02:11:22:33:44:76",
		IP:            "192.0.2.50",
		Hostname:      "freshness-gateway",
		Model:         profile.Model,
		ModelDisplay:  profile.ModelDisplay,
		DeviceType:    profile.DeviceType,
		Version:       profile.Version,
		Serial:        "021122334476",
		InformURL:     "http://192.0.2.10:8080/inform",
		UptimeSeconds: 3661,
	}, device.SwitchPortsWithOptions(profile.Ports, profile.PortOptions()))
	if err != nil {
		t.Fatal(err)
	}

	var doc struct {
		Uptime    int    `json:"uptime"`
		TimeMS    int64  `json:"time_ms"`
		Timestamp string `json:"timestamp"`
		UptimeStr string `json:"uptime_str"`
	}
	if err := json.Unmarshal(payload, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.Uptime != 3661 {
		t.Fatalf("uptime = %d, want 3661", doc.Uptime)
	}
	if doc.TimeMS <= 0 {
		t.Fatalf("time_ms = %d, want positive wall-clock milliseconds", doc.TimeMS)
	}
	if doc.Timestamp == "" {
		t.Fatal("timestamp is empty")
	}
	if doc.UptimeStr != "1h1m1s" {
		t.Fatalf("uptime_str = %q, want 1h1m1s", doc.UptimeStr)
	}
}

// TestMinimalSwitchPayloadReportsAdoptedState verifies that adoption state maps
// to the controller-facing state fields.
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

// TestMinimalSwitchPayloadReportsTenGigUplink verifies that uplink speed and
// media metadata are rendered for high-speed switch profiles.
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

// TestMinimalSwitchPayloadReportsManagementVLAN verifies the duplicated
// management VLAN fields expected by controller versions.
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

// TestSwitchPortsWithProfilePortGroups verifies contiguous profile port groups
// produce the expected generated port layout.
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

// TestSwitchPortsWithAggregationProPortGroups verifies mixed media and speed
// groups for aggregation-style switch profiles.
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

// TestGatewayProfileReportsDeviceTypeAndPortNames verifies gateway profile
// identity and port naming across gateway payload tables.
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

// TestTenGigGatewayProfileReportsPortLayout verifies high-speed gateway profile
// roles, media, and table synchronization.
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
		DeviceType        string           `json:"type"`
		InformIP          string           `json:"inform_ip"`
		NumPort           int              `json:"num_port"`
		Uplink            string           `json:"uplink"`
		ConfigPortTable   []map[string]any `json:"config_port_table"`
		EthernetOverrides []map[string]any `json:"ethernet_overrides"`
		IfTable           []map[string]any `json:"if_table"`
		NetworkTable      []map[string]any `json:"network_table"`
		PortTable         []map[string]any `json:"port_table"`
		ReportedNetworks  []map[string]any `json:"reported_networks"`
		UplinkTable       []map[string]any `json:"uplink_table"`
		ConfigNetworkWAN  map[string]any   `json:"config_network_wan"`
		ConfigNetworkWAN2 map[string]any   `json:"config_network_wan2"`
		WAN1              map[string]any   `json:"wan1"`
		WAN2              map[string]any   `json:"wan2"`
		OutletEnabled     bool             `json:"outlet_enabled"`
		OutletOverrides   []map[string]any `json:"outlet_overrides"`
		OutletTable       []map[string]any `json:"outlet_table"`
	}
	if err := json.Unmarshal(payload, &doc); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"ethernet_table", "internet", "port_overrides"} {
		if _, ok := raw[key]; ok {
			t.Fatalf("gateway payload contains unsupported table/key %q", key)
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
	if len(doc.ConfigPortTable) != 4 {
		t.Fatalf("config_port_table length = %d, want 4", len(doc.ConfigPortTable))
	}
	if len(doc.EthernetOverrides) != 4 {
		t.Fatalf("ethernet_overrides length = %d, want 4", len(doc.EthernetOverrides))
	}
	if len(doc.ReportedNetworks) != 4 {
		t.Fatalf("reported_networks length = %d, want 4", len(doc.ReportedNetworks))
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
	if len(doc.PortTable) != 4 {
		t.Fatalf("port_table length = %d, want 4", len(doc.PortTable))
	}
	if got := doc.NetworkTable[2]["networkgroup"].(string); got != "WAN2" {
		t.Fatalf("network_table port 3 networkgroup = %q, want WAN2", got)
	}
	if got := doc.PortTable[2]["networkgroup"].(string); got != "WAN2" {
		t.Fatalf("port_table port 3 networkgroup = %q, want WAN2", got)
	}
	if got := doc.PortTable[2]["media"].(string); got != "SFP+" {
		t.Fatalf("port_table port 3 media = %q, want SFP+", got)
	}
	if _, ok := doc.PortTable[2]["native_networkconf_id"]; ok {
		t.Fatalf("port_table unexpectedly sets native_networkconf_id: %#v", doc.PortTable[2]["native_networkconf_id"])
	}
	if _, ok := doc.PortTable[2]["portconf_id"]; ok {
		t.Fatalf("port_table unexpectedly sets portconf_id: %#v", doc.PortTable[2]["portconf_id"])
	}
	if doc.OutletEnabled {
		t.Fatal("outlet_enabled = true, want false for gateway stub")
	}
	if len(doc.OutletOverrides) != 0 {
		t.Fatalf("outlet_overrides length = %d, want 0", len(doc.OutletOverrides))
	}
	if len(doc.OutletTable) != 0 {
		t.Fatalf("outlet_table length = %d, want 0", len(doc.OutletTable))
	}
	assertGatewayConfigNetwork(t, doc.ConfigNetworkWAN, "eth0", "WAN", "wan", 1)
	assertGatewayConfigNetwork(t, doc.ConfigNetworkWAN2, "eth2", "WAN2", "wan2", 3)
	assertGatewayWANStatus(t, doc.WAN1, "eth0", "WAN", "wan", 1)
	assertGatewayWANStatus(t, doc.WAN2, "eth2", "WAN2", "wan2", 3)
}

// TestGatewayPayloadReportsManagementVLANOnUplink verifies that management VLAN
// metadata is attached to the resolved gateway uplink.
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

// TestCloudGatewayFiberProfileReportsGatewayPayload verifies UCG-Fiber profile
// data renders through the gateway payload path.
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

	var doc struct {
		DeviceType        string           `json:"type"`
		NumPort           int              `json:"num_port"`
		Uplink            string           `json:"uplink"`
		ConfigPortTable   []map[string]any `json:"config_port_table"`
		EthernetOverrides []map[string]any `json:"ethernet_overrides"`
		IfTable           []map[string]any `json:"if_table"`
		NetworkTable      []map[string]any `json:"network_table"`
		PortTable         []map[string]any `json:"port_table"`
		ReportedNetworks  []map[string]any `json:"reported_networks"`
		UplinkTable       []map[string]any `json:"uplink_table"`
		ConfigNetworkWAN  map[string]any   `json:"config_network_wan"`
		ConfigNetworkWAN2 map[string]any   `json:"config_network_wan2"`
	}
	if err := json.Unmarshal(payload, &doc); err != nil {
		t.Fatal(err)
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
	if len(doc.ConfigPortTable) != 7 {
		t.Fatalf("config_port_table length = %d, want 7", len(doc.ConfigPortTable))
	}
	if len(doc.EthernetOverrides) != 7 {
		t.Fatalf("ethernet_overrides length = %d, want 7", len(doc.EthernetOverrides))
	}
	if len(doc.ReportedNetworks) != 7 {
		t.Fatalf("reported_networks length = %d, want 7", len(doc.ReportedNetworks))
	}
	if len(doc.PortTable) != 7 {
		t.Fatalf("port_table length = %d, want 7", len(doc.PortTable))
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
	assertGatewayConfigNetwork(t, doc.ConfigNetworkWAN, "eth5", "WAN", "wan", 6)
	assertGatewayConfigNetwork(t, doc.ConfigNetworkWAN2, "eth4", "WAN2", "wan2", 5)
}

// TestGatewayPortAssignmentsCanBeOverriddenFromConfigModel verifies that config
// overrides can adjust gateway port role and network assignment.
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
		ConfigPortTable []map[string]any `json:"config_port_table"`
		NetworkTable    []map[string]any `json:"network_table"`
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

// TestGatewayPayloadReportsHostTableClientMetadata verifies MAC-table client
// metadata appears in gateway host-table output.
func TestGatewayPayloadReportsHostTableClientMetadata(t *testing.T) {
	profile, ok := device.LookupProfile("ugw3")
	if !ok {
		t.Fatal("profile not found")
	}
	ports := device.SwitchPortsWithOptions(profile.Ports, profile.PortOptions())
	ports[1].MACs = []device.MacTableEntry{
		{
			MAC:      "02:00:5e:00:53:03",
			Hostname: "lab-host-2",
			IP:       "192.0.2.52",
			Age:      4,
			Uptime:   1200,
		},
	}
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
		ConfigPortTable []map[string]any `json:"config_port_table"`
		NetworkTable    []map[string]any `json:"network_table"`
	}
	if err := json.Unmarshal(payload, &doc); err != nil {
		t.Fatal(err)
	}
	hosts, ok := doc.NetworkTable[1]["host_table"].([]any)
	if !ok || len(hosts) != 1 {
		t.Fatalf("LAN host_table = %#v", doc.NetworkTable[1]["host_table"])
	}
	host := hosts[0].(map[string]any)
	if host["hostname"] != "lab-host-2" || host["ip"] != "192.0.2.52" {
		t.Fatalf("host_table metadata = %#v", host)
	}
	if hosts, ok := doc.NetworkTable[0]["host_table"].([]any); ok && len(hosts) != 0 {
		t.Fatalf("gateway WAN host_table should not expose uplink neighbor: %#v", hosts)
	}
	connection := doc.ConfigPortTable[1]["last_connection"].(map[string]any)
	if connection["hostname"] != "lab-host-2" || connection["ip"] != "192.0.2.52" {
		t.Fatalf("config_port_table last_connection metadata = %#v", connection)
	}
	if got, ok := doc.ConfigPortTable[1]["connected"].(bool); !ok || !got {
		t.Fatalf("config_port_table connected = %#v, want true", doc.ConfigPortTable[1]["connected"])
	}
}

// TestGatewayPayloadReportsDownstreamDeviceOnLANHostTable verifies downstream
// LAN clients are reported on the LAN host table.
func TestGatewayPayloadReportsDownstreamDeviceOnLANHostTable(t *testing.T) {
	profile, ok := device.LookupProfile("uxgpro")
	if !ok {
		t.Fatal("profile not found")
	}
	ports := device.SwitchPortsWithOptions(profile.Ports, profile.PortOptions())
	ports[3].MACs = []device.MacTableEntry{
		{
			MAC:      "28:70:4e:c3:b7:b8",
			Hostname: "management-downlink",
			IP:       "10.10.0.21",
			VLAN:     1001,
			Type:     "usw",
			Static:   true,
		},
	}
	payload, err := device.MinimalSwitchPayload(device.Identity{
		MAC:          "02:17:05:10:01:21",
		IP:           "10.10.0.29",
		Hostname:     "opnsense",
		Model:        profile.Model,
		ModelDisplay: profile.ModelDisplay,
		DeviceType:   profile.DeviceType,
		Version:      profile.Version,
		Serial:       "021705100121",
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
	hosts, ok := doc.NetworkTable[3]["host_table"].([]any)
	if !ok || len(hosts) != 1 {
		t.Fatalf("LAN2 host_table = %#v", doc.NetworkTable[3]["host_table"])
	}
	host := hosts[0].(map[string]any)
	if host["mac"] != "28:70:4e:c3:b7:b8" || host["ip"] != "10.10.0.21" || host["type"] != "usw" {
		t.Fatalf("downstream host_table metadata = %#v", host)
	}
	if got := int(host["vlan"].(float64)); got != 1001 {
		t.Fatalf("downstream host_table vlan = %d, want 1001", got)
	}
}

// TestSwitchPortsCanOverrideAggregationUplinkToTenGigPort verifies explicit
// uplink selection on grouped aggregation profiles.
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
	if ports[0].ProfileUplink {
		t.Fatal("port 1 should be the active uplink, not a profile uplink cage")
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
	if !ports[28].ProfileUplink {
		t.Fatal("port 29 lost its profile uplink marker")
	}
	if ports[28].Speed != 25000 || ports[28].Media != "SFP28" {
		t.Fatalf("port 29 = speed %d media %q, want 25000 SFP28", ports[28].Speed, ports[28].Media)
	}
}

// TestApplyPortOverridesChangesSpeedAndLinkState verifies operator overrides
// change generated link metadata without changing unrelated ports.
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

// TestApplyPortOverridesKeepsMediaOrderAndCounters verifies override setter
// ordering for speed-derived media, explicit media, and counters.
func TestApplyPortOverridesKeepsMediaOrderAndCounters(t *testing.T) {
	ports := device.ApplyPortOverrides(device.SwitchPorts(8), []device.PortOverride{
		{
			Port:      2,
			Speed:     2500,
			Media:     "SFP+",
			RXBytes:   100,
			TXBytes:   200,
			RXPackets: 3,
			TXPackets: 4,
			RXErrors:  5,
			TXErrors:  6,
		},
	})
	port := ports[1]
	if port.Speed != 2500 {
		t.Fatalf("Speed = %d, want 2500", port.Speed)
	}
	if port.Media != "SFP+" {
		t.Fatalf("Media = %q, want explicit SFP+", port.Media)
	}
	if port.RXBytes != 100 ||
		port.TXBytes != 200 ||
		port.RXPackets != 3 ||
		port.TXPackets != 4 ||
		port.RXErrors != 5 ||
		port.TXErrors != 6 {
		t.Fatalf("counters = %+v", port)
	}
}

// TestGatewayPayloadReportsPortOverrideMACs verifies MAC/IP override data is
// reflected in gateway interface rows.
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

// TestUXGGatewayPayloadUsesInterfaceOverrideData verifies gateway payloads use
// source-interface override data consistently.
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
		WAN1         map[string]any   `json:"wan1"`
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

// TestGatewayPayloadReportsExplicitTrafficRates verifies observed or configured
// byte rates are rendered without synthetic fallback rates.
func TestGatewayPayloadReportsExplicitTrafficRates(t *testing.T) {
	profile, ok := device.LookupProfile("uxg-lite")
	if !ok {
		t.Fatal("profile not found")
	}
	ports := device.ApplyPortOverrides(device.SwitchPortsWithOptions(profile.Ports, profile.PortOptions()), []device.PortOverride{
		{
			Port:            1,
			Name:            "LAN",
			Interface:       "vtnet0",
			MAC:             "02:00:5e:00:53:02",
			IP:              "192.0.2.1",
			Netmask:         "255.255.255.0",
			RXBytes:         1000,
			TXBytes:         2000,
			RXPackets:       101,
			TXPackets:       201,
			RXErrors:        1,
			TXErrors:        2,
			RXBytesRate:     10,
			TXBytesRate:     20,
			TrafficRatesSet: true,
		},
		{
			Port:            2,
			Name:            "WAN",
			Interface:       "ixl0",
			MAC:             "02:00:5e:00:53:01",
			IP:              "198.51.100.9",
			Netmask:         "255.255.255.0",
			RXBytes:         3000,
			TXBytes:         4000,
			RXPackets:       301,
			TXPackets:       401,
			RXErrors:        3,
			TXErrors:        4,
			RXBytesRate:     30,
			TXBytesRate:     40,
			TrafficRatesSet: true,
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
		PortTable    []map[string]any `json:"port_table"`
		UplinkTable  []map[string]any `json:"uplink_table"`
		WAN1         map[string]any   `json:"wan1"`
	}
	if err := json.Unmarshal(payload, &doc); err != nil {
		t.Fatal(err)
	}
	if got := int64(doc.IfTable[0]["rx_bytes-r"].(float64)); got != 10 {
		t.Fatalf("LAN if_table rx rate = %d, want 10", got)
	}
	if got := int64(doc.IfTable[1]["tx_packets"].(float64)); got != 401 {
		t.Fatalf("WAN if_table tx_packets = %d, want 401", got)
	}
	wanStats := doc.NetworkTable[1]["stats"].(map[string]any)
	if got := int64(wanStats["tx_bytes-r"].(float64)); got != 40 {
		t.Fatalf("WAN network_table tx rate = %d, want 40", got)
	}
	if got := int64(wanStats["rx_packets"].(float64)); got != 301 {
		t.Fatalf("WAN network_table rx_packets = %d, want 301", got)
	}
	if got := int64(doc.UplinkTable[0]["rx_bytes-r"].(float64)); got != 30 {
		t.Fatalf("uplink rx rate = %d, want 30", got)
	}
	if got := int64(doc.UplinkTable[0]["tx_errors"].(float64)); got != 4 {
		t.Fatalf("uplink tx_errors = %d, want 4", got)
	}
	if got := int64(doc.WAN1["rx_bytes-r"].(float64)); got != 30 {
		t.Fatalf("wan1 rx rate = %d, want 30", got)
	}
	if got := int64(doc.WAN1["tx_bytes"].(float64)); got != 4000 {
		t.Fatalf("wan1 tx_bytes = %d, want 4000", got)
	}
	if got := int64(doc.PortTable[1]["rx_bytes-r"].(float64)); got != 30 {
		t.Fatalf("gateway port_table WAN rx rate = %d, want 30", got)
	}
	if got := int64(doc.PortTable[1]["tx_bytes"].(float64)); got != 4000 {
		t.Fatalf("gateway port_table WAN tx_bytes = %d, want 4000", got)
	}
	assertNoExperimentalRateFields(t, doc.IfTable[0])
	assertNoExperimentalRateFields(t, wanStats)
	assertNoExperimentalRateFields(t, doc.PortTable[1])
	assertNoExperimentalRateFields(t, doc.UplinkTable[0])
	assertNoExperimentalRateFields(t, doc.WAN1)
}

// TestGatewayPayloadSynchronizesResolvedPortTables verifies gateway tables all
// consume the same resolved PortView data.
func TestGatewayPayloadSynchronizesResolvedPortTables(t *testing.T) {
	profile, ok := device.LookupProfile("uxg-lite")
	if !ok {
		t.Fatal("profile not found")
	}
	ports := device.ApplyPortOverrides(device.SwitchPortsWithOptions(profile.Ports, profile.PortOptions()), []device.PortOverride{
		{
			Port:         1,
			Name:         "LAN",
			Interface:    "vtnet0",
			MAC:          "02:00:5e:00:53:11",
			IP:           "192.0.2.1",
			Netmask:      "255.255.255.0",
			Role:         "lan",
			NetworkGroup: "LAN",
			Speed:        2500,
		},
		{
			Port:         2,
			Name:         "WAN",
			Interface:    "ixl0",
			MAC:          "02:00:5e:00:53:12",
			IP:           "198.51.100.9",
			Netmask:      "255.255.255.0",
			Role:         "wan",
			NetworkGroup: "WAN",
			Speed:        10000,
			Media:        "SFP+",
		},
	})
	payload, err := device.MinimalSwitchPayload(device.Identity{
		MAC:          "02:00:5e:00:53:10",
		IP:           "192.0.2.1",
		Hostname:     "opnsense",
		Model:        profile.Model,
		ModelDisplay: profile.ModelDisplay,
		DeviceType:   profile.DeviceType,
		Version:      profile.Version,
		Serial:       "02005E005310",
		InformURL:    "http://192.0.2.10:8080/inform",
	}, ports)
	if err != nil {
		t.Fatal(err)
	}

	var doc struct {
		ConfigPortTable   []map[string]any `json:"config_port_table"`
		EthernetOverrides []map[string]any `json:"ethernet_overrides"`
		IfTable           []map[string]any `json:"if_table"`
		NetworkTable      []map[string]any `json:"network_table"`
		PortTable         []map[string]any `json:"port_table"`
		ReportedNetworks  []map[string]any `json:"reported_networks"`
		ConfigNetworkWAN  map[string]any   `json:"config_network_wan"`
		ConfigNetworkWAN2 map[string]any   `json:"config_network_wan2,omitempty"`
		WAN1              map[string]any   `json:"wan1"`
		WAN2              map[string]any   `json:"wan2,omitempty"`
	}
	if err := json.Unmarshal(payload, &doc); err != nil {
		t.Fatal(err)
	}
	assertGatewayPortSync := func(index int, ifname, mac, networkGroup, sourceInterface string, speed int) {
		t.Helper()
		rows := []struct {
			name string
			row  map[string]any
		}{
			{"config_port_table", doc.ConfigPortTable[index-1]},
			{"ethernet_overrides", doc.EthernetOverrides[index-1]},
			{"if_table", doc.IfTable[index-1]},
			{"network_table", doc.NetworkTable[index-1]},
			{"port_table", doc.PortTable[index-1]},
			{"reported_networks", doc.ReportedNetworks[index-1]},
		}
		for _, item := range rows {
			if got := item.row["ifname"].(string); got != ifname {
				t.Fatalf("%s port %d ifname = %q, want %q", item.name, index, got, ifname)
			}
			if got := item.row["networkgroup"].(string); got != networkGroup {
				t.Fatalf("%s port %d networkgroup = %q, want %q", item.name, index, got, networkGroup)
			}
			if got := item.row["source_interface"].(string); got != sourceInterface {
				t.Fatalf("%s port %d source_interface = %q, want %q", item.name, index, got, sourceInterface)
			}
		}
		for _, item := range []struct {
			name string
			row  map[string]any
		}{
			{"ethernet_overrides", doc.EthernetOverrides[index-1]},
			{"if_table", doc.IfTable[index-1]},
			{"network_table", doc.NetworkTable[index-1]},
			{"port_table", doc.PortTable[index-1]},
		} {
			if got := item.row["mac"].(string); got != mac {
				t.Fatalf("%s port %d mac = %q, want %q", item.name, index, got, mac)
			}
		}
		if got := int(doc.IfTable[index-1]["speed"].(float64)); got != speed {
			t.Fatalf("if_table port %d speed = %d, want %d", index, got, speed)
		}
		if got := doc.NetworkTable[index-1]["max_speed"].(string); got != strconv.Itoa(speed) {
			t.Fatalf("network_table port %d max_speed = %q, want %d", index, got, speed)
		}
	}

	assertGatewayPortSync(1, "eth0", "02:00:5e:00:53:11", "LAN", "vtnet0", 2500)
	assertGatewayPortSync(2, "eth1", "02:00:5e:00:53:12", "WAN", "ixl0", 10000)
	assertGatewayConfigNetwork(t, doc.ConfigNetworkWAN, "eth1", "WAN", "wan", 2)
	assertGatewayWANStatus(t, doc.WAN1, "eth1", "WAN", "wan", 2)
	if doc.ConfigNetworkWAN2 != nil {
		t.Fatalf("config_network_wan2 = %#v, want omitted", doc.ConfigNetworkWAN2)
	}
	if doc.WAN2 != nil {
		t.Fatalf("wan2 = %#v, want omitted", doc.WAN2)
	}
}

// TestCustomGatewayPayloadUsesProfileRolesWithoutModelSpecialCase verifies
// gateway behavior follows profile roles instead of hard-coded model names.
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

// TestApplyUplinkNeighborAddsConfiguredNeighbor verifies configured uplink
// topology hints are inserted into the uplink MAC table.
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

// TestApplyPortNeighborsAddsConfiguredMacTableEntry verifies configured
// per-port neighbor metadata is added to the target port.
func TestApplyPortNeighborsAddsConfiguredMacTableEntry(t *testing.T) {
	profile, ok := device.LookupProfile("ugw3")
	if !ok {
		t.Fatal("profile not found")
	}
	ports := device.ApplyPortNeighbors(device.SwitchPortsWithOptions(profile.Ports, profile.PortOptions()), []device.PortNeighbor{
		{
			Port: 2,
			Entry: device.MacTableEntry{
				MAC:      "02:00:5e:00:53:03",
				Hostname: "lab-host-2",
				IP:       "192.0.2.52",
				VLAN:     1,
				Static:   true,
				Type:     "usw",
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
	if entry.Hostname != "lab-host-2" || entry.IP != "192.0.2.52" || !entry.Static {
		t.Fatalf("port 2 neighbor metadata = %+v", entry)
	}
	if entry.Age == 0 || entry.Uptime == 0 {
		t.Fatalf("port 2 neighbor missing defaults: %+v", entry)
	}
	if len(ports[0].MACs) == 0 {
		t.Fatal("port 1 lost its generated uplink MAC table")
	}
}

// TestApplyPortNeighborsDefaultsToClientType verifies per-port neighbors
// default to client topology type.
func TestApplyPortNeighborsDefaultsToClientType(t *testing.T) {
	ports := device.ApplyPortNeighbors(device.SwitchPorts(4), []device.PortNeighbor{
		{
			Port: 2,
			Entry: device.MacTableEntry{
				MAC:      "02:00:5e:00:53:03",
				Hostname: "lab-host-2",
				IP:       "192.0.2.52",
			},
		},
	})

	entry := ports[1].MACs[0]
	if entry.Type != "client" {
		t.Fatalf("port neighbor default type = %q, want client", entry.Type)
	}
}

// TestApplyPortNeighborsPreservesObservedClientIP verifies configured neighbor
// data can merge with observed client IP metadata.
func TestApplyPortNeighborsPreservesObservedClientIP(t *testing.T) {
	ports := device.SwitchPorts(4)
	ports[1].MACs = []device.MacTableEntry{
		{MAC: "02:00:5e:00:53:03", IP: "192.0.2.52", VLAN: 20, Age: 4, Uptime: 1200, Type: "client"},
	}

	ports = device.ApplyPortNeighbors(ports, []device.PortNeighbor{
		{
			Port: 2,
			Entry: device.MacTableEntry{
				MAC:      "02:00:5e:00:53:03",
				Hostname: "lab-host-2",
				Type:     "client",
			},
		},
	})

	entry := ports[1].MACs[0]
	if entry.Hostname != "lab-host-2" || entry.IP != "192.0.2.52" || entry.VLAN != 20 {
		t.Fatalf("merged neighbor = %+v", entry)
	}
}

// TestSwitchPayloadReportsNeighborClientMetadata verifies switch port tables
// render configured neighbor client metadata.
func TestSwitchPayloadReportsNeighborClientMetadata(t *testing.T) {
	ports := device.ApplyPortNeighbors(device.SwitchPorts(4), []device.PortNeighbor{
		{
			Port: 2,
			Entry: device.MacTableEntry{
				MAC:      "02:00:5e:00:53:03",
				Hostname: "lab-host-2",
				IP:       "192.0.2.52",
				VLAN:     1,
				Static:   true,
				Type:     "client",
			},
		},
	})
	payload, err := device.MinimalSwitchPayload(device.Identity{
		MAC:          "02:11:22:33:44:60",
		IP:           "192.0.2.50",
		Hostname:     "unifi-stubd-lab",
		Model:        "US8",
		ModelDisplay: "UniFi Switch 8",
		Version:      "7.4.1.16850",
		Serial:       "021122334460",
		InformURL:    "http://192.0.2.10:8080/inform",
	}, ports)
	if err != nil {
		t.Fatal(err)
	}

	var doc struct {
		PortTable []map[string]any `json:"port_table"`
	}
	if err := json.Unmarshal(payload, &doc); err != nil {
		t.Fatal(err)
	}
	macs, ok := doc.PortTable[1]["mac_table"].([]any)
	if !ok || len(macs) != 1 {
		t.Fatalf("port 2 mac_table = %#v", doc.PortTable[1]["mac_table"])
	}
	entry := macs[0].(map[string]any)
	if entry["hostname"] != "lab-host-2" || entry["ip"] != "192.0.2.52" || entry["static"] != true {
		t.Fatalf("mac_table metadata = %#v", entry)
	}
}

// TestMinimalSwitchPayloadReportsPortOverrideLinkDown verifies link-down
// overrides clear live link fields in switch payloads.
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
	if enabled := port["enable"].(bool); !enabled {
		t.Fatal("link-down port is administratively disabled")
	}
}

// TestMinimalSwitchPayloadReportsDisabledPort verifies disabled ports render as
// down, zero-speed, and without learned MACs.
func TestMinimalSwitchPayloadReportsDisabledPort(t *testing.T) {
	payload, err := device.MinimalSwitchPayload(device.Identity{
		MAC:          "02:11:22:33:44:61",
		IP:           "192.0.2.50",
		Hostname:     "unifi-stubd-lab",
		Model:        "US8",
		ModelDisplay: "UniFi Switch 8",
		Version:      "7.4.1.16850",
		Serial:       "021122334461",
		InformURL:    "http://192.0.2.10:8080/inform",
	}, device.ApplyPortOverrides(device.SwitchPorts(8), []device.PortOverride{
		{Port: 1, Disabled: true},
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
	port := doc.PortTable[0]
	if up := port["up"].(bool); up {
		t.Fatal("disabled port is up")
	}
	if enabled := port["enable"].(bool); enabled {
		t.Fatal("disabled port is administratively enabled")
	}
	if got := int(port["speed"].(float64)); got != 0 {
		t.Fatalf("disabled port speed = %d, want 0", got)
	}
	if macs, ok := port["mac_table"].([]any); ok && len(macs) != 0 {
		t.Fatalf("disabled port MAC table length = %d, want 0", len(macs))
	}
}

// TestMinimalSwitchPayloadReportsGroupedUplinkSpeed verifies grouped profile
// uplink speed reaches switch payload rows.
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

// TestMinimalSwitchPayloadReportsObservedCounters verifies observed interface
// counters are copied into switch payload rows.
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

// TestMinimalSwitchPayloadPrefersExplicitTrafficRates verifies explicit traffic
// rates take precedence over synthetic heartbeat rates.
func TestMinimalSwitchPayloadPrefersExplicitTrafficRates(t *testing.T) {
	payload, err := device.MinimalSwitchPayload(device.Identity{
		MAC:          "02:11:22:33:44:60",
		IP:           "192.0.2.50",
		Hostname:     "unifi-stubd-lab",
		Model:        "US8",
		ModelDisplay: "UniFi Switch 8",
		Version:      "7.4.1.16850",
		Serial:       "021122334460",
		InformURL:    "http://192.0.2.10:8080/inform",
	}, []device.Port{
		{
			Index:           1,
			Name:            "Port 1",
			Media:           "GE",
			Uplink:          true,
			Up:              true,
			Speed:           1000,
			RXBytesRate:     123,
			TXBytesRate:     456,
			TrafficRatesSet: true,
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
	if got := int64(port["rx_bytes-r"].(float64)); got != 123 {
		t.Fatalf("rx_bytes-r = %d, want 123", got)
	}
	if got := int64(port["tx_bytes-r"].(float64)); got != 456 {
		t.Fatalf("tx_bytes-r = %d, want 456", got)
	}
}

// TestMinimalSwitchPayloadSuppressesSyntheticRatesWhenTrafficRatesEnabledWithoutSource
// verifies rate fields stay zero when tracking is enabled without a source.
func TestMinimalSwitchPayloadSuppressesSyntheticRatesWhenTrafficRatesEnabledWithoutSource(t *testing.T) {
	payload, err := device.MinimalSwitchPayload(device.Identity{
		MAC:          "02:11:22:33:44:61",
		IP:           "192.0.2.50",
		Hostname:     "unifi-stubd-lab",
		Model:        "US8",
		ModelDisplay: "UniFi Switch 8",
		Version:      "7.4.1.16850",
		Serial:       "021122334461",
		InformURL:    "http://192.0.2.10:8080/inform",
	}, []device.Port{
		{
			Index:               1,
			Name:                "Port 1",
			Media:               "GE",
			Uplink:              true,
			Up:                  true,
			Speed:               1000,
			TrafficRatesEnabled: true,
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
	if got := int64(port["rx_bytes-r"].(float64)); got != 0 {
		t.Fatalf("rx_bytes-r = %d, want 0", got)
	}
	if got := int64(port["tx_bytes-r"].(float64)); got != 0 {
		t.Fatalf("tx_bytes-r = %d, want 0", got)
	}
	assertNoExperimentalRateFields(t, port)
}

// assertNoExperimentalRateFields verifies payload rows avoid legacy
// experimental rate keys.
func assertNoExperimentalRateFields(t *testing.T, row map[string]any) {
	t.Helper()
	for _, key := range []string{"rx_packets-r", "tx_packets-r", "rx_errors-r", "tx_errors-r"} {
		if _, ok := row[key]; ok {
			t.Fatalf("unexpected experimental rate field %q in %#v", key, row)
		}
	}
}

// assertGatewayConfigNetwork checks the shared gateway config-network row
// contract used across WAN and LAN tests.
func assertGatewayConfigNetwork(t *testing.T, row map[string]any, ifname, networkGroup, role string, portIndex int) {
	t.Helper()
	if got := row["type"].(string); got != "dhcp" {
		t.Fatalf("config network type = %q, want dhcp", got)
	}
	if got := row["ifname"].(string); got != ifname {
		t.Fatalf("config network ifname = %q, want %q", got, ifname)
	}
	if got := row["networkgroup"].(string); got != networkGroup {
		t.Fatalf("config network networkgroup = %q, want %q", got, networkGroup)
	}
	if got := row["role"].(string); got != role {
		t.Fatalf("config network role = %q, want %q", got, role)
	}
	if got := int(row["port_idx"].(float64)); got != portIndex {
		t.Fatalf("config network port_idx = %d, want %d", got, portIndex)
	}
}

// assertGatewayWANStatus checks the shared gateway WAN status row contract.
func assertGatewayWANStatus(t *testing.T, row map[string]any, ifname, networkGroup, role string, portIndex int) {
	t.Helper()
	if row == nil {
		t.Fatal("wan status is nil")
	}
	assertGatewayConfigNetwork(t, row, ifname, networkGroup, role, portIndex)
	if got, ok := row["up"].(bool); !ok || !got {
		t.Fatalf("wan status up = %#v, want true", row["up"])
	}
	if got := int(row["uptime"].(float64)); got < 1 {
		t.Fatalf("wan status uptime = %d, want >= 1", got)
	}
	if got := int(row["latency"].(float64)); got != 0 {
		t.Fatalf("wan status latency = %d, want 0", got)
	}
}

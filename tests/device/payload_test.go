//nolint:goconst // Repeated payload fixture literals document expected UniFi shapes.
package device_test

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/konstruktor1/unifi-stubd/internal/device"
	"github.com/konstruktor1/unifi-stubd/internal/device/payload"
)

func switchPayload(id device.Identity, ports []device.Port) ([]byte, error) {
	doc, err := payload.Build(device.Profile{}, id, ports)
	if err != nil {
		return nil, fmt.Errorf("build minimal switch payload: %w", err)
	}
	return doc, nil
}

func buildPayload(profile device.Profile, id device.Identity, ports []device.Port) ([]byte, error) {
	doc, err := payload.Build(profile, id, ports)
	if err != nil {
		return nil, fmt.Errorf("build payload: %w", err)
	}
	return doc, nil
}

// TestSwitchPayloadReportsPortCount verifies that switch payloads expose
// generated port counts consistently across controller tables.
func TestSwitchPayloadReportsPortCount(t *testing.T) {
	payload, err := switchPayload(device.Identity{
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
	payload, err := switchPayload(device.Identity{
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
	payload, err := buildPayload(profile, device.Identity{
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
	}, device.BuildPorts(profile, device.PortBuildOptions{}))
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

// TestSwitchPayloadReportsAdoptedState verifies that adoption state maps
// to the controller-facing state fields.
func TestSwitchPayloadReportsAdoptedState(t *testing.T) {
	payload, err := switchPayload(device.Identity{
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

// TestSwitchPayloadReportsTenGigUplink verifies that uplink speed and
// media metadata are rendered for high-speed switch profiles.
func TestSwitchPayloadReportsTenGigUplink(t *testing.T) {
	payload, err := switchPayload(device.Identity{
		MAC:          "02:11:22:33:44:57",
		IP:           "192.0.2.50",
		Hostname:     "unifi-stubd-lab",
		Model:        "US16XG",
		ModelDisplay: "UniFi Switch 16 XG",
		Version:      "7.4.1.16850",
		Serial:       "021122334457",
		InformURL:    "http://192.0.2.10:8080/inform",
	}, device.BuildPorts(device.Profile{
		Ports:       16,
		PortSpeed:   10000,
		UplinkSpeed: 10000,
		PortMedia:   "SFP+",
		UplinkMedia: "SFP+",
	}, device.PortBuildOptions{}))
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

// TestSwitchPayloadReportsManagementVLAN verifies the duplicated
// management VLAN fields expected by controller versions.
func TestSwitchPayloadReportsManagementVLAN(t *testing.T) {
	payload, err := switchPayload(device.Identity{
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
	ports := device.BuildPorts(profile, device.PortBuildOptions{})

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
	ports := device.BuildPorts(profile, device.PortBuildOptions{})

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
	ports := device.BuildPorts(profile, device.PortBuildOptions{})
	payload, err := switchPayload(device.Identity{
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
// media and controller-facing gateway port tables.
func TestTenGigGatewayProfileReportsPortLayout(t *testing.T) {
	profile, ok := device.LookupProfile("uxgpro")
	if !ok {
		t.Fatal("profile not found")
	}
	linkDown := false
	ports := device.ApplyPortOverrides(device.BuildPorts(profile, device.PortBuildOptions{UplinkPort: 3}), []device.PortOverride{
		{Port: 1, Up: &linkDown},
		{Port: 2, Up: &linkDown},
		{Port: 3, Role: "wan", NetworkGroup: "WAN"},
	})
	payload, err := switchPayload(device.Identity{
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
		DeviceType       string           `json:"type"`
		FWCaps           int              `json:"fw_caps"`
		InformIP         string           `json:"inform_ip"`
		NumPort          int              `json:"num_port"`
		UDAPICaps        int              `json:"udapi_caps"`
		UDAPIVersion     map[string]any   `json:"udapi_version"`
		Uplink           string           `json:"uplink"`
		UplinkDepth      int              `json:"uplink_depth"`
		LastUplink       any              `json:"last_uplink"`
		IfTable          []map[string]any `json:"if_table"`
		NetworkTable     []map[string]any `json:"network_table"`
		ConfigPortTable  []map[string]any `json:"config_port_table"`
		EthernetTable    []map[string]any `json:"ethernet_table"`
		EthernetOverride []map[string]any `json:"ethernet_overrides"`
		PortTable        []map[string]any `json:"port_table"`
		ConfigNetworkWAN map[string]any   `json:"config_network_wan"`
		OutletEnable     bool             `json:"outlet_enabled"`
		OutletTable      []map[string]any `json:"outlet_table"`
		OutletOvr        []map[string]any `json:"outlet_overrides"`
		UplinkTable      []map[string]any `json:"uplink_table"`
		WAN1             map[string]any   `json:"wan1"`
	}
	if err := json.Unmarshal(payload, &doc); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{
		"internet",
		"port_overrides",
	} {
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
	if doc.UplinkDepth != 0 || doc.LastUplink != nil {
		t.Fatalf("gateway root hints = uplink_depth %d last_uplink %#v, want 0/null", doc.UplinkDepth, doc.LastUplink)
	}
	if doc.FWCaps == 0 || doc.UDAPICaps == 0 {
		t.Fatalf("gateway caps fw=%d udapi=%d, want non-zero", doc.FWCaps, doc.UDAPICaps)
	}
	if doc.UDAPIVersion["version"] == nil {
		t.Fatalf("udapi_version = %#v, want version metadata", doc.UDAPIVersion)
	}
	if len(doc.ConfigPortTable) != 4 {
		t.Fatalf("config_port_table length = %d, want 4", len(doc.ConfigPortTable))
	}
	if len(doc.EthernetTable) != 4 {
		t.Fatalf("ethernet_table length = %d, want 4", len(doc.EthernetTable))
	}
	if len(doc.EthernetOverride) != 2 {
		t.Fatalf("ethernet_overrides length = %d, want 2", len(doc.EthernetOverride))
	}
	if len(doc.PortTable) != 4 {
		t.Fatalf("port_table length = %d, want 4", len(doc.PortTable))
	}
	if got := doc.PortTable[2]["ifname"].(string); got != "eth2" {
		t.Fatalf("port_table port 3 ifname = %q, want eth2", got)
	}
	if got := doc.PortTable[2]["media"].(string); got != "SFP+" {
		t.Fatalf("port_table port 3 media = %q, want SFP+", got)
	}
	if got := doc.PortTable[2]["networkgroup"].(string); got != "WAN" {
		t.Fatalf("port_table port 3 networkgroup = %q, want WAN", got)
	}
	if got := doc.PortTable[2]["network_name"].(string); got != "wan" {
		t.Fatalf("port_table port 3 network_name = %q, want wan", got)
	}
	if got := int(doc.PortTable[2]["max_speed"].(float64)); got != 10000 {
		t.Fatalf("port_table port 3 max_speed = %d, want 10000", got)
	}
	ifRow := rowByPortIndex(t, doc.IfTable, 3)
	if got := ifRow["ifname"].(string); got != "eth2" {
		t.Fatalf("if_table port 3 ifname = %q, want eth2", got)
	}
	if got := int(ifRow["speed"].(float64)); got != 10000 {
		t.Fatalf("if_table port 3 speed = %d, want 10000", got)
	}
	if doc.OutletEnable || len(doc.OutletTable) != 0 || len(doc.OutletOvr) != 0 {
		t.Fatalf("outlet fields = enabled %t table %d overrides %d, want disabled empty", doc.OutletEnable, len(doc.OutletTable), len(doc.OutletOvr))
	}
	if got := doc.ConfigNetworkWAN["type"].(string); got != "dhcp" {
		t.Fatalf("config_network_wan type = %q, want dhcp", got)
	}
	assertGatewayConfigNetworkWAN(t, doc.ConfigNetworkWAN, "192.0.2.2", "255.255.255.0")
	if got := doc.ConfigNetworkWAN["ifname"].(string); got != "eth2" {
		t.Fatalf("config_network_wan ifname = %q, want eth2", got)
	}
	if got := int(doc.ConfigNetworkWAN["port_idx"].(float64)); got != 3 {
		t.Fatalf("config_network_wan port_idx = %d, want 3", got)
	}
	if len(doc.UplinkTable) != 1 {
		t.Fatalf("uplink_table length = %d, want 1", len(doc.UplinkTable))
	}
	if got := doc.WAN1["ifname"].(string); got != "eth2" {
		t.Fatalf("wan1 ifname = %q, want eth2", got)
	}
	if got := int(doc.WAN1["port_idx"].(float64)); got != 3 {
		t.Fatalf("wan1 port_idx = %d, want 3", got)
	}
	if got := doc.WAN1["media"].(string); got != "SFP+" {
		t.Fatalf("wan1 media = %q, want SFP+", got)
	}
	if got := int(doc.WAN1["max_speed"].(float64)); got != 10000 {
		t.Fatalf("wan1 max_speed = %d, want 10000", got)
	}
	if doc.Uplink != "eth2" {
		t.Fatalf("uplink = %q, want eth2", doc.Uplink)
	}
	if len(doc.NetworkTable) != 4 {
		t.Fatalf("network_table length = %d, want 4", len(doc.NetworkTable))
	}
	networkWAN := rowByPortIndex(t, doc.NetworkTable, 3)
	if got := networkWAN["ifname"].(string); got != "eth2" {
		t.Fatalf("network_table WAN ifname = %q, want eth2", got)
	}
	if got := networkWAN["max_speed"].(string); got != "10000" {
		t.Fatalf("network_table WAN max_speed = %q, want 10000", got)
	}
}

// TestGatewayNetworkBindingsKeepProfilePortInterfaces verifies that explicit
// controller assignment metadata changes the WAN/LAN function without renaming
// the profile port interface that Network uses to identify the connector.
func TestGatewayNetworkBindingsKeepProfilePortInterfaces(t *testing.T) {
	profile, ok := device.LookupProfile("uxgpro")
	if !ok {
		t.Fatal("profile not found")
	}
	ports := device.ApplyPortOverrides(device.BuildPorts(profile, device.PortBuildOptions{UplinkPort: 3}), []device.PortOverride{
		{Port: 1, Role: "unassigned", NetworkGroup: "Unassigned"},
		{Port: 2, Role: "unassigned", NetworkGroup: "Unassigned"},
		{
			Port:                3,
			Role:                "wan",
			NetworkGroup:        "WAN",
			NetworkConfID:       "wan-network-id",
			NativeNetworkConfID: "wan-network-id",
		},
		{
			Port:                4,
			Role:                "lan",
			NetworkGroup:        "LAN",
			Interface:           "vtnet0",
			NetworkConfID:       "lan-network-id",
			NativeNetworkConfID: "lan-network-id",
		},
	})
	payload, err := switchPayload(device.Identity{
		MAC:          "02:11:22:33:44:68",
		IP:           "192.0.2.50",
		Hostname:     "unifi-stubd-uxg",
		Model:        profile.Model,
		ModelDisplay: profile.ModelDisplay,
		DeviceType:   profile.DeviceType,
		Version:      profile.Version,
		Serial:       "021122334468",
		InformURL:    "http://192.0.2.10:8080/inform",
	}, ports)
	if err != nil {
		t.Fatal(err)
	}

	var doc struct {
		Uplink           string                    `json:"uplink"`
		IfTable          []map[string]any          `json:"if_table"`
		NetworkTable     []map[string]any          `json:"network_table"`
		PortTable        []map[string]any          `json:"port_table"`
		ConfigNetworkWAN map[string]any            `json:"config_network_wan"`
		ConfigNetworkLAN map[string]any            `json:"config_network_lan"`
		UptimeStats      map[string]map[string]any `json:"uptime_stats"`
		WAN1             map[string]any            `json:"wan1"`
		Wans             []map[string]any          `json:"wans"`
		EthernetOverride []map[string]any          `json:"ethernet_overrides"`
		LANIP            string                    `json:"lan_ip"`
		HasEth1          bool                      `json:"has_eth1"`
	}
	if err := json.Unmarshal(payload, &doc); err != nil {
		t.Fatal(err)
	}
	for name, row := range map[string]map[string]any{
		"if_table":           rowByPortIndex(t, doc.IfTable, 3),
		"network_table":      rowByPortIndex(t, doc.NetworkTable, 3),
		"config_network_wan": doc.ConfigNetworkWAN,
		"wan1":               doc.WAN1,
		"uptime_stats":       doc.UptimeStats["WAN"],
		"wans":               doc.Wans[0],
	} {
		got := row["ifname"]
		if name == "wans" {
			got = row["interface"]
		} else if got == nil {
			t.Fatalf("%s has no ifname: %#v", name, row)
		}
		if got != "eth2" {
			t.Fatalf("%s WAN interface = %q, want eth2", name, got)
		}
	}
	if doc.Uplink != "eth2" {
		t.Fatalf("uplink = %q, want eth2", doc.Uplink)
	}
	if len(doc.EthernetOverride) != 4 {
		t.Fatalf("ethernet_overrides length = %d, want 4", len(doc.EthernetOverride))
	}
	for index, want := range map[int]string{1: "Unassigned", 2: "Unassigned", 3: "WAN", 4: "LAN"} {
		row := rowByPortIndex(t, doc.EthernetOverride, index)
		if got := row["networkgroup"].(string); got != want {
			t.Fatalf("ethernet_overrides port %d networkgroup = %q, want %q", index, got, want)
		}
	}
	for _, index := range []int{1, 2} {
		if got := rowByPortIndex(t, doc.EthernetOverride, index)["disabled"].(bool); !got {
			t.Fatalf("ethernet_overrides port %d disabled = false, want true", index)
		}
	}
	for _, index := range []int{3, 4} {
		if _, ok := rowByPortIndex(t, doc.EthernetOverride, index)["disabled"]; ok {
			t.Fatalf("ethernet_overrides port %d has disabled field", index)
		}
	}
	if got := rowByPortIndex(t, doc.EthernetOverride, 3)["ifname"].(string); got != "eth2" {
		t.Fatalf("ethernet_overrides WAN ifname = %q, want eth2", got)
	}
	port := rowByPortIndex(t, doc.PortTable, 3)
	if got := port["name"].(string); got != "eth2" {
		t.Fatalf("port_table port 3 name = %q, want eth2", got)
	}
	for _, index := range []int{1, 2} {
		if got, ok := rowByPortIndex(t, doc.PortTable, index)["ip"]; ok {
			t.Fatalf("port_table port %d has IP %v, want no IP on unassigned port", index, got)
		}
	}
	lan := rowByPortIndex(t, doc.NetworkTable, 4)
	if got := lan["ifname"].(string); got != "eth3" {
		t.Fatalf("LAN network_table ifname = %q, want eth3", got)
	}
	if got := doc.ConfigNetworkLAN["ifname"].(string); got != "eth3" {
		t.Fatalf("config_network_lan ifname = %q, want eth3", got)
	}
	if got := int(doc.ConfigNetworkLAN["port_idx"].(float64)); got != 4 {
		t.Fatalf("config_network_lan port_idx = %d, want 4", got)
	}
	if got := rowByPortIndex(t, doc.PortTable, 4)["ifname"].(string); got != "eth3" {
		t.Fatalf("port_table LAN ifname = %q, want eth3", got)
	}
	if got := rowByPortIndex(t, doc.EthernetOverride, 4)["ifname"].(string); got != "eth3" {
		t.Fatalf("ethernet_overrides LAN ifname = %q, want eth3", got)
	}
	lanOverride := rowByPortIndex(t, doc.EthernetOverride, 4)
	lanPort := firstRowByString(t, doc.PortTable, "ifname", lanOverride["ifname"].(string))
	if got := int(lanPort["port_idx"].(float64)); got != 4 {
		t.Fatalf("controller LAN lookup port_idx = %d, want 4", got)
	}
	if got := lanPort["ip"].(string); got != "192.0.2.50" {
		t.Fatalf("controller LAN lookup ip = %q, want 192.0.2.50", got)
	}
	if doc.LANIP != "192.0.2.50" {
		t.Fatalf("lan_ip = %q, want 192.0.2.50", doc.LANIP)
	}
	if !doc.HasEth1 {
		t.Fatal("has_eth1 = false, want true")
	}
}

// TestGatewayPayloadReportsManagementVLANOnUplink verifies that management VLAN
// metadata is attached to the resolved gateway uplink.
func TestGatewayPayloadReportsManagementVLANOnUplink(t *testing.T) {
	profile, ok := device.LookupProfile("uxgpro")
	if !ok {
		t.Fatal("profile not found")
	}
	ports := device.BuildPorts(profile, device.PortBuildOptions{})
	payload, err := switchPayload(device.Identity{
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
	ports := device.BuildPorts(profile, device.PortBuildOptions{})
	payload, err := switchPayload(device.Identity{
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
		DeviceType       string           `json:"type"`
		NumPort          int              `json:"num_port"`
		Uplink           string           `json:"uplink"`
		IfTable          []map[string]any `json:"if_table"`
		NetworkTable     []map[string]any `json:"network_table"`
		ConfigPortTable  []map[string]any `json:"config_port_table"`
		EthernetTable    []map[string]any `json:"ethernet_table"`
		EthernetOverride []map[string]any `json:"ethernet_overrides"`
		PortTable        []map[string]any `json:"port_table"`
		ConfigNetworkWAN map[string]any   `json:"config_network_wan"`
		UplinkTable      []map[string]any `json:"uplink_table"`
		WAN1             map[string]any   `json:"wan1"`
	}
	if err := json.Unmarshal(payload, &doc); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"internet", "port_overrides"} {
		if _, ok := raw[key]; ok {
			t.Fatalf("gateway payload contains unexpected top-level key %q", key)
		}
	}
	if doc.DeviceType != "udm" {
		t.Fatalf("type = %q, want udm", doc.DeviceType)
	}
	if doc.NumPort != 7 {
		t.Fatalf("num_port = %d, want 7", doc.NumPort)
	}
	if len(doc.ConfigPortTable) != 7 {
		t.Fatalf("config_port_table length = %d, want 7", len(doc.ConfigPortTable))
	}
	if len(doc.EthernetTable) != 7 {
		t.Fatalf("ethernet_table length = %d, want 7", len(doc.EthernetTable))
	}
	if len(doc.EthernetOverride) != 5 {
		t.Fatalf("ethernet_overrides length = %d, want 5", len(doc.EthernetOverride))
	}
	if len(doc.PortTable) != 7 {
		t.Fatalf("port_table length = %d, want 7", len(doc.PortTable))
	}
	if len(doc.UplinkTable) != 1 {
		t.Fatalf("uplink_table length = %d, want 1", len(doc.UplinkTable))
	}
	if got := doc.ConfigNetworkWAN["type"].(string); got != "dhcp" {
		t.Fatalf("config_network_wan type = %q, want dhcp", got)
	}
	assertGatewayConfigNetworkWAN(t, doc.ConfigNetworkWAN, "192.0.2.2", "255.255.255.0")
	if got := doc.ConfigNetworkWAN["ifname"].(string); got != "eth5" {
		t.Fatalf("config_network_wan ifname = %q, want eth5", got)
	}
	if got := doc.WAN1["ifname"].(string); got != "eth5" {
		t.Fatalf("wan1 ifname = %q, want eth5", got)
	}
	if doc.Uplink != "eth5" {
		t.Fatalf("uplink = %q, want eth5", doc.Uplink)
	}
	if len(doc.NetworkTable) != 7 {
		t.Fatalf("network_table length = %d, want 7", len(doc.NetworkTable))
	}
}

// TestGatewayPortOverridesOnlyAffectObservedLinkFacts verifies that interface
// overrides can attach host observations without changing status-row IDs.
func TestGatewayPortOverridesOnlyAffectObservedLinkFacts(t *testing.T) {
	profile, ok := device.LookupProfile("uxgpro")
	if !ok {
		t.Fatal("profile not found")
	}
	ports := device.ApplyPortOverrides(device.BuildPorts(profile, device.PortBuildOptions{}), []device.PortOverride{
		{Port: 1, Name: "WAN uplink", Role: "wan", NetworkGroup: "WAN", Interface: "ixl0"},
		{Port: 2, Name: "LAN bridge", Role: "lan", NetworkGroup: "LAN", Interface: "vtnet0"},
		{Port: 3, Name: "backup_wan", Role: "wan2", NetworkGroup: "WAN2", Interface: "vlan09"},
		{Port: 4, Name: "unused_lab_lan", Role: "lan2", NetworkGroup: "LAN", Interface: "vlan10"},
	})
	payload, err := switchPayload(device.Identity{
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
}

// TestGatewayWANHealthHintsFollowPortOverrides verifies that WAN SLA-style
// telemetry stays deterministic and config-driven.
func TestGatewayWANHealthHintsFollowPortOverrides(t *testing.T) {
	profile, ok := device.LookupProfile("uxgpro")
	if !ok {
		t.Fatal("profile not found")
	}
	connected := true
	disconnected := false
	ports := device.ApplyPortOverrides(device.BuildPorts(profile, device.PortBuildOptions{}), []device.PortOverride{
		{
			Port:               1,
			Role:               "wan",
			NetworkGroup:       "WAN",
			WANUptimePercent:   float64Ref(99.5),
			WANLatencyMS:       7,
			WANDowntimeSeconds: 30,
			WANConnected:       &connected,
		},
		{
			Port:               3,
			Role:               "wan2",
			NetworkGroup:       "WAN2",
			WANUptimePercent:   float64Ref(0),
			WANLatencyMS:       0,
			WANDowntimeSeconds: 3600,
			WANConnected:       &disconnected,
		},
	})
	payload, err := switchPayload(device.Identity{
		MAC:           "02:11:22:33:44:67",
		IP:            "192.0.2.50",
		Hostname:      "unifi-stubd-uxg",
		Model:         profile.Model,
		ModelDisplay:  profile.ModelDisplay,
		DeviceType:    profile.DeviceType,
		Version:       profile.Version,
		Serial:        "021122334467",
		InformURL:     "http://192.0.2.10:8080/inform",
		UptimeSeconds: 3600,
	}, ports)
	if err != nil {
		t.Fatal(err)
	}

	var doc struct {
		UptimeStats  map[string]map[string]any `json:"uptime_stats"`
		IfTable      []map[string]any          `json:"if_table"`
		NetworkTable []map[string]any          `json:"network_table"`
		PortTable    []map[string]any          `json:"port_table"`
		Speedtest    map[string]any            `json:"speedtest-status"`
		Internet     map[string]any            `json:"internet_health"`
		LastWAN      map[string]string         `json:"last_wan_status"`
		LastWANIP    string                    `json:"last_wan_ip"`
		UplinkTable  []map[string]any          `json:"uplink_table"`
		WAN1         map[string]any            `json:"wan1"`
		WAN2         map[string]any            `json:"wan2"`
	}
	if err := json.Unmarshal(payload, &doc); err != nil {
		t.Fatal(err)
	}
	if got := doc.UptimeStats["WAN"]["uptime"].(float64); got != 99.5 {
		t.Fatalf("WAN uptime_stats uptime = %.1f, want 99.5", got)
	}
	if got := doc.UptimeStats["WAN"]["availability"].(float64); got != 99.5 {
		t.Fatalf("WAN availability = %.1f, want 99.5", got)
	}
	if got := int(doc.UptimeStats["WAN"]["latency"].(float64)); got != 7 {
		t.Fatalf("WAN latency = %d, want 7", got)
	}
	if got := int(doc.UptimeStats["WAN"]["downtime"].(float64)); got != 30 {
		t.Fatalf("WAN downtime = %d, want 30", got)
	}
	if got := doc.UptimeStats["WAN"]["isWanConnected"].(bool); !got {
		t.Fatal("WAN isWanConnected = false, want true")
	}
	if got := doc.UptimeStats["WAN"]["uplink_ifname"].(string); got != "eth0" {
		t.Fatalf("WAN uplink_ifname = %q, want eth0", got)
	}
	if got := int(doc.WAN1["latency"].(float64)); got != 7 {
		t.Fatalf("wan1 latency = %d, want 7", got)
	}
	if got := doc.WAN1["uplink_ifname"].(string); got != "eth0" {
		t.Fatalf("wan1 uplink_ifname = %q, want eth0", got)
	}
	for name, row := range map[string]map[string]any{
		"if_table":      rowByPortIndex(t, doc.IfTable, 1),
		"network_table": rowByPortIndex(t, doc.NetworkTable, 1),
		"port_table":    rowByPortIndex(t, doc.PortTable, 1),
	} {
		if got := int(row["latency"].(float64)); got != 7 {
			t.Fatalf("%s WAN latency = %d, want 7", name, got)
		}
		if got := row["isWanConnected"].(bool); !got {
			t.Fatalf("%s isWanConnected = false, want true", name)
		}
	}
	uplinkIf := rowByPortIndex(t, doc.IfTable, 1)
	if got := int(uplinkIf["uptime"].(float64)); got != 3600 {
		t.Fatalf("if_table uplink uptime = %d, want 3600", got)
	}
	if got := uplinkIf["speedtest_status"].(string); got != "Success" {
		t.Fatalf("if_table uplink speedtest_status = %q, want Success", got)
	}
	if got := int(uplinkIf["speedtest_ping"].(float64)); got != 7 {
		t.Fatalf("if_table uplink speedtest_ping = %d, want 7", got)
	}
	if got := int(uplinkIf["speedtest_lastrun"].(float64)); got <= 0 {
		t.Fatalf("if_table uplink speedtest_lastrun = %d, want > 0", got)
	}
	if got := int(doc.Speedtest["latency"].(float64)); got != 7 {
		t.Fatalf("speedtest-status latency = %d, want 7", got)
	}
	if got := int(doc.Speedtest["status_summary"].(float64)); got != 2 {
		t.Fatalf("speedtest-status status_summary = %d, want 2", got)
	}
	if got := doc.Internet["status"].(string); got != "ok" {
		t.Fatalf("internet_health status = %q, want ok", got)
	}
	if got := int(doc.Internet["latency"].(float64)); got != 7 {
		t.Fatalf("internet_health latency = %d, want 7", got)
	}
	if got := doc.LastWAN["WAN"]; got != "online" {
		t.Fatalf("last_wan_status WAN = %q, want online", got)
	}
	if got := doc.LastWANIP; got != "192.0.2.2" {
		t.Fatalf("last_wan_ip = %q, want 192.0.2.2", got)
	}
	if len(doc.UplinkTable) != 1 {
		t.Fatalf("uplink_table length = %d, want 1", len(doc.UplinkTable))
	}
	if got := int(doc.UplinkTable[0]["uptime"].(float64)); got != 3600 {
		t.Fatalf("uplink_table uptime = %d, want 3600", got)
	}
	if got := doc.UplinkTable[0]["speedtest_status"].(string); got != "Success" {
		t.Fatalf("uplink_table speedtest_status = %q, want Success", got)
	}
	if got := int(doc.UplinkTable[0]["speedtest_ping"].(float64)); got != 7 {
		t.Fatalf("uplink_table speedtest_ping = %d, want 7", got)
	}
	if got := int(doc.UplinkTable[0]["speedtest_lastrun"].(float64)); got <= 0 {
		t.Fatalf("uplink_table speedtest_lastrun = %d, want > 0", got)
	}
	if got := doc.UptimeStats["WAN2"]["isWanConnected"].(bool); got {
		t.Fatal("WAN2 isWanConnected = true, want false")
	}
	if got := int(doc.UptimeStats["WAN2"]["downtime"].(float64)); got != 3600 {
		t.Fatalf("WAN2 downtime = %d, want 3600", got)
	}
	if got := doc.WAN2["uplink_ifname"].(string); got != "eth2" {
		t.Fatalf("wan2 uplink_ifname = %q, want eth2", got)
	}
}

// TestWANHealthResultsBecomeHealthOnlyOverrides verifies active probe results
// cannot rewrite assignment, addressing, VLAN, or link-state fields.
func TestWANHealthResultsBecomeHealthOnlyOverrides(t *testing.T) {
	overrides := device.WANHealthOverrides([]device.WANHealthResult{
		{Port: 3, Connected: true, LatencyMS: 8, DowntimeSeconds: 0, UptimePercent: 100},
	})
	if len(overrides) != 1 {
		t.Fatalf("len(overrides) = %d, want 1", len(overrides))
	}
	override := overrides[0]
	if override.Port != 3 ||
		override.WANConnected == nil ||
		!*override.WANConnected ||
		override.WANLatencyMS != 8 ||
		override.WANDowntimeSeconds != 0 ||
		override.WANUptimePercent == nil ||
		*override.WANUptimePercent != 100 {
		t.Fatalf("health fields = %+v", override)
	}
	if override.Role != "" ||
		override.NetworkGroup != "" ||
		override.IP != "" ||
		override.VLAN != 0 ||
		override.PortConfID != "" ||
		override.NetworkConfID != "" ||
		override.Up != nil {
		t.Fatalf("non-health field was set: %+v", override)
	}
}

// TestGatewayPayloadReportsHostTableClientMetadata verifies MAC-table client
// metadata appears in gateway host-table output.
func TestGatewayPayloadReportsHostTableClientMetadata(t *testing.T) {
	profile, ok := device.LookupProfile("ugw3")
	if !ok {
		t.Fatal("profile not found")
	}
	ports := device.BuildPorts(profile, device.PortBuildOptions{})
	ports[1].MACs = []device.MacTableEntry{
		{
			MAC:      "02:00:5e:00:53:03",
			Hostname: "lab-host-2",
			IP:       "192.0.2.52",
			Age:      4,
			Uptime:   1200,
		},
	}
	payload, err := switchPayload(device.Identity{
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
		NetworkTable []map[string]any `json:"network_table"`
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
}

// TestGatewayLANPortReportsSwitchLikeNeighborWithoutLANHostTable verifies LAN
// gateway ports behave like physical switchports without turning downstream
// infrastructure into routed LAN clients.
func TestGatewayLANPortReportsSwitchLikeNeighborWithoutLANHostTable(t *testing.T) {
	profile, ok := device.LookupProfile("uxgpro")
	if !ok {
		t.Fatal("profile not found")
	}
	ports := device.BuildPorts(profile, device.PortBuildOptions{})
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
	payload, err := switchPayload(device.Identity{
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
		PortTable    []map[string]any `json:"port_table"`
	}
	if err := json.Unmarshal(payload, &doc); err != nil {
		t.Fatal(err)
	}
	if hosts, ok := doc.NetworkTable[3]["host_table"].([]any); ok && len(hosts) != 0 {
		t.Fatalf("LAN2 host_table exposes infrastructure device: %#v", hosts)
	}
	port := rowByPortIndex(t, doc.PortTable, 4)
	lastConnection, ok := port["last_connection"].(map[string]any)
	if !ok {
		t.Fatalf("LAN port last_connection = %#v, want switch-like neighbor", port["last_connection"])
	}
	if got := lastConnection["hostname"]; got != "management-downlink" {
		t.Fatalf("LAN port last_connection hostname = %v, want management-downlink", got)
	}
	if got := lastConnection["type"]; got != "usw" {
		t.Fatalf("LAN port last_connection type = %v, want usw", got)
	}
	macs, ok := port["mac_table"].([]any)
	if !ok || len(macs) != 1 {
		t.Fatalf("LAN port mac_table = %#v, want one infrastructure neighbor", port["mac_table"])
	}
	entry := macs[0].(map[string]any)
	if got := entry["hostname"]; got != "management-downlink" {
		t.Fatalf("LAN port mac_table hostname = %v, want management-downlink", got)
	}
	if got := entry["type"]; got != "usw" {
		t.Fatalf("LAN port mac_table type = %v, want usw", got)
	}
}

// TestSwitchPortsCanOverrideAggregationUplinkToTenGigPort verifies explicit
// uplink selection on grouped aggregation profiles.
func TestSwitchPortsCanOverrideAggregationUplinkToTenGigPort(t *testing.T) {
	profile, ok := device.LookupProfile("usaggpro")
	if !ok {
		t.Fatal("profile not found")
	}
	ports := device.BuildPorts(profile, device.PortBuildOptions{UplinkPort: 1})

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
	ports := device.ApplyPortOverrides(device.BuildPorts(profile, device.PortBuildOptions{}), []device.PortOverride{
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

// TestValidatePortOverrideAcceptsUnassignedRole keeps active gateway configs
// with disabled or unused ports loadable under strict YAML validation.
func TestValidatePortOverrideAcceptsUnassignedRole(t *testing.T) {
	err := device.ValidatePortOverride(device.PortOverride{
		Port:         1,
		Name:         "Port 1",
		Role:         "unassigned",
		NetworkGroup: "Unassigned",
	}, 4)
	if err != nil {
		t.Fatal(err)
	}
}

// TestGatewayPayloadReportsPortOverrideMACs verifies MAC/IP override data is
// reflected in gateway interface rows.
func TestGatewayPayloadReportsPortOverrideMACs(t *testing.T) {
	profile, ok := device.LookupProfile("ugw3")
	if !ok {
		t.Fatal("profile not found")
	}
	ports := device.ApplyPortOverrides(device.BuildPorts(profile, device.PortBuildOptions{Count: 2}), []device.PortOverride{
		{Port: 1, Name: "WAN", MAC: "02:00:5e:00:53:01", IP: "192.0.2.2", Netmask: "255.255.255.0"},
		{Port: 2, Name: "LAN", MAC: "02:00:5e:00:53:02", IP: "192.0.2.1", Netmask: "255.255.255.0"},
	})
	payload, err := switchPayload(device.Identity{
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
		IfTable          []map[string]any `json:"if_table"`
		NetworkTable     []map[string]any `json:"network_table"`
		ConfigPortTable  []map[string]any `json:"config_port_table"`
		EthernetOverride []map[string]any `json:"ethernet_overrides"`
		PortTable        []map[string]any `json:"port_table"`
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

// TestGatewayPayloadCanMirrorControllerPortAssignments verifies controller
// port assignment metadata is emitted only when an operator explicitly mirrors it.
func TestGatewayPayloadCanMirrorControllerPortAssignments(t *testing.T) {
	profile, ok := device.LookupProfile("uxgpro")
	if !ok {
		t.Fatal("profile not found")
	}
	ports := device.ApplyPortOverrides(device.BuildPorts(profile, device.PortBuildOptions{UplinkPort: 3}), []device.PortOverride{
		{
			Port:                3,
			Role:                "wan",
			NetworkGroup:        "WAN",
			PortConfID:          "portconf-real-wan",
			NetworkConfID:       "network-real-wan",
			NativeNetworkConfID: "network-real-wan",
			NetworkName:         "real_wan",
			VLAN:                3,
		},
	})
	payload, err := switchPayload(device.Identity{
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
		IfTable          []map[string]any `json:"if_table"`
		ConfigPortTable  []map[string]any `json:"config_port_table"`
		EthernetOverride []map[string]any `json:"ethernet_overrides"`
		PortTable        []map[string]any `json:"port_table"`
	}
	if err := json.Unmarshal(payload, &doc); err != nil {
		t.Fatal(err)
	}
	for _, table := range [][]map[string]any{doc.ConfigPortTable, doc.PortTable} {
		row := table[2]
		if got := row["portconf_id"].(string); got != "portconf-real-wan" {
			t.Fatalf("portconf_id = %q", got)
		}
		if got := row["networkconf_id"].(string); got != "network-real-wan" {
			t.Fatalf("networkconf_id = %q", got)
		}
		if got := row["native_networkconf_id"].(string); got != "network-real-wan" {
			t.Fatalf("native_networkconf_id = %q", got)
		}
		if got := row["network_name"].(string); got != "real_wan" {
			t.Fatalf("network_name = %q", got)
		}
		if got := int(row["vlan"].(float64)); got != 3 {
			t.Fatalf("vlan = %d, want 3", got)
		}
	}
}

// TestUXGGatewayPayloadUsesInterfaceOverrideData verifies gateway payloads use
// source-interface override data consistently.
func TestUXGGatewayPayloadUsesInterfaceOverrideData(t *testing.T) {
	profile, ok := device.LookupProfile("uxg-lite")
	if !ok {
		t.Fatal("profile not found")
	}
	ports := device.ApplyPortOverrides(device.BuildPorts(profile, device.PortBuildOptions{}), []device.PortOverride{
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
	payload, err := switchPayload(device.Identity{
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
	ports := device.ApplyPortOverrides(device.BuildPorts(profile, device.PortBuildOptions{}), []device.PortOverride{
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
	payload, err := switchPayload(device.Identity{
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
		Bytes       int64 `json:"bytes"`
		RXBytes     int64 `json:"rx_bytes"`
		TXBytes     int64 `json:"tx_bytes"`
		BytesRate   int64 `json:"bytes-r"`
		RXBytesRate int64 `json:"rx_bytes-r"`
		TXBytesRate int64 `json:"tx_bytes-r"`
		RXRate      int64 `json:"rx_rate"`
		TXRate      int64 `json:"tx_rate"`

		IfTable      []map[string]any `json:"if_table"`
		NetworkTable []map[string]any `json:"network_table"`
		UplinkTable  []map[string]any `json:"uplink_table"`
	}
	if err := json.Unmarshal(payload, &doc); err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err := json.Unmarshal(payload, &raw); err != nil {
		t.Fatal(err)
	}
	if _, ok := raw["stat"]; ok {
		t.Fatalf("gateway payload contains removed stat block: %#v", raw["stat"])
	}
	if doc.Bytes != 7000 {
		t.Fatalf("gateway root bytes = %d, want 7000", doc.Bytes)
	}
	if doc.RXBytes != 4000 || doc.TXBytes != 3000 {
		t.Fatalf("gateway root rx/tx = %d/%d, want 4000/3000", doc.RXBytes, doc.TXBytes)
	}
	if doc.BytesRate != 70 || doc.RXBytesRate != 40 || doc.TXBytesRate != 30 {
		t.Fatalf("gateway root byte rates = %d/%d/%d, want 70/40/30", doc.BytesRate, doc.RXBytesRate, doc.TXBytesRate)
	}
	if doc.RXRate != 320 || doc.TXRate != 240 {
		t.Fatalf("gateway root bit rates = %d/%d, want 320/240", doc.RXRate, doc.TXRate)
	}
	if got := int64(doc.IfTable[0]["rx_bytes-r"].(float64)); got != 10 {
		t.Fatalf("LAN if_table rx rate = %d, want 10", got)
	}
	if got := int64(doc.IfTable[0]["rx_rate"].(float64)); got != 80 {
		t.Fatalf("LAN if_table gateway rx rate = %d, want 80", got)
	}
	if got := int64(doc.IfTable[0]["bytes-r"].(float64)); got != 30 {
		t.Fatalf("LAN if_table byte rate = %d, want 30", got)
	}
	if got := int64(doc.IfTable[1]["tx_packets"].(float64)); got != 401 {
		t.Fatalf("WAN if_table tx_packets = %d, want 401", got)
	}
	wanStats := doc.NetworkTable[1]["stats"].(map[string]any)
	if got := int64(wanStats["tx_bytes-r"].(float64)); got != 40 {
		t.Fatalf("WAN network_table tx rate = %d, want 40", got)
	}
	if got := int64(wanStats["tx_rate"].(float64)); got != 320 {
		t.Fatalf("WAN network_table gateway tx rate = %d, want 320", got)
	}
	if got := int64(wanStats["bytes-r"].(float64)); got != 70 {
		t.Fatalf("WAN network_table byte rate = %d, want 70", got)
	}
	if got := int64(wanStats["rx_packets"].(float64)); got != 301 {
		t.Fatalf("WAN network_table rx_packets = %d, want 301", got)
	}
	if got := int64(doc.UplinkTable[0]["rx_bytes-r"].(float64)); got != 30 {
		t.Fatalf("uplink rx rate = %d, want 30", got)
	}
	if got := int64(doc.UplinkTable[0]["rx_rate"].(float64)); got != 240 {
		t.Fatalf("uplink gateway rx rate = %d, want 240", got)
	}
	if got := int64(doc.UplinkTable[0]["bytes-r"].(float64)); got != 70 {
		t.Fatalf("uplink byte rate = %d, want 70", got)
	}
	if got := int64(doc.UplinkTable[0]["tx_errors"].(float64)); got != 4 {
		t.Fatalf("uplink tx_errors = %d, want 4", got)
	}
	assertNoRateFields(t, doc.IfTable[0])
	assertNoRateFields(t, wanStats)
	assertNoRateFields(t, doc.UplinkTable[0])
}

// TestGatewayPayloadSynchronizesResolvedTables verifies gateway tables all
// consume the same resolved PortView data.
func TestGatewayPayloadSynchronizesResolvedTables(t *testing.T) {
	profile, ok := device.LookupProfile("uxg-lite")
	if !ok {
		t.Fatal("profile not found")
	}
	ports := device.ApplyPortOverrides(device.BuildPorts(profile, device.PortBuildOptions{}), []device.PortOverride{
		{
			Port:         1,
			Name:         "LAN",
			Interface:    "vtnet0",
			MAC:          "02:00:5e:00:53:11",
			IP:           "192.0.2.1",
			Netmask:      "255.255.255.0",
			IPv6:         []string{"2001:db8:10::1/64"},
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
			IPv6:         []string{"2001:db8:100::9/64", "2001:db8:100::2cf/128"},
			Role:         "wan",
			NetworkGroup: "WAN",
			Speed:        10000,
			Media:        "SFP+",
		},
	})
	payload, err := switchPayload(device.Identity{
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
		IfTable          []map[string]any `json:"if_table"`
		NetworkTable     []map[string]any `json:"network_table"`
		ConfigPortTable  []map[string]any `json:"config_port_table"`
		EthernetOverride []map[string]any `json:"ethernet_overrides"`
		InternetHealth   map[string]any   `json:"internet_health"`
		PortTable        []map[string]any `json:"port_table"`
		ReportedNetworks []map[string]any `json:"reported_networks"`
		WAN1             map[string]any   `json:"wan1"`
		Wans             []map[string]any `json:"wans"`
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
			{"if_table", doc.IfTable[index-1]},
			{"network_table", doc.NetworkTable[index-1]},
			{"config_port_table", doc.ConfigPortTable[index-1]},
			{"port_table", doc.PortTable[index-1]},
		}
		for _, item := range rows {
			if got := item.row["ifname"].(string); got != ifname {
				t.Fatalf("%s port %d ifname = %q, want %q", item.name, index, got, ifname)
			}
			if got := item.row["source_interface"].(string); got != sourceInterface {
				t.Fatalf("%s port %d source_interface = %q, want %q", item.name, index, got, sourceInterface)
			}
			if got := item.row["networkgroup"].(string); got != networkGroup {
				t.Fatalf("%s port %d networkgroup = %q, want %q", item.name, index, got, networkGroup)
			}
		}
		for _, item := range []struct {
			name string
			row  map[string]any
		}{
			{"if_table", doc.IfTable[index-1]},
			{"network_table", doc.NetworkTable[index-1]},
			{"port_table", doc.PortTable[index-1]},
		} {
			got, ok := item.row["mac"].(string)
			if !ok {
				continue
			}
			if got != mac {
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
	assertStringSlice(t, doc.NetworkTable[0], "addresses", []string{"192.0.2.1/24", "2001:db8:10::1/64"})
	assertStringSlice(t, doc.NetworkTable[1], "addresses", []string{
		"198.51.100.9/24",
		"2001:db8:100::9/64",
		"2001:db8:100::2cf/128",
	})
	assertStringSlice(t, doc.IfTable[1], "ipv6", []string{"2001:db8:100::9/64", "2001:db8:100::2cf/128"})
	assertStringSlice(t, doc.ReportedNetworks[1], "addresses", []string{
		"198.51.100.9/24",
		"2001:db8:100::9/64",
		"2001:db8:100::2cf/128",
	})
	assertStringSlice(t, doc.WAN1, "ipv6", []string{"2001:db8:100::9/64", "2001:db8:100::2cf/128"})
	assertStringSlice(t, doc.Wans[0], "ipv6", []string{"2001:db8:100::9/64", "2001:db8:100::2cf/128"})
	assertStringSlice(t, doc.InternetHealth, "ipv6", []string{"2001:db8:100::9/64", "2001:db8:100::2cf/128"})
}

func assertStringSlice(t *testing.T, row map[string]any, key string, want []string) {
	t.Helper()
	values, ok := row[key].([]any)
	if !ok {
		t.Fatalf("%s = %#v, want string slice", key, row[key])
	}
	if len(values) != len(want) {
		t.Fatalf("%s length = %d, want %d (%#v)", key, len(values), len(want), values)
	}
	for index, value := range values {
		got, ok := value.(string)
		if !ok {
			t.Fatalf("%s[%d] = %#v, want string", key, index, value)
		}
		if got != want[index] {
			t.Fatalf("%s[%d] = %q, want %q", key, index, got, want[index])
		}
	}
}

// TestGatewayRoleRemapKeepsProfilePortInterface verifies role overrides keep
// the physical profile interface and only change the controller-visible role.
func TestGatewayRoleRemapKeepsProfilePortInterface(t *testing.T) {
	profile, ok := device.LookupProfile("uxgpro")
	if !ok {
		t.Fatal("profile not found")
	}
	ports := device.ApplyPortOverrides(device.BuildPorts(profile, device.PortBuildOptions{}), []device.PortOverride{
		{Port: 4, Role: "lan", NetworkGroup: "LAN", Interface: "vtnet0"},
	})
	payload, err := switchPayload(device.Identity{
		MAC:          "02:00:5e:00:53:20",
		IP:           "192.0.2.1",
		Hostname:     "opnsense",
		Model:        profile.Model,
		ModelDisplay: profile.ModelDisplay,
		DeviceType:   profile.DeviceType,
		Version:      profile.Version,
		Serial:       "02005E005320",
		InformURL:    "http://192.0.2.10:8080/inform",
	}, ports)
	if err != nil {
		t.Fatal(err)
	}

	var doc struct {
		IfTable      []map[string]any `json:"if_table"`
		NetworkTable []map[string]any `json:"network_table"`
		PortTable    []map[string]any `json:"port_table"`
	}
	if err := json.Unmarshal(payload, &doc); err != nil {
		t.Fatal(err)
	}
	ifRow := rowByPortIndex(t, doc.IfTable, 4)
	if got := ifRow["ifname"].(string); got != "eth3" {
		t.Fatalf("port 4 LAN if_table ifname = %q, want eth3", got)
	}
	if got := ifRow["source_interface"].(string); got != "vtnet0" {
		t.Fatalf("port 4 if_table source_interface = %q, want vtnet0", got)
	}
	networkRow := rowByPortIndex(t, doc.NetworkTable, 4)
	if got := networkRow["ifname"].(string); got != "eth3" {
		t.Fatalf("port 4 LAN network_table ifname = %q, want eth3", got)
	}
	for _, row := range []map[string]any{networkRow} {
		if got := row["source_interface"].(string); got != "vtnet0" {
			t.Fatalf("port 4 source_interface = %q, want vtnet0", got)
		}
	}
	portRow := doc.PortTable[3]
	if got := portRow["ifname"].(string); got != "eth3" {
		t.Fatalf("port_table port 4 ifname = %q, want eth3", got)
	}
	if got := portRow["source_interface"].(string); got != "vtnet0" {
		t.Fatalf("port_table port 4 source_interface = %q, want vtnet0", got)
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
	ports := device.BuildPorts(profile, device.PortBuildOptions{})
	payload, err := buildPayload(profile, device.Identity{
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
	ports := device.ApplyUplinkNeighbor(device.BuildPorts(profile, device.PortBuildOptions{UplinkPort: 1}), &device.MacTableEntry{
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

// TestSwitchPayloadSkipsGatewayUplinkNeighborMacTable avoids making the
// upstream gateway look like a downstream station on the switch uplink.
func TestSwitchPayloadSkipsGatewayUplinkNeighborMacTable(t *testing.T) {
	profile, ok := device.LookupProfile("us48p500")
	if !ok {
		t.Fatal("profile not found")
	}
	ports := device.BuildPorts(profile, device.PortBuildOptions{UplinkPort: 49})
	ports[48].MACs = nil
	ports = device.ApplyUplinkNeighbor(ports, &device.MacTableEntry{
		MAC:  "02:aa:bb:cc:dd:01",
		IP:   "192.0.2.1",
		Type: "uxg",
	})

	if len(ports[48].MACs) != 1 {
		t.Fatalf("gateway neighbor was not retained in port data: %+v", ports[48].MACs)
	}
	payload, err := switchPayload(device.Identity{
		MAC:          "02:11:22:33:44:60",
		IP:           "192.0.2.50",
		Hostname:     "server-lan1",
		Model:        profile.Model,
		ModelDisplay: profile.ModelDisplay,
		Version:      profile.Version,
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
	port := rowByPortIndex(t, doc.PortTable, 49)
	if macs, ok := port["mac_table"].([]any); ok && len(macs) != 0 {
		t.Fatalf("gateway uplink mac_table = %#v, want empty", macs)
	}
	if port["last_connection"] != nil {
		t.Fatalf("gateway uplink last_connection = %#v, want null", port["last_connection"])
	}
}

// TestApplyPortNeighborsAddsConfiguredMacTableEntry verifies configured
// per-port neighbor metadata is added to the target port.
func TestApplyPortNeighborsAddsConfiguredMacTableEntry(t *testing.T) {
	profile, ok := device.LookupProfile("ugw3")
	if !ok {
		t.Fatal("profile not found")
	}
	ports := device.ApplyPortNeighbors(device.BuildPorts(profile, device.PortBuildOptions{}), []device.PortNeighbor{
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
	payload, err := switchPayload(device.Identity{
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

// TestSwitchPayloadReportsPortOverrideLinkDown verifies link-down
// overrides clear live link fields in switch payloads.
func TestSwitchPayloadReportsPortOverrideLinkDown(t *testing.T) {
	linkDown := false
	payload, err := switchPayload(device.Identity{
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

// TestSwitchPayloadReportsDisabledPort verifies disabled ports render as
// down, zero-speed, and without learned MACs.
func TestSwitchPayloadReportsDisabledPort(t *testing.T) {
	payload, err := switchPayload(device.Identity{
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

// TestSwitchPayloadReportsGroupedUplinkSpeed verifies grouped profile
// uplink speed reaches switch payload rows.
func TestSwitchPayloadReportsGroupedUplinkSpeed(t *testing.T) {
	profile, ok := device.LookupProfile("usw-pro-xg-48")
	if !ok {
		t.Fatal("profile not found")
	}
	payload, err := switchPayload(device.Identity{
		MAC:          "02:11:22:33:44:58",
		IP:           "192.0.2.50",
		Hostname:     "unifi-stubd-lab",
		Model:        profile.Model,
		ModelDisplay: profile.ModelDisplay,
		Version:      profile.Version,
		Serial:       "021122334458",
		InformURL:    "http://192.0.2.10:8080/inform",
	}, device.BuildPorts(profile, device.PortBuildOptions{}))
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

// TestSwitchPayloadReportsObservedCounters verifies observed interface
// counters are copied into switch payload rows.
func TestSwitchPayloadReportsObservedCounters(t *testing.T) {
	payload, err := switchPayload(device.Identity{
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

// TestSwitchPayloadPrefersExplicitTrafficRates verifies explicit traffic
// rates take precedence over synthetic heartbeat rates.
func TestSwitchPayloadPrefersExplicitTrafficRates(t *testing.T) {
	payload, err := switchPayload(device.Identity{
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
	if got := int64(port["bytes-r"].(float64)); got != 579 {
		t.Fatalf("bytes-r = %d, want 579", got)
	}
}

// TestSwitchPayloadSuppressesSyntheticRatesWhenTrafficRatesEnabledWithoutSource
// verifies rate fields stay zero when tracking is enabled without a source.
func TestSwitchPayloadSuppressesSyntheticRatesWhenTrafficRatesEnabledWithoutSource(t *testing.T) {
	payload, err := switchPayload(device.Identity{
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
	if got := int64(port["bytes-r"].(float64)); got != 0 {
		t.Fatalf("bytes-r = %d, want 0", got)
	}
	assertNoRateFields(t, port)
}

// assertNoRateFields verifies payload rows avoid legacy
// experimental rate keys.
func assertNoRateFields(t *testing.T, row map[string]any) {
	t.Helper()
	for _, key := range []string{"rx_packets-r", "tx_packets-r", "rx_errors-r", "tx_errors-r"} {
		if _, ok := row[key]; ok {
			t.Fatalf("unexpected experimental rate field %q in %#v", key, row)
		}
	}
}

// assertGatewayConfigNetworkWAN checks the UXG-style WAN config block. Port
// anchoring is reported through uplink, if_table, and port tables.
func assertGatewayConfigNetworkWAN(t *testing.T, row map[string]any, ip, netmask string) {
	t.Helper()
	if got := row["ip"].(string); got != ip {
		t.Fatalf("config_network_wan ip = %q, want %q", got, ip)
	}
	if got := row["netmask"].(string); got != netmask {
		t.Fatalf("config_network_wan netmask = %q, want %q", got, netmask)
	}
	if got := row["speed"].(string); got != "auto" {
		t.Fatalf("config_network_wan speed = %q, want auto", got)
	}
	if got := row["autoneg"].(bool); !got {
		t.Fatal("config_network_wan autoneg = false, want true")
	}
	if got := row["full_duplex"].(bool); !got {
		t.Fatal("config_network_wan full_duplex = false, want true")
	}
}

func rowByPortIndex(t *testing.T, rows []map[string]any, portIndex int) map[string]any {
	t.Helper()
	for _, row := range rows {
		if got, ok := row["port_idx"].(float64); ok && int(got) == portIndex {
			return row
		}
	}
	t.Fatalf("no row with port_idx %d in %#v", portIndex, rows)
	return nil
}

func firstRowByString(t *testing.T, rows []map[string]any, key, value string) map[string]any {
	t.Helper()
	for _, row := range rows {
		if got, ok := row[key].(string); ok && got == value {
			return row
		}
	}
	t.Fatalf("no row with %s=%q in %#v", key, value, rows)
	return nil
}

func float64Ref(value float64) *float64 {
	return &value
}

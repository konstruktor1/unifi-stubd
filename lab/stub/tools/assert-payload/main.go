// Command assert-payload verifies selected dry-run payload invariants.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
)

const payloadMarker = "minimal_inform_payload_json:\n"

const (
	fieldHostname            = "hostname"
	fieldIfname              = "ifname"
	fieldNativeNetworkconfID = "native_networkconf_id"
	fieldNetworkconfID       = "networkconf_id"
	fieldNetworkName         = "network_name"
	fieldPortconfID          = "portconf_id"
	fieldSourceInterface     = "source_interface"
	fieldVLAN                = "vlan"
	tableConfigPort          = "config_port_table"
	tableEthernetOverrides   = "ethernet_overrides"
	tableEthernet            = "ethernet_table"
	tableIf                  = "if_table"
	tableNetwork             = "network_table"
	tablePort                = "port_table"
	tableReportedNetworks    = "reported_networks"
	tableUplink              = "uplink_table"
	controllerEth2           = "eth2"
	controllerEth3           = "eth3"
	sourceIXL0               = "ixl0"
	sourceVTNet0             = "vtnet0"
)

var bridgeClients = map[int]map[string]string{
	2: {
		"mac":         "02:00:5e:10:01:01",
		fieldHostname: "lab-client-101",
		"ip":          "192.0.2.101",
	},
	3: {
		"mac":         "02:00:5e:10:02:01",
		fieldHostname: "lab-client-102",
		"ip":          "192.0.2.102",
	},
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "payload assertion failed: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: assert-payload bridge-observe|port-map|gateway-smoke|management-lan|opnsense-uxg <dry-run-output>")
		os.Exit(2)
	}
	payload, err := extractPayload(args[1])
	if err != nil {
		return err
	}
	switch args[0] {
	case "bridge-observe":
		return assertBridgeObserve(payload)
	case "port-map":
		return assertPortMap(payload)
	case "gateway-smoke":
		return assertGatewaySmoke(payload)
	case "management-lan":
		return assertManagementLAN(payload)
	case "opnsense-uxg":
		return assertOPNsenseUXG(payload)
	default:
		fmt.Fprintf(os.Stderr, "unknown assertion mode %q\n", args[0])
		os.Exit(2)
	}
	return nil
}

func extractPayload(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read dry-run output: %w", err)
	}
	parts := strings.SplitN(string(data), payloadMarker, 2)
	if len(parts) != 2 {
		return nil, errors.New("dry-run output does not contain inform payload marker")
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(parts[1]), &payload); err != nil {
		return nil, fmt.Errorf("decode dry-run payload: %w", err)
	}
	return payload, nil
}

func assertBridgeObserve(payload map[string]any) error {
	if payload[fieldHostname] != "stub-bridge-observe" {
		return fmt.Errorf("hostname = %q, want stub-bridge-observe", payload[fieldHostname])
	}
	ports, err := portsByIndex(payload)
	if err != nil {
		return err
	}
	for index, expected := range bridgeClients {
		entries := map[string]map[string]any{}
		for _, entry := range list(ports[index]["mac_table"]) {
			if obj, ok := entry.(map[string]any); ok {
				entries[stringValue(obj["mac"])] = obj
			}
		}
		client := entries[expected["mac"]]
		if client == nil {
			return fmt.Errorf("port %d mac_table %v does not contain %s", index, mapKeys(entries), expected["mac"])
		}
		for _, field := range []string{fieldHostname, "ip"} {
			if client[field] != expected[field] {
				return fmt.Errorf("port %d client %s = %q, want %q", index, field, client[field], expected[field])
			}
		}
	}
	if ports[1]["is_uplink"] != true {
		return errors.New("port 1 should remain the bridge-observe uplink")
	}
	return nil
}

func assertManagementLAN(payload map[string]any) error {
	if err := assertBridgeObserve(payload); err != nil {
		return err
	}
	expectedIP := os.Getenv("UNIFI_STUB_BRIDGE_IP")
	if payload["ip"] != expectedIP {
		return fmt.Errorf("payload ip = %q, want %q", payload["ip"], expectedIP)
	}
	if intValue(payload["management_vlan"]) != 42 {
		return fmt.Errorf("management_vlan = %v, want 42", payload["management_vlan"])
	}
	ifTable := list(payload["if_table"])
	if len(ifTable) == 0 {
		return errors.New("payload has no if_table")
	}
	management, ok := ifTable[0].(map[string]any)
	if !ok {
		return fmt.Errorf("if_table[0] is not an object: %v", ifTable[0])
	}
	if intValue(management["management_vlan"]) != 42 {
		return fmt.Errorf("if_table management_vlan = %v, want 42", management["management_vlan"])
	}
	if intValue(management["vlan"]) != 42 {
		return fmt.Errorf("if_table vlan = %v, want 42", management["vlan"])
	}
	return nil
}

func assertPortMap(payload map[string]any) error {
	if payload[fieldHostname] != "stub-port-map" {
		return fmt.Errorf("hostname = %q, want stub-port-map", payload[fieldHostname])
	}
	ports, err := portsByIndex(payload)
	if err != nil {
		return err
	}
	if ports[1][fieldSourceInterface] != "pmeth1" {
		return fmt.Errorf("port 1 source_interface = %q", ports[1][fieldSourceInterface])
	}
	if ports[2][fieldSourceInterface] != "pmeth2" {
		return fmt.Errorf("port 2 source_interface = %q", ports[2][fieldSourceInterface])
	}
	if ports[3]["up"] != false || intValue(ports[3]["speed"]) != 0 {
		return fmt.Errorf("port 3 should be disabled, got %v", ports[3])
	}
	if stringValue(ports[4][fieldSourceInterface]) != "" {
		return fmt.Errorf("port 4 should be unmapped, got %v", ports[4])
	}
	if ports[4]["up"] != true || intValue(ports[4]["speed"]) == 0 {
		return fmt.Errorf("port 4 should keep profile defaults, got %v", ports[4])
	}
	return nil
}

func assertGatewaySmoke(payload map[string]any) error {
	if payload[fieldHostname] != "stub-gateway-smoke" {
		return fmt.Errorf("hostname = %q, want stub-gateway-smoke", payload[fieldHostname])
	}
	if payload["type"] != "uxg" {
		return fmt.Errorf("gateway type = %q, want uxg", payload["type"])
	}
	if payload["model"] != "UXG" {
		return fmt.Errorf("gateway model = %q, want UXG", payload["model"])
	}
	if _, ok := payload["gw_caps"].(map[string]any); !ok {
		return fmt.Errorf("gateway gw_caps = %v, want object", payload["gw_caps"])
	}
	for _, tableName := range []string{tableEthernet, tableIf, tableNetwork, tableConfigPort, tableEthernetOverrides, tablePort, tableReportedNetworks, tableUplink} {
		if len(list(payload[tableName])) == 0 {
			return fmt.Errorf("gateway table %s is missing or empty", tableName)
		}
	}
	for _, tableName := range []string{"internet", "port_overrides", "wan2"} {
		if _, ok := payload[tableName]; ok {
			return fmt.Errorf("gateway payload should not define switch-style %s", tableName)
		}
	}
	if payload["outlet_enabled"] != false {
		return fmt.Errorf("gateway outlet_enabled = %v, want false", payload["outlet_enabled"])
	}
	for _, tableName := range []string{"outlet_table", "outlet_overrides"} {
		if len(list(payload[tableName])) != 0 {
			return fmt.Errorf("gateway %s = %v, want empty list", tableName, payload[tableName])
		}
	}
	for _, tableName := range []string{tableIf, tableNetwork, tableConfigPort, tablePort, tableReportedNetworks} {
		table, err := rowsByIndex(payload, tableName, 2)
		if err != nil {
			return err
		}
		if err := assertGatewayRow(tableName, table[1], "eth0", "pmeth1"); err != nil {
			return err
		}
		if err := assertGatewayRow(tableName, table[2], "eth1", "pmeth2"); err != nil {
			return err
		}
	}
	for _, tableName := range []string{tableConfigPort, tablePort} {
		table, err := rowsByIndex(payload, tableName, 2)
		if err != nil {
			return err
		}
		if err := assertGatewayAssignment(tableName, table[2], map[string]any{
			fieldPortconfID:          "portconf-gateway-wan",
			fieldNetworkconfID:       "network-gateway-wan",
			fieldNativeNetworkconfID: "network-gateway-wan",
			fieldNetworkName:         "gateway_wan",
			fieldVLAN:                3,
		}); err != nil {
			return err
		}
	}
	configNetworkWAN, ok := payload["config_network_wan"].(map[string]any)
	if !ok {
		return errors.New("gateway payload has no config_network_wan")
	}
	if configNetworkWAN["type"] != "dhcp" {
		return fmt.Errorf("gateway config_network_wan = %v, want DHCP", configNetworkWAN)
	}
	if err := assertGatewayWANConfig("config_network_wan", configNetworkWAN); err != nil {
		return err
	}
	wan1, ok := payload["wan1"].(map[string]any)
	if !ok {
		return errors.New("gateway payload has no wan1 status row")
	}
	if err := assertGatewayRow("wan1", wan1, "eth1", "pmeth2"); err != nil {
		return err
	}
	if intValue(payload["uptime"]) < 1 {
		return fmt.Errorf("gateway uptime = %v", payload["uptime"])
	}
	if intValue(payload["time_ms"]) <= 0 {
		return fmt.Errorf("gateway time_ms = %v", payload["time_ms"])
	}
	return nil
}

func assertOPNsenseUXG(payload map[string]any) error {
	hostname := stringValue(payload[fieldHostname])
	if !strings.HasPrefix(hostname, "opnsense-uxg") {
		return fmt.Errorf("opnsense gateway hostname = %q, want opnsense-uxg*", hostname)
	}
	if payload["type"] != "uxg" {
		return fmt.Errorf("opnsense gateway type = %q, want uxg", payload["type"])
	}
	if payload["model"] != "UXGPRO" {
		return fmt.Errorf("opnsense gateway model = %q, want UXGPRO", payload["model"])
	}
	if intValue(payload["num_port"]) != 4 {
		return fmt.Errorf("opnsense gateway num_port = %v, want 4", payload["num_port"])
	}
	for _, tableName := range []string{tableEthernet, tableIf, tableNetwork, tableConfigPort, tableEthernetOverrides, tablePort, tableReportedNetworks, tableUplink} {
		if len(list(payload[tableName])) == 0 {
			return fmt.Errorf("opnsense gateway table %s is missing or empty", tableName)
		}
	}
	for _, tableName := range []string{"internet", "port_overrides"} {
		if _, ok := payload[tableName]; ok {
			return fmt.Errorf("opnsense gateway payload should not define switch-style %s", tableName)
		}
	}
	rows, err := rowsByPortIdx(payload, tableIf)
	if err != nil {
		return err
	}
	if err := assertGatewayRow(tableIf, rows[3], controllerEth2, sourceIXL0); err != nil {
		return err
	}
	if err := assertGatewayRow(tableIf, rows[4], controllerEth3, sourceVTNet0); err != nil {
		return err
	}
	for _, tableName := range []string{tableNetwork, tableReportedNetworks} {
		rows, err := rowsByPortIdx(payload, tableName)
		if err != nil {
			return err
		}
		if err := assertGatewayRow(tableName, rows[3], controllerEth2, sourceIXL0); err != nil {
			return err
		}
		if err := assertGatewayRow(tableName, rows[4], controllerEth3, sourceVTNet0); err != nil {
			return err
		}
	}
	for _, tableName := range []string{tableEthernet, tableConfigPort, tablePort} {
		table, err := rowsByIndex(payload, tableName, 4)
		if err != nil {
			return err
		}
		if table[3][fieldIfname] != controllerEth2 {
			return fmt.Errorf("%s port 3 ifname = %q, want eth2", tableName, table[3][fieldIfname])
		}
		if table[4][fieldIfname] != controllerEth3 {
			return fmt.Errorf("%s port 4 ifname = %q, want eth3", tableName, table[4][fieldIfname])
		}
		if tableName == tableConfigPort || tableName == tableEthernetOverrides || tableName == tablePort {
			if table[3][fieldSourceInterface] != sourceIXL0 {
				return fmt.Errorf("%s port 3 source_interface = %q, want ixl0", tableName, table[3][fieldSourceInterface])
			}
			if table[4][fieldSourceInterface] != sourceVTNet0 {
				return fmt.Errorf("%s port 4 source_interface = %q, want vtnet0", tableName, table[4][fieldSourceInterface])
			}
		}
	}
	portTable, err := rowsByIndex(payload, tablePort, 4)
	if err != nil {
		return err
	}
	if err := assertGatewaySwitchportNeighbor("port_table", portTable[4], "server-lan1"); err != nil {
		return err
	}
	ethernetOverrides, err := rowsByPortIdx(payload, tableEthernetOverrides)
	if err != nil {
		return err
	}
	expectedOverrideGroups := map[int][]string{
		1: {"eth0", "Unassigned", ""},
		2: {"eth1", "Unassigned", ""},
		3: {controllerEth2, "WAN", sourceIXL0},
		4: {controllerEth3, "LAN", sourceVTNet0},
	}
	if !sameIntKeys(ethernetOverrides, expectedOverrideGroups) {
		return fmt.Errorf("ethernet_overrides ports = %v, want [1 2 3 4]", sortedIntKeys(ethernetOverrides))
	}
	for index, expected := range expectedOverrideGroups {
		row := ethernetOverrides[index]
		if row[fieldIfname] != expected[0] {
			return fmt.Errorf("ethernet_overrides port %d ifname = %q, want %s", index, row[fieldIfname], expected[0])
		}
		if row["networkgroup"] != expected[1] {
			return fmt.Errorf("ethernet_overrides port %d networkgroup = %q, want %s", index, row["networkgroup"], expected[1])
		}
		if expected[2] != "" && row[fieldSourceInterface] != expected[2] {
			return fmt.Errorf("ethernet_overrides port %d source_interface = %q, want %s", index, row[fieldSourceInterface], expected[2])
		}
		if (index == 1 || index == 2) && row["disabled"] != true {
			return fmt.Errorf("ethernet_overrides port %d disabled = %v, want true", index, row["disabled"])
		}
		if (index == 3 || index == 4) && hasKey(row, "disabled") {
			return fmt.Errorf("ethernet_overrides port %d should not be disabled: %v", index, row)
		}
	}
	configNetworkLAN, ok := payload["config_network_lan"].(map[string]any)
	if !ok {
		return errors.New("opnsense gateway payload has no config_network_lan")
	}
	lanIP := stringValue(payload["lan_ip"])
	if lanIP == "" {
		return fmt.Errorf("opnsense lan_ip = %q, want populated LAN IP", payload["lan_ip"])
	}
	if configNetworkLAN["cidr"] != lanIP+"/24" {
		return fmt.Errorf("config_network_lan cidr = %q, want %s/24", configNetworkLAN["cidr"], lanIP)
	}
	if configNetworkLAN[fieldIfname] != controllerEth3 {
		return fmt.Errorf("config_network_lan ifname = %q, want eth3", configNetworkLAN[fieldIfname])
	}
	if intValue(configNetworkLAN["port_idx"]) != 4 {
		return fmt.Errorf("config_network_lan port_idx = %v, want 4", configNetworkLAN["port_idx"])
	}
	if payload["has_eth1"] != true {
		return fmt.Errorf("opnsense has_eth1 = %v, want true", payload["has_eth1"])
	}
	assignments := map[int]map[string]any{
		3: {
			fieldPortconfID:          "portconf-opnsense-wan",
			fieldNetworkconfID:       "network-opnsense-wan",
			fieldNativeNetworkconfID: "network-opnsense-wan",
			fieldNetworkName:         "opnsense_wan",
			fieldVLAN:                3,
		},
		4: {
			fieldPortconfID:          "portconf-opnsense-lan",
			fieldNetworkconfID:       "network-opnsense-lan",
			fieldNativeNetworkconfID: "network-opnsense-lan",
			fieldNetworkName:         "opnsense_lan",
			fieldVLAN:                1,
		},
	}
	for _, tableName := range []string{tableConfigPort, tablePort} {
		table, err := rowsByIndex(payload, tableName, 4)
		if err != nil {
			return err
		}
		for port, expected := range assignments {
			if hostname == "opnsense-uxg-sfp-lab" {
				if err := assertGatewayAssignment(tableName, table[port], expected); err != nil {
					return err
				}
			}
		}
	}
	configNetworkWAN, ok := payload["config_network_wan"].(map[string]any)
	if !ok {
		return errors.New("opnsense gateway payload has no config_network_wan")
	}
	if err := assertGatewayWANConfig("config_network_wan", configNetworkWAN); err != nil {
		return err
	}
	if configNetworkWAN[fieldIfname] != controllerEth2 {
		return fmt.Errorf("config_network_wan ifname = %q, want eth2", configNetworkWAN[fieldIfname])
	}
	wan1, ok := payload["wan1"].(map[string]any)
	if !ok {
		return errors.New("opnsense gateway payload has no wan1 status row")
	}
	if err := assertGatewayRow("wan1", wan1, controllerEth2, sourceIXL0); err != nil {
		return err
	}
	wans := list(payload["wans"])
	if len(wans) == 0 {
		return errors.New("opnsense gateway payload has no wans inventory")
	}
	wan, ok := wans[0].(map[string]any)
	if !ok || wan["interface"] != "eth2" || intValue(wan["port"]) != 3 {
		return fmt.Errorf("opnsense wans[0] = %v, want interface eth2 port 3", wans[0])
	}
	if payload["uplink"] != controllerEth2 {
		return fmt.Errorf("opnsense uplink = %q, want eth2", payload["uplink"])
	}
	return assertNoHostIfnames(payload, map[string]bool{sourceIXL0: true, sourceVTNet0: true, "igb0": true})
}

func assertGatewayRow(tableName string, row map[string]any, ifname, source string) error {
	if row == nil {
		return fmt.Errorf("%s row is missing", tableName)
	}
	if row[fieldIfname] != ifname {
		return fmt.Errorf("%s ifname = %q, want %s", tableName, row[fieldIfname], ifname)
	}
	if row[fieldSourceInterface] != source {
		return fmt.Errorf("%s source_interface = %q, want %s", tableName, row[fieldSourceInterface], source)
	}
	return nil
}

func assertGatewayWANConfig(tableName string, row map[string]any) error {
	if row["type"] != "dhcp" {
		return fmt.Errorf("%s type = %q, want dhcp", tableName, row["type"])
	}
	if stringValue(row["ip"]) == "" || stringValue(row["netmask"]) == "" {
		return fmt.Errorf("%s has no ip/netmask: %v", tableName, row)
	}
	if row["speed"] != "auto" {
		return fmt.Errorf("%s speed = %q, want auto", tableName, row["speed"])
	}
	if row["autoneg"] != true || row["full_duplex"] != true {
		return fmt.Errorf("%s autoneg/full_duplex invalid: %v", tableName, row)
	}
	if ifname := stringValue(row[fieldIfname]); ifname != "" && !allowedControllerIfname(ifname) {
		return fmt.Errorf("%s ifname = %q, want controller ethN", tableName, ifname)
	}
	if _, ok := row[fieldSourceInterface]; ok {
		return fmt.Errorf("%s must not expose host source interface, got %v", tableName, row)
	}
	return nil
}

func assertGatewayAssignment(tableName string, row map[string]any, expected map[string]any) error {
	for key, want := range expected {
		got := row[key]
		if key == fieldVLAN {
			if intValue(got) != intValue(want) {
				return fmt.Errorf("%s %s = %v, want %v", tableName, key, got, want)
			}
			continue
		}
		if got != want {
			return fmt.Errorf("%s %s = %q, want %q", tableName, key, got, want)
		}
	}
	return nil
}

func assertGatewaySwitchportNeighbor(tableName string, row map[string]any, hostname string) error {
	lastConnection, ok := row["last_connection"].(map[string]any)
	if !ok {
		return fmt.Errorf("%s last_connection = %v, want %s", tableName, row["last_connection"], hostname)
	}
	if lastConnection["hostname"] != hostname {
		return fmt.Errorf("%s last_connection hostname = %q, want %s", tableName, lastConnection["hostname"], hostname)
	}
	if lastConnection["type"] != "usw" {
		return fmt.Errorf("%s last_connection type = %q, want usw", tableName, lastConnection["type"])
	}
	macTable := list(row["mac_table"])
	if len(macTable) == 0 {
		return fmt.Errorf("%s mac_table = %v, want switchport neighbor", tableName, row["mac_table"])
	}
	for _, entry := range macTable {
		obj, ok := entry.(map[string]any)
		if ok && obj["hostname"] == hostname && obj["type"] == "usw" {
			return nil
		}
	}
	return fmt.Errorf("%s mac_table lacks %s: %v", tableName, hostname, macTable)
}

func assertNoHostIfnames(payload map[string]any, blocked map[string]bool) error {
	stack := []any{payload}
	for len(stack) > 0 {
		value := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		switch v := value.(type) {
		case map[string]any:
			if blocked[stringValue(v[fieldIfname])] {
				return fmt.Errorf("host interface leaked into ifname: %v", v)
			}
			for _, child := range v {
				stack = append(stack, child)
			}
		case []any:
			stack = append(stack, v...)
		}
	}
	return nil
}

func portsByIndex(payload map[string]any) (map[int]map[string]any, error) {
	return rowsByIndex(payload, "port_table", 8)
}

func rowsByPortIdx(payload map[string]any, tableName string) (map[int]map[string]any, error) {
	table := list(payload[tableName])
	if table == nil {
		return nil, fmt.Errorf("%s is not a list", tableName)
	}
	out := map[int]map[string]any{}
	for _, row := range table {
		obj, ok := row.(map[string]any)
		if !ok {
			continue
		}
		index := intValue(obj["port_idx"])
		if index > 0 {
			out[index] = obj
		}
	}
	return out, nil
}

func rowsByIndex(payload map[string]any, tableName string, expectedPorts int) (map[int]map[string]any, error) {
	table := list(payload[tableName])
	if table == nil {
		return nil, fmt.Errorf("payload has no %s list", tableName)
	}
	out := map[int]map[string]any{}
	for _, row := range table {
		obj, ok := row.(map[string]any)
		if !ok {
			continue
		}
		index := intValue(obj["port_idx"])
		if index > 0 {
			out[index] = obj
		}
	}
	var missing []int
	for index := 1; index <= expectedPorts; index++ {
		if _, ok := out[index]; !ok {
			missing = append(missing, index)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing %s rows %v", tableName, missing)
	}
	return out, nil
}

func list(value any) []any {
	values, _ := value.([]any)
	return values
}

func stringValue(value any) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}

func intValue(value any) int {
	switch v := value.(type) {
	case int:
		return v
	case float64:
		return int(v)
	default:
		return 0
	}
}

func allowedControllerIfname(ifname string) bool {
	switch ifname {
	case "eth0", "eth1", "eth2", "eth3", "eth4", "eth5":
		return true
	default:
		return false
	}
}

func hasKey(row map[string]any, key string) bool {
	_, ok := row[key]
	return ok
}

func mapKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedIntKeys[T any](m map[int]T) []int {
	keys := make([]int, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Ints(keys)
	return keys
}

func sameIntKeys[A, B any](left map[int]A, right map[int]B) bool {
	leftKeys := sortedIntKeys(left)
	rightKeys := sortedIntKeys(right)
	if len(leftKeys) != len(rightKeys) {
		return false
	}
	for index := range leftKeys {
		if leftKeys[index] != rightKeys[index] {
			return false
		}
	}
	return true
}

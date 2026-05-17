package device

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	jsonKeyAutoneg    = "autoneg"
	jsonKeyEnable     = "enable"
	jsonKeyFullDuplex = "full_duplex"
	jsonKeyIfName     = "ifname"
	jsonKeyL1Up       = "l1up"
	jsonKeyMAC        = "mac"
	jsonKeyMaxSpeed   = "max_speed"
	jsonKeyMedia      = "media"
	jsonKeyName       = "name"
	jsonKeyNetworkGrp = "networkgroup"
	jsonKeyNetmask    = "netmask"
	jsonKeyNumPort    = "num_port"
	jsonKeyPortIdx    = "port_idx"
	jsonKeyRXBytes    = "rx_bytes"
	jsonKeyRXErrors   = "rx_errors"
	jsonKeyRXPackets  = "rx_packets"
	jsonKeySpeed      = "speed"
	jsonKeySpeedCaps  = "speed_caps"
	jsonKeySourceIf   = "source_interface"
	jsonKeyTXBytes    = "tx_bytes"
	jsonKeyTXErrors   = "tx_errors"
	jsonKeyTXPackets  = "tx_packets"
	jsonKeyType       = "type"
	jsonKeyUp         = "up"
	jsonKeyUptime     = "uptime"

	gatewayNetworkGroupLAN  = "LAN"
	gatewayNetworkGroupWAN  = "WAN"
	gatewayNetworkGroupWAN2 = "WAN2"
	gatewayPortRoleLAN      = "lan"
	gatewayPortRoleLAN2     = "lan2"
	gatewayPortRoleWAN      = "wan"
	gatewayPortRoleWAN2     = "wan2"
	payloadModeSwitch       = "switch"
)

// Identity contains the device attributes reported in inform payloads.
type Identity struct {
	// MAC is the fake device MAC address in controller-facing text form.
	MAC string
	// IP is the device management IP address reported to UniFi.
	IP string
	// Hostname is the device name reported to UniFi.
	Hostname string
	// Model is the UniFi model identifier.
	Model string
	// ModelDisplay is the human-readable UniFi model name.
	ModelDisplay string
	// DeviceType is the controller-facing UniFi device family.
	DeviceType string
	// Version is the firmware version reported by the stub.
	Version string
	// Serial is the serial number reported by the stub.
	Serial string
	// InformURL is the controller inform URL currently known by the device.
	InformURL string
	// InformIP is the numeric controller inform endpoint address reported to UniFi.
	InformIP string
	// CFGVersion is the controller configuration version applied to the device.
	CFGVersion string
	// Adopted reports whether the stub should present itself as adopted.
	Adopted bool
}

// MacTableEntry represents a learned MAC entry for a switch port.
type MacTableEntry struct {
	// MAC is the learned client or neighbor MAC address.
	MAC string `json:"mac"`
	// Age is the controller-facing age counter for this entry.
	Age int `json:"age"`
	// Uptime is the number of seconds the entry has been visible.
	Uptime int `json:"uptime"`
	// VLAN is the optional VLAN associated with the entry.
	VLAN int `json:"vlan,omitempty"`
	// Type describes the learned device type when known.
	Type string `json:"type,omitempty"`
}

// Port describes one fake switch port in the UniFi payload.
type Port struct {
	// Index is the one-based UniFi port index.
	Index int
	// Name is the display name reported for the port.
	Name string
	// Interface is the optional host interface that supplied this port's data.
	Interface string
	// MAC is the optional interface MAC address reported for this port.
	MAC string
	// IP is the optional IPv4 address reported for this port.
	IP string
	// Netmask is the optional IPv4 netmask reported for this port.
	Netmask string
	// Role is the gateway-facing role, such as wan, lan, wan2, or lan2.
	Role string
	// NetworkGroup is the UniFi network group, such as WAN, WAN2, or LAN.
	NetworkGroup string
	// Media is the UniFi media label, such as GE or SFP+.
	Media string
	// Uplink marks the port as the upstream connection.
	Uplink bool
	// Up reports whether link is up.
	Up bool
	// Speed is the negotiated speed in Mbps.
	Speed int
	// RXBytes is the receive byte counter.
	RXBytes int64
	// TXBytes is the transmit byte counter.
	TXBytes int64
	// RXPackets is the receive packet counter.
	RXPackets int64
	// TXPackets is the transmit packet counter.
	TXPackets int64
	// RXErrors is the receive error counter.
	RXErrors int64
	// TXErrors is the transmit error counter.
	TXErrors int64
	// MACs contains learned MAC entries for this port.
	MACs []MacTableEntry
}

// PortGroup describes one contiguous block in a switch port layout.
type PortGroup struct {
	// Count is the number of ports in this block.
	Count int
	// Speed is the negotiated speed in Mbps for ports in this block.
	Speed int
	// Media is the UniFi media label for ports in this block.
	Media string
	// Uplink marks the first port in this block as the upstream connection.
	Uplink bool
}

// PortOptions configures generated switch port defaults.
type PortOptions struct {
	// Speed is the default access port speed in Mbps.
	Speed int
	// UplinkSpeed is the uplink port speed in Mbps.
	UplinkSpeed int
	// Media is the default access port media label.
	Media string
	// UplinkMedia is the uplink port media label.
	UplinkMedia string
	// UplinkPort overrides the generated uplink port when positive.
	UplinkPort int
	// PortGroups optionally describe a non-uniform physical port layout.
	PortGroups []PortGroup
	// PortNames optionally override one-based port display labels.
	PortNames []string
	// PortRoles optionally assign one-based gateway roles.
	PortRoles []string
	// PortNetworkGroups optionally assign one-based UniFi network groups.
	PortNetworkGroups []string
}

// PortOverride describes one per-port runtime override.
type PortOverride struct {
	// Port is the one-based switch port index.
	Port int
	// Name overrides the controller-facing port label when set.
	Name string
	// Interface names the optional host interface used as a passive source.
	Interface string
	// MAC overrides the controller-facing interface MAC when set.
	MAC string
	// IP overrides the controller-facing interface IPv4 address when set.
	IP string
	// Netmask overrides the controller-facing interface IPv4 netmask when set.
	Netmask string
	// Role overrides the gateway-facing role when set.
	Role string
	// NetworkGroup overrides the UniFi network group when set.
	NetworkGroup string
	// Speed overrides the negotiated speed in Mbps when positive.
	Speed int
	// Media overrides the controller-facing media label when set.
	Media string
	// Up overrides link state when set.
	Up *bool
	// RXBytes overrides the receive byte counter when non-zero.
	RXBytes int64
	// TXBytes overrides the transmit byte counter when non-zero.
	TXBytes int64
	// RXPackets overrides the receive packet counter when non-zero.
	RXPackets int64
	// TXPackets overrides the transmit packet counter when non-zero.
	TXPackets int64
	// RXErrors overrides the receive error counter when non-zero.
	RXErrors int64
	// TXErrors overrides the transmit error counter when non-zero.
	TXErrors int64
}

// PortNeighbor describes one configured MAC-table entry on a specific port.
type PortNeighbor struct {
	// Port is the one-based switch port index.
	Port int
	// Entry is the controller-facing MAC table entry to expose.
	Entry MacTableEntry
}

// MinimalSwitchPayload returns a JSON inform payload with a switch-shaped port table.
func MinimalSwitchPayload(id Identity, ports []Port) ([]byte, error) {
	now := time.Now().Unix()
	numPorts := len(ports)
	informURL := id.InformURL
	if informURL == "" {
		informURL = "http://unifi:8080/inform"
	}
	cfgVersion := id.CFGVersion
	if cfgVersion == "" {
		cfgVersion = "?"
	}
	ifSpeed := 1000
	if speed := managementInterfaceSpeed(ports); speed > 0 {
		ifSpeed = speed
	}
	deviceType := deviceTypeOrDefault(id.DeviceType)

	payload := map[string]any{
		jsonKeyMAC:           id.MAC,
		"ip":                 id.IP,
		"hostname":           id.Hostname,
		"model":              id.Model,
		"model_display":      id.ModelDisplay,
		jsonKeyType:          deviceType,
		"version":            id.Version,
		"serial":             id.Serial,
		jsonKeyNumPort:       numPorts,
		"state":              informState(id.Adopted),
		"adopted":            id.Adopted,
		"default":            !id.Adopted,
		"discovery_response": true,
		"required_version":   "5.0.0",
		"cfgversion":         cfgVersion,
		jsonKeyUptime:        1,
		"time":               now,
		"inform_url":         informURL,
		"sys_stats":          sysStats(),
		"system-stats":       map[string]any{"cpu": 1.0, "mem": 10.0, jsonKeyUptime: 1},
	}
	if id.InformIP != "" {
		payload["inform_ip"] = id.InformIP
	}
	if isGatewayDeviceType(deviceType) {
		applyGatewayPayload(payload, id, ports)
	} else {
		applySwitchPayload(payload, id, ports, numPorts, ifSpeed)
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal switch payload: %w", err)
	}
	return data, nil
}

func applySwitchPayload(payload map[string]any, id Identity, ports []Port, numPorts int, ifSpeed int) {
	payload["if_table"] = []map[string]any{
		{
			jsonKeyName:       "eth0",
			jsonKeyMAC:        id.MAC,
			"ip":              id.IP,
			jsonKeyNumPort:    numPorts,
			"up":              true,
			jsonKeySpeed:      ifSpeed,
			jsonKeyFullDuplex: true,
		},
	}
	payload["ethernet_table"] = []map[string]any{
		{
			jsonKeyName:    "eth0",
			jsonKeyMAC:     id.MAC,
			jsonKeyNumPort: numPorts,
		},
		{
			jsonKeyName: "srv0",
			jsonKeyMAC:  incrementMAC(id.MAC),
		},
	}
	payload["port_table"] = portTable(ports)
}

func informState(adopted bool) int {
	if adopted {
		return 2
	}
	return 1
}

func isGatewayDeviceType(deviceType string) bool {
	switch strings.TrimSpace(deviceType) {
	case deviceTypeUGW, deviceTypeUXG, deviceTypeUDM:
		return true
	default:
		return false
	}
}

func applyGatewayPayload(payload map[string]any, id Identity, ports []Port) {
	applyGatewayTelemetry(payload, id)
	payload["if_table"] = gatewayIfTable(id, ports)
	payload["network_table"] = gatewayNetworkTable(id, ports)
	payload["uplink"] = gatewayInterfaceName(gatewayUplinkPortIndex(ports))
	payload["uplink_table"] = gatewayUplinkTable(id, ports)
	payload["has_eth1"] = len(ports) > 1
	payload["has_dpi"] = false
	payload["config_network_wan"] = map[string]any{jsonKeyType: "dhcp"}
	if len(ports) > 2 {
		payload["config_network_wan2"] = map[string]any{jsonKeyType: "dhcp"}
	}
}

func applyGatewayTelemetry(payload map[string]any, id Identity) {
	cfgVersion, _ := payload["cfgversion"].(string)
	if cfgVersion == "" {
		cfgVersion = "?"
	}
	payload["anon_id"] = ""
	payload["architecture"] = "aarch64"
	payload["ble_caps"] = 0
	payload["board_rev"] = 1
	payload["bomrev"] = "unknown"
	payload["bomrev_id"] = "00000000"
	payload["boot"] = map[string]any{}
	payload["bootid"] = -1
	payload["bootrom_version"] = "unknown"
	payload["cfgversion_effective"] = cfgVersion
	payload["connections"] = []map[string]any{}
	payload["content_filtering_status"] = map[string]any{"feature_status": "UNAVAILABLE_NO_SUBSCRIPTION"}
	payload["dns_shield"] = map[string]any{"hash": ""}
	payload["dpi_stats"] = []map[string]any{}
	payload["dualboot"] = false
	payload["ever_crash"] = false
	payload["fingerprint"] = "00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00"
	payload["fingerprints"] = []map[string]any{}
	payload["fw2_caps"] = 0
	payload["fw_caps"] = 0
	payload["guest_kicks"] = 0
	payload["guest_token"] = ""
	payload["gw_caps"] = map[string]any{}
	payload["hardware_uuid"] = "00000000-0000-4000-8000-000000000000"
	payload["has_default_route_distance"] = true
	payload["has_speaker"] = false
	payload["has_ssh_disable"] = true
	payload["has_vti"] = true
	payload["hw_caps"] = 0
	payload["ids_ips_rule"] = map[string]any{"rule_count": 0, "sha256": "", "signature_type": "", "update_time": ""}
	payload["inform_min_interval"] = 1
	payload["ipv4_active_leases"] = []map[string]any{}
	payload["isolated"] = false
	payload["kernel_version"] = "6.12.0-stubd"
	payload["last_error_conns"] = []map[string]any{}
	payload["led_state"] = map[string]any{"pattern": "0", "tempo": 120}
	payload["lldp_table"] = []map[string]any{}
	payload["locating"] = false
	payload["manufacturer_id"] = 61
	payload["netmask"] = "255.255.255.0"
	payload["outlet_table"] = []map[string]any{}
	payload["pingtest-status"] = []map[string]any{}
	payload["qrid"] = ""
	payload["reboot_duration"] = 30
	payload["selfrun_beacon"] = true
	payload["speedtest-status"] = gatewaySpeedtestStatus()
	payload["speedtest-status-udapi"] = []map[string]any{}
	payload["ssh_session_table"] = []map[string]any{}
	payload["stats_inform_interval"] = 0
	payload["switch_caps"] = map[string]any{"feature_caps": 1048576, "max_aggregate_sessions": 0, "max_mirror_sessions": 1}
	payload["sys_error_caps"] = 0
	payload["sysid"] = gatewaySysID(id.MAC)
	payload["teleport_version"] = 1
	payload["time_ms"] = 0
	payload["timestamp"] = time.Now().UTC().Format("2006-01-02T15:04:05")
	payload["tm_ready"] = false
	payload["triggers"] = []map[string]any{}
	payload["triggers_dns_filter"] = []map[string]any{}
	payload["triggers_geo"] = []map[string]any{}
	payload["udapi_caps"] = 0
	payload["udapi_version"] = map[string]any{}
	payload["upgrade_duration"] = 150
	payload["uptime_str"] = "1s"
	payload["usg2_caps"] = 0
	payload["usg_caps"] = 0
	payload["wifi_caps"] = 0
}

func gatewaySpeedtestStatus() map[string]any {
	return map[string]any{
		"latency":         0,
		"rundate":         0,
		"runtime":         0,
		"server":          map[string]any{"cc": "", "city": "", "country": "", "lat": 0.0, "lon": 0.0, "provider": "", "provider_url": ""},
		jsonKeySourceIf:   "",
		"status_download": 0,
		"status_ping":     0,
		"status_summary":  0,
		"status_upload":   0,
		"xput_download":   0.0,
		"xput_upload":     0.0,
	}
}

func gatewaySysID(macText string) int {
	mac, err := net.ParseMAC(macText)
	if err != nil || len(mac) < 2 {
		return 42615
	}
	return int(mac[len(mac)-2])<<8 | int(mac[len(mac)-1])
}

func gatewayUplinkPortIndex(ports []Port) int {
	for _, port := range ports {
		if port.Uplink {
			return port.Index
		}
	}
	return 1
}

// SwitchPorts returns count generated switch ports with profile-neutral defaults.
func SwitchPorts(count int) []Port {
	return SwitchPortsWithOptions(count, PortOptions{})
}

// SwitchPortsWithOptions returns count generated switch ports using options.
func SwitchPortsWithOptions(count int, options PortOptions) []Port {
	if count < 1 {
		count = 1
	}
	options = normalizePortOptions(options)
	if ports := groupedSwitchPorts(count, options); len(ports) > 0 {
		return applyUplinkPort(ports, options.UplinkPort)
	}

	ports := make([]Port, 0, count)
	for i := 1; i <= count; i++ {
		speed := options.Speed
		media := options.Media
		if i == 1 {
			speed = options.UplinkSpeed
			media = options.UplinkMedia
		}
		ports = append(ports, generatedPort(i, speed, media, i == 1, options.PortNames, options.PortRoles, options.PortNetworkGroups))
	}
	return applyUplinkPort(ports, options.UplinkPort)
}

// ApplyPortOverrides applies per-port overrides to ports.
func ApplyPortOverrides(ports []Port, overrides []PortOverride) []Port {
	if len(overrides) == 0 || len(ports) == 0 {
		return ports
	}
	for _, override := range overrides {
		if override.Port < 1 || override.Port > len(ports) {
			continue
		}
		port := &ports[override.Port-1]
		if name := strings.TrimSpace(override.Name); name != "" {
			port.Name = name
		}
		if iface := strings.TrimSpace(override.Interface); iface != "" {
			port.Interface = iface
		}
		if mac := strings.TrimSpace(override.MAC); mac != "" {
			port.MAC = strings.ToLower(mac)
		}
		if ip := strings.TrimSpace(override.IP); ip != "" {
			port.IP = ip
		}
		if netmask := strings.TrimSpace(override.Netmask); netmask != "" {
			port.Netmask = netmask
		}
		if role := normalizeGatewayRole(override.Role); role != "" {
			port.Role = role
		}
		if networkGroup := normalizeGatewayNetworkGroup(override.NetworkGroup); networkGroup != "" {
			port.NetworkGroup = networkGroup
		}
		if override.Speed > 0 {
			port.Speed = override.Speed
			if strings.TrimSpace(override.Media) == "" {
				port.Media = mediaForSpeed(override.Speed)
			}
		}
		if override.RXBytes != 0 {
			port.RXBytes = override.RXBytes
		}
		if override.TXBytes != 0 {
			port.TXBytes = override.TXBytes
		}
		if override.RXPackets != 0 {
			port.RXPackets = override.RXPackets
		}
		if override.TXPackets != 0 {
			port.TXPackets = override.TXPackets
		}
		if override.RXErrors != 0 {
			port.RXErrors = override.RXErrors
		}
		if override.TXErrors != 0 {
			port.TXErrors = override.TXErrors
		}
		if media := strings.TrimSpace(override.Media); media != "" {
			port.Media = media
		}
		if override.Up != nil {
			port.Up = *override.Up
			if !*override.Up && override.Speed <= 0 {
				port.Speed = 0
			}
		}
	}
	return ports
}

// ApplyUplinkNeighbor adds a configured neighbor entry to the uplink port.
func ApplyUplinkNeighbor(ports []Port, neighbor *MacTableEntry) []Port {
	if neighbor == nil || strings.TrimSpace(neighbor.MAC) == "" {
		return ports
	}
	entry := normalizeMacTableEntry(*neighbor)
	for index := range ports {
		if !ports[index].Uplink {
			continue
		}
		for macIndex := range ports[index].MACs {
			if strings.EqualFold(ports[index].MACs[macIndex].MAC, entry.MAC) {
				ports[index].MACs[macIndex] = entry
				return ports
			}
		}
		ports[index].MACs = append([]MacTableEntry{entry}, ports[index].MACs...)
		return ports
	}
	return ports
}

// ApplyPortNeighbors adds configured MAC-table entries to their target ports.
func ApplyPortNeighbors(ports []Port, neighbors []PortNeighbor) []Port {
	if len(neighbors) == 0 || len(ports) == 0 {
		return ports
	}
	for _, neighbor := range neighbors {
		if neighbor.Port < 1 || neighbor.Port > len(ports) || strings.TrimSpace(neighbor.Entry.MAC) == "" {
			continue
		}
		entry := normalizeMacTableEntry(neighbor.Entry)
		port := &ports[neighbor.Port-1]
		replaced := false
		for index := range port.MACs {
			if strings.EqualFold(port.MACs[index].MAC, entry.MAC) {
				port.MACs[index] = entry
				replaced = true
				break
			}
		}
		if !replaced {
			port.MACs = append(port.MACs, entry)
		}
	}
	return ports
}

func normalizeMacTableEntry(entry MacTableEntry) MacTableEntry {
	entry.MAC = strings.ToLower(strings.TrimSpace(entry.MAC))
	if entry.Age == 0 {
		entry.Age = 4
	}
	if entry.Uptime == 0 {
		entry.Uptime = 1200
	}
	if strings.TrimSpace(entry.Type) == "" {
		entry.Type = deviceTypeUSW
	}
	return entry
}

func groupedSwitchPorts(count int, options PortOptions) []Port {
	if len(options.PortGroups) == 0 {
		return nil
	}
	total := 0
	uplinkIndex := 0
	for _, group := range options.PortGroups {
		if group.Count < 1 {
			return nil
		}
		if group.Uplink && uplinkIndex == 0 {
			uplinkIndex = total + 1
		}
		total += group.Count
	}
	if total != count {
		return nil
	}
	if uplinkIndex == 0 {
		uplinkIndex = 1
	}

	ports := make([]Port, 0, count)
	index := 0
	for _, group := range options.PortGroups {
		speed := group.Speed
		if speed <= 0 {
			speed = options.Speed
		}
		media := group.Media
		if media == "" {
			media = mediaForSpeed(speed)
		}
		for range group.Count {
			index++
			isUplink := index == uplinkIndex
			portSpeed := speed
			portMedia := media
			if isUplink {
				portSpeed = options.UplinkSpeed
				portMedia = options.UplinkMedia
			}
			ports = append(ports, generatedPort(
				index,
				portSpeed,
				portMedia,
				isUplink,
				options.PortNames,
				options.PortRoles,
				options.PortNetworkGroups,
			))
		}
	}
	return ports
}

func generatedPort(index, speed int, media string, uplink bool, names, roles, networkGroups []string) Port {
	port := Port{
		Index:        index,
		Name:         portName(index, names),
		Role:         normalizeGatewayRole(oneBasedString(index, roles)),
		NetworkGroup: normalizeGatewayNetworkGroup(oneBasedString(index, networkGroups)),
		Media:        media,
		Uplink:       uplink,
		Up:           true,
		Speed:        speed,
		RXBytes:      int64(1000 * index),
		TXBytes:      int64(900 * index),
		RXPackets:    1,
		TXPackets:    1,
	}
	if uplink {
		port.MACs = []MacTableEntry{
			{MAC: "02:aa:bb:cc:dd:01", Age: 4, Uptime: 1200, VLAN: 1, Type: deviceTypeUSW},
		}
	}
	return port
}

func applyUplinkPort(ports []Port, uplinkPort int) []Port {
	if uplinkPort <= 0 {
		return ports
	}
	if uplinkPort > len(ports) {
		return ports
	}
	targetIndex := uplinkPort - 1
	var uplinkMACs []MacTableEntry
	for index := range ports {
		if ports[index].Uplink && len(ports[index].MACs) > 0 {
			uplinkMACs = append([]MacTableEntry{}, ports[index].MACs...)
		}
		ports[index].Uplink = false
		if index != targetIndex {
			ports[index].MACs = nil
		}
	}
	ports[targetIndex].Uplink = true
	if len(ports[targetIndex].MACs) == 0 {
		ports[targetIndex].MACs = uplinkMACs
	}
	return ports
}

func portName(index int, names []string) string {
	if index < 1 {
		index = 1
	}
	if index <= len(names) {
		if name := strings.TrimSpace(names[index-1]); name != "" {
			return name
		}
	}
	return "Port " + strconv.Itoa(index)
}

func oneBasedString(index int, values []string) string {
	if index < 1 || index > len(values) {
		return ""
	}
	return strings.TrimSpace(values[index-1])
}

func incrementMAC(macText string) string {
	mac, err := net.ParseMAC(macText)
	if err != nil || len(mac) == 0 {
		return macText
	}
	out := append(net.HardwareAddr{}, mac...)
	out[len(out)-1]++
	return out.String()
}

func portTable(ports []Port) []map[string]any {
	out := make([]map[string]any, 0, len(ports))
	for _, p := range ports {
		speed := p.Speed
		if p.Up && speed <= 0 {
			speed = 1000
		}
		media := p.Media
		if media == "" && speed > 0 {
			media = mediaForSpeed(speed)
		}
		out = append(out, map[string]any{
			jsonKeyPortIdx:    p.Index,
			jsonKeyIfName:     gatewayInterfaceName(p.Index),
			jsonKeyName:       p.Name,
			jsonKeyMedia:      media,
			jsonKeyEnable:     true,
			jsonKeyUp:         p.Up,
			"is_uplink":       p.Uplink,
			"op_mode":         payloadModeSwitch,
			jsonKeySpeed:      speed,
			jsonKeyMaxSpeed:   speed,
			jsonKeySpeedCaps:  speedCaps(speed, media),
			jsonKeyFullDuplex: true,
			jsonKeyAutoneg:    true,
			"flowctrl_rx":     false,
			"flowctrl_tx":     false,
			"port_poe":        false,
			"poe_enable":      false,
			"poe_caps":        0,
			jsonKeyRXBytes:    p.RXBytes,
			"rx_bytes-r":      0,
			jsonKeyTXBytes:    p.TXBytes,
			"tx_bytes-r":      0,
			jsonKeyRXPackets:  firstNonZeroInt64(p.RXPackets, 1),
			jsonKeyTXPackets:  firstNonZeroInt64(p.TXPackets, 1),
			jsonKeyRXErrors:   p.RXErrors,
			"rx_dropped":      0,
			jsonKeyTXErrors:   p.TXErrors,
			"tx_dropped":      0,
			"satisfaction":    100,
			"stp_state":       "forwarding",
			"stp_pathcost":    20000,
			"mac_table":       p.MACs,
			jsonKeySourceIf:   p.Interface,
		})
	}
	return out
}

func gatewayIfTable(id Identity, ports []Port) []map[string]any {
	out := make([]map[string]any, 0, len(ports))
	for _, port := range ports {
		speed := port.Speed
		if port.Up && speed <= 0 {
			speed = 1000
		}
		ip := gatewayInterfaceIP(id, port)
		netmask := gatewayInterfaceNetmask(port)
		out = append(out, map[string]any{
			jsonKeyName:       gatewayInterfaceName(port.Index),
			jsonKeyIfName:     gatewayInterfaceName(port.Index),
			"comment":         port.Name,
			jsonKeyPortIdx:    port.Index,
			jsonKeyMAC:        gatewayPortMAC(id.MAC, port),
			"ip":              ip,
			jsonKeyNetmask:    netmask,
			jsonKeyNumPort:    1,
			jsonKeyUp:         port.Up,
			jsonKeyEnable:     true,
			jsonKeySpeed:      speed,
			jsonKeyMaxSpeed:   speed,
			jsonKeySpeedCaps:  speedCaps(speed, port.Media),
			jsonKeyMedia:      port.Media,
			jsonKeyNetworkGrp: gatewayNetworkGroup(id.Model, port),
			jsonKeyFullDuplex: true,
			jsonKeyRXBytes:    port.RXBytes,
			jsonKeyTXBytes:    port.TXBytes,
			jsonKeyRXPackets:  firstNonZeroInt64(port.RXPackets, 1),
			jsonKeyTXPackets:  firstNonZeroInt64(port.TXPackets, 1),
			jsonKeyRXErrors:   port.RXErrors,
			jsonKeyTXErrors:   port.TXErrors,
			jsonKeySourceIf:   port.Interface,
		})
	}
	return out
}

func gatewayNetworkTable(id Identity, ports []Port) []map[string]any {
	out := make([]map[string]any, 0, len(ports))
	for _, port := range ports {
		speed := gatewayPortSpeed(port)
		ip := gatewayInterfaceIP(id, port)
		netmask := gatewayInterfaceNetmask(port)
		address := interfaceAddressCIDR(ip, netmask)
		entry := map[string]any{
			jsonKeyName:       gatewayInterfaceName(port.Index),
			jsonKeyIfName:     gatewayInterfaceName(port.Index),
			jsonKeyPortIdx:    port.Index,
			jsonKeyMAC:        gatewayPortMAC(id.MAC, port),
			jsonKeyNetworkGrp: gatewayNetworkGroup(id.Model, port),
			"ip":              ip,
			jsonKeyNetmask:    netmask,
			"address":         address,
			"addresses": []string{
				address,
			},
			jsonKeyUp:       boolText(port.Up),
			jsonKeyL1Up:     boolText(port.Up),
			jsonKeyAutoneg:  "true",
			"duplex":        "full",
			jsonKeySpeed:    strconv.Itoa(speed),
			jsonKeyMaxSpeed: strconv.Itoa(speed),
			"mtu":           "1500",
			jsonKeySourceIf: port.Interface,
			"stats": map[string]any{
				jsonKeyRXBytes:   port.RXBytes,
				jsonKeyTXBytes:   port.TXBytes,
				jsonKeyRXPackets: firstNonZeroInt64(port.RXPackets, 1),
				jsonKeyTXPackets: firstNonZeroInt64(port.TXPackets, 1),
				jsonKeyRXErrors:  port.RXErrors,
				jsonKeyTXErrors:  port.TXErrors,
			},
		}
		if hosts := gatewayHostTable(port); len(hosts) > 0 {
			entry["host_table"] = hosts
		}
		out = append(out, entry)
	}
	return out
}

func gatewayHostTable(port Port) []map[string]any {
	out := make([]map[string]any, 0, len(port.MACs))
	for _, entry := range port.MACs {
		out = append(out, map[string]any{
			jsonKeyMAC:       strings.ToLower(strings.TrimSpace(entry.MAC)),
			"age":            entry.Age,
			"authorized":     true,
			jsonKeyRXBytes:   port.RXBytes,
			jsonKeyTXBytes:   port.TXBytes,
			jsonKeyRXPackets: firstNonZeroInt64(port.RXPackets, 1),
			jsonKeyTXPackets: firstNonZeroInt64(port.TXPackets, 1),
			jsonKeyUptime:    firstNonZero(entry.Uptime, 1200),
		})
	}
	return out
}

func gatewayUplinkTable(id Identity, ports []Port) []map[string]any {
	uplinkIndex := gatewayUplinkPortIndex(ports)
	for _, port := range ports {
		if port.Index != uplinkIndex {
			continue
		}
		speed := gatewayPortSpeed(port)
		return []map[string]any{
			{
				jsonKeyName:       gatewayInterfaceName(port.Index),
				jsonKeyIfName:     gatewayInterfaceName(port.Index),
				jsonKeyPortIdx:    port.Index,
				jsonKeyMAC:        gatewayPortMAC(id.MAC, port),
				jsonKeySpeed:      speed,
				jsonKeyMaxSpeed:   speed,
				jsonKeySpeedCaps:  speedCaps(speed, port.Media),
				jsonKeyType:       "wire",
				jsonKeyMedia:      port.Media,
				jsonKeyUp:         port.Up,
				jsonKeyEnable:     true,
				jsonKeyFullDuplex: true,
				jsonKeyRXBytes:    port.RXBytes,
				jsonKeyTXBytes:    port.TXBytes,
				jsonKeyRXPackets:  firstNonZeroInt64(port.RXPackets, 1),
				jsonKeyTXPackets:  firstNonZeroInt64(port.TXPackets, 1),
				jsonKeyRXErrors:   port.RXErrors,
				jsonKeyTXErrors:   port.TXErrors,
				jsonKeySourceIf:   port.Interface,
			},
		}
	}
	return nil
}

func gatewayPortRole(model string, port Port) string {
	if role := normalizeGatewayRole(port.Role); role != "" {
		return role
	}
	if strings.EqualFold(model, "UXG") {
		switch port.Index {
		case 1:
			return gatewayPortRoleLAN
		case 2:
			return gatewayPortRoleWAN
		}
	}
	if strings.EqualFold(model, "UXGPRO") {
		switch port.Index {
		case 1:
			return gatewayPortRoleWAN
		case 2:
			return gatewayPortRoleLAN
		case 3:
			return gatewayPortRoleWAN2
		case 4:
			return gatewayPortRoleLAN2
		}
	}
	switch port.Index {
	case 1:
		return gatewayPortRoleWAN
	case 2:
		return gatewayPortRoleLAN
	case 3:
		return gatewayPortRoleWAN2
	default:
		return gatewayPortRoleLAN
	}
}

func gatewayNetworkGroup(model string, port Port) string {
	if networkGroup := normalizeGatewayNetworkGroup(port.NetworkGroup); networkGroup != "" {
		return networkGroup
	}
	if strings.EqualFold(model, "UXG") {
		switch port.Index {
		case 2:
			return gatewayNetworkGroupWAN
		default:
			return gatewayNetworkGroupLAN
		}
	}
	if strings.EqualFold(model, "UXGPRO") {
		switch port.Index {
		case 1:
			return gatewayNetworkGroupWAN
		case 3:
			return gatewayNetworkGroupWAN2
		default:
			return gatewayNetworkGroupLAN
		}
	}
	switch gatewayPortRole(model, port) {
	case gatewayPortRoleWAN:
		return gatewayNetworkGroupWAN
	case gatewayPortRoleWAN2:
		return gatewayNetworkGroupWAN2
	default:
		return gatewayNetworkGroupLAN
	}
}

func normalizeGatewayRole(role string) string {
	return strings.ToLower(strings.TrimSpace(role))
}

func normalizeGatewayNetworkGroup(networkGroup string) string {
	return strings.TrimSpace(networkGroup)
}

func gatewayPortSpeed(port Port) int {
	speed := port.Speed
	if port.Up && speed <= 0 {
		return 1000
	}
	return speed
}

func speedCaps(speed int, media string) []int {
	media = strings.ToUpper(strings.TrimSpace(media))
	switch {
	case speed >= 25000 || strings.Contains(media, "SFP28"):
		return []int{1000, 10000, 25000}
	case speed >= 10000 || strings.Contains(media, "SFP+"):
		return []int{1000, 10000}
	case speed >= 2500:
		return []int{10, 100, 1000, 2500}
	default:
		return []int{10, 100, 1000}
	}
}

func gatewayInterfaceName(portIndex int) string {
	if portIndex < 1 {
		portIndex = 1
	}
	return "eth" + strconv.Itoa(portIndex-1)
}

func gatewayPortMAC(baseMAC string, port Port) string {
	if mac := strings.TrimSpace(port.MAC); mac != "" {
		return strings.ToLower(mac)
	}
	return gatewayInterfaceMAC(baseMAC, port.Index)
}

func gatewayInterfaceMAC(baseMAC string, portIndex int) string {
	mac, err := net.ParseMAC(baseMAC)
	if err != nil || len(mac) == 0 {
		return baseMAC
	}
	out := append(net.HardwareAddr{}, mac...)
	out[len(out)-1] += byte(portIndex - 1)
	return out.String()
}

func gatewayInterfaceIP(id Identity, port Port) string {
	if ip := strings.TrimSpace(port.IP); ip != "" {
		return ip
	}
	switch gatewayPortRole(id.Model, port) {
	case gatewayPortRoleLAN, gatewayPortRoleLAN2:
		return id.IP
	case gatewayPortRoleWAN, gatewayPortRoleWAN2:
		return "192.0.2.2"
	}
	return "0.0.0.0"
}

func gatewayInterfaceNetmask(port Port) string {
	if netmask := strings.TrimSpace(port.Netmask); netmask != "" {
		return netmask
	}
	return "255.255.255.0"
}

func interfaceAddressCIDR(ip, netmask string) string {
	prefix := netmaskPrefixLength(netmask)
	if prefix < 0 {
		prefix = 24
	}
	return strings.TrimSpace(ip) + "/" + strconv.Itoa(prefix)
}

func netmaskPrefixLength(netmask string) int {
	parsed := net.ParseIP(strings.TrimSpace(netmask)).To4()
	if parsed == nil {
		return -1
	}
	ones, bits := net.IPMask(parsed).Size()
	if bits != 32 {
		return -1
	}
	return ones
}

func boolText(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func firstNonZeroInt64(values ...int64) int64 {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}

func managementInterfaceSpeed(ports []Port) int {
	for _, port := range ports {
		if port.Uplink && port.Speed > 0 {
			return port.Speed
		}
	}
	if len(ports) > 0 && ports[0].Speed > 0 {
		return ports[0].Speed
	}
	return 0
}

func normalizePortOptions(options PortOptions) PortOptions {
	if options.Speed <= 0 {
		options.Speed = 1000
	}
	if options.UplinkSpeed <= 0 {
		options.UplinkSpeed = options.Speed
	}
	if options.Media == "" {
		options.Media = mediaForSpeed(options.Speed)
	}
	if options.UplinkMedia == "" {
		options.UplinkMedia = mediaForSpeed(options.UplinkSpeed)
	}
	return options
}

func mediaForSpeed(speed int) string {
	if speed >= 10000 {
		return mediaSFPPlus
	}
	return "GE"
}

func sysStats() map[string]any {
	return map[string]any{
		"loadavg_1":  0.01,
		"loadavg_5":  0.01,
		"loadavg_15": 0.01,
		"mem_total":  536870912,
		"mem_used":   67108864,
		"mem_buffer": 0,
	}
}

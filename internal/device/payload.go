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
	jsonKeyMedia      = "media"
	jsonKeyName       = "name"
	jsonKeyNetworkGrp = "networkgroup"
	jsonKeyNumPort    = "num_port"
	jsonKeyPortIdx    = "port_idx"
	jsonKeyRXBytes    = "rx_bytes"
	jsonKeyRXErrors   = "rx_errors"
	jsonKeyRXPackets  = "rx_packets"
	jsonKeySpeed      = "speed"
	jsonKeySpeedCaps  = "speed_caps"
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
}

// PortOverride describes one per-port runtime override.
type PortOverride struct {
	// Port is the one-based switch port index.
	Port int
	// Name overrides the controller-facing port label when set.
	Name string
	// Speed overrides the negotiated speed in Mbps when positive.
	Speed int
	// Media overrides the controller-facing media label when set.
	Media string
	// Up overrides link state when set.
	Up *bool
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
		"state":              1,
		"default":            !id.Adopted,
		"discovery_response": true,
		"required_version":   "5.0.0",
		"cfgversion":         cfgVersion,
		jsonKeyUptime:        1,
		"time":               now,
		"inform_url":         informURL,
		"if_table": []map[string]any{
			{
				jsonKeyName:       "eth0",
				jsonKeyMAC:        id.MAC,
				"ip":              id.IP,
				jsonKeyNumPort:    numPorts,
				"up":              true,
				jsonKeySpeed:      ifSpeed,
				jsonKeyFullDuplex: true,
			},
		},
		"ethernet_table": []map[string]any{
			{
				jsonKeyName:    "eth0",
				jsonKeyMAC:     id.MAC,
				jsonKeyNumPort: numPorts,
			},
			{
				jsonKeyName: "srv0",
				jsonKeyMAC:  incrementMAC(id.MAC),
			},
		},
		"port_table":   portTable(ports),
		"sys_stats":    sysStats(),
		"system-stats": map[string]any{"cpu": 1.0, "mem": 10.0, jsonKeyUptime: 1},
	}
	if isGatewayDeviceType(deviceType) {
		applyGatewayPayload(payload, id, ports)
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal switch payload: %w", err)
	}
	return data, nil
}

func isGatewayDeviceType(deviceType string) bool {
	switch strings.TrimSpace(deviceType) {
	case deviceTypeUGW, deviceTypeUXG:
		return true
	default:
		return false
	}
}

func applyGatewayPayload(payload map[string]any, id Identity, ports []Port) {
	payload["if_table"] = gatewayIfTable(id, ports)
	payload["ethernet_table"] = gatewayEthernetTable(id, ports)
	payload["config_port_table"] = gatewayConfigPortTable(id.Model, ports)
	payload["ethernet_overrides"] = gatewayEthernetOverrides(id.Model, ports)
	payload["port_overrides"] = gatewayPortOverrides(id.Model, ports)
	payload["network_table"] = gatewayNetworkTable(id, ports)
	payload["uplink"] = gatewayInterfaceName(gatewayUplinkPortIndex(ports))
	payload["uplink_table"] = gatewayUplinkTable(ports)
	payload["has_eth1"] = len(ports) > 1
	payload["has_dpi"] = false
	payload["config_network_wan"] = map[string]any{jsonKeyType: "dhcp"}
	if len(ports) > 2 {
		payload["config_network_wan2"] = map[string]any{jsonKeyType: "dhcp"}
	}
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
		ports = append(ports, generatedPort(i, speed, media, i == 1, options.PortNames))
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
		if override.Speed > 0 {
			port.Speed = override.Speed
			if strings.TrimSpace(override.Media) == "" {
				port.Media = mediaForSpeed(override.Speed)
			}
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
			ports = append(ports, generatedPort(index, portSpeed, portMedia, isUplink, options.PortNames))
		}
	}
	return ports
}

func generatedPort(index, speed int, media string, uplink bool, names []string) Port {
	port := Port{
		Index:     index,
		Name:      portName(index, names),
		Media:     media,
		Uplink:    uplink,
		Up:        true,
		Speed:     speed,
		RXBytes:   int64(1000 * index),
		TXBytes:   int64(900 * index),
		RXPackets: 1,
		TXPackets: 1,
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
		out = append(out, map[string]any{
			jsonKeyName:       gatewayInterfaceName(port.Index),
			jsonKeyMAC:        gatewayInterfaceMAC(id.MAC, port.Index),
			"ip":              gatewayInterfaceIP(id.IP, port),
			"netmask":         "255.255.255.0",
			jsonKeyNumPort:    1,
			jsonKeyUp:         port.Up,
			jsonKeyEnable:     true,
			jsonKeySpeed:      speed,
			jsonKeySpeedCaps:  speedCaps(speed, port.Media),
			jsonKeyMedia:      port.Media,
			jsonKeyFullDuplex: true,
			jsonKeyRXBytes:    port.RXBytes,
			jsonKeyTXBytes:    port.TXBytes,
			jsonKeyRXPackets:  firstNonZeroInt64(port.RXPackets, 1),
			jsonKeyTXPackets:  firstNonZeroInt64(port.TXPackets, 1),
			jsonKeyRXErrors:   port.RXErrors,
			jsonKeyTXErrors:   port.TXErrors,
		})
	}
	return out
}

func gatewayConfigPortTable(model string, ports []Port) []map[string]any {
	out := make([]map[string]any, 0, len(ports))
	for _, port := range ports {
		out = append(out, map[string]any{
			jsonKeyIfName: gatewayInterfaceName(port.Index),
			jsonKeyName:   gatewayPortRole(model, port),
		})
	}
	return out
}

func gatewayEthernetTable(id Identity, ports []Port) []map[string]any {
	out := make([]map[string]any, 0, len(ports))
	for _, port := range ports {
		speed := gatewayPortSpeed(port)
		out = append(out, map[string]any{
			jsonKeyName:       gatewayInterfaceName(port.Index),
			jsonKeyIfName:     gatewayInterfaceName(port.Index),
			jsonKeyMAC:        gatewayInterfaceMAC(id.MAC, port.Index),
			jsonKeyNumPort:    1,
			jsonKeyPortIdx:    port.Index,
			jsonKeySpeed:      speed,
			jsonKeyUp:         port.Up,
			jsonKeyMedia:      port.Media,
			jsonKeyNetworkGrp: gatewayNetworkGroup(id.Model, port),
			jsonKeySpeedCaps:  speedCaps(speed, port.Media),
		})
	}
	return out
}

func gatewayEthernetOverrides(model string, ports []Port) []map[string]any {
	out := make([]map[string]any, 0, len(ports))
	for _, port := range ports {
		speed := gatewayPortSpeed(port)
		out = append(out, map[string]any{
			jsonKeyIfName:     gatewayInterfaceName(port.Index),
			jsonKeyName:       port.Name,
			jsonKeyPortIdx:    port.Index,
			jsonKeyNetworkGrp: gatewayNetworkGroup(model, port),
			jsonKeySpeed:      speed,
			jsonKeyFullDuplex: true,
			jsonKeyAutoneg:    true,
		})
	}
	return out
}

func gatewayPortOverrides(model string, ports []Port) []map[string]any {
	out := make([]map[string]any, 0, len(ports))
	for _, port := range ports {
		speed := gatewayPortSpeed(port)
		out = append(out, map[string]any{
			jsonKeyPortIdx:    port.Index,
			jsonKeyIfName:     gatewayInterfaceName(port.Index),
			jsonKeyName:       port.Name,
			jsonKeyNetworkGrp: gatewayNetworkGroup(model, port),
			"op_mode":         payloadModeSwitch,
			jsonKeyMedia:      port.Media,
			jsonKeySpeed:      speed,
			jsonKeySpeedCaps:  speedCaps(speed, port.Media),
			jsonKeyEnable:     true,
			jsonKeyUp:         port.Up,
			jsonKeyFullDuplex: true,
			jsonKeyAutoneg:    true,
		})
	}
	return out
}

func gatewayNetworkTable(id Identity, ports []Port) []map[string]any {
	out := make([]map[string]any, 0, len(ports))
	for _, port := range ports {
		speed := gatewayPortSpeed(port)
		entry := map[string]any{
			jsonKeyName: gatewayInterfaceName(port.Index),
			jsonKeyMAC:  gatewayInterfaceMAC(id.MAC, port.Index),
			"address":   gatewayInterfaceIP(id.IP, port) + "/24",
			"addresses": []string{
				gatewayInterfaceIP(id.IP, port) + "/24",
			},
			jsonKeyUp:      boolText(port.Up),
			jsonKeyL1Up:    boolText(port.Up),
			jsonKeyAutoneg: "true",
			"duplex":       "full",
			jsonKeySpeed:   strconv.Itoa(speed),
			"mtu":          "1500",
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

func gatewayUplinkTable(ports []Port) []map[string]any {
	uplinkIndex := gatewayUplinkPortIndex(ports)
	for _, port := range ports {
		if port.Index != uplinkIndex {
			continue
		}
		return []map[string]any{
			{
				jsonKeyName:    gatewayInterfaceName(port.Index),
				jsonKeyIfName:  gatewayInterfaceName(port.Index),
				jsonKeyPortIdx: port.Index,
				jsonKeySpeed:   gatewayPortSpeed(port),
				jsonKeyType:    "wire",
				jsonKeyMedia:   port.Media,
			},
		}
	}
	return nil
}

func gatewayPortRole(model string, port Port) string {
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

func gatewayInterfaceMAC(baseMAC string, portIndex int) string {
	mac, err := net.ParseMAC(baseMAC)
	if err != nil || len(mac) == 0 {
		return baseMAC
	}
	out := append(net.HardwareAddr{}, mac...)
	out[len(out)-1] += byte(portIndex - 1)
	return out.String()
}

func gatewayInterfaceIP(baseIP string, port Port) string {
	if port.Uplink {
		return "192.0.2.2"
	}
	if port.Index == 2 || strings.Contains(strings.ToLower(port.Name), "downlink") {
		return baseIP
	}
	if port.Index == 4 {
		return baseIP
	}
	return "0.0.0.0"
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

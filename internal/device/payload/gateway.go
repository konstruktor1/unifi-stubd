package payload

import (
	"net"
	"strconv"
	"strings"
)

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
func gatewayUplinkPortIndex(ports []Port) int {
	for _, port := range ports {
		if port.Uplink {
			return port.Index
		}
	}
	return 1
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

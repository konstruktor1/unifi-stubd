package payload

// This file renders gateway-specific inform tables from generated ports.

import (
	"net"
	"strconv"
	"strings"
)

// applyGatewayPayload fills the gateway-specific tables expected by UniFi.
func applyGatewayPayload(payload map[string]any, profile Profile, id Identity, ports []Port) {
	applyGatewayTelemetry(payload, id)
	payload["if_table"] = gatewayIfTable(profile, id, ports)
	payload["network_table"] = gatewayNetworkTable(profile, id, ports)
	payload["uplink"] = gatewayInterfaceName(profile, gatewayUplinkPortIndex(ports))
	payload["uplink_table"] = gatewayUplinkTable(profile, id, ports)
	payload["has_eth1"] = len(ports) > 1
	payload["has_dpi"] = profile.HasDPI
	payload["config_network_wan"] = map[string]any{jsonKeyType: "dhcp"}
	if len(ports) > 2 {
		payload["config_network_wan2"] = map[string]any{jsonKeyType: "dhcp"}
	}
}

// gatewayUplinkPortIndex returns the one-based uplink port index.
func gatewayUplinkPortIndex(ports []Port) int {
	for _, port := range ports {
		if port.Uplink {
			return port.Index
		}
	}
	return 1
}

// gatewayIfTable renders physical interfaces for gateway inform payloads.
func gatewayIfTable(profile Profile, id Identity, ports []Port) []map[string]any {
	out := make([]map[string]any, 0, len(ports))
	for _, port := range ports {
		speed := port.Speed
		if port.Up && speed <= 0 {
			speed = 1000
		}
		ip := gatewayInterfaceIP(id, port)
		netmask := gatewayInterfaceNetmask(port)
		ifaceName := gatewayInterfaceName(profile, port.Index)
		row := map[string]any{
			jsonKeyName:       ifaceName,
			jsonKeyIfName:     ifaceName,
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
			jsonKeyNetworkGrp: gatewayNetworkGroup(port),
			jsonKeyFullDuplex: true,
			jsonKeyRXBytes:    port.RXBytes,
			jsonKeyTXBytes:    port.TXBytes,
			jsonKeyRXPackets:  firstNonZeroInt64(port.RXPackets, 1),
			jsonKeyTXPackets:  firstNonZeroInt64(port.TXPackets, 1),
			jsonKeyRXErrors:   port.RXErrors,
			jsonKeyTXErrors:   port.TXErrors,
			jsonKeySourceIf:   port.Interface,
		}
		if port.Index == gatewayUplinkPortIndex(ports) {
			addManagementVLAN(row, id.ManagementVLAN)
		}
		out = append(out, row)
	}
	return out
}

// gatewayNetworkTable renders the routed network view for each gateway port.
func gatewayNetworkTable(profile Profile, id Identity, ports []Port) []map[string]any {
	out := make([]map[string]any, 0, len(ports))
	for _, port := range ports {
		speed := gatewayPortSpeed(port)
		ip := gatewayInterfaceIP(id, port)
		netmask := gatewayInterfaceNetmask(port)
		address := interfaceAddressCIDR(ip, netmask)
		ifaceName := gatewayInterfaceName(profile, port.Index)
		entry := map[string]any{
			jsonKeyName:       ifaceName,
			jsonKeyIfName:     ifaceName,
			jsonKeyPortIdx:    port.Index,
			jsonKeyMAC:        gatewayPortMAC(id.MAC, port),
			jsonKeyNetworkGrp: gatewayNetworkGroup(port),
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

// gatewayHostTable renders learned client MACs for one gateway port.
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

// gatewayUplinkTable renders the controller-facing uplink entry.
func gatewayUplinkTable(profile Profile, id Identity, ports []Port) []map[string]any {
	uplinkIndex := gatewayUplinkPortIndex(ports)
	for _, port := range ports {
		if port.Index != uplinkIndex {
			continue
		}
		speed := gatewayPortSpeed(port)
		ifaceName := gatewayInterfaceName(profile, port.Index)
		row := map[string]any{
			jsonKeyName:       ifaceName,
			jsonKeyIfName:     ifaceName,
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
		}
		addManagementVLAN(row, id.ManagementVLAN)
		return []map[string]any{row}
	}
	return nil
}

// gatewayPortRole returns profile or override data before generic fallback roles.
func gatewayPortRole(port Port) string {
	if role := normalizeGatewayRole(port.Role); role != "" {
		return role
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

// gatewayNetworkGroup maps a gateway role into the UniFi network group label.
func gatewayNetworkGroup(port Port) string {
	if networkGroup := normalizeGatewayNetworkGroup(port.NetworkGroup); networkGroup != "" {
		return networkGroup
	}
	switch gatewayPortRole(port) {
	case gatewayPortRoleWAN:
		return gatewayNetworkGroupWAN
	case gatewayPortRoleWAN2:
		return gatewayNetworkGroupWAN2
	default:
		return gatewayNetworkGroupLAN
	}
}

// normalizeGatewayRole normalizes configured gateway role labels.
func normalizeGatewayRole(role string) string {
	return strings.ToLower(strings.TrimSpace(role))
}

// normalizeGatewayNetworkGroup normalizes configured network group labels.
func normalizeGatewayNetworkGroup(networkGroup string) string {
	return strings.TrimSpace(networkGroup)
}

// gatewayPortSpeed keeps an up gateway port from reporting an invalid speed.
func gatewayPortSpeed(port Port) int {
	speed := port.Speed
	if port.Up && speed <= 0 {
		return 1000
	}
	return speed
}

// gatewayInterfaceName maps a one-based port index to a profile-selected interface prefix.
func gatewayInterfaceName(profile Profile, portIndex int) string {
	if portIndex < 1 {
		portIndex = 1
	}
	prefix := strings.TrimSpace(profile.GatewayInterfacePrefix)
	if prefix == "" {
		prefix = "eth"
	}
	return prefix + strconv.Itoa(portIndex-1)
}

// gatewayPortMAC returns a configured port MAC or derives one from the device MAC.
func gatewayPortMAC(baseMAC string, port Port) string {
	if mac := strings.TrimSpace(port.MAC); mac != "" {
		return strings.ToLower(mac)
	}
	return gatewayInterfaceMAC(baseMAC, port.Index)
}

// gatewayInterfaceMAC derives a stable per-interface MAC from the base address.
func gatewayInterfaceMAC(baseMAC string, portIndex int) string {
	mac, err := net.ParseMAC(baseMAC)
	if err != nil || len(mac) == 0 {
		return baseMAC
	}
	out := append(net.HardwareAddr{}, mac...)
	out[len(out)-1] += byte(portIndex - 1)
	return out.String()
}

// gatewayInterfaceIP chooses the management or documentation WAN address for a port.
func gatewayInterfaceIP(id Identity, port Port) string {
	if ip := strings.TrimSpace(port.IP); ip != "" {
		return ip
	}
	switch gatewayPortRole(port) {
	case gatewayPortRoleLAN, gatewayPortRoleLAN2:
		return id.IP
	case gatewayPortRoleWAN, gatewayPortRoleWAN2:
		return "192.0.2.2"
	}
	return "0.0.0.0"
}

// gatewayInterfaceNetmask returns an override or the lab default netmask.
func gatewayInterfaceNetmask(port Port) string {
	if netmask := strings.TrimSpace(port.Netmask); netmask != "" {
		return netmask
	}
	return "255.255.255.0"
}

// interfaceAddressCIDR combines dotted netmask data into controller CIDR form.
func interfaceAddressCIDR(ip, netmask string) string {
	prefix := netmaskPrefixLength(netmask)
	if prefix < 0 {
		prefix = 24
	}
	return strings.TrimSpace(ip) + "/" + strconv.Itoa(prefix)
}

// netmaskPrefixLength converts a dotted IPv4 netmask to a prefix length.
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

// boolText returns the string form used by gateway network table fields.
func boolText(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

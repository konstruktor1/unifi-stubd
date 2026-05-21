// Package payload derives gateway if_table, network_table, and WAN/LAN state
// from profile metadata and port roles. Model names do not decide gateway
// behavior.
package payload

import (
	"net"
	"strconv"
	"strings"
	"time"
)

// applyGatewayPayload fills the gateway-specific tables expected by UniFi.
func applyGatewayPayload(payload map[string]any, profile Profile, id Identity, ports []PortView, now time.Time, uptime int) {
	applyGatewayTelemetry(payload, id, now, uptime)
	payload["if_table"] = gatewayIfTable(profile, id, ports)
	payload["network_table"] = gatewayNetworkTable(profile, id, ports)
	payload["config_port_table"] = gatewayConfigPortTable(ports)
	payload["ethernet_overrides"] = gatewayEthernetOverrides(ports)
	payload["reported_networks"] = gatewayReportedNetworks(ports)
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
func gatewayUplinkPortIndex(ports []PortView) int {
	for _, port := range ports {
		if port.Uplink {
			return port.Index
		}
	}
	return 1
}

// gatewayIfTable renders physical interfaces for gateway inform payloads.
func gatewayIfTable(_ Profile, id Identity, ports []PortView) []map[string]any {
	out := make([]map[string]any, 0, len(ports))
	uplinkIndex := gatewayUplinkPortIndex(ports)
	for _, view := range ports {
		iface := view.GatewayInterface
		row := map[string]any{
			jsonKeyName:       iface.Name,
			jsonKeyIfName:     iface.IfName,
			"comment":         iface.Comment,
			jsonKeyPortIdx:    view.Index,
			jsonKeyMAC:        iface.MAC,
			"ip":              iface.IP,
			jsonKeyNetmask:    iface.Netmask,
			jsonKeyNumPort:    1,
			jsonKeyUp:         view.Up,
			jsonKeyEnable:     view.Enabled,
			jsonKeyNetworkGrp: iface.NetworkGroup,
			jsonKeyFullDuplex: true,
		}
		addFields(row, portLinkFields(view.Speed, view.Media), portCounterFields(view.Port), sourceFields(view.SourceInterface))
		if view.Index == uplinkIndex {
			addManagementVLAN(row, id.ManagementVLAN)
		}
		out = append(out, row)
	}
	return out
}

// gatewayNetworkTable renders the routed network view for each gateway port.
func gatewayNetworkTable(_ Profile, _ Identity, ports []PortView) []map[string]any {
	out := make([]map[string]any, 0, len(ports))
	for _, view := range ports {
		iface := view.GatewayInterface
		counters := portCounterFields(view.Port)
		entry := map[string]any{
			jsonKeyName:       iface.Name,
			jsonKeyIfName:     iface.IfName,
			jsonKeyPortIdx:    view.Index,
			jsonKeyMAC:        iface.MAC,
			jsonKeyNetworkGrp: iface.NetworkGroup,
			"ip":              iface.IP,
			jsonKeyNetmask:    iface.Netmask,
			"address":         iface.Address,
			"addresses": []string{
				iface.Address,
			},
			jsonKeyUp:       boolText(view.Up),
			jsonKeyL1Up:     boolText(view.Up),
			jsonKeyAutoneg:  "true",
			"duplex":        "full",
			jsonKeySpeed:    strconv.Itoa(view.Speed),
			jsonKeyMaxSpeed: strconv.Itoa(view.Speed),
			"mtu":           "1500",
			"stats":         counters,
		}
		addFields(entry, sourceFields(view.SourceInterface))
		if hosts := gatewayHostTable(view.Port); len(hosts) > 0 {
			entry["host_table"] = hosts
		}
		out = append(out, entry)
	}
	return out
}

// gatewayConfigPortTable renders gateway WAN/LAN port assignments from the
// same resolved port view used by interface and network tables.
func gatewayConfigPortTable(ports []PortView) []map[string]any {
	out := make([]map[string]any, 0, len(ports))
	for _, view := range ports {
		iface := view.GatewayInterface
		row := map[string]any{
			jsonKeyName:       view.Name,
			jsonKeyIfName:     iface.IfName,
			jsonKeyPortIdx:    view.Index,
			jsonKeyNetworkGrp: view.NetworkGroup,
			"role":            view.Role,
			jsonKeyUp:         view.Up,
			jsonKeyEnable:     view.Enabled,
			"is_uplink":       view.Uplink,
		}
		addFields(row, portLinkFields(view.Speed, view.Media), sourceFields(view.SourceInterface))
		out = append(out, row)
	}
	return out
}

// gatewayEthernetOverrides renders the interface binding data that gateway
// controllers use for port remapping and visual WAN/LAN state.
func gatewayEthernetOverrides(ports []PortView) []map[string]any {
	out := make([]map[string]any, 0, len(ports))
	for _, view := range ports {
		iface := view.GatewayInterface
		row := map[string]any{
			jsonKeyName:       iface.Name,
			jsonKeyIfName:     iface.IfName,
			jsonKeyPortIdx:    view.Index,
			jsonKeyMAC:        iface.MAC,
			jsonKeyNetworkGrp: view.NetworkGroup,
			"role":            view.Role,
			jsonKeyUp:         view.Up,
			jsonKeyEnable:     view.Enabled,
		}
		addFields(row, portLinkFields(view.Speed, view.Media), sourceFields(view.SourceInterface))
		out = append(out, row)
	}
	return out
}

// gatewayReportedNetworks renders a read-only network summary per gateway
// port. It mirrors network_table values without inventing host configuration.
func gatewayReportedNetworks(ports []PortView) []map[string]any {
	out := make([]map[string]any, 0, len(ports))
	for _, view := range ports {
		iface := view.GatewayInterface
		row := map[string]any{
			jsonKeyName:       iface.NetworkGroup,
			jsonKeyIfName:     iface.IfName,
			jsonKeyPortIdx:    view.Index,
			jsonKeyNetworkGrp: iface.NetworkGroup,
			jsonKeyType:       view.Role,
			"ip":              iface.IP,
			jsonKeyNetmask:    iface.Netmask,
			"address":         iface.Address,
			"addresses":       []string{iface.Address},
			jsonKeyUp:         view.Up,
		}
		addFields(row, sourceFields(view.SourceInterface))
		out = append(out, row)
	}
	return out
}

// gatewayHostTable renders learned client MACs for one gateway port.
func gatewayHostTable(port Port) []map[string]any {
	out := make([]map[string]any, 0, len(port.MACs))
	for _, entry := range port.MACs {
		row := map[string]any{
			jsonKeyMAC:       strings.ToLower(strings.TrimSpace(entry.MAC)),
			"age":            entry.Age,
			"authorized":     true,
			jsonKeyRXBytes:   port.RXBytes,
			jsonKeyTXBytes:   port.TXBytes,
			jsonKeyRXPackets: firstNonZeroInt64(port.RXPackets, 1),
			jsonKeyTXPackets: firstNonZeroInt64(port.TXPackets, 1),
			jsonKeyUptime:    firstNonZero(entry.Uptime, 1200),
		}
		if hostname := strings.TrimSpace(entry.Hostname); hostname != "" {
			row["hostname"] = hostname
		}
		if ip := strings.TrimSpace(entry.IP); ip != "" {
			row["ip"] = ip
		}
		out = append(out, row)
	}
	return out
}

// gatewayUplinkTable renders the controller-facing uplink entry.
func gatewayUplinkTable(_ Profile, id Identity, ports []PortView) []map[string]any {
	uplinkIndex := gatewayUplinkPortIndex(ports)
	for _, view := range ports {
		if view.Index != uplinkIndex {
			continue
		}
		iface := view.GatewayInterface
		row := map[string]any{
			jsonKeyName:       iface.Name,
			jsonKeyIfName:     iface.IfName,
			jsonKeyPortIdx:    view.Index,
			jsonKeyMAC:        iface.MAC,
			jsonKeyType:       "wire",
			jsonKeyUp:         view.Up,
			jsonKeyEnable:     view.Enabled,
			jsonKeyFullDuplex: true,
		}
		addFields(row, portLinkFields(view.Speed, view.Media), portCounterFields(view.Port), sourceFields(view.SourceInterface))
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

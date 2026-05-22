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
	payload["port_table"] = gatewayPortTable(ports)
	payload["uplink"] = gatewayInterfaceName(profile, gatewayUplinkPortIndex(ports))
	payload["uplink_table"] = gatewayUplinkTable(profile, id, ports)
	payload["has_eth1"] = len(ports) > 1
	payload["has_dpi"] = profile.HasDPI
	payload["config_network_wan"] = gatewayConfigNetwork(ports, gatewayPortRoleWAN)
	if wan1 := gatewayWANStatus(ports, gatewayPortRoleWAN, uptime); len(wan1) > 0 {
		payload["wan1"] = wan1
	}
	if wan2 := gatewayConfigNetwork(ports, gatewayPortRoleWAN2); len(wan2) > 1 {
		payload["config_network_wan2"] = wan2
	}
	if wan2 := gatewayWANStatus(ports, gatewayPortRoleWAN2, uptime); len(wan2) > 0 {
		payload["wan2"] = wan2
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
			"physical_ports":  []int{view.Index},
		}
		addFields(row, portLinkFields(view.Speed, view.Media), portCounterFields(view.Port), explicitPortRateFields(view.Port), sourceFields(view.SourceInterface), gatewayConnectionFields(view))
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
		addFields(counters, explicitPortRateFields(view.Port))
		entry := map[string]any{
			jsonKeyName:       iface.Name,
			jsonKeyIfName:     iface.IfName,
			jsonKeyPortIdx:    view.Index,
			jsonKeyMAC:        iface.MAC,
			jsonKeyNetworkGrp: iface.NetworkGroup,
			"ip":              iface.IP,
			jsonKeyNetmask:    iface.Netmask,
			jsonKeyAddress:    iface.Address,
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
		addFields(entry, sourceFields(view.SourceInterface), gatewayConnectedField(view))
		if hosts := gatewayHostTable(view.Port, view.Uplink); len(hosts) > 0 {
			entry["host_table"] = hosts
		}
		out = append(out, entry)
	}
	return out
}

// gatewayConfigNetwork renders the controller-owned network assignment view
// for the first port matching the requested WAN/LAN role.
func gatewayConfigNetwork(ports []PortView, role string) map[string]any {
	for _, view := range ports {
		if gatewayPortRole(view.Port) != role {
			continue
		}
		iface := view.GatewayInterface
		row := map[string]any{
			jsonKeyType:       payloadTypeDHCP,
			jsonKeyName:       iface.NetworkGroup,
			jsonKeyIfName:     iface.IfName,
			jsonKeyPortIdx:    view.Index,
			jsonKeyNetworkGrp: iface.NetworkGroup,
			jsonKeyRole:       view.Role,
			jsonKeyMAC:        iface.MAC,
			"ip":              iface.IP,
			jsonKeyNetmask:    iface.Netmask,
			jsonKeyAddress:    iface.Address,
			jsonKeyUp:         view.Up,
			jsonKeyEnable:     view.Enabled,
		}
		addFields(row, sourceFields(view.SourceInterface))
		return row
	}
	return map[string]any{jsonKeyType: payloadTypeDHCP}
}

// gatewayWANStatus renders live WAN-like state from the same resolved port view
// used by the gateway interface tables.
func gatewayWANStatus(ports []PortView, role string, uptime int) map[string]any {
	for _, view := range ports {
		if gatewayPortRole(view.Port) != role {
			continue
		}
		iface := view.GatewayInterface
		row := map[string]any{
			jsonKeyType:       payloadTypeDHCP,
			jsonKeyName:       iface.NetworkGroup,
			jsonKeyIfName:     iface.IfName,
			jsonKeyPortIdx:    view.Index,
			jsonKeyNetworkGrp: iface.NetworkGroup,
			jsonKeyRole:       view.Role,
			jsonKeyMAC:        iface.MAC,
			"ip":              iface.IP,
			jsonKeyNetmask:    iface.Netmask,
			jsonKeyAddress:    iface.Address,
			jsonKeyUp:         view.Up,
			jsonKeyEnable:     view.Enabled,
			jsonKeyUptime:     uptime,
			jsonKeyLatency:    0,
		}
		addFields(row, portLinkFields(view.Speed, view.Media), portCounterFields(view.Port), explicitPortRateFields(view.Port), sourceFields(view.SourceInterface))
		return row
	}
	return nil
}

// gatewayConnectedField uses the resolved port view as the single source for
// gateway connection state.
func gatewayConnectedField(view PortView) map[string]any {
	return map[string]any{"connected": view.Up}
}

// gatewayConnectionFields derives controller topology hints from the first
// visible MAC-table entry on a connected port.
func gatewayConnectionFields(view PortView) map[string]any {
	out := gatewayConnectedField(view)
	if !view.Up || len(view.MACs) == 0 {
		return out
	}
	// Controllers use last_connection as a topology hint. The first MAC entry
	// is therefore treated as metadata about the visible neighbor, not as host
	// configuration to apply.
	entry := view.MACs[0]
	connection := map[string]any{
		jsonKeyMAC: strings.ToLower(strings.TrimSpace(entry.MAC)),
		"source":   jsonKeyMACTable,
	}
	if ip := strings.TrimSpace(entry.IP); ip != "" {
		connection["ip"] = ip
	}
	if hostname := strings.TrimSpace(entry.Hostname); hostname != "" {
		connection["hostname"] = hostname
	}
	if entryType := strings.TrimSpace(entry.Type); entryType != "" {
		connection[jsonKeyType] = entryType
	}
	out["last_connection"] = connection
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
			jsonKeyRole:       view.Role,
			jsonKeyUp:         view.Up,
			jsonKeyEnable:     view.Enabled,
			jsonKeyIsUplink:   view.Uplink,
		}
		addFields(row, portLinkFields(view.Speed, view.Media), sourceFields(view.SourceInterface), gatewayConnectionFields(view))
		out = append(out, row)
	}
	return out
}

// gatewayPortTable renders physical gateway ports for controller views that
// treat UXG/UDM ports as switch-like rows. It intentionally avoids port profile
// or VLAN assignments; those remain controller-owned configuration.
func gatewayPortTable(ports []PortView) []map[string]any {
	out := make([]map[string]any, 0, len(ports))
	for _, view := range ports {
		iface := view.GatewayInterface
		row := map[string]any{
			jsonKeyPortIdx:    view.Index,
			jsonKeyIfName:     iface.IfName,
			jsonKeyName:       view.Name,
			jsonKeyEnable:     view.Enabled,
			jsonKeyUp:         view.Up,
			jsonKeyNetworkGrp: iface.NetworkGroup,
			jsonKeyRole:       view.Role,
			jsonKeyIsUplink:   view.Uplink,
			"op_mode":         payloadKindGateway,
			jsonKeyFullDuplex: true,
			jsonKeyAutoneg:    true,
			"flowctrl_rx":     false,
			"flowctrl_tx":     false,
			jsonKeyMACTable:   view.MACs,
			"rx_dropped":      0,
			"tx_dropped":      0,
		}
		if iface.MAC != "" {
			row[jsonKeyMAC] = iface.MAC
		}
		if iface.IP != "" {
			row["ip"] = iface.IP
		}
		addFields(row, portLinkFields(view.Speed, view.Media), portCounterFields(view.Port), explicitPortRateFields(view.Port), sourceFields(view.SourceInterface), gatewayConnectionFields(view))
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
			jsonKeyRole:       view.Role,
			jsonKeyUp:         view.Up,
			jsonKeyEnable:     view.Enabled,
		}
		addFields(row, portLinkFields(view.Speed, view.Media), sourceFields(view.SourceInterface), gatewayConnectionFields(view))
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
			jsonKeyAddress:    iface.Address,
			"addresses":       []string{iface.Address},
			jsonKeyUp:         view.Up,
		}
		addFields(row, sourceFields(view.SourceInterface), gatewayConnectedField(view))
		out = append(out, row)
	}
	return out
}

// gatewayHostTable renders learned downstream MACs for one gateway port.
func gatewayHostTable(port Port, uplink bool) []map[string]any {
	out := make([]map[string]any, 0, len(port.MACs))
	for _, entry := range port.MACs {
		entryType := strings.TrimSpace(entry.Type)
		if uplink && entryType != "" && entryType != "client" {
			// Uplink neighbor metadata belongs in uplink/last_connection fields.
			// The gateway host table should contain downstream client-like MACs.
			continue
		}
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
		if entryType != "" {
			row[jsonKeyType] = entryType
		}
		if entry.VLAN > 0 {
			row["vlan"] = entry.VLAN
		}
		if entry.Static {
			row["static"] = true
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
		addFields(row, portLinkFields(view.Speed, view.Media), portCounterFields(view.Port), explicitPortRateFields(view.Port), sourceFields(view.SourceInterface), gatewayConnectionFields(view))
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
	// Gateway WAN fallbacks use documentation addresses so payload examples do
	// not leak or invent real lab network data.
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

// Package payload derives gateway if_table, network_table, and WAN/LAN state
// from profile metadata and port roles. Model names do not decide gateway
// behavior.
package payload

import (
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/konstruktor1/unifi-stubd/internal/device"
)

type gatewayPayload struct {
	basePayload
	gatewayTelemetry
	IfTable           []gatewayIfRow               `json:"if_table"`
	NetworkTable      []gatewayNetworkRow          `json:"network_table"`
	ConfigPortTable   []gatewayConfigPortRow       `json:"config_port_table"`
	EthernetOverrides []gatewayEthernetOverrideRow `json:"ethernet_overrides"`
	ReportedNetworks  []gatewayReportedNetworkRow  `json:"reported_networks"`
	PortTable         []gatewayPortRow             `json:"port_table"`
	Uplink            string                       `json:"uplink"`
	UplinkTable       []gatewayUplinkRow           `json:"uplink_table"`
	HasEth1           bool                         `json:"has_eth1"`
	HasDPI            bool                         `json:"has_dpi"`
	ConfigNetworkWAN  gatewayConfigNetworkRow      `json:"config_network_wan"`
	WAN1              *gatewayWANStatusRow         `json:"wan1,omitempty"`
	ConfigNetworkWAN2 *gatewayConfigNetworkRow     `json:"config_network_wan2,omitempty"`
	WAN2              *gatewayWANStatusRow         `json:"wan2,omitempty"`
}

type connectionFields struct {
	Connected      bool                   `json:"connected"`
	LastConnection *gatewayLastConnection `json:"last_connection,omitempty"`
}

type gatewayLastConnection struct {
	MAC      string `json:"mac"`
	Source   string `json:"source"`
	IP       string `json:"ip,omitempty"`
	Hostname string `json:"hostname,omitempty"`
	Type     string `json:"type,omitempty"`
}

type gatewayNetworkStats struct {
	counterFields
	optionalRateFields
}

type gatewayIfRow struct {
	Name           string `json:"name"`
	IfName         string `json:"ifname"`
	Comment        string `json:"comment"`
	PortIdx        int    `json:"port_idx"`
	MAC            string `json:"mac"`
	IP             string `json:"ip"`
	Netmask        string `json:"netmask"`
	NumPort        int    `json:"num_port"`
	Up             bool   `json:"up"`
	Enable         bool   `json:"enable"`
	NetworkGroup   string `json:"networkgroup"`
	FullDuplex     bool   `json:"full_duplex"`
	PhysicalPorts  []int  `json:"physical_ports"`
	VLAN           int    `json:"vlan,omitempty"`
	ManagementVLAN int    `json:"management_vlan,omitempty"`
	linkFields
	counterFields
	optionalRateFields
	SourceInterface string `json:"source_interface"`
	connectionFields
}

type gatewayNetworkRow struct {
	Name            string              `json:"name"`
	IfName          string              `json:"ifname"`
	PortIdx         int                 `json:"port_idx"`
	MAC             string              `json:"mac"`
	NetworkGroup    string              `json:"networkgroup"`
	IP              string              `json:"ip"`
	Netmask         string              `json:"netmask"`
	Address         string              `json:"address"`
	Addresses       []string            `json:"addresses"`
	Up              string              `json:"up"`
	L1Up            string              `json:"l1up"`
	Autoneg         string              `json:"autoneg"`
	Duplex          string              `json:"duplex"`
	Speed           string              `json:"speed"`
	MaxSpeed        string              `json:"max_speed"`
	MTU             string              `json:"mtu"`
	Stats           gatewayNetworkStats `json:"stats"`
	SourceInterface string              `json:"source_interface"`
	connectionFields
	HostTable []gatewayHostRow `json:"host_table,omitempty"`
}

type gatewayConfigNetworkRow struct {
	Type            string  `json:"type"`
	Name            *string `json:"name,omitempty"`
	IfName          *string `json:"ifname,omitempty"`
	PortIdx         *int    `json:"port_idx,omitempty"`
	NetworkGroup    *string `json:"networkgroup,omitempty"`
	Role            *string `json:"role,omitempty"`
	MAC             *string `json:"mac,omitempty"`
	IP              *string `json:"ip,omitempty"`
	Netmask         *string `json:"netmask,omitempty"`
	Address         *string `json:"address,omitempty"`
	Up              *bool   `json:"up,omitempty"`
	Enable          *bool   `json:"enable,omitempty"`
	SourceInterface *string `json:"source_interface,omitempty"`
}

type gatewayWANStatusRow struct {
	Type         string `json:"type"`
	Name         string `json:"name"`
	IfName       string `json:"ifname"`
	PortIdx      int    `json:"port_idx"`
	NetworkGroup string `json:"networkgroup"`
	Role         string `json:"role"`
	MAC          string `json:"mac"`
	IP           string `json:"ip"`
	Netmask      string `json:"netmask"`
	Address      string `json:"address"`
	Up           bool   `json:"up"`
	Enable       bool   `json:"enable"`
	Uptime       int    `json:"uptime"`
	Latency      int    `json:"latency"`
	linkFields
	counterFields
	optionalRateFields
	SourceInterface string `json:"source_interface"`
}

type gatewayConfigPortRow struct {
	Name         string `json:"name"`
	IfName       string `json:"ifname"`
	PortIdx      int    `json:"port_idx"`
	NetworkGroup string `json:"networkgroup"`
	Role         string `json:"role"`
	Up           bool   `json:"up"`
	Enable       bool   `json:"enable"`
	IsUplink     bool   `json:"is_uplink"`
	linkFields
	SourceInterface string `json:"source_interface"`
	connectionFields
}

type gatewayPortRow struct {
	PortIdx      int                    `json:"port_idx"`
	IfName       string                 `json:"ifname"`
	Name         string                 `json:"name"`
	Enable       bool                   `json:"enable"`
	Up           bool                   `json:"up"`
	NetworkGroup string                 `json:"networkgroup"`
	Role         string                 `json:"role"`
	IsUplink     bool                   `json:"is_uplink"`
	OpMode       string                 `json:"op_mode"`
	FullDuplex   bool                   `json:"full_duplex"`
	Autoneg      bool                   `json:"autoneg"`
	FlowctrlRX   bool                   `json:"flowctrl_rx"`
	FlowctrlTX   bool                   `json:"flowctrl_tx"`
	MACTable     []device.MacTableEntry `json:"mac_table"`
	RXDropped    int                    `json:"rx_dropped"`
	TXDropped    int                    `json:"tx_dropped"`
	MAC          string                 `json:"mac,omitempty"`
	IP           string                 `json:"ip,omitempty"`
	linkFields
	counterFields
	optionalRateFields
	SourceInterface string `json:"source_interface"`
	connectionFields
}

type gatewayEthernetOverrideRow struct {
	Name         string `json:"name"`
	IfName       string `json:"ifname"`
	PortIdx      int    `json:"port_idx"`
	MAC          string `json:"mac"`
	NetworkGroup string `json:"networkgroup"`
	Role         string `json:"role"`
	Up           bool   `json:"up"`
	Enable       bool   `json:"enable"`
	linkFields
	SourceInterface string `json:"source_interface"`
	connectionFields
}

type gatewayReportedNetworkRow struct {
	Name            string   `json:"name"`
	IfName          string   `json:"ifname"`
	PortIdx         int      `json:"port_idx"`
	NetworkGroup    string   `json:"networkgroup"`
	Type            string   `json:"type"`
	IP              string   `json:"ip"`
	Netmask         string   `json:"netmask"`
	Address         string   `json:"address"`
	Addresses       []string `json:"addresses"`
	Up              bool     `json:"up"`
	SourceInterface string   `json:"source_interface"`
	connectionFields
}

type gatewayHostRow struct {
	MAC        string `json:"mac"`
	Age        int    `json:"age"`
	Authorized bool   `json:"authorized"`
	RXBytes    int64  `json:"rx_bytes"`
	TXBytes    int64  `json:"tx_bytes"`
	RXPackets  int64  `json:"rx_packets"`
	TXPackets  int64  `json:"tx_packets"`
	Uptime     int    `json:"uptime"`
	Hostname   string `json:"hostname,omitempty"`
	IP         string `json:"ip,omitempty"`
	Type       string `json:"type,omitempty"`
	VLAN       int    `json:"vlan,omitempty"`
	Static     bool   `json:"static,omitempty"`
}

type gatewayUplinkRow struct {
	Name           string `json:"name"`
	IfName         string `json:"ifname"`
	PortIdx        int    `json:"port_idx"`
	MAC            string `json:"mac"`
	Type           string `json:"type"`
	Up             bool   `json:"up"`
	Enable         bool   `json:"enable"`
	FullDuplex     bool   `json:"full_duplex"`
	VLAN           int    `json:"vlan,omitempty"`
	ManagementVLAN int    `json:"management_vlan,omitempty"`
	linkFields
	counterFields
	optionalRateFields
	SourceInterface string `json:"source_interface"`
	connectionFields
}

// buildGatewayPayload fills the gateway-specific tables expected by UniFi.
func buildGatewayPayload(base basePayload, profile device.Profile, id device.Identity, ports []PortView, now time.Time, uptime int) gatewayPayload {
	configWAN, _ := gatewayConfigNetwork(ports, gatewayPortRoleWAN)
	configWAN2, hasWAN2 := gatewayConfigNetwork(ports, gatewayPortRoleWAN2)
	payload := gatewayPayload{
		basePayload:       base,
		gatewayTelemetry:  newGatewayTelemetry(id, now, uptime, base.CFGVersion),
		IfTable:           gatewayIfTable(profile, id, ports),
		NetworkTable:      gatewayNetworkTable(profile, id, ports),
		ConfigPortTable:   gatewayConfigPortTable(ports),
		EthernetOverrides: gatewayEthernetOverrides(ports),
		ReportedNetworks:  gatewayReportedNetworks(ports),
		PortTable:         gatewayPortTable(ports),
		Uplink:            gatewayInterfaceName(profile, gatewayUplinkPortIndex(ports)),
		UplinkTable:       gatewayUplinkTable(profile, id, ports),
		HasEth1:           len(ports) > 1,
		HasDPI:            profile.Payload.HasDPI,
		ConfigNetworkWAN:  configWAN,
		WAN1:              gatewayWANStatus(ports, gatewayPortRoleWAN, uptime),
		WAN2:              gatewayWANStatus(ports, gatewayPortRoleWAN2, uptime),
	}
	if hasWAN2 {
		payload.ConfigNetworkWAN2 = &configWAN2
	}
	return payload
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
func gatewayIfTable(_ device.Profile, id device.Identity, ports []PortView) []gatewayIfRow {
	out := make([]gatewayIfRow, 0, len(ports))
	uplinkIndex := gatewayUplinkPortIndex(ports)
	for _, view := range ports {
		iface := view.GatewayInterface
		row := gatewayIfRow{
			Name:               iface.Name,
			IfName:             iface.IfName,
			Comment:            iface.Comment,
			PortIdx:            view.Index,
			MAC:                iface.MAC,
			IP:                 iface.IP,
			Netmask:            iface.Netmask,
			NumPort:            1,
			Up:                 view.Up,
			Enable:             view.Enabled,
			NetworkGroup:       iface.NetworkGroup,
			FullDuplex:         true,
			PhysicalPorts:      []int{view.Index},
			linkFields:         portLinkFields(view.Speed, view.Media),
			counterFields:      portCounterFields(view.Port),
			optionalRateFields: explicitPortRateFields(view.Port),
			SourceInterface:    view.SourceInterface,
			connectionFields:   gatewayConnectionFields(view),
		}
		if view.Index == uplinkIndex {
			row.VLAN = id.ManagementVLAN
			row.ManagementVLAN = id.ManagementVLAN
		}
		out = append(out, row)
	}
	return out
}

// gatewayNetworkTable renders the routed network view for each gateway port.
func gatewayNetworkTable(_ device.Profile, _ device.Identity, ports []PortView) []gatewayNetworkRow {
	out := make([]gatewayNetworkRow, 0, len(ports))
	for _, view := range ports {
		iface := view.GatewayInterface
		entry := gatewayNetworkRow{
			Name:             iface.Name,
			IfName:           iface.IfName,
			PortIdx:          view.Index,
			MAC:              iface.MAC,
			NetworkGroup:     iface.NetworkGroup,
			IP:               iface.IP,
			Netmask:          iface.Netmask,
			Address:          iface.Address,
			Addresses:        []string{iface.Address},
			Up:               boolText(view.Up),
			L1Up:             boolText(view.Up),
			Autoneg:          "true",
			Duplex:           "full",
			Speed:            strconv.Itoa(view.Speed),
			MaxSpeed:         strconv.Itoa(view.Speed),
			MTU:              "1500",
			Stats:            gatewayNetworkStats{counterFields: portCounterFields(view.Port), optionalRateFields: explicitPortRateFields(view.Port)},
			SourceInterface:  view.SourceInterface,
			connectionFields: gatewayConnectionFields(view),
		}
		if hosts := gatewayHostTable(view.Port, view.Uplink); len(hosts) > 0 {
			entry.HostTable = hosts
		}
		out = append(out, entry)
	}
	return out
}

// gatewayConfigNetwork renders the controller-owned network assignment view
// for the first port matching the requested WAN/LAN role.
func gatewayConfigNetwork(ports []PortView, role string) (gatewayConfigNetworkRow, bool) {
	for _, view := range ports {
		if gatewayPortRole(view.Port) != role {
			continue
		}
		iface := view.GatewayInterface
		return gatewayConfigNetworkRow{
			Type:            payloadTypeDHCP,
			Name:            stringRef(iface.NetworkGroup),
			IfName:          stringRef(iface.IfName),
			PortIdx:         intRef(view.Index),
			NetworkGroup:    stringRef(iface.NetworkGroup),
			Role:            stringRef(view.Role),
			MAC:             stringRef(iface.MAC),
			IP:              stringRef(iface.IP),
			Netmask:         stringRef(iface.Netmask),
			Address:         stringRef(iface.Address),
			Up:              boolRef(view.Up),
			Enable:          boolRef(view.Enabled),
			SourceInterface: stringRef(view.SourceInterface),
		}, true
	}
	return gatewayConfigNetworkRow{Type: payloadTypeDHCP}, false
}

// gatewayWANStatus renders live WAN-like state from the same resolved port view
// used by the gateway interface tables.
func gatewayWANStatus(ports []PortView, role string, uptime int) *gatewayWANStatusRow {
	for _, view := range ports {
		if gatewayPortRole(view.Port) != role {
			continue
		}
		iface := view.GatewayInterface
		return &gatewayWANStatusRow{
			Type:               payloadTypeDHCP,
			Name:               iface.NetworkGroup,
			IfName:             iface.IfName,
			PortIdx:            view.Index,
			NetworkGroup:       iface.NetworkGroup,
			Role:               view.Role,
			MAC:                iface.MAC,
			IP:                 iface.IP,
			Netmask:            iface.Netmask,
			Address:            iface.Address,
			Up:                 view.Up,
			Enable:             view.Enabled,
			Uptime:             uptime,
			Latency:            0,
			linkFields:         portLinkFields(view.Speed, view.Media),
			counterFields:      portCounterFields(view.Port),
			optionalRateFields: explicitPortRateFields(view.Port),
			SourceInterface:    view.SourceInterface,
		}
	}
	return nil
}

// gatewayConnectionFields uses the resolved port view as the single source for
// gateway connection and topology state.
func gatewayConnectionFields(view PortView) connectionFields {
	out := connectionFields{Connected: view.Up}
	if !view.Up || len(view.MACs) == 0 {
		return out
	}
	// Controllers use last_connection as a topology hint. The first MAC entry
	// is therefore treated as metadata about the visible neighbor, not as host
	// configuration to apply.
	entry := view.MACs[0]
	connection := gatewayLastConnection{
		MAC:    strings.ToLower(strings.TrimSpace(entry.MAC)),
		Source: jsonKeyMACTable,
	}
	if ip := strings.TrimSpace(entry.IP); ip != "" {
		connection.IP = ip
	}
	if hostname := strings.TrimSpace(entry.Hostname); hostname != "" {
		connection.Hostname = hostname
	}
	if entryType := strings.TrimSpace(entry.Type); entryType != "" {
		connection.Type = entryType
	}
	out.LastConnection = &connection
	return out
}

// gatewayConfigPortTable renders gateway WAN/LAN port assignments from the
// same resolved port view used by interface and network tables.
func gatewayConfigPortTable(ports []PortView) []gatewayConfigPortRow {
	out := make([]gatewayConfigPortRow, 0, len(ports))
	for _, view := range ports {
		iface := view.GatewayInterface
		row := gatewayConfigPortRow{
			Name:             view.Name,
			IfName:           iface.IfName,
			PortIdx:          view.Index,
			NetworkGroup:     view.NetworkGroup,
			Role:             view.Role,
			Up:               view.Up,
			Enable:           view.Enabled,
			IsUplink:         view.Uplink,
			linkFields:       portLinkFields(view.Speed, view.Media),
			SourceInterface:  view.SourceInterface,
			connectionFields: gatewayConnectionFields(view),
		}
		out = append(out, row)
	}
	return out
}

// gatewayPortTable renders physical gateway ports for controller views that
// treat UXG/UDM ports as switch-like rows. It intentionally avoids port profile
// or VLAN assignments; those remain controller-owned configuration.
func gatewayPortTable(ports []PortView) []gatewayPortRow {
	out := make([]gatewayPortRow, 0, len(ports))
	for _, view := range ports {
		iface := view.GatewayInterface
		row := gatewayPortRow{
			PortIdx:            view.Index,
			IfName:             iface.IfName,
			Name:               view.Name,
			Enable:             view.Enabled,
			Up:                 view.Up,
			NetworkGroup:       iface.NetworkGroup,
			Role:               view.Role,
			IsUplink:           view.Uplink,
			OpMode:             payloadKindGateway,
			FullDuplex:         true,
			Autoneg:            true,
			FlowctrlRX:         false,
			FlowctrlTX:         false,
			MACTable:           view.MACs,
			RXDropped:          0,
			TXDropped:          0,
			MAC:                iface.MAC,
			IP:                 iface.IP,
			linkFields:         portLinkFields(view.Speed, view.Media),
			counterFields:      portCounterFields(view.Port),
			optionalRateFields: explicitPortRateFields(view.Port),
			SourceInterface:    view.SourceInterface,
			connectionFields:   gatewayConnectionFields(view),
		}
		out = append(out, row)
	}
	return out
}

// gatewayEthernetOverrides renders the interface binding data that gateway
// controllers use for port remapping and visual WAN/LAN state.
func gatewayEthernetOverrides(ports []PortView) []gatewayEthernetOverrideRow {
	out := make([]gatewayEthernetOverrideRow, 0, len(ports))
	for _, view := range ports {
		iface := view.GatewayInterface
		row := gatewayEthernetOverrideRow{
			Name:             iface.Name,
			IfName:           iface.IfName,
			PortIdx:          view.Index,
			MAC:              iface.MAC,
			NetworkGroup:     view.NetworkGroup,
			Role:             view.Role,
			Up:               view.Up,
			Enable:           view.Enabled,
			linkFields:       portLinkFields(view.Speed, view.Media),
			SourceInterface:  view.SourceInterface,
			connectionFields: gatewayConnectionFields(view),
		}
		out = append(out, row)
	}
	return out
}

// gatewayReportedNetworks renders a read-only network summary per gateway
// port. It mirrors network_table values without inventing host configuration.
func gatewayReportedNetworks(ports []PortView) []gatewayReportedNetworkRow {
	out := make([]gatewayReportedNetworkRow, 0, len(ports))
	for _, view := range ports {
		iface := view.GatewayInterface
		row := gatewayReportedNetworkRow{
			Name:             iface.NetworkGroup,
			IfName:           iface.IfName,
			PortIdx:          view.Index,
			NetworkGroup:     iface.NetworkGroup,
			Type:             view.Role,
			IP:               iface.IP,
			Netmask:          iface.Netmask,
			Address:          iface.Address,
			Addresses:        []string{iface.Address},
			Up:               view.Up,
			SourceInterface:  view.SourceInterface,
			connectionFields: gatewayConnectionFields(view),
		}
		out = append(out, row)
	}
	return out
}

// gatewayHostTable renders learned downstream MACs for one gateway port.
func gatewayHostTable(port device.Port, uplink bool) []gatewayHostRow {
	out := make([]gatewayHostRow, 0, len(port.MACs))
	for _, entry := range port.MACs {
		entryType := strings.TrimSpace(entry.Type)
		if uplink && entryType != "" && entryType != "client" {
			// Uplink neighbor metadata belongs in uplink/last_connection fields.
			// The gateway host table should contain downstream client-like MACs.
			continue
		}
		row := gatewayHostRow{
			MAC:        strings.ToLower(strings.TrimSpace(entry.MAC)),
			Age:        entry.Age,
			Authorized: true,
			RXBytes:    port.RXBytes,
			TXBytes:    port.TXBytes,
			RXPackets:  firstNonZeroInt64(port.RXPackets, 1),
			TXPackets:  firstNonZeroInt64(port.TXPackets, 1),
			Uptime:     firstNonZero(entry.Uptime, 1200),
		}
		if hostname := strings.TrimSpace(entry.Hostname); hostname != "" {
			row.Hostname = hostname
		}
		if ip := strings.TrimSpace(entry.IP); ip != "" {
			row.IP = ip
		}
		if entryType != "" {
			row.Type = entryType
		}
		if entry.VLAN > 0 {
			row.VLAN = entry.VLAN
		}
		if entry.Static {
			row.Static = true
		}
		out = append(out, row)
	}
	return out
}

// gatewayUplinkTable renders the controller-facing uplink entry.
func gatewayUplinkTable(_ device.Profile, id device.Identity, ports []PortView) []gatewayUplinkRow {
	uplinkIndex := gatewayUplinkPortIndex(ports)
	for _, view := range ports {
		if view.Index != uplinkIndex {
			continue
		}
		iface := view.GatewayInterface
		row := gatewayUplinkRow{
			Name:               iface.Name,
			IfName:             iface.IfName,
			PortIdx:            view.Index,
			MAC:                iface.MAC,
			Type:               "wire",
			Up:                 view.Up,
			Enable:             view.Enabled,
			FullDuplex:         true,
			VLAN:               id.ManagementVLAN,
			ManagementVLAN:     id.ManagementVLAN,
			linkFields:         portLinkFields(view.Speed, view.Media),
			counterFields:      portCounterFields(view.Port),
			optionalRateFields: explicitPortRateFields(view.Port),
			SourceInterface:    view.SourceInterface,
			connectionFields:   gatewayConnectionFields(view),
		}
		return []gatewayUplinkRow{row}
	}
	return nil
}

// gatewayPortRole returns profile or override data before generic fallback roles.
func gatewayPortRole(port device.Port) string {
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
func gatewayNetworkGroup(port device.Port) string {
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
func gatewayInterfaceName(profile device.Profile, portIndex int) string {
	if portIndex < 1 {
		portIndex = 1
	}
	prefix := strings.TrimSpace(profile.Payload.GatewayInterfacePrefix)
	if prefix == "" {
		prefix = "eth"
	}
	return prefix + strconv.Itoa(portIndex-1)
}

// gatewayPortMAC returns a configured port MAC or derives one from the device MAC.
func gatewayPortMAC(baseMAC string, port device.Port) string {
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
func gatewayInterfaceIP(id device.Identity, port device.Port) string {
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
func gatewayInterfaceNetmask(port device.Port) string {
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

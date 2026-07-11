// Package payload converts generated ports into UniFi switch tables such as
// port_table and if_table. Profile selection is complete before switch rendering
// runs, so this code only handles switch payload shape.
package payload

import (
	"net"
	"strconv"
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/device"
)

type switchInterfaceRow struct {
	Name           string `json:"name"`
	MAC            string `json:"mac"`
	IP             string `json:"ip"`
	NumPort        int    `json:"num_port"`
	Up             bool   `json:"up"`
	Speed          int    `json:"speed"`
	FullDuplex     bool   `json:"full_duplex"`
	VLAN           int    `json:"vlan,omitempty"`
	ManagementVLAN int    `json:"management_vlan,omitempty"`
}

type switchEthernetRow struct {
	Name    string `json:"name"`
	MAC     string `json:"mac"`
	NumPort *int   `json:"num_port,omitempty"`
}

type switchPortRow struct {
	PortIdx        int                    `json:"port_idx"`
	IfName         string                 `json:"ifname"`
	Name           string                 `json:"name"`
	Enable         bool                   `json:"enable"`
	Up             bool                   `json:"up"`
	IsUplink       bool                   `json:"is_uplink"`
	OpMode         string                 `json:"op_mode"`
	FullDuplex     bool                   `json:"full_duplex"`
	Autoneg        bool                   `json:"autoneg"`
	FlowctrlRX     bool                   `json:"flowctrl_rx"`
	FlowctrlTX     bool                   `json:"flowctrl_tx"`
	PortPOE        bool                   `json:"port_poe"`
	POEEnable      bool                   `json:"poe_enable"`
	POECaps        int                    `json:"poe_caps"`
	RXDropped      int                    `json:"rx_dropped"`
	TXDropped      int                    `json:"tx_dropped"`
	Satisfaction   int                    `json:"satisfaction"`
	STPState       string                 `json:"stp_state"`
	STPPathcost    int                    `json:"stp_pathcost"`
	MACTable       []device.MacTableEntry `json:"mac_table"`
	LastConnection *switchLastConnection  `json:"last_connection"`
	linkFields
	counterFields
	BytesRate       int64  `json:"bytes-r"`
	RXBytesRate     int64  `json:"rx_bytes-r"`
	TXBytesRate     int64  `json:"tx_bytes-r"`
	SourceInterface string `json:"source_interface"`
}

type switchLastConnection struct {
	Connected bool   `json:"connected"`
	MAC       string `json:"mac"`
	IP        string `json:"ip,omitempty"`
	Hostname  string `json:"hostname,omitempty"`
	Type      string `json:"type,omitempty"`
}

type switchUplinkRow struct {
	Name             string `json:"name"`
	IP               string `json:"ip"`
	MAC              string `json:"mac"`
	NumPort          int    `json:"num_port"`
	Type             string `json:"type"`
	PortIdx          int    `json:"port_idx"`
	Up               bool   `json:"up"`
	FullDuplex       bool   `json:"full_duplex"`
	UplinkMAC        string `json:"uplink_mac"`
	UplinkDeviceName string `json:"uplink_device_name,omitempty"`
	UplinkRemotePort int    `json:"uplink_remote_port,omitempty"`
	UplinkSource     string `json:"uplink_source,omitempty"`
	linkFields
	counterFields
	BytesRate   int64 `json:"bytes-r"`
	RXBytesRate int64 `json:"rx_bytes-r"`
	TXBytesRate int64 `json:"tx_bytes-r"`
}

type switchLastUplinkRow struct {
	PortIdx          int    `json:"port_idx"`
	UplinkMAC        string `json:"uplink_mac"`
	UplinkRemotePort int    `json:"uplink_remote_port,omitempty"`
	Type             string `json:"type"`
}

type switchPayload struct {
	basePayload
	IfTable       []switchInterfaceRow `json:"if_table"`
	EthernetTable []switchEthernetRow  `json:"ethernet_table"`
	PortTable     []switchPortRow      `json:"port_table"`
	Uplink        *switchUplinkRow     `json:"uplink,omitempty"`
	LastUplink    *switchLastUplinkRow `json:"last_uplink,omitempty"`
	LLDPTable     []lldpRow            `json:"lldp_table"`
}

// buildSwitchPayload fills the tables expected by UniFi switch devices.
func buildSwitchPayload(base basePayload, profile device.Profile, id device.Identity, ports []PortView, numPorts int, ifSpeed int) switchPayload {
	ifaceName := profile.Payload.ManagementInterface
	iface := switchInterfaceRow{
		Name:           ifaceName,
		MAC:            id.MAC,
		IP:             id.IP,
		NumPort:        numPorts,
		Up:             true,
		Speed:          ifSpeed,
		FullDuplex:     true,
		VLAN:           id.ManagementVLAN,
		ManagementVLAN: id.ManagementVLAN,
	}
	uplink, lastUplink := switchUplinkRows(profile, id, ports, numPorts)
	return switchPayload{
		basePayload: base,
		IfTable:     []switchInterfaceRow{iface},
		EthernetTable: []switchEthernetRow{
			{
				Name:    ifaceName,
				MAC:     id.MAC,
				NumPort: intRef(numPorts),
			},
			{
				// srv0 is a synthetic secondary interface seen by controllers on
				// switch-like payloads. It is derived from the fake MAC and does not
				// represent a host interface.
				Name: "srv0",
				MAC:  incrementMAC(id.MAC),
			},
		},
		PortTable:  portTable(ports),
		Uplink:     uplink,
		LastUplink: lastUplink,
		LLDPTable:  lldpTable(ports),
	}
}

func switchUplinkRows(profile device.Profile, id device.Identity, ports []PortView, numPorts int) (*switchUplinkRow, *switchLastUplinkRow) {
	for _, port := range ports {
		if port.TopologyNeighbor == nil || port.TopologyDirection != device.DirectionUpstream {
			continue
		}
		neighbor := *port.TopologyNeighbor
		rxRate, txRate := portRateFields(port.Port)
		name := strings.TrimSpace(profile.Payload.ManagementInterface)
		if name == "" {
			name = "eth0"
		}
		source := ""
		if neighbor.Source == device.NeighborSourceLLDP {
			source = "lldp_uplink"
		}
		uplink := &switchUplinkRow{
			Name:             name,
			IP:               id.IP,
			MAC:              strings.ToLower(strings.TrimSpace(id.MAC)),
			NumPort:          numPorts,
			Type:             payloadTypeWire,
			PortIdx:          port.Index,
			Up:               port.Up,
			FullDuplex:       true,
			UplinkMAC:        strings.ToLower(strings.TrimSpace(neighbor.MAC)),
			UplinkDeviceName: strings.TrimSpace(neighbor.Hostname),
			UplinkRemotePort: neighbor.RemotePort,
			UplinkSource:     source,
			linkFields:       portLinkFields(port.Speed, port.Media),
			counterFields:    portCounterFields(port.Port),
			BytesRate:        rxRate + txRate,
			RXBytesRate:      rxRate,
			TXBytesRate:      txRate,
		}
		last := &switchLastUplinkRow{
			PortIdx:          port.Index,
			UplinkMAC:        uplink.UplinkMAC,
			UplinkRemotePort: neighbor.RemotePort,
			Type:             payloadTypeWire,
		}
		return uplink, last
	}
	return nil, nil
}

// incrementMAC derives the secondary switch interface MAC from the device MAC.
func incrementMAC(macText string) string {
	mac, err := net.ParseMAC(macText)
	if err != nil || len(mac) == 0 {
		return macText
	}
	out := append(net.HardwareAddr{}, mac...)
	out[len(out)-1]++
	return out.String()
}

// portTable renders switch port rows in the shape expected by UniFi Network.
func portTable(ports []PortView) []switchPortRow {
	out := make([]switchPortRow, 0, len(ports))
	for _, p := range ports {
		rxRate, txRate := portRateFields(p.Port)
		macTable := downstreamMACTable(p)
		// Every switch row is rendered from PortView so switch and gateway
		// payloads agree on observed counters, source interface, link state, and
		// MAC-table metadata.
		row := switchPortRow{
			PortIdx:         p.Index,
			IfName:          p.SwitchInterfaceName,
			Name:            p.Name,
			Enable:          p.Enabled,
			Up:              p.Up,
			IsUplink:        p.Uplink,
			OpMode:          payloadModeSwitch,
			FullDuplex:      true,
			Autoneg:         true,
			FlowctrlRX:      false,
			FlowctrlTX:      false,
			PortPOE:         false,
			POEEnable:       false,
			POECaps:         0,
			RXDropped:       0,
			TXDropped:       0,
			Satisfaction:    100,
			STPState:        "forwarding",
			STPPathcost:     20000,
			MACTable:        macTable,
			LastConnection:  switchLastConnectionFor(macTable),
			linkFields:      portLinkFields(p.Speed, p.Media),
			counterFields:   portCounterFields(p.Port),
			BytesRate:       rxRate + txRate,
			RXBytesRate:     rxRate,
			TXBytesRate:     txRate,
			SourceInterface: p.SourceInterface,
		}
		out = append(out, row)
	}
	return out
}

func switchLastConnectionFor(entries []device.MacTableEntry) *switchLastConnection {
	if len(entries) == 0 {
		return nil
	}
	entry := entries[0]
	return &switchLastConnection{
		Connected: true,
		MAC:       entry.MAC,
		IP:        entry.IP,
		Hostname:  entry.Hostname,
		Type:      entry.Type,
	}
}

// managementInterfaceSpeed chooses a stable management speed from generated ports.
func managementInterfaceSpeed(ports []device.Port) int {
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

func managementSpeed(ports []device.Port) int {
	if speed := managementInterfaceSpeed(ports); speed > 0 {
		return speed
	}
	return 1000
}

// switchInterfaceName maps one-based UniFi ports to zero-based eth names used
// in switch payload rows.
func switchInterfaceName(portIndex int) string {
	if portIndex < 1 {
		portIndex = 1
	}
	return "eth" + strconv.Itoa(portIndex-1)
}

// speedCaps returns controller speed capabilities implied by speed and media.
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

// firstNonZeroInt64 returns the first non-zero value from a fallback list.
func firstNonZeroInt64(values ...int64) int64 {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}

// firstNonZero returns the first non-zero value from a fallback list.
func firstNonZero(values ...int) int {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}

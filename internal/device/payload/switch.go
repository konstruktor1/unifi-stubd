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
	counterFields
	optionalRateFields
	gatewayRateFields
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
	Satisfaction   int                    `json:"satisfaction"`
	STPState       string                 `json:"stp_state"`
	STPPathcost    int                    `json:"stp_pathcost"`
	MACTable       []device.MacTableEntry `json:"mac_table"`
	LastConnection *switchLastConnection  `json:"last_connection"`
	linkFields
	counterFields
	optionalRateFields
	gatewayRateFields
	SourceInterface string `json:"source_interface,omitempty"`
}

type switchLastConnection struct {
	Connected bool   `json:"connected"`
	MAC       string `json:"mac"`
	IP        string `json:"ip,omitempty"`
	Hostname  string `json:"hostname,omitempty"`
	Type      string `json:"type,omitempty"`
}

type switchLastUplinkRow struct {
	PortIdx          int    `json:"port_idx"`
	UplinkMAC        string `json:"uplink_mac"`
	UplinkRemotePort int    `json:"uplink_remote_port,omitempty"`
	Type             string `json:"type"`
}

// switchTrafficSummary is optional because switch root traffic is meaningful
// only when the selected uplink has explicit rate reporting enabled. Pointers
// preserve explicit zero samples while omitting the entire summary otherwise.
type switchTrafficSummary struct {
	Bytes       *int64 `json:"bytes,omitempty"`
	RXBytes     *int64 `json:"rx_bytes,omitempty"`
	TXBytes     *int64 `json:"tx_bytes,omitempty"`
	BytesRate   *int64 `json:"bytes-r,omitempty"`
	RXBytesRate *int64 `json:"rx_bytes-r,omitempty"`
	TXBytesRate *int64 `json:"tx_bytes-r,omitempty"`
	RXRate      *int64 `json:"rx_rate,omitempty"`
	TXRate      *int64 `json:"tx_rate,omitempty"`
}

type switchPayload struct {
	basePayload
	switchTrafficSummary
	IfTable       []switchInterfaceRow `json:"if_table"`
	EthernetTable []switchEthernetRow  `json:"ethernet_table"`
	PortTable     []switchPortRow      `json:"port_table"`
	Uplink        string               `json:"uplink"`
	LastUplink    *switchLastUplinkRow `json:"last_uplink,omitempty"`
	LLDPTable     []lldpRow            `json:"lldp_table"`
}

// buildSwitchPayload fills the tables expected by UniFi switch devices.
func buildSwitchPayload(base basePayload, profile device.Profile, id device.Identity, ports []PortView, numPorts int, ifSpeed int) switchPayload {
	ifaceName := profile.Payload.ManagementInterface
	if strings.TrimSpace(ifaceName) == "" {
		ifaceName = "eth0"
	}
	iface := switchManagementInterfaceRow(ifaceName, id, ports, numPorts, ifSpeed)
	lastUplink := switchLastUplinkRowFor(ports)
	return switchPayload{
		basePayload:          base,
		switchTrafficSummary: switchTrafficSummaryFor(ports),
		IfTable:              []switchInterfaceRow{iface},
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
		Uplink:     ifaceName,
		LastUplink: lastUplink,
		LLDPTable:  lldpTable(ports),
	}
}

func switchManagementInterfaceRow(ifaceName string, id device.Identity, ports []PortView, numPorts int, ifSpeed int) switchInterfaceRow {
	byteRates, bitRates := switchInterfaceRatesFor(ports)
	return switchInterfaceRow{
		Name:               ifaceName,
		MAC:                id.MAC,
		IP:                 id.IP,
		NumPort:            numPorts,
		Up:                 true,
		Speed:              ifSpeed,
		FullDuplex:         true,
		VLAN:               id.ManagementVLAN,
		ManagementVLAN:     id.ManagementVLAN,
		counterFields:      switchInterfaceCountersFor(ports),
		optionalRateFields: byteRates,
		gatewayRateFields:  bitRates,
	}
}

// switchInterfaceCountersFor mirrors the selected physical uplink counters
// into the management interface named by the wire-level uplink selector. This
// matches real switch if_table payloads: physical direction and cumulative
// byte, packet, and error counters, without port-local rate extensions.
func switchInterfaceCountersFor(ports []PortView) counterFields {
	for _, view := range ports {
		if view.Uplink {
			return portCounterFields(view.Port)
		}
	}
	return counterFields{}
}

// switchInterfaceRatesFor projects explicit selected-uplink rates onto the
// interface named by uplink. Network 10.5 reads its normalized live uplink
// rate from that interface path. The cumulative counter shape remains the
// native switch contract; these rate fields are an additive compatibility
// projection and are never synthesized.
func switchInterfaceRatesFor(ports []PortView) (optionalRateFields, gatewayRateFields) {
	for _, view := range ports {
		if !view.Uplink {
			continue
		}
		port := view.Port
		if !port.TrafficRatesSet {
			return optionalRateFields{}, gatewayRateFields{}
		}
		rxByteRate, txByteRate := portRateFields(port)
		return optionalRateFields{
				BytesRate:   int64Ref(rxByteRate + txByteRate),
				RXBytesRate: int64Ref(rxByteRate),
				TXBytesRate: int64Ref(txByteRate),
			}, gatewayRateFields{
				RXRate: int64Ref(bitsPerSecond(rxByteRate)),
				TXRate: int64Ref(bitsPerSecond(txByteRate)),
			}
	}
	return optionalRateFields{}, gatewayRateFields{}
}

func switchLastUplinkRowFor(ports []PortView) *switchLastUplinkRow {
	for _, port := range ports {
		if port.TopologyNeighbor == nil || port.TopologyDirection != device.DirectionUpstream {
			continue
		}
		neighbor := *port.TopologyNeighbor
		return &switchLastUplinkRow{
			PortIdx:          port.Index,
			UplinkMAC:        strings.ToLower(strings.TrimSpace(neighbor.MAC)),
			UplinkRemotePort: neighbor.RemotePort,
			Type:             payloadTypeWire,
		}
	}
	return nil
}

// switchTrafficSummaryFor reports only the selected uplink. Per-port rows use
// interface-local RX/TX direction, while the controller root convention is
// device-facing: physical uplink RX is Download/root TX and physical uplink TX
// is Upload/root RX. No synthetic heartbeat rates are used at root level.
func switchTrafficSummaryFor(ports []PortView) switchTrafficSummary {
	for _, view := range ports {
		if !view.Uplink {
			continue
		}
		port := view.Port
		if !port.CountersSet && !port.TrafficRatesSet {
			return switchTrafficSummary{}
		}
		out := switchTrafficSummary{}
		if port.CountersSet {
			out.Bytes = int64Ref(port.RXBytes + port.TXBytes)
			out.RXBytes = int64Ref(port.TXBytes)
			out.TXBytes = int64Ref(port.RXBytes)
		}
		if port.TrafficRatesSet {
			rxByteRate := int64(0)
			txByteRate := int64(0)
			if view.Up {
				rxByteRate = port.TXBytesRate
				txByteRate = port.RXBytesRate
			}
			out.BytesRate = int64Ref(rxByteRate + txByteRate)
			out.RXBytesRate = int64Ref(rxByteRate)
			out.TXBytesRate = int64Ref(txByteRate)
			out.RXRate = int64Ref(bitsPerSecond(rxByteRate))
			out.TXRate = int64Ref(bitsPerSecond(txByteRate))
		}
		return out
	}
	return switchTrafficSummary{}
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
		byteRates := explicitPortRateFields(p.Port)
		bitRates := gatewayRateFields{}
		if p.Uplink && p.Port.TrafficRatesSet {
			rxRate, txRate := portRateFields(p.Port)
			bitRates = gatewayRateFields{
				RXRate: int64Ref(bitsPerSecond(rxRate)),
				TXRate: int64Ref(bitsPerSecond(txRate)),
			}
		}
		macTable := downstreamMACTable(p)
		// Every switch row is rendered from PortView so switch and gateway
		// payloads agree on observed counters, source interface, link state, and
		// MAC-table metadata.
		link := portLinkFields(p.Speed, p.Media)
		if p.Uplink && p.Port.Speed <= 0 {
			// The controller normalizes its current uplink from the selected
			// port_table row. Preserve an unknown negotiated speed there instead
			// of leaking the generic active-port fallback into uplink.speed.
			link = portLinkFields(0, effectivePortMedia(p.Port, 0))
		}
		row := switchPortRow{
			PortIdx:            p.Index,
			IfName:             p.SwitchInterfaceName,
			Name:               p.Name,
			Enable:             p.Enabled,
			Up:                 p.Up,
			IsUplink:           p.Uplink,
			OpMode:             payloadModeSwitch,
			FullDuplex:         true,
			Autoneg:            true,
			FlowctrlRX:         false,
			FlowctrlTX:         false,
			PortPOE:            false,
			POEEnable:          false,
			POECaps:            0,
			Satisfaction:       100,
			STPState:           "forwarding",
			STPPathcost:        20000,
			MACTable:           macTable,
			LastConnection:     switchLastConnectionFor(macTable),
			linkFields:         link,
			counterFields:      portCounterFields(p.Port),
			optionalRateFields: byteRates,
			gatewayRateFields:  bitRates,
			SourceInterface:    p.SourceInterface,
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

// firstNonZero returns the first non-zero value from a fallback list.
func firstNonZero(values ...int) int {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}

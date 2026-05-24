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
	PortIdx      int                    `json:"port_idx"`
	IfName       string                 `json:"ifname"`
	Name         string                 `json:"name"`
	Enable       bool                   `json:"enable"`
	Up           bool                   `json:"up"`
	IsUplink     bool                   `json:"is_uplink"`
	OpMode       string                 `json:"op_mode"`
	FullDuplex   bool                   `json:"full_duplex"`
	Autoneg      bool                   `json:"autoneg"`
	FlowctrlRX   bool                   `json:"flowctrl_rx"`
	FlowctrlTX   bool                   `json:"flowctrl_tx"`
	PortPOE      bool                   `json:"port_poe"`
	POEEnable    bool                   `json:"poe_enable"`
	POECaps      int                    `json:"poe_caps"`
	RXDropped    int                    `json:"rx_dropped"`
	TXDropped    int                    `json:"tx_dropped"`
	Satisfaction int                    `json:"satisfaction"`
	STPState     string                 `json:"stp_state"`
	STPPathcost  int                    `json:"stp_pathcost"`
	MACTable     []device.MacTableEntry `json:"mac_table"`
	linkFields
	counterFields
	RXBytesRate     int64  `json:"rx_bytes-r"`
	TXBytesRate     int64  `json:"tx_bytes-r"`
	SourceInterface string `json:"source_interface"`
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
			MACTable:        p.MACs,
			linkFields:      portLinkFields(p.Speed, p.Media),
			counterFields:   portCounterFields(p.Port),
			RXBytesRate:     rxRate,
			TXBytesRate:     txRate,
			SourceInterface: p.SourceInterface,
		}
		out = append(out, row)
	}
	return out
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

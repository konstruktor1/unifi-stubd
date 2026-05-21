// Package payload converts generated ports into UniFi switch tables such as
// port_table and if_table. Profile selection is complete before switch rendering
// runs, so this code only handles switch payload shape.
package payload

import (
	"net"
	"strconv"
	"strings"
)

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
func portTable(ports []PortView) []map[string]any {
	out := make([]map[string]any, 0, len(ports))
	for _, p := range ports {
		row := map[string]any{
			jsonKeyPortIdx:    p.Index,
			jsonKeyIfName:     p.SwitchInterfaceName,
			jsonKeyName:       p.Name,
			jsonKeyEnable:     p.Enabled,
			jsonKeyUp:         p.Up,
			jsonKeyIsUplink:   p.Uplink,
			"op_mode":         payloadModeSwitch,
			jsonKeyFullDuplex: true,
			jsonKeyAutoneg:    true,
			"flowctrl_rx":     false,
			"flowctrl_tx":     false,
			"port_poe":        false,
			"poe_enable":      false,
			"poe_caps":        0,
			"rx_dropped":      0,
			"tx_dropped":      0,
			"satisfaction":    100,
			"stp_state":       "forwarding",
			"stp_pathcost":    20000,
			jsonKeyMACTable:   p.MACs,
		}
		addFields(row, portLinkFields(p.Speed, p.Media), portCounterFields(p.Port), portRateFields(p.Port), sourceFields(p.SourceInterface))
		out = append(out, row)
	}
	return out
}

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

package payload

// This file renders switch-specific inform tables from generated ports.

import (
	"net"
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
			jsonKeyMaxSpeed:   speed,
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
			jsonKeySourceIf:   p.Interface,
		})
	}
	return out
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

// Package payload normalizes link, counter, and source-interface fragments for
// both switch and gateway renderers. Centralizing them keeps renderer logic
// data-driven rather than model-name driven.
package payload

import "strings"

func effectivePortSpeed(port Port) int {
	speed := port.Speed
	if port.Up && speed <= 0 {
		return 1000
	}
	return speed
}

func effectivePortMedia(port Port, speed int) string {
	media := strings.TrimSpace(port.Media)
	if media == "" && speed > 0 {
		return mediaForSpeed(speed)
	}
	return media
}

func addFields(row map[string]any, fields ...map[string]any) {
	for _, fieldSet := range fields {
		for key, value := range fieldSet {
			row[key] = value
		}
	}
}

func portLinkFields(speed int, media string) map[string]any {
	return map[string]any{
		jsonKeySpeed:     speed,
		jsonKeyMaxSpeed:  speed,
		jsonKeySpeedCaps: speedCaps(speed, media),
		jsonKeyMedia:     media,
	}
}

func portCounterFields(port Port) map[string]any {
	return map[string]any{
		jsonKeyRXBytes:   port.RXBytes,
		jsonKeyTXBytes:   port.TXBytes,
		jsonKeyRXPackets: firstNonZeroInt64(port.RXPackets, 1),
		jsonKeyTXPackets: firstNonZeroInt64(port.TXPackets, 1),
		jsonKeyRXErrors:  port.RXErrors,
		jsonKeyTXErrors:  port.TXErrors,
	}
}

func portRateFields(port Port) map[string]any {
	rxRate := int64(0)
	txRate := int64(0)
	if port.Up && effectivePortSpeed(port) > 0 {
		rxRate = 64 + int64(port.Index)
		txRate = 48 + int64(port.Index)
	}
	return map[string]any{
		"rx_bytes-r": rxRate,
		"tx_bytes-r": txRate,
	}
}

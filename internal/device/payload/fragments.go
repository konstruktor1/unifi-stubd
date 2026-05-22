// Package payload normalizes link, counter, and source-interface fragments for
// both switch and gateway renderers. Centralizing them keeps renderer logic
// data-driven rather than model-name driven.
package payload

import "strings"

// effectivePortSpeed keeps connected synthetic ports from rendering as
// zero-speed links.
func effectivePortSpeed(port Port) int {
	speed := port.Speed
	if port.Up && speed <= 0 {
		return 1000
	}
	return speed
}

// effectivePortMedia derives a controller media label from speed unless the
// profile or override supplied one explicitly.
func effectivePortMedia(port Port, speed int) string {
	media := strings.TrimSpace(port.Media)
	if media == "" && speed > 0 {
		return mediaForSpeed(speed)
	}
	return media
}

// addFields merges optional renderer fragments into one payload row.
func addFields(row map[string]any, fields ...map[string]any) {
	for _, fieldSet := range fields {
		for key, value := range fieldSet {
			row[key] = value
		}
	}
}

// portLinkFields renders the common link speed, capability, and media fields
// shared by switch and gateway tables.
func portLinkFields(speed int, media string) map[string]any {
	return map[string]any{
		jsonKeySpeed:     speed,
		jsonKeyMaxSpeed:  speed,
		jsonKeySpeedCaps: speedCaps(speed, media),
		jsonKeyMedia:     media,
	}
}

// portCounterFields renders raw counters plus packet fallbacks shared by switch
// and gateway payload rows.
func portCounterFields(port Port) map[string]any {
	// Packet counters stay non-zero for synthetic connected ports because some
	// controller views treat all-zero rows as stale.
	return map[string]any{
		jsonKeyRXBytes:   port.RXBytes,
		jsonKeyTXBytes:   port.TXBytes,
		jsonKeyRXPackets: firstNonZeroInt64(port.RXPackets, 1),
		jsonKeyTXPackets: firstNonZeroInt64(port.TXPackets, 1),
		jsonKeyRXErrors:  port.RXErrors,
		jsonKeyTXErrors:  port.TXErrors,
	}
}

// portRateFields returns explicit observed rates when available, otherwise a
// synthetic low heartbeat for connected synthetic ports.
func portRateFields(port Port) map[string]any {
	if port.TrafficRatesSet || port.TrafficRatesEnabled {
		return map[string]any{
			jsonKeyRXBytesRate: port.RXBytesRate,
			jsonKeyTXBytesRate: port.TXBytesRate,
		}
	}
	rxRate := int64(0)
	txRate := int64(0)
	if port.Up && effectivePortSpeed(port) > 0 {
		// Synthetic rates provide a small heartbeat when no real traffic source
		// is enabled. Observed or explicitly configured rates bypass this path.
		rxRate = 64 + int64(port.Index)
		txRate = 48 + int64(port.Index)
	}
	return map[string]any{
		jsonKeyRXBytesRate: rxRate,
		jsonKeyTXBytesRate: txRate,
	}
}

// explicitPortRateFields returns rate fields only when observation or operator
// input made rates explicit.
func explicitPortRateFields(port Port) map[string]any {
	if !port.TrafficRatesSet && !port.TrafficRatesEnabled {
		return nil
	}
	return map[string]any{
		jsonKeyRXBytesRate: port.RXBytesRate,
		jsonKeyTXBytesRate: port.TXBytesRate,
	}
}

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

func lastInformTrafficFromPayload(payload []byte) *lastInformTrafficStatus {
	var doc map[string]any
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	if err := decoder.Decode(&doc); err != nil {
		return nil
	}
	status := lastInformTrafficStatus{
		Root: trafficRatesFromMap(doc),
	}
	for _, table := range []string{"if_table", "network_table", "uplink_table", "port_table", "port_stats"} {
		status.Rows = append(status.Rows, trafficRowsFromTable(table, doc[table])...)
	}
	if !status.Root.hasAny() && len(status.Rows) == 0 {
		return nil
	}
	return &status
}

func trafficRowsFromTable(table string, value any) []lastInformTrafficRow {
	rows, ok := value.([]any)
	if !ok {
		return nil
	}
	out := make([]lastInformTrafficRow, 0, len(rows))
	for _, item := range rows {
		row, ok := item.(map[string]any)
		if !ok {
			continue
		}
		rates := trafficRatesFromMap(row)
		if stats, ok := row["stats"].(map[string]any); ok {
			rates = trafficRatesFromMap(stats)
		}
		if !rates.hasAny() {
			continue
		}
		out = append(out, lastInformTrafficRow{
			Table:           table,
			PortIdx:         intFromAny(row["port_idx"]),
			IfName:          stringFromAny(row["ifname"]),
			SourceInterface: stringFromAny(row["source_interface"]),
			Role:            stringFromAny(row["role"]),
			NetworkGroup:    stringFromAny(row["networkgroup"]),
			Up:              boolRefFromAny(row["up"]),
			Rates:           rates,
		})
	}
	return out
}

func trafficRatesFromMap(row map[string]any) lastInformTrafficRates {
	return lastInformTrafficRates{
		Bytes:                     int64PointerFromMap(row, "bytes"),
		RXBytes:                   int64PointerFromMap(row, "rx_bytes"),
		TXBytes:                   int64PointerFromMap(row, "tx_bytes"),
		BytesRateBytesPerSecond:   int64PointerFromMap(row, "bytes-r"),
		RXBytesRateBytesPerSecond: int64PointerFromMap(row, "rx_bytes-r"),
		TXBytesRateBytesPerSecond: int64PointerFromMap(row, "tx_bytes-r"),
		RXRateBitsPerSecond:       int64PointerFromMap(row, "rx_rate"),
		TXRateBitsPerSecond:       int64PointerFromMap(row, "tx_rate"),
	}
}

func (r lastInformTrafficRates) hasAny() bool {
	return r.Bytes != nil ||
		r.RXBytes != nil ||
		r.TXBytes != nil ||
		r.BytesRateBytesPerSecond != nil ||
		r.RXBytesRateBytesPerSecond != nil ||
		r.TXBytesRateBytesPerSecond != nil ||
		r.RXRateBitsPerSecond != nil ||
		r.TXRateBitsPerSecond != nil
}

func int64PointerFromMap(row map[string]any, key string) *int64 {
	value, exists := row[key]
	if !exists {
		return nil
	}
	parsed, ok := int64FromAny(value)
	if !ok {
		return nil
	}
	return &parsed
}

func intFromAny(value any) int {
	parsed, ok := int64FromAny(value)
	if !ok {
		return 0
	}
	return int(parsed)
}

func int64FromAny(value any) (int64, bool) {
	switch v := value.(type) {
	case json.Number:
		if parsed, err := v.Int64(); err == nil {
			return parsed, true
		}
		parsed, err := strconv.ParseFloat(v.String(), 64)
		if err != nil {
			return 0, false
		}
		return int64(parsed), true
	case float64:
		return int64(v), true
	case int64:
		return v, true
	case int:
		return int64(v), true
	default:
		return 0, false
	}
}

func stringFromAny(value any) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func boolRefFromAny(value any) *bool {
	switch v := value.(type) {
	case bool:
		return &v
	case string:
		parsed, err := strconv.ParseBool(strings.TrimSpace(v))
		if err != nil {
			return nil
		}
		return &parsed
	default:
		return nil
	}
}

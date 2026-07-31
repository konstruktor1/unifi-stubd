package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

func payloadTraffic(payload []byte) *lastInformTrafficStatus {
	var doc map[string]any
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	if err := decoder.Decode(&doc); err != nil {
		return nil
	}
	status := lastInformTrafficStatus{
		Root: trafficRatesFromMap(doc),
	}
	if uplink, ok := doc["uplink"].(map[string]any); ok {
		if row, ok := trafficRowFromMap("uplink", uplink); ok {
			status.Rows = append(status.Rows, row)
		}
	} else if stringFromAny(doc["type"]) == "usw" {
		if row, ok := trafficRowForSelectedSwitchUplink(doc["port_table"]); ok {
			status.Rows = append(status.Rows, row)
		}
	}
	for _, table := range []string{"if_table", "network_table", "uplink_table", "port_table", "port_stats"} {
		status.Rows = append(status.Rows, trafficRowsFromTable(table, doc[table])...)
	}
	if !status.Root.hasAny() && len(status.Rows) == 0 {
		return nil
	}
	return &status
}

func trafficRowForSelectedSwitchUplink(value any) (lastInformTrafficRow, bool) {
	rows, ok := value.([]any)
	if !ok {
		return lastInformTrafficRow{}, false
	}
	for _, value := range rows {
		row, ok := value.(map[string]any)
		if !ok || !boolFromAny(row["is_uplink"]) {
			continue
		}
		return trafficRowFromMap("uplink", row)
	}
	return lastInformTrafficRow{}, false
}

func boolFromAny(value any) bool {
	valueBool, _ := value.(bool)
	return valueBool
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
		if parsed, ok := trafficRowFromMap(table, row); ok {
			out = append(out, parsed)
		}
	}
	return out
}

func trafficRowFromMap(table string, row map[string]any) (lastInformTrafficRow, bool) {
	rates := trafficRatesFromMap(row)
	if stats, ok := row["stats"].(map[string]any); ok {
		rates = trafficRatesFromMap(stats)
	}
	if !rates.hasAny() {
		return lastInformTrafficRow{}, false
	}
	ifName := stringFromAny(row["ifname"])
	if ifName == "" {
		ifName = stringFromAny(row["name"])
	}
	return lastInformTrafficRow{
		Table:           table,
		PortIdx:         intFromAny(row["port_idx"]),
		IfName:          ifName,
		SourceInterface: stringFromAny(row["source_interface"]),
		Role:            stringFromAny(row["role"]),
		NetworkGroup:    stringFromAny(row["networkgroup"]),
		Up:              boolRefFromAny(row["up"]),
		Rates:           rates,
	}, true
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

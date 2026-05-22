// Package platform imports passive LLDP neighbor data from lldpd. The stub
// never sends LLDP frames itself; this file only normalizes lldpcli JSON into
// portable neighbor facts that payload/status code can consume.
package platform

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
)

// LLDP reads lldpd neighbor data through lldpcli when enabled and normalizes
// command or parse failures into warnings for status/payload code.
func (p hostPlatform) LLDP(ctx context.Context, cfg LLDPConfig) ([]LLDPNeighbor, []error) {
	source := normalizedSource(cfg.Source)
	if source == SourceOff {
		return nil, nil
	}
	if source != LLDPSourceLLDPD {
		return nil, []error{fmt.Errorf("unsupported lldp source %q", source)}
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = p.cfg.CommandTimeout
	}
	out, err := commandContext(ctx, timeout, "lldpcli", "-f", "json", "show", "neighbors")
	if err != nil {
		return nil, []error{err}
	}
	neighbors, parseErr := ParseLLDPCLIJSON(out)
	if parseErr != nil {
		return neighbors, []error{parseErr}
	}
	return neighbors, nil
}

// ParseLLDPCLIJSON parses lldpcli JSON output into portable neighbors.
func ParseLLDPCLIJSON(data []byte) ([]LLDPNeighbor, error) {
	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse lldpcli json: %w", err)
	}
	root, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("parse lldpcli json: root is not an object")
	}
	lldpRoot := objectValue(root, "lldp")
	if lldpRoot == nil {
		lldpRoot = root
	}
	interfaces := objectValue(lldpRoot, "interface")
	if interfaces == nil {
		return nil, nil
	}
	var neighbors []LLDPNeighbor
	for iface, value := range interfaces {
		for _, item := range normalizeObjectList(value) {
			// lldpcli JSON changes shape between versions: an interface may be a
			// single object or a list. Normalize it before extracting portable
			// neighbor fields.
			neighbor := parseLLDPInterface(iface, item)
			if neighbor.Interface != "" {
				neighbors = append(neighbors, neighbor)
			}
		}
	}
	return neighbors, nil
}

// parseLLDPInterface extracts one neighbor from lldpcli's nested interface
// object, accepting the field-name variants seen across lldpd versions.
func parseLLDPInterface(iface string, raw map[string]any) LLDPNeighbor {
	neighbor := LLDPNeighbor{Interface: strings.TrimSpace(iface)}
	if chassis := firstNestedObject(raw, "chassis"); chassis != nil {
		neighbor.ChassisID = firstValueByKeys(chassis, "id", "chassis-id")
		neighbor.ChassisMAC = firstMAC(neighbor.ChassisID, firstValueByKeys(chassis, "mac"))
		neighbor.SystemName = firstValueByKeys(chassis, "name", "sysname", "system-name")
		neighbor.ManagementIP = firstValueByKeys(chassis, "mgmt-ip", "mgmt_ip", "management-ip")
		neighbor.Capabilities = valuesByKey(chassis, "capability")
	}
	if port := firstNestedObject(raw, "port"); port != nil {
		neighbor.PortID = firstValueByKeys(port, "id", "port-id", "ifname")
		neighbor.PortDescription = firstValueByKeys(port, "descr", "description", "port-description")
	}
	if neighbor.ChassisMAC == "" {
		neighbor.ChassisMAC = firstMAC(firstValueByKeys(raw, "chassis", "chassis-id", "chassis_id"))
	}
	return neighbor
}

// objectValue returns a nested object only when lldpcli encoded one at key.
func objectValue(values map[string]any, key string) map[string]any {
	value, ok := values[key]
	if !ok {
		return nil
	}
	object, _ := value.(map[string]any)
	return object
}

// normalizeObjectList accepts either a singleton object or a list of objects,
// which lets one parser handle both lldpcli JSON shapes.
func normalizeObjectList(value any) []map[string]any {
	switch typed := value.(type) {
	case map[string]any:
		return []map[string]any{typed}
	case []any:
		out := make([]map[string]any, 0, len(typed))
		for _, item := range typed {
			if object, ok := item.(map[string]any); ok {
				out = append(out, object)
			}
		}
		return out
	default:
		return nil
	}
}

// firstNestedObject unwraps lldpcli containers that may either contain leaf
// fields directly or one anonymous nested object.
func firstNestedObject(values map[string]any, key string) map[string]any {
	container := objectValue(values, key)
	if container == nil {
		return nil
	}
	if looksLikeLeaf(container) {
		return container
	}
	for _, value := range container {
		if object, ok := value.(map[string]any); ok {
			return object
		}
	}
	return container
}

// looksLikeLeaf detects lldpcli containers that already hold neighbor fields.
func looksLikeLeaf(values map[string]any) bool {
	for _, key := range []string{"id", "value", "name", "descr", "mgmt-ip", "capability"} {
		if _, ok := values[key]; ok {
			return true
		}
	}
	return false
}

// firstValueByKeys reads the first matching scalar across lldpcli field-name
// variants.
func firstValueByKeys(values map[string]any, keys ...string) string {
	for _, key := range keys {
		if value := scalarString(values[key]); value != "" {
			return value
		}
		if object := objectValue(values, key); object != nil {
			if value := scalarString(object["value"]); value != "" {
				return value
			}
		}
	}
	return ""
}

// valuesByKey reads lldpcli fields that may be a scalar, a list of scalars, or
// a list of typed objects.
func valuesByKey(values map[string]any, key string) []string {
	value, ok := values[key]
	if !ok {
		return nil
	}
	switch typed := value.(type) {
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			if text := scalarString(item); text != "" {
				out = append(out, text)
				continue
			}
			if object, ok := item.(map[string]any); ok {
				text := firstValueByKeys(object, "type", "value")
				if text != "" {
					out = append(out, text)
				}
			}
		}
		return out
	default:
		if text := scalarString(value); text != "" {
			return []string{text}
		}
	}
	return nil
}

// scalarString normalizes lldpcli scalar wrappers into trimmed text.
func scalarString(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case map[string]any:
		return scalarString(typed["value"])
	default:
		if typed == nil {
			return ""
		}
		return strings.TrimSpace(fmt.Sprint(typed))
	}
}

// firstMAC extracts the first parseable MAC from lldpcli chassis identifiers.
func firstMAC(values ...string) string {
	for _, value := range values {
		fields := strings.FieldsFunc(value, func(r rune) bool {
			return r == ' ' || r == ',' || r == ';'
		})
		for _, field := range fields {
			field = strings.TrimPrefix(strings.TrimSpace(field), "mac:")
			if mac, err := net.ParseMAC(field); err == nil {
				return mac.String()
			}
		}
	}
	return ""
}

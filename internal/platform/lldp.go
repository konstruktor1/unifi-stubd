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
			neighbor := parseLLDPInterface(iface, item)
			if neighbor.Interface != "" {
				neighbors = append(neighbors, neighbor)
			}
		}
	}
	return neighbors, nil
}

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

func objectValue(values map[string]any, key string) map[string]any {
	value, ok := values[key]
	if !ok {
		return nil
	}
	object, _ := value.(map[string]any)
	return object
}

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

func looksLikeLeaf(values map[string]any) bool {
	for _, key := range []string{"id", "value", "name", "descr", "mgmt-ip", "capability"} {
		if _, ok := values[key]; ok {
			return true
		}
	}
	return false
}

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

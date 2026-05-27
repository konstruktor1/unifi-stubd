package adoption

import (
	"encoding/json"
	"sort"
	"strings"
)

// jsonString reads optional controller fields defensively and trims whitespace.
func jsonString(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var value string
	if err := json.Unmarshal(raw, &value); err == nil {
		return strings.TrimSpace(value)
	}
	return ""
}

// jsonInt reads optional numeric controller fields without failing the whole
// response parse.
func jsonInt(raw json.RawMessage) int {
	if len(raw) == 0 {
		return 0
	}
	var value int
	if err := json.Unmarshal(raw, &value); err == nil {
		return value
	}
	return 0
}

// jsonStringSlice reads optional controller string lists and drops empty items.
func jsonStringSlice(raw json.RawMessage) []string {
	if len(raw) == 0 {
		return nil
	}
	var values []string
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			out = append(out, value)
		}
	}
	return out
}

// summarizeSystemCFG records only size and top-level keys for controller
// provisioning data that the stub intentionally refuses to apply.
func summarizeSystemCFG(raw json.RawMessage) (int, []string) {
	raw = []byte(strings.TrimSpace(string(raw)))
	if len(raw) == 0 {
		return 0, nil
	}
	payload := raw
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		payload = []byte(strings.TrimSpace(text))
	}
	keys := topLevelJSONKeys(payload)
	return len(payload), keys
}

// topLevelJSONKeys makes ignored provisioning blocks inspectable without
// storing their full contents.
func topLevelJSONKeys(data []byte) []string {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(data, &object); err != nil {
		return nil
	}
	keys := make([]string, 0, len(object))
	for key := range object {
		if key = strings.TrimSpace(key); key != "" {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}

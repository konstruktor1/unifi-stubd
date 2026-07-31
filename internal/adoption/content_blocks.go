package adoption

import (
	"encoding/json"
	"fmt"
	"strconv"
	"unicode"
	"unicode/utf8"
)

const (
	maxControllerBlockCount = 256
	maxControllerBlockName  = 128
)

type contentBlockItem struct {
	stringID string
	objectID string
}

// jsonContentBlockStringIDs returns the exact IDs whose last normalized value
// is JSON true. Object descriptors replace duplicate string entries but do not
// enable include/exclude producer predicates.
func jsonContentBlockStringIDs(raw json.RawMessage) ([]string, error) {
	items, err := jsonContentBlockItems(raw)
	if err != nil {
		return nil, err
	}
	order := make([]string, 0, len(items))
	isString := make(map[string]bool, len(items))
	for _, item := range items {
		id := item.stringID
		stringValue := true
		if id == "" {
			id = item.objectID
			stringValue = false
		}
		if id == "" {
			continue
		}
		if _, exists := isString[id]; !exists {
			order = append(order, id)
		}
		isString[id] = stringValue
	}
	ids := make([]string, 0, len(order))
	for _, id := range order {
		if isString[id] {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

// jsonContentBlockIDs returns a bounded inventory of string sentinels and
// object descriptor IDs. No descriptor method, endpoint, payload, or filter is
// retained or executed by the core stub.
func jsonContentBlockIDs(raw json.RawMessage) ([]string, error) {
	items, err := jsonContentBlockItems(raw)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		id := item.stringID
		if id == "" {
			id = item.objectID
		}
		if id == "" {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return ids, nil
}

func jsonContentBlockItems(raw json.RawMessage) ([]contentBlockItem, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var encoded []json.RawMessage
	if err := json.Unmarshal(raw, &encoded); err != nil {
		// Firmware replaces the old map even when the new value is not an array.
		return nil, nil //nolint:nilerr // The non-array value intentionally becomes an empty selection.
	}
	if len(encoded) > maxControllerBlockCount {
		return nil, fmt.Errorf("block list has %d entries, limit is %d", len(encoded), maxControllerBlockCount)
	}
	items := make([]contentBlockItem, 0, len(encoded))
	for _, rawItem := range encoded {
		var stringID string
		if err := json.Unmarshal(rawItem, &stringID); err == nil {
			if err := validateContentBlockID(stringID); err != nil {
				return nil, err
			}
			items = append(items, contentBlockItem{stringID: stringID})
			continue
		}
		var object map[string]json.RawMessage
		if err := json.Unmarshal(rawItem, &object); err != nil || object == nil {
			continue
		}
		var objectID string
		if err := json.Unmarshal(object["id"], &objectID); err != nil || objectID == "" {
			continue
		}
		if err := validateContentBlockID(objectID); err != nil {
			return nil, err
		}
		items = append(items, contentBlockItem{objectID: objectID})
	}
	return items, nil
}

func validateContentBlockID(id string) error {
	if len(id) > maxControllerBlockName {
		return fmt.Errorf("block name exceeds %d bytes", maxControllerBlockName)
	}
	if !utf8.ValidString(id) {
		return fmt.Errorf("block name is not valid UTF-8")
	}
	for _, value := range id {
		if unicode.IsControl(value) {
			return fmt.Errorf("block name contains control characters")
		}
	}
	return nil
}

func jsonBool(raw json.RawMessage) bool {
	var value bool
	_ = json.Unmarshal(raw, &value)
	return value
}

// jsonStringInt64Prefix mirrors strtoll-style controller-time parsing. Only a
// JSON string is considered; leading ASCII whitespace and a sign are accepted,
// trailing bytes are ignored, no digits produce zero, and overflow saturates.
func jsonStringInt64Prefix(raw json.RawMessage) (int64, bool) {
	var parsed *string
	if err := json.Unmarshal(raw, &parsed); err != nil || parsed == nil {
		return 0, false
	}
	text := *parsed
	index := 0
	for index < len(text) && isASCIISpace(text[index]) {
		index++
	}
	start := index
	if index < len(text) && (text[index] == '+' || text[index] == '-') {
		index++
	}
	digits := index
	for index < len(text) && text[index] >= '0' && text[index] <= '9' {
		index++
	}
	if index == digits {
		return 0, true
	}
	number, _ := strconv.ParseInt(text[start:index], 10, 64)
	return number, true
}

func isASCIISpace(value byte) bool {
	switch value {
	case ' ', '\t', '\n', '\v', '\f', '\r':
		return true
	default:
		return false
	}
}

package adoption

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

// Controller JSON limits bound every response before the sanitizer allocates
// maps, slices, or nested provisioning metadata from controller input.
const (
	// MaxControllerJSONBytes is the largest decoded controller JSON document.
	MaxControllerJSONBytes = 4 * 1024 * 1024
	// MaxControllerJSONDepth is the deepest accepted object/array nesting.
	MaxControllerJSONDepth = 64
	// MaxControllerJSONTokens bounds names, values, and structural delimiters.
	MaxControllerJSONTokens = 65536
)

type jsonLimitError struct {
	message string
}

func (limitErr *jsonLimitError) Error() string {
	return limitErr.message
}

// validateJSONLimits tokenizes one JSON value before any typed or map decode.
// Token count includes field names and delimiters so both wide objects and
// arrays have a fixed allocation-independent ceiling.
func validateJSONLimits(data []byte, subject string) error {
	if len(data) > MaxControllerJSONBytes {
		return &jsonLimitError{message: fmt.Sprintf(
			"%s exceeds %d bytes",
			subject,
			MaxControllerJSONBytes,
		)}
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	depth := 0
	tokens := 0
	rootComplete := false
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("invalid %s: %w", subject, err)
		}
		if rootComplete {
			return fmt.Errorf("invalid %s: multiple top-level values", subject)
		}

		tokens++
		if tokens > MaxControllerJSONTokens {
			return &jsonLimitError{message: fmt.Sprintf(
				"%s exceeds %d tokens",
				subject,
				MaxControllerJSONTokens,
			)}
		}

		delim, isDelim := token.(json.Delim)
		if !isDelim {
			if depth == 0 {
				rootComplete = true
			}
			continue
		}
		switch delim {
		case '{', '[':
			depth++
			if depth > MaxControllerJSONDepth {
				return &jsonLimitError{message: fmt.Sprintf(
					"%s exceeds maximum depth %d",
					subject,
					MaxControllerJSONDepth,
				)}
			}
		case '}', ']':
			depth--
			if depth == 0 {
				rootComplete = true
			}
		}
	}
	if !rootComplete {
		return fmt.Errorf("invalid %s: empty input", subject)
	}
	return nil
}

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

// jsonExactString reads fixed controller protocol vocabulary without
// normalizing it into a different recognized value.
func jsonExactString(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var value string
	if err := json.Unmarshal(raw, &value); err == nil {
		return value
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

// summarizeSystemCFG records only size and top-level keys for controller
// provisioning data that the stub intentionally refuses to apply.
func summarizeSystemCFG(raw json.RawMessage) (int, []string, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return 0, nil, nil
	}
	payload := raw
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		payload = bytes.TrimSpace([]byte(text))
	}
	if err := validateJSONLimits(payload, "controller system_cfg JSON"); err != nil {
		var limitErr *jsonLimitError
		if errors.As(err, &limitErr) {
			return len(payload), nil, err
		}
		// system_cfg is metadata-only. Preserve the established behavior for a
		// syntactically invalid provisioning string while retaining no content.
		return len(payload), nil, nil
	}
	keys := topLevelJSONKeys(payload)
	return len(payload), keys, nil
}

// topLevelJSONKeys retains only the two observed fixed namespaces. Unknown
// controller-selected names are not status vocabulary.
func topLevelJSONKeys(data []byte) []string {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(data, &object); err != nil {
		return nil
	}
	keys := make([]string, 0, 2)
	for _, key := range []string{"ubntconf", "udapi"} {
		if _, present := object[key]; present {
			keys = append(keys, key)
		}
	}
	return keys
}

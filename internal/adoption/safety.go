package adoption

import (
	"encoding/json"
	"strings"
)

// isUnsafeControllerCommand identifies response types that may imply restart,
// firmware, shell, or host changes and must stay metadata-only.
func isUnsafeControllerCommand(responseType string) bool {
	switch strings.TrimSpace(responseType) {
	case "cmd", "exec", "restart", "reboot", "restore-default", "shell", "syswrapper", "upgrade":
		return true
	default:
		return false
	}
}

// isResetControllerCommand recognizes controller removal commands that should
// reset only local adoption state.
func isResetControllerCommand(responseType string) bool {
	switch strings.TrimSpace(responseType) {
	case "delete", "forget", "remove", "restore-default", "setdefault":
		return true
	default:
		return false
	}
}

// responseHasResetCommand scans non-type fields for reset-like command text
// seen in controller response variants.
func responseHasResetCommand(raw map[string]json.RawMessage) bool {
	for key, value := range raw {
		if key == "_type" {
			continue
		}
		if jsonRawContainsResetCommand(value) {
			return true
		}
	}
	return false
}

// jsonRawContainsResetCommand searches strings and string lists without
// executing or interpreting arbitrary controller command payloads.
func jsonRawContainsResetCommand(raw json.RawMessage) bool {
	if len(raw) == 0 {
		return false
	}
	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return textContainsResetCommand(text)
	}
	var values []string
	if err := json.Unmarshal(raw, &values); err == nil {
		for _, value := range values {
			if textContainsResetCommand(value) {
				return true
			}
		}
	}
	return textContainsResetCommand(string(raw))
}

// textContainsResetCommand matches reset command fragments used by UniFi shell
// wrappers while keeping the action local to the adoption store.
func textContainsResetCommand(value string) bool {
	value = strings.ToLower(value)
	return strings.Contains(value, "restore-default") ||
		strings.Contains(value, "reset2defaults") ||
		strings.Contains(value, "setdefault")
}

// resetReason turns a reset-like controller response into a status-safe audit
// message.
func resetReason(responseType string) string {
	responseType = strings.TrimSpace(responseType)
	if responseType == "" {
		return "controller reset command"
	}
	return "controller " + responseType + " command"
}

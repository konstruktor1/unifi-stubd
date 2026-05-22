// Package adoption persists the minimal controller state needed for later inform
// requests. The env mapping table intentionally whitelists accepted mgmt_cfg
// keys instead of applying arbitrary controller provisioning data.
package adoption

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// storeField maps one whitelisted adoption value between env files and
// controller mgmt_cfg keys.
type storeField struct {
	envKey  string
	mgmtKey string
	get     func(Store) string
	set     func(*Store, string)
}

// storeFields is the adoption-state allowlist accepted from controller
// responses.
var storeFields = []storeField{
	{
		envKey: "STATE",
		get:    func(store Store) string { return string(store.State) },
		set:    func(store *Store, value string) { store.State = State(value) },
	},
	{
		envKey:  "INFORM_URL",
		mgmtKey: "inform_url",
		get:     func(store Store) string { return store.InformURL },
		set:     func(store *Store, value string) { store.InformURL = value },
	},
	{
		envKey:  "AUTHKEY",
		mgmtKey: "authkey",
		get:     func(store Store) string { return store.AuthKey },
		set:     func(store *Store, value string) { store.AuthKey = value },
	},
	{
		envKey:  "CFGVERSION",
		mgmtKey: "cfgversion",
		get:     func(store Store) string { return store.CFGVersion },
		set:     func(store *Store, value string) { store.CFGVersion = value },
	},
	{
		envKey:  "USE_AES_GCM",
		mgmtKey: "use_aes_gcm",
		get: func(store Store) string {
			if store.UseAESGCM {
				return "true"
			}
			return ""
		},
		set: func(store *Store, value string) {
			store.UseAESGCM, _ = strconv.ParseBool(value)
		},
	},
	{
		envKey: "VERSION",
		get:    func(store Store) string { return store.Version },
		set:    func(store *Store, value string) { store.Version = value },
	},
}

// LoadEnv reads adoption state from a key-value environment file.
func LoadEnv(path string) (Store, error) {
	var store Store
	data, err := os.ReadFile(path)
	if err != nil {
		return store, fmt.Errorf("read adoption state %s: %w", path, err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		key, value, ok := strings.Cut(strings.TrimSpace(line), "=")
		if !ok {
			continue
		}
		if field, ok := storeFieldByEnvKey(key); ok {
			field.set(&store, value)
		}
	}
	return store, nil
}

// SaveEnv writes adoption state to a key-value environment file.
func SaveEnv(path string, store Store) error {
	if path == "" {
		return errors.New("state path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create adoption state directory: %w", err)
	}
	var b strings.Builder
	for _, field := range storeFields {
		if value := field.get(store); value != "" {
			b.WriteString(field.envKey + "=" + value + "\n")
		}
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o600); err != nil {
		return fmt.Errorf("write adoption state %s: %w", path, err)
	}
	return nil
}

// ResetEnv clears adoption data and persists the stub in factory state.
func ResetEnv(path string) (Store, error) {
	store := Store{State: StateFactory}
	if err := SaveEnv(path, store); err != nil {
		return store, err
	}
	return store, nil
}

// Merge applies non-empty fields from update to base and reports changes.
func Merge(base, update Store) (Store, bool) {
	changed := false
	for _, field := range storeFields {
		if value := field.get(update); value != "" && field.get(base) != value {
			field.set(&base, value)
			changed = true
		}
	}
	return base, changed
}

// Controller response types handled specially by the adoption sanitizer.
const (
	responseTypeNoop     = "noop"
	responseTypeSetParam = "setparam"
	responseTypeUpgrade  = "upgrade"
)

// ParseSetParamResponse extracts adoption settings from a setparam response.
func ParseSetParamResponse(data []byte) (Store, bool, error) {
	store, kind, ok, err := ParseControllerResponse(data)
	return store, ok && kind == responseTypeSetParam, err
}

// ParseControllerResponse extracts adoption state from a controller response.
func ParseControllerResponse(data []byte) (Store, string, bool, error) {
	info, err := ParseControllerResponseInfo(data)
	if err != nil {
		return Store{}, "", false, err
	}
	return info.Store, info.Type, info.HasStateUpdate, nil
}

// ParseControllerResponseInfo returns a sanitized controller response summary.
func ParseControllerResponseInfo(data []byte) (ControllerResponse, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return ControllerResponse{}, fmt.Errorf("parse controller response: %w", err)
	}
	response := ControllerResponse{Type: jsonString(raw["_type"])}
	// Only a narrow mgmt_cfg allowlist is persisted. Provisioning blocks,
	// firmware actions, shell commands, and restart-like requests are reported
	// as metadata or local stub resets, never executed on the host.
	switch response.Type {
	case responseTypeSetParam:
		if mgmtCFG := jsonString(raw["mgmt_cfg"]); mgmtCFG != "" {
			response.HasMgmtCFG = true
			response.Store = parseMgmtCFG(mgmtCFG)
			response.HasStateUpdate = storeHasStateUpdate(response.Store)
		}
		if systemCFG, ok := raw["system_cfg"]; ok {
			response.HasSystemCFG = true
			response.SystemCFGBytes, response.SystemCFGKeys = summarizeSystemCFG(systemCFG)
			response.Ignored = true
			response.IgnoredReason = "system_cfg provisioning is recorded as metadata only"
		}
	case responseTypeUpgrade:
		version := jsonString(raw["version"])
		if version != "" {
			response.Store = Store{
				State:   StateProvisioning,
				Version: version,
			}
			response.HasStateUpdate = true
		}
		response.Ignored = true
		response.IgnoredReason = "firmware upgrade request ignored by safety policy"
	case responseTypeNoop:
		response.Store = Store{State: StateConnected}
		response.HasStateUpdate = true
		response.IntervalSeconds = jsonInt(raw["interval"])
		response.IncludeBlocks = jsonStringSlice(raw["include_blocks"])
	default:
		if isResetControllerCommand(response.Type) || responseHasResetCommand(raw) {
			response.Store = Store{State: StateFactory}
			response.HasStateUpdate = true
			response.ResetRequested = true
			response.ResetReason = resetReason(response.Type)
			return response, nil
		}
		if isUnsafeControllerCommand(response.Type) {
			response.Store = Store{State: StateProvisioning}
			response.HasStateUpdate = true
			response.Ignored = true
			response.IgnoredReason = "controller command ignored by safety policy"
		}
	}
	return response, nil
}

// parseMgmtCFG accepts only whitelisted mgmt_cfg keys that affect future
// inform identity; other controller provisioning keys are ignored.
func parseMgmtCFG(mgmtCFG string) Store {
	store := Store{State: StateProvisioning}
	for _, line := range strings.Split(mgmtCFG, "\n") {
		key, value, ok := strings.Cut(strings.TrimSpace(line), "=")
		if !ok {
			continue
		}
		// Unknown controller keys are ignored by design. The storeFields table is
		// the policy boundary for adoption data accepted from the controller.
		if field, ok := storeFieldByMgmtKey(key); ok {
			field.set(&store, value)
		}
	}
	return store
}

// storeHasStateUpdate reports whether parsing found any durable adoption field
// worth persisting.
func storeHasStateUpdate(store Store) bool {
	for _, field := range storeFields {
		if field.get(store) != "" {
			return true
		}
	}
	return false
}

// storeFieldByEnvKey resolves persisted environment keys through the same field
// table used for saving adoption state.
func storeFieldByEnvKey(key string) (storeField, bool) {
	for _, field := range storeFields {
		if field.envKey == key {
			return field, true
		}
	}
	return storeField{}, false
}

// storeFieldByMgmtKey is the allowlist boundary for controller-provided
// mgmt_cfg fields.
func storeFieldByMgmtKey(key string) (storeField, bool) {
	for _, field := range storeFields {
		if field.mgmtKey != "" && field.mgmtKey == key {
			return field, true
		}
	}
	return storeField{}, false
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

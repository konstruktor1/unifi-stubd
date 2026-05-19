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
		switch key {
		case "STATE":
			store.State = State(value)
		case "INFORM_URL":
			store.InformURL = value
		case "AUTHKEY":
			store.AuthKey = value
		case "CFGVERSION":
			store.CFGVersion = value
		case "USE_AES_GCM":
			store.UseAESGCM, _ = strconv.ParseBool(value)
		case "VERSION":
			store.Version = value
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
	if store.State != "" {
		b.WriteString("STATE=" + string(store.State) + "\n")
	}
	if store.InformURL != "" {
		b.WriteString("INFORM_URL=" + store.InformURL + "\n")
	}
	if store.AuthKey != "" {
		b.WriteString("AUTHKEY=" + store.AuthKey + "\n")
	}
	if store.CFGVersion != "" {
		b.WriteString("CFGVERSION=" + store.CFGVersion + "\n")
	}
	if store.UseAESGCM {
		b.WriteString("USE_AES_GCM=true\n")
	}
	if store.Version != "" {
		b.WriteString("VERSION=" + store.Version + "\n")
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
	setString := func(dst *string, src string) {
		if src != "" && *dst != src {
			*dst = src
			changed = true
		}
	}
	setString(&base.InformURL, update.InformURL)
	setString(&base.AuthKey, update.AuthKey)
	setString(&base.CFGVersion, update.CFGVersion)
	setString(&base.Version, update.Version)
	if update.UseAESGCM && !base.UseAESGCM {
		base.UseAESGCM = true
		changed = true
	}
	if update.State != "" && base.State != update.State {
		base.State = update.State
		changed = true
	}
	return base, changed
}

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

func parseMgmtCFG(mgmtCFG string) Store {
	store := Store{State: StateProvisioning}
	for _, line := range strings.Split(mgmtCFG, "\n") {
		key, value, ok := strings.Cut(strings.TrimSpace(line), "=")
		if !ok {
			continue
		}
		switch key {
		case "inform_url":
			store.InformURL = value
		case "authkey":
			store.AuthKey = value
		case "cfgversion":
			store.CFGVersion = value
		case "use_aes_gcm":
			store.UseAESGCM, _ = strconv.ParseBool(value)
		}
	}
	return store
}

func storeHasStateUpdate(store Store) bool {
	return store.InformURL != "" ||
		store.AuthKey != "" ||
		store.CFGVersion != "" ||
		store.UseAESGCM ||
		store.Version != "" ||
		store.State != ""
}

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

func isUnsafeControllerCommand(responseType string) bool {
	switch strings.TrimSpace(responseType) {
	case "cmd", "exec", "restart", "reboot", "restore-default", "shell", "syswrapper", "upgrade":
		return true
	default:
		return false
	}
}

func isResetControllerCommand(responseType string) bool {
	switch strings.TrimSpace(responseType) {
	case "delete", "forget", "remove", "restore-default":
		return true
	default:
		return false
	}
}

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

func textContainsResetCommand(value string) bool {
	value = strings.ToLower(value)
	return strings.Contains(value, "restore-default") || strings.Contains(value, "reset2defaults")
}

func resetReason(responseType string) string {
	responseType = strings.TrimSpace(responseType)
	if responseType == "" {
		return "controller reset command"
	}
	return "controller " + responseType + " command"
}

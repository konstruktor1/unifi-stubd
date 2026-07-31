package adoption

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

// Controller response types handled specially by the adoption sanitizer.
const (
	responseTypeCommand    = "cmd"
	responseTypeNoop       = "noop"
	responseTypeSetDefault = "setdefault"
	responseTypeSetParam   = "setparam"
	responseTypeUpgrade    = "upgrade"
)

// ParseControllerResponseInfo returns a sanitized controller response summary.
func ParseControllerResponseInfo(data []byte) (ControllerResponse, error) {
	if err := validateJSONLimits(data, "controller response JSON"); err != nil {
		return ControllerResponse{}, fmt.Errorf("parse controller response: %w", err)
	}
	raw, isObject, err := decodeUniqueTopLevelObject(data)
	if err != nil {
		return ControllerResponse{}, fmt.Errorf("parse controller response: %w", err)
	}
	if !isObject {
		return ControllerResponse{}, errors.New("parse controller response: top-level value must be an object")
	}
	rawType, present := raw["_type"]
	if !present {
		return ControllerResponse{}, errors.New("parse controller response: missing _type string")
	}
	var responseType *string
	if err := json.Unmarshal(rawType, &responseType); err != nil || responseType == nil {
		return ControllerResponse{}, errors.New("parse controller response: _type must be a string")
	}
	if !knownControllerResponseType(*responseType) {
		return ControllerResponse{}, errors.New("parse controller response: unsupported _type")
	}
	response := ControllerResponse{Type: *responseType}
	// Only a narrow mgmt_cfg allowlist is persisted. Provisioning blocks,
	// firmware actions, shell commands, and restart-like requests are reported
	// as metadata or local stub resets, never executed on the host.
	switch response.Type {
	case responseTypeSetParam:
		// Firmware enters Managed on every successful setparam before applying
		// response-specific identity handling. Merge returns an authkey-bearing
		// response or a changed URL to Adopting and enforces keyless safety.
		response.Store = Store{State: StateConnected}
		response.HasStateUpdate = true
		if mgmtCFG := jsonString(raw["mgmt_cfg"]); mgmtCFG != "" {
			response.HasMgmtCFG = true
			store, err := parseMgmtCFG(mgmtCFG)
			if err != nil {
				return ControllerResponse{}, fmt.Errorf("parse controller mgmt_cfg: %w", err)
			}
			store.State = StateConnected
			response.Store = store
		}
		if systemCFG, ok := raw["system_cfg"]; ok {
			response.HasSystemCFG = true
			var err error
			response.SystemCFGBytes, response.SystemCFGKeys, err = summarizeSystemCFG(systemCFG)
			if err != nil {
				return ControllerResponse{}, fmt.Errorf("parse controller response: %w", err)
			}
			response.Ignored = true
			response.IgnoredReason = "system_cfg provisioning is recorded as metadata only"
		}
	case responseTypeUpgrade:
		response.Ignored = true
		response.IgnoredReason = "firmware upgrade request retained as status only"
	case responseTypeNoop:
		response.Store = Store{State: StateConnected}
		response.HasStateUpdate = true
		response.IntervalSeconds = jsonInt(raw["interval"])
		response.ServerTimeUTC, response.ServerTimeSet = jsonStringInt64Prefix(raw["server_time_in_utc"])
		response.BlockSelectionSet = true
		var parseErr error
		response.IncludeBlocks, parseErr = jsonContentBlockStringIDs(raw["include_blocks"])
		if parseErr != nil {
			response.IncludeBlocks = nil
			response.BlockSelectionLimited = true
		}
		response.ExcludeBlocks, parseErr = jsonContentBlockStringIDs(raw["exclude_blocks"])
		if parseErr != nil {
			response.ExcludeBlocks = nil
			response.BlockSelectionLimited = true
		}
		response.UDAPIBlocks, parseErr = jsonContentBlockIDs(raw["udapi_blocks"])
		if parseErr != nil {
			response.UDAPIBlocks = nil
			response.BlockSelectionLimited = true
		}
		response.Immediate = jsonBool(raw["immediate"])
	case responseTypeSetDefault:
		response.Store = Store{State: StateFactory}
		response.HasStateUpdate = true
		response.ResetRequested = true
		response.ResetReason = "controller setdefault command"
	case responseTypeCommand:
		response.Ignored = true
		response.IgnoredReason = "controller command ignored by safety policy"
	default:
		if isUnsafeControllerCommand(response.Type) {
			response.Ignored = true
			response.IgnoredReason = "controller command ignored by safety policy"
		}
	}
	return response, nil
}

// decodeUniqueTopLevelObject rejects ambiguous controller responses before
// their discriminator or any state-bearing field can be dispatched.
func decodeUniqueTopLevelObject(data []byte) (map[string]json.RawMessage, bool, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	first, err := decoder.Token()
	if err != nil {
		return nil, false, fmt.Errorf("read first controller response token: %w", err)
	}
	delim, ok := first.(json.Delim)
	if !ok || delim != '{' {
		return nil, false, nil
	}

	raw := make(map[string]json.RawMessage)
	for decoder.More() {
		keyToken, err := decoder.Token()
		if err != nil {
			return nil, false, fmt.Errorf("read controller response object key: %w", err)
		}
		key, ok := keyToken.(string)
		if !ok {
			return nil, false, errors.New("top-level JSON object key must be a string")
		}
		if _, duplicate := raw[key]; duplicate {
			return nil, false, errors.New("duplicate top-level JSON key")
		}
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return nil, false, fmt.Errorf("decode controller response object value: %w", err)
		}
		raw[key] = value
	}
	if _, err := decoder.Token(); err != nil {
		return nil, false, fmt.Errorf("read controller response object end: %w", err)
	}
	return raw, true, nil
}

package adoption

import (
	"encoding/json"
	"fmt"
)

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

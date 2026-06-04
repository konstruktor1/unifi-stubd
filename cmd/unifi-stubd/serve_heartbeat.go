package main

import (
	"log"
	"net"

	"github.com/konstruktor1/unifi-stubd/internal/adoption"
	"github.com/konstruktor1/unifi-stubd/internal/discovery"
)

// sendDiscovery sends the already-encoded UDP announcement and treats send
// failures as heartbeat warnings, not daemon-fatal errors.
func sendDiscovery(packet []byte, hostname string, mac net.HardwareAddr, skip bool, iface string, targets []string) {
	if skip {
		return
	}
	if err := discovery.SendToInterface(packet, targets, iface); err != nil {
		log.Printf("discovery send failed: %v", err)
		return
	}
	log.Printf("sent discovery announcement for %s (%s)", hostname, mac)
}

// sendInformHeartbeat sends one inform packet, persists sanitized exchange
// status, and applies only safe adoption-state updates from decoded controller
// responses.
func sendInformHeartbeat(mac net.HardwareAddr, informURL, statePath, statusPath string, store adoption.Store, payload []byte, sourceIP net.IP) {
	if informURL == "" {
		return
	}
	resp, cipher, err := sendInform(mac, informURL, store, payload, sourceIP)
	if err != nil {
		recordLastInform(statusPath, newLastInformStatus(informURL, store, payload), 0, "", cipher, 0, 0, err)
		log.Printf("inform send failed: %v", err)
		return
	}
	last := newLastInformStatus(informURL, store, payload)
	last.StatusCode = resp.StatusCode
	last.AttemptedAESGCM = cipher.AttemptedAESGCM
	last.UsedAESGCM = cipher.UsedAESGCM
	last.FallbackToCBC = cipher.FallbackToCBC
	last.RawBytes = len(resp.RawBody)
	last.JSONBytes = len(resp.JSONBody)
	if len(resp.JSONBody) > 0 {
		controllerResponse, parseErr := adoption.ParseControllerResponseInfo(resp.JSONBody)
		if parseErr != nil {
			last.Error = parseErr.Error()
			log.Printf("controller response parse failed: %v", parseErr)
		} else {
			store = updateAdoptionState(statePath, store, controllerResponse, cipher.UsedAESGCM)
			if controllerResponse.ResetRequested && store.State == adoption.StateFactory && store.AuthKey == "" {
				controllerResponse.ResetApplied = true
			}
			last.ControllerState = adoptionStateText(store)
			last.CFGVersion = store.CFGVersion
			last.Version = store.Version
			applyResponseStatus(&last, controllerResponse)
			logInformResponse(resp, controllerResponse, store, cipher)
		}
		recordLastInform(statusPath, last, resp.StatusCode, last.ResponseType, cipher, len(resp.RawBody), len(resp.JSONBody), nil)
		return
	}
	recordLastInform(statusPath, last, resp.StatusCode, "", cipher, len(resp.RawBody), 0, nil)
	log.Printf("inform response status=%d raw_bytes=%d cipher=%s", resp.StatusCode, len(resp.RawBody), cipherStatusText(cipher))
}

// applyResponseStatus copies the sanitized controller response into
// persisted status, excluding raw provisioning bodies.
func applyResponseStatus(last *lastInformStatus, response adoption.ControllerResponse) {
	last.ResponseType = response.Type
	last.IntervalSeconds = response.IntervalSeconds
	last.IncludeBlocks = cloneStrings(response.IncludeBlocks)
	last.ResetRequested = response.ResetRequested
	last.ResetApplied = response.ResetApplied
	last.ResetReason = response.ResetReason
	last.HasMgmtCFG = response.HasMgmtCFG
	last.HasSystemCFG = response.HasSystemCFG
	last.SystemCFGBytes = response.SystemCFGBytes
	last.SystemCFGKeys = cloneStrings(response.SystemCFGKeys)
	last.Ignored = response.Ignored
	last.IgnoredReason = response.IgnoredReason
}

// recordLastInform writes the same sanitized controller-exchange summary used
// by --status, keeping auth keys and raw controller payloads out of the status
// file.
func recordLastInform(statusPath string, last lastInformStatus, statusCode int, responseType string, cipher informCipherStatus, rawBytes, jsonBytes int, err error) {
	last.StatusCode = statusCode
	last.ResponseType = responseType
	last.AttemptedAESGCM = cipher.AttemptedAESGCM
	last.UsedAESGCM = cipher.UsedAESGCM
	last.FallbackToCBC = cipher.FallbackToCBC
	last.RawBytes = rawBytes
	last.JSONBytes = jsonBytes
	if err != nil {
		last.Error = err.Error()
	}
	if saveErr := saveLastInformStatus(statusPath, last); saveErr != nil {
		log.Printf("runtime status write failed: %v", saveErr)
	}
}

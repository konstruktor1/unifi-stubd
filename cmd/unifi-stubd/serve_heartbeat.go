package main

import (
	"log"
	"net"
	"time"

	"github.com/konstruktor1/unifi-stubd/internal/adoption"
	"github.com/konstruktor1/unifi-stubd/internal/discovery"
	"github.com/konstruktor1/unifi-stubd/internal/inform"
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
func sendInformHeartbeat(mac net.HardwareAddr, informURL, statePath, statusPath string, store adoption.Store, payload []byte, sourceIP net.IP, commands *adoption.CommandLifecycle, schedule *inform.Schedule) inform.ScheduleDecision {
	if informURL == "" {
		return schedule.Current()
	}
	resultAttached := false
	if commands != nil {
		payloadWithResult, attached, attachErr := commands.Attach(payload)
		if attachErr != nil {
			log.Printf("controller command result attach failed: %v", attachErr)
		} else {
			payload = payloadWithResult
			resultAttached = attached
		}
	}
	resp, cipher, err := sendInform(mac, informURL, store, payload, sourceIP)
	if err != nil {
		decision := schedule.Failure()
		last := newLastInformStatus(informURL, store, payload)
		applyCommandLifecycleStatus(&last, commands)
		applyInformScheduleStatus(&last, decision)
		recordLastInform(statusPath, last, 0, "", cipher, 0, 0, err)
		log.Printf("inform send failed: %v retry_in=%s consecutive_failures=%d", err, decision.NextDelay, decision.ConsecutiveFailures)
		return decision
	}
	if commands != nil && resultAttached && resp.StatusCode == 200 {
		commands.MarkResultSent()
		log.Printf("controller command result delivered: name=wg-genkeypair host_mutation=false")
	}
	last := newLastInformStatus(informURL, store, payload)
	applyCommandLifecycleStatus(&last, commands)
	last.StatusCode = resp.StatusCode
	last.AttemptedAESGCM = cipher.AttemptedAESGCM
	last.UsedAESGCM = cipher.UsedAESGCM
	last.FallbackToCBC = cipher.FallbackToCBC
	last.RawBytes = len(resp.RawBody)
	last.JSONBytes = len(resp.JSONBody)
	if len(resp.JSONBody) == 0 {
		decision := schedule.Success(0)
		applyInformScheduleStatus(&last, decision)
		recordLastInform(statusPath, last, resp.StatusCode, "", cipher, len(resp.RawBody), 0, nil)
		log.Printf("inform response status=%d raw_bytes=%d cipher=%s", resp.StatusCode, len(resp.RawBody), cipherStatusText(cipher))
		return decision
	}
	controllerResponse, parseErr := adoption.ParseControllerResponseInfo(resp.JSONBody)
	if parseErr != nil {
		decision := schedule.Success(0)
		applyInformScheduleStatus(&last, decision)
		last.Error = parseErr.Error()
		log.Printf("controller response parse failed: %v", parseErr)
		recordLastInform(statusPath, last, resp.StatusCode, "", cipher, len(resp.RawBody), len(resp.JSONBody), nil)
		return decision
	}
	decision := schedule.Success(controllerIntervalDuration(controllerResponse.IntervalSeconds))
	applyInformScheduleStatus(&last, decision)
	handleControllerCommand(&last, commands, &controllerResponse)
	store = updateAdoptionState(statePath, store, controllerResponse, cipher.UsedAESGCM)
	if controllerResponse.ResetRequested && store.State == adoption.StateFactory && store.AuthKey == "" {
		controllerResponse.ResetApplied = true
	}
	last.ControllerState = adoptionStateText(store)
	last.CFGVersion = store.CFGVersion
	last.Version = store.Version
	applyResponseStatus(&last, controllerResponse)
	logInformResponse(resp, controllerResponse, store, cipher)
	recordLastInform(statusPath, last, resp.StatusCode, last.ResponseType, cipher, len(resp.RawBody), len(resp.JSONBody), nil)
	return decision
}

func controllerIntervalDuration(seconds int) time.Duration {
	if seconds <= 0 {
		return 0
	}
	const maximumDuration = time.Duration(1<<63 - 1)
	if int64(seconds) > int64(maximumDuration/time.Second) {
		return maximumDuration
	}
	return time.Duration(seconds) * time.Second
}

func applyInformScheduleStatus(last *lastInformStatus, decision inform.ScheduleDecision) {
	last.Schedule = lastInformScheduleStatus{
		EffectiveInterval:     decision.EffectiveInterval.String(),
		NextDelay:             decision.NextDelay.String(),
		ConsecutiveFailures:   decision.ConsecutiveFailures,
		Backoff:               decision.Backoff,
		ControllerHintApplied: decision.ControllerHintApplied,
	}
}

func handleControllerCommand(last *lastInformStatus, commands *adoption.CommandLifecycle, response *adoption.ControllerResponse) {
	if commands == nil {
		return
	}
	if err := commands.Handle(response); err != nil {
		last.Error = err.Error()
		log.Printf("controller command handling failed: %v", err)
	}
	applyCommandLifecycleStatus(last, commands)
}

func applyCommandLifecycleStatus(last *lastInformStatus, commands *adoption.CommandLifecycle) {
	if commands == nil {
		return
	}
	last.ControllerCommands = commands.Summary()
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
	last.CommandName = response.CommandName
	last.CommandRequestIDSet = response.CommandRequestIDSet
	last.CommandHandled = response.CommandHandled
	last.CommandResultQueued = response.CommandResultQueued
	last.CommandDuplicate = response.CommandDuplicate
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

package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/konstruktor1/unifi-stubd/internal/adoption"
	"github.com/konstruktor1/unifi-stubd/internal/device"
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
func sendInformHeartbeat(mac net.HardwareAddr, informURL, statePath, statusPath string, store adoption.Store, payload []byte, sourceIP net.IP, blocks *adoption.BlockSelection, schedule *inform.Schedule, controllerClock *inform.ControllerClock, reportingPolicy device.ControllerReportingPolicy) (inform.ScheduleDecision, error) {
	if informURL == "" {
		return schedule.Current(), nil
	}
	resp, cipher, err := sendInform(mac, informURL, store, payload, sourceIP)
	if err != nil {
		decision := schedule.Failure()
		last := newLastInformStatus(informURL, store)
		applyControllerClockStatus(&last, controllerClock)
		applyInformScheduleStatus(&last, decision)
		statusCode, rawBytes, jsonBytes := 0, 0, 0
		if resp != nil {
			statusCode = resp.StatusCode
			rawBytes = len(resp.RawBody)
			jsonBytes = len(resp.JSONBody)
		}
		recordLastInform(statusPath, last, statusCode, "", cipher, rawBytes, jsonBytes, err)
		log.Printf("inform send failed: %v retry_in=%s consecutive_failures=%d", err, decision.NextDelay, decision.ConsecutiveFailures)
		return decision, nil
	}
	if resp.StatusCode != http.StatusOK {
		decision := schedule.Failure()
		last := newLastInformStatus(informURL, store)
		applyControllerClockStatus(&last, controllerClock)
		applyInformScheduleStatus(&last, decision)
		responseErr := fmt.Errorf("inform response status %d", resp.StatusCode)
		recordLastInform(statusPath, last, resp.StatusCode, "", cipher, len(resp.RawBody), 0, responseErr)
		log.Printf("inform response status=%d raw_bytes=%d cipher=%s retry_in=%s consecutive_failures=%d", resp.StatusCode, len(resp.RawBody), cipherStatusText(cipher), decision.NextDelay, decision.ConsecutiveFailures)
		return decision, nil
	}
	last := newLastInformStatus(informURL, store)
	applyControllerClockStatus(&last, controllerClock)
	last.StatusCode = resp.StatusCode
	last.RawBytes = len(resp.RawBody)
	last.JSONBytes = len(resp.JSONBody)
	if len(resp.RawBody) == 0 {
		decision := schedule.Success(0)
		applyInformScheduleStatus(&last, decision)
		recordLastInform(statusPath, last, resp.StatusCode, "", cipher, len(resp.RawBody), 0, nil)
		log.Printf("inform response status=%d raw_bytes=%d cipher=%s", resp.StatusCode, len(resp.RawBody), cipherStatusText(cipher))
		return decision, nil
	}
	controllerResponse, parseErr := adoption.ParseControllerResponseInfo(resp.JSONBody)
	if parseErr != nil {
		decision := schedule.Failure()
		applyInformScheduleStatus(&last, decision)
		log.Printf("controller response parse failed: %v retry_in=%s consecutive_failures=%d", parseErr, decision.NextDelay, decision.ConsecutiveFailures)
		recordLastInform(statusPath, last, resp.StatusCode, "", cipher, len(resp.RawBody), len(resp.JSONBody), parseErr)
		return decision, nil
	}
	last.ResponseJSONParsed = true
	updatedStore, stateErr := updateAdoptionState(
		statePath,
		store,
		&controllerResponse,
		cipher.RequestFlags&inform.FlagEncryptedGCM != 0,
	)
	if stateErr != nil {
		decision := schedule.Failure()
		applyInformScheduleStatus(&last, decision)
		last.ControllerState = adoptionStateText(store)
		last.CFGVersion = store.CFGVersion
		last.Version = store.Version
		applyResponseStatus(&last, controllerResponse)
		applyControllerClockStatus(&last, controllerClock)
		logInformResponse(resp, controllerResponse, store, cipher)
		recordLastInform(statusPath, last, resp.StatusCode, last.ResponseType, cipher, len(resp.RawBody), len(resp.JSONBody), stateErr)
		return decision, fmt.Errorf("apply controller lifecycle response: %w", stateErr)
	}
	store = updatedStore
	decision := schedule.SuccessWithOptions(
		controllerIntervalDuration(controllerResponse.IntervalSeconds),
		controllerResponse.Immediate,
	)
	applyInformScheduleStatus(&last, decision)
	if controllerResponse.ResetApplied {
		if blocks != nil {
			*blocks = adoption.BlockSelection{}
		}
		if controllerClock != nil {
			controllerClock.Reset()
		}
		decision = schedule.Reset()
		applyInformScheduleStatus(&last, decision)
	}
	last.ControllerState = adoptionStateText(store)
	last.CFGVersion = store.CFGVersion
	last.Version = store.Version
	applyResponseStatus(&last, controllerResponse)
	if reportingPolicy.ControllerClock && controllerResponse.Type == "noop" && controllerResponse.ServerTimeSet && controllerClock != nil {
		controllerClock.Update(controllerResponse.ServerTimeUTC, time.Now())
	}
	applyControllerClockStatus(&last, controllerClock)
	if reportingPolicy.BlockSelection && blocks != nil && controllerResponse.BlockSelectionSet {
		*blocks = controllerResponse.BlockSelection()
	}
	logInformResponse(resp, controllerResponse, store, cipher)
	recordLastInform(statusPath, last, resp.StatusCode, last.ResponseType, cipher, len(resp.RawBody), len(resp.JSONBody), nil)
	return decision, nil
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
		ControllerHintIgnored: decision.ControllerHintIgnored,
		ControllerHints:       decision.ControllerHints,
		ImmediateApplied:      decision.ImmediateApplied,
		ImmediateSuppressed:   decision.ImmediateSuppressed,
	}
}

func applyControllerClockStatus(last *lastInformStatus, controllerClock *inform.ControllerClock) {
	if controllerClock == nil {
		return
	}
	last.ControllerTimeOffset = controllerClock.OffsetSeconds()
	last.ControllerTimeSet = controllerClock.Set()
}

// applyResponseStatus copies the sanitized controller response into
// persisted status, excluding raw provisioning bodies.
func applyResponseStatus(last *lastInformStatus, response adoption.ControllerResponse) {
	last.ResponseType = response.Type
	last.IntervalSeconds = response.IntervalSeconds
	last.IncludeBlocks = cloneNonEmptySlice(response.IncludeBlocks)
	last.ExcludeBlocks = cloneNonEmptySlice(response.ExcludeBlocks)
	last.UDAPIBlocks = cloneNonEmptySlice(response.UDAPIBlocks)
	last.BlockSelectionLimited = response.BlockSelectionLimited
	last.Immediate = response.Immediate
	last.ServerTimeUTC = response.ServerTimeUTC
	last.ServerTimeSet = response.ServerTimeSet
	last.ResetRequested = response.ResetRequested
	last.ResetApplied = response.ResetApplied
	last.ResetReason = response.ResetReason
	last.ResetDisposition = response.ResetDisposition
	last.HasMgmtCFG = response.HasMgmtCFG
	last.HasSystemCFG = response.HasSystemCFG
	last.SystemCFGBytes = response.SystemCFGBytes
	last.SystemCFGKeys = cloneNonEmptySlice(response.SystemCFGKeys)
	last.Ignored = response.Ignored
	last.IgnoredReason = response.IgnoredReason
}

// recordLastInform writes the same sanitized controller-exchange summary used
// by --status, keeping auth keys and raw controller payloads out of the status
// file.
func recordLastInform(statusPath string, last lastInformStatus, statusCode int, responseType string, cipher informCipherStatus, rawBytes, jsonBytes int, err error) {
	last.StatusCode = statusCode
	last.ResponseType = responseType
	last.Cipher = cipherStatusText(cipher)
	last.RequestFlags = cipher.RequestFlags
	last.KeySource = cipher.KeySource
	last.ResponseBodyPresent = cipher.ResponseBodyPresent
	last.ResponsePacketDecoded = cipher.ResponsePacketDecoded
	last.ResponseFlags = cipher.ResponseFlags
	last.RawBytes = rawBytes
	last.JSONBytes = jsonBytes
	if err != nil {
		last.Error = err.Error()
	}
	if saveErr := saveLastInformStatus(statusPath, last); saveErr != nil {
		log.Printf("runtime status write failed: %v", saveErr)
	}
}

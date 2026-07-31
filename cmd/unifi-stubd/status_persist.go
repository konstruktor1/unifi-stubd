package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/konstruktor1/unifi-stubd/internal/adoption"
	"github.com/konstruktor1/unifi-stubd/internal/atomicfile"
	"github.com/konstruktor1/unifi-stubd/internal/inform"
)

const maxInformSessionEvents = 16

// loadPersistedRunStatus reads the sanitized last-inform status written by the
// daemon loop for later status commands.
func loadPersistedRunStatus(path string) (persistedRunStatus, error) {
	var status persistedRunStatus
	data, err := os.ReadFile(path)
	if err != nil {
		return status, fmt.Errorf("read runtime status %s: %w", path, err)
	}
	if err := json.Unmarshal(data, &status); err != nil {
		return status, fmt.Errorf("parse runtime status %s: %w", path, err)
	}
	return status, nil
}

// saveLastInformStatus persists only the last sanitized inform summary and
// counter metadata, keeping runtime status narrow and safe to inspect.
func saveLastInformStatus(path string, last lastInformStatus, session informSessionStatus) error {
	if path == "" {
		return errors.New("status path is required")
	}
	data, err := json.MarshalIndent(persistedRunStatus{
		LastInform:    last,
		InformSession: session,
	}, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal runtime status: %w", err)
	}
	data = append(data, '\n')
	if err := atomicfile.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write runtime status %s: %w", path, err)
	}
	return nil
}

// newInformSessionStatus creates a fresh evidence window for one daemon run.
// A restart therefore cannot accidentally combine exchanges from two builds.
func newInformSessionStatus() informSessionStatus {
	return informSessionStatus{StartedAt: time.Now().Format(time.RFC3339)}
}

// reset starts a new bounded evidence window after local de-adoption so status
// cannot conflate exchanges from two controller lifecycle epochs.
func (session *informSessionStatus) reset() {
	if session == nil {
		return
	}
	*session = newInformSessionStatus()
}

// observe records sanitized metadata for actual HTTP requests and updates
// exact mode counters. The bounded event list proves ordering without turning
// a failure before HTTP into synthetic wire evidence.
func (session *informSessionStatus) observe(last lastInformStatus, attempts []informAttemptStatus) {
	if session == nil || len(attempts) == 0 {
		return
	}
	if last.FallbackToCBC {
		session.CBCFallbacks++
	}
	for index, attempt := range attempts {
		session.Attempts++
		if attempt.StatusCode == 200 {
			session.HTTP200++
		}
		if attempt.ResponsePacketDecoded {
			session.DecodedResponses++
		}
		if attempt.RequestFlags&inform.FlagEncryptedGCM != 0 {
			session.GCMRequests++
		} else if attempt.RequestFlags&inform.FlagEncrypted != 0 {
			session.CBCRequests++
		}

		finalAttempt := index == len(attempts)-1
		event := informExchangeStatus{
			Sequence:              session.Attempts,
			RequestFlags:          attempt.RequestFlags,
			KeySource:             attempt.KeySource,
			StatusCode:            attempt.StatusCode,
			ResponseBodyPresent:   attempt.ResponseBodyPresent,
			ResponsePacketDecoded: attempt.ResponsePacketDecoded,
			ResponseFlags:         attempt.ResponseFlags,
			UsedAESGCM:            attempt.UsedAESGCM,
			Failed:                attempt.Failed,
		}
		if finalAttempt {
			event.ResponseJSONParsed = last.ResponseJSONParsed
			event.ResponseType = last.ResponseType
			event.HasMgmtCFG = last.HasMgmtCFG
			event.FallbackToCBC = last.FallbackToCBC
			event.Failed = event.Failed || last.Error != ""
			if last.ResponseJSONParsed {
				session.ParsedResponses++
				switch last.ResponseType {
				case "setparam":
					session.SetParamResponses++
				case "noop":
					session.NoopResponses++
				}
			}
		}
		session.Events = append(session.Events, event)
	}
	if len(session.Events) > maxInformSessionEvents {
		session.Events = append([]informExchangeStatus(nil), session.Events[len(session.Events)-maxInformSessionEvents:]...)
	}
}

// newLastInformStatus starts a sanitized status record before the controller
// response is decoded.
func newLastInformStatus(url string, store adoption.Store) lastInformStatus {
	wireState := controllerInformState(store)
	wireStateValue := int(wireState)
	return lastInformStatus{
		Time:            time.Now().Format(time.RFC3339),
		URL:             url,
		ControllerState: adoptionStateText(store),
		WireState:       &wireStateValue,
		WireStateName:   wireState.String(),
		CFGVersion:      store.CFGVersion,
		Version:         store.Version,
	}
}

// Controller adoption responses are reduced to local stub-state transitions.
// Provisioning, reset, and upgrade requests are summarized or persisted only
// when they are safe for future inform requests; host state is never changed.
package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/konstruktor1/unifi-stubd/internal/adoption"
	"github.com/konstruktor1/unifi-stubd/internal/inform"
)

const (
	resetDispositionApplied        = "applied"
	resetDispositionAlreadyFactory = "already_factory"
	resetDispositionFailed         = "failed"
)

// loadAdoptionState reads persisted controller adoption data and treats a
// missing state file as factory state so first boot stays non-fatal. A present
// but invalid file fails closed instead of being mistaken for a new device.
func loadAdoptionState(path string) (adoption.Store, error) {
	store, err := adoption.LoadEnv(path)
	if err == nil {
		return store, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return adoption.Store{State: adoption.StateFactory}, nil
	}
	return adoption.Store{}, fmt.Errorf("load adoption state: %w", err)
}

// effectiveInformURL prefers the adopted controller URL because set-inform
// responses are the controller's durable endpoint for future heartbeats.
func effectiveInformURL(fallback string, store adoption.Store) string {
	if store.InformURL != "" {
		return store.InformURL
	}
	return fallback
}

// updateAdoptionState applies only sanitized controller response data to the
// local adoption store. Controller reset and upgrade requests never mutate host
// services or firmware.
func updateAdoptionState(path string, store adoption.Store, response *adoption.ControllerResponse, usedGCM bool) (adoption.Store, error) {
	if response == nil {
		return store, nil
	}
	if response.ResetRequested {
		alreadyFactory := adoptionStoreAlreadyFactory(store)
		resetStore, err := adoption.ResetEnv(path)
		if err != nil {
			response.ResetApplied = false
			response.ResetDisposition = resetDispositionFailed
			log.Printf("adoption state reset failed: %v", err)
			return store, fmt.Errorf("persist controller reset: %w", err)
		}
		response.ResetApplied = true
		response.ResetDisposition = resetDispositionApplied
		if alreadyFactory {
			response.ResetDisposition = resetDispositionAlreadyFactory
		}
		log.Printf("controller reset applied to stub adoption state only: %s", response.ResetReason)
		return resetStore, nil
	}
	update := response.Store
	if !response.HasStateUpdate {
		if usedGCM && store.AuthKey != "" && !store.UseAESGCM {
			updated := store
			updated.UseAESGCM = true
			if err := adoption.SaveEnv(path, updated); err != nil {
				log.Printf("adoption state write failed: %v", err)
				return store, fmt.Errorf("persist adoption cipher state: %w", err)
			}
			return updated, nil
		}
		return store, nil
	}
	if usedGCM {
		update.UseAESGCM = true
	}
	updated, changed := adoption.Merge(store, update)
	if changed {
		if err := adoption.SaveEnv(path, updated); err != nil {
			log.Printf("adoption state write failed: %v", err)
			return store, fmt.Errorf("persist controller adoption state: %w", err)
		}
	}
	return updated, nil
}

func adoptionStoreAlreadyFactory(store adoption.Store) bool {
	return store.AuthKey == "" && store.InformURL == "" && store.CFGVersion == "" &&
		!store.UseAESGCM && store.Version == "" &&
		(store.State == "" || store.State == adoption.StateFactory)
}

// logInformResponse writes a compact audit line for the controller response so
// operators can see accepted, ignored, and reset-like actions without raw JSON.
func logInformResponse(resp *inform.Response, response adoption.ControllerResponse, store adoption.Store, cipher informCipherStatus) {
	if response.Type != "" {
		if response.Type == "setparam" {
			log.Printf(
				"inform response status=%d setparam cfgversion=%q inform_url=%q use_aes_gcm=%t authkey_set=%t mgmt_cfg=%t system_cfg=%t system_cfg_bytes=%d ignored=%t cipher=%s",
				resp.StatusCode,
				store.CFGVersion,
				store.InformURL,
				store.UseAESGCM,
				store.AuthKey != "",
				response.HasMgmtCFG,
				response.HasSystemCFG,
				response.SystemCFGBytes,
				response.Ignored,
				cipherStatusText(cipher),
			)
			return
		}
		if response.ResetRequested {
			log.Printf(
				"inform response status=%d type=%s reset_requested=true reset_applied=%t reset_disposition=%s reason=%q state=%q cipher=%s",
				resp.StatusCode,
				response.Type,
				response.ResetApplied,
				response.ResetDisposition,
				response.ResetReason,
				store.State,
				cipherStatusText(cipher),
			)
			return
		}
		if response.Ignored {
			log.Printf(
				"inform response status=%d type=%s ignored=true reason=%q state=%q version=%q cipher=%s",
				resp.StatusCode,
				response.Type,
				response.IgnoredReason,
				store.State,
				store.Version,
				cipherStatusText(cipher),
			)
			return
		}
		log.Printf(
			"inform response status=%d type=%s state=%q version=%q interval=%d immediate=%t block_selection_limited=%t include_blocks=%q exclude_blocks=%q udapi_blocks=%q cipher=%s",
			resp.StatusCode,
			response.Type,
			store.State,
			store.Version,
			response.IntervalSeconds,
			response.Immediate,
			response.BlockSelectionLimited,
			response.IncludeBlocks,
			response.ExcludeBlocks,
			response.UDAPIBlocks,
			cipherStatusText(cipher),
		)
		return
	}
	log.Printf("inform response status=%d decoded_json_bytes=%d type=unknown cipher=%s", resp.StatusCode, len(resp.JSONBody), cipherStatusText(cipher))
}

// cipherStatusText compresses inform cipher negotiation into stable audit and
// status labels.
func cipherStatusText(cipher informCipherStatus) string {
	switch {
	case cipher.UsedAESGCM:
		return "aes-gcm"
	case cipher.FallbackToCBC:
		return "aes-cbc-fallback"
	case cipher.AttemptedAESGCM:
		return "aes-cbc-after-gcm-attempt"
	default:
		return "aes-cbc"
	}
}

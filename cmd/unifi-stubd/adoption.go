package main

import (
	"errors"
	"log"
	"os"

	"github.com/konstruktor1/unifi-stubd/internal/adoption"
	"github.com/konstruktor1/unifi-stubd/internal/inform"
)

func loadAdoptionState(path string) adoption.Store {
	store, err := adoption.LoadEnv(path)
	if err == nil {
		return store
	}
	if !errors.Is(err, os.ErrNotExist) {
		log.Printf("adoption state read failed: %v", err)
	}
	return adoption.Store{}
}

func effectiveInformURL(fallback string, store adoption.Store) string {
	if store.InformURL != "" {
		return store.InformURL
	}
	return fallback
}

func updateAdoptionState(path string, store adoption.Store, response adoption.ControllerResponse, usedGCM bool) adoption.Store {
	update := response.Store
	if !response.HasStateUpdate {
		if usedGCM && store.AuthKey != "" && !store.UseAESGCM {
			store.UseAESGCM = true
			if err := adoption.SaveEnv(path, store); err != nil {
				log.Printf("adoption state write failed: %v", err)
			}
		}
		return store
	}
	if usedGCM {
		update.UseAESGCM = true
	}
	store, changed := adoption.Merge(store, update)
	if changed {
		if err := adoption.SaveEnv(path, store); err != nil {
			log.Printf("adoption state write failed: %v", err)
		}
	}
	if response.Type == "upgrade" {
		log.Printf("controller requested firmware version %q; reporting it from next inform", store.Version)
	}
	return store
}

func logInformResponse(resp *inform.Response, response adoption.ControllerResponse, store adoption.Store) {
	if response.Type != "" {
		if response.Type == "setparam" {
			log.Printf(
				"inform response status=%d setparam cfgversion=%q inform_url=%q use_aes_gcm=%t authkey_set=%t mgmt_cfg=%t system_cfg=%t system_cfg_bytes=%d ignored=%t",
				resp.StatusCode,
				store.CFGVersion,
				store.InformURL,
				store.UseAESGCM,
				store.AuthKey != "",
				response.HasMgmtCFG,
				response.HasSystemCFG,
				response.SystemCFGBytes,
				response.Ignored,
			)
			return
		}
		if response.Ignored {
			log.Printf(
				"inform response status=%d type=%s ignored=true reason=%q state=%q version=%q",
				resp.StatusCode,
				response.Type,
				response.IgnoredReason,
				store.State,
				store.Version,
			)
			return
		}
		log.Printf(
			"inform response status=%d type=%s state=%q version=%q interval=%d include_blocks=%v",
			resp.StatusCode,
			response.Type,
			store.State,
			store.Version,
			response.IntervalSeconds,
			response.IncludeBlocks,
		)
		return
	}
	log.Printf("inform response status=%d decoded_json_bytes=%d type=unknown", resp.StatusCode, len(resp.JSONBody))
}

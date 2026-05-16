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

func updateAdoptionState(path string, store adoption.Store, body []byte, usedGCM bool) adoption.Store {
	update, kind, ok, err := adoption.ParseControllerResponse(body)
	if err != nil {
		log.Printf("controller response parse failed: %v", err)
		return store
	}
	if !ok {
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
	if kind == "upgrade" {
		log.Printf("controller requested firmware version %q; reporting it from next inform", store.Version)
	}
	return store
}

func logInformResponse(resp *inform.Response, store adoption.Store) {
	if _, kind, ok, _ := adoption.ParseControllerResponse(resp.JSONBody); ok {
		if kind == "setparam" {
			log.Printf(
				"inform response status=%d setparam cfgversion=%q inform_url=%q use_aes_gcm=%t authkey_set=%t",
				resp.StatusCode,
				store.CFGVersion,
				store.InformURL,
				store.UseAESGCM,
				store.AuthKey != "",
			)
			return
		}
		log.Printf(
			"inform response status=%d type=%s state=%q version=%q",
			resp.StatusCode,
			kind,
			store.State,
			store.Version,
		)
		return
	}
	log.Printf("inform response status=%d body=%s", resp.StatusCode, string(resp.JSONBody))
}

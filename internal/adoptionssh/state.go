package adoptionssh

import (
	"errors"
	"log"
	"os"

	"github.com/konstruktor1/unifi-stubd/internal/adoption"
)

// saveState persists adoption values as local stub state, never as shell
// instructions or audit payloads.
func (h *Handler) saveState(informURL, authKey string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	statePath := h.config.StatePath
	if statePath == "" {
		statePath = defaultStatePath
	}
	store, err := adoption.LoadEnv(statePath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Printf("adoption state read failed: %v", err)
	}
	update := adoption.Store{
		State:     adoption.StateAdopting,
		InformURL: informURL,
		AuthKey:   authKey,
	}
	store, changed := adoption.Merge(store, update)
	if err := adoption.SaveEnv(statePath, store); err != nil {
		log.Printf("adoption state write failed: %v", err)
		return
	}
	if changed {
		log.Printf("adoption state updated")
	}
}

// resetState clears only the adoption store, making the next inform look
// factory-default without resetting the host.
func (h *Handler) resetState() {
	h.mu.Lock()
	defer h.mu.Unlock()

	statePath := h.config.StatePath
	if statePath == "" {
		statePath = defaultStatePath
	}
	if _, err := adoption.ResetEnv(statePath); err != nil {
		log.Printf("adoption state reset failed: %v", err)
		return
	}
	log.Printf("adoption state reset accepted")
}

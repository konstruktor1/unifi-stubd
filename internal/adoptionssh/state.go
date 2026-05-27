package adoptionssh

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/konstruktor1/unifi-stubd/internal/adoption"
)

// saveState persists adoption commands as local stub state and audit metadata,
// never as shell instructions to execute.
func (h *Handler) saveState(informURL, authKey, command string) {
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
	store, _ = adoption.Merge(store, update)
	if err := adoption.SaveEnv(statePath, store); err != nil {
		log.Printf("adoption state write failed: %v", err)
	}
	if command != "" {
		log.Printf("adoption state command accepted at %s: %s", time.Now().UTC().Format(time.RFC3339), command)
	}
}

// resetState clears only the adoption store, making the next inform look
// factory-default without resetting the host.
func (h *Handler) resetState(command string) {
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
	if command != "" {
		log.Printf("adoption state reset accepted at %s: %s", time.Now().UTC().Format(time.RFC3339), command)
	}
}

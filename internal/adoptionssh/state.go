package adoptionssh

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/konstruktor1/unifi-stubd/internal/adoption"
)

// saveState persists adoption values as local stub state, never as shell
// instructions or audit payloads.
func (h *Handler) saveState(informURL, authKey string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if informURL == "" && authKey == "" {
		return nil
	}

	statePath := h.config.StatePath
	if statePath == "" {
		statePath = defaultStatePath
	}
	store, err := adoption.LoadEnv(statePath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Printf("adoption state read failed")
		return fmt.Errorf("load adoption state: %w", err)
	}
	state := adoption.StateProvisioning
	if authKey != "" {
		state = adoption.StateAdopting
	}
	update := adoption.Store{
		State:     state,
		InformURL: informURL,
		AuthKey:   authKey,
	}
	store, changed := adoption.Merge(store, update)
	if err := adoption.SaveEnv(statePath, store); err != nil {
		log.Printf("adoption state write failed")
		return fmt.Errorf("save adoption state: %w", err)
	}
	if changed {
		log.Printf("adoption state updated")
	}
	return nil
}

// resetState clears only the adoption store, making the next inform look
// factory-default without resetting the host.
func (h *Handler) resetState() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	statePath := h.config.StatePath
	if statePath == "" {
		statePath = defaultStatePath
	}
	if _, err := adoption.ResetEnv(statePath); err != nil {
		log.Printf("adoption state reset failed")
		return fmt.Errorf("reset adoption state: %w", err)
	}
	log.Printf("adoption state reset accepted")
	return nil
}

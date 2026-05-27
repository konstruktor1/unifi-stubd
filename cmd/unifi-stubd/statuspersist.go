package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/konstruktor1/unifi-stubd/internal/adoption"
)

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

// saveLastInformStatus persists only the last sanitized inform summary, keeping
// runtime status narrow and safe to inspect.
func saveLastInformStatus(path string, last lastInformStatus) error {
	if path == "" {
		return errors.New("status path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create runtime status directory: %w", err)
	}
	data, err := json.MarshalIndent(persistedRunStatus{LastInform: last}, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal runtime status: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write runtime status %s: %w", path, err)
	}
	return nil
}

// newLastInformStatus starts a sanitized status record before the controller
// response is decoded.
func newLastInformStatus(url string, store adoption.Store) lastInformStatus {
	return lastInformStatus{
		Time:            time.Now().Format(time.RFC3339),
		URL:             url,
		ControllerState: adoptionStateText(store),
		CFGVersion:      store.CFGVersion,
		Version:         store.Version,
	}
}

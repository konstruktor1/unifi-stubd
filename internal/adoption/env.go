// Package adoption persists the minimal controller state needed for later inform
// requests. The env mapping table intentionally whitelists accepted mgmt_cfg
// keys instead of applying arbitrary controller provisioning data.
package adoption

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LoadEnv reads adoption state from a key-value environment file.
func LoadEnv(path string) (Store, error) {
	var store Store
	data, err := os.ReadFile(path)
	if err != nil {
		return store, fmt.Errorf("read adoption state %s: %w", path, err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		key, value, ok := strings.Cut(strings.TrimSpace(line), "=")
		if !ok {
			continue
		}
		if field, ok := storeFieldByEnvKey(key); ok {
			field.set(&store, value)
		}
	}
	return store, nil
}

// SaveEnv writes adoption state to a key-value environment file.
func SaveEnv(path string, store Store) error {
	if path == "" {
		return errors.New("state path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create adoption state directory: %w", err)
	}
	var b strings.Builder
	for _, field := range storeFields {
		if value := field.get(store); value != "" {
			b.WriteString(field.envKey + "=" + value + "\n")
		}
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o600); err != nil {
		return fmt.Errorf("write adoption state %s: %w", path, err)
	}
	return nil
}

// ResetEnv clears adoption data and persists the stub in factory state.
func ResetEnv(path string) (Store, error) {
	store := Store{State: StateFactory}
	if err := SaveEnv(path, store); err != nil {
		return store, err
	}
	return store, nil
}

// Merge applies non-empty fields from update to base and reports changes.
func Merge(base, update Store) (Store, bool) {
	changed := false
	for _, field := range storeFields {
		if value := field.get(update); value != "" && field.get(base) != value {
			field.set(&base, value)
			changed = true
		}
	}
	return base, changed
}

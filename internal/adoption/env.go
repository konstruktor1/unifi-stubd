// Package adoption persists the minimal controller state needed for later inform
// requests. The env mapping table intentionally whitelists accepted mgmt_cfg
// keys instead of applying arbitrary controller provisioning data.
package adoption

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/atomicfile"
	"github.com/konstruktor1/unifi-stubd/internal/inform"
)

// LoadEnv reads adoption state from a key-value environment file.
func LoadEnv(path string) (Store, error) {
	var store Store
	data, err := os.ReadFile(path)
	if err != nil {
		return store, fmt.Errorf("read adoption state %s: %w", path, err)
	}
	seen := make(map[string]struct{}, len(storeFields))
	recognized := 0
	for lineIndex, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if line != "" {
				return Store{}, fmt.Errorf("invalid persisted adoption line %d: whitespace-only line", lineIndex+1)
			}
			continue
		}
		if line != trimmed {
			return Store{}, fmt.Errorf("invalid persisted adoption line %d: non-canonical surrounding whitespace", lineIndex+1)
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return Store{}, fmt.Errorf("invalid persisted adoption line %d: expected KEY=VALUE", lineIndex+1)
		}
		if key == "" || key != strings.TrimSpace(key) {
			return Store{}, fmt.Errorf("invalid persisted adoption line %d: malformed key", lineIndex+1)
		}
		if field, ok := storeFieldByEnvKey(key); ok {
			if _, duplicate := seen[key]; duplicate {
				return Store{}, fmt.Errorf("invalid persisted adoption field %s: duplicate field", key)
			}
			seen[key] = struct{}{}
			recognized++
			if value == "" {
				return Store{}, fmt.Errorf("invalid persisted adoption field %s: value must not be empty", key)
			}
			if err := validatePersistedValue(value); err != nil {
				return Store{}, fmt.Errorf("invalid persisted adoption field %s: %w", key, err)
			}
			if err := validatePersistedField(field, value); err != nil {
				return Store{}, fmt.Errorf("invalid persisted adoption field %s: %w", key, err)
			}
			field.set(&store, value)
		}
	}
	if recognized == 0 {
		return Store{}, errors.New("invalid persisted adoption state: no supported fields")
	}
	if err := validatePersistedStore(store); err != nil {
		return Store{}, fmt.Errorf("invalid persisted adoption state: %w", err)
	}
	return store, nil
}

func validatePersistedValue(value string) error {
	if value != strings.TrimSpace(value) {
		return errors.New("value has non-canonical surrounding whitespace")
	}
	for _, char := range []byte(value) {
		if char < 0x20 || char == 0x7f {
			return errors.New("value contains a control byte")
		}
	}
	return nil
}

func validatePersistedField(field storeField, value string) error {
	switch field.envKey {
	case "AUTHKEY":
		return validateIndividualAuthKeyText(value, true)
	case "CFGVERSION":
		return validateMgmtField("cfgversion", value)
	case "INFORM_URL":
		return validateMgmtField("inform_url", value)
	case "USE_AES_GCM":
		if _, err := strconv.ParseBool(value); err != nil {
			return errors.New("USE_AES_GCM must be boolean")
		}
	}
	return nil
}

func validatePersistedStore(store Store) error {
	switch store.State {
	case "", StateFactory, StateDiscovered, StateAdopting, StateProvisioning, StateConnected, StateFailed:
	default:
		return errors.New("unknown STATE value")
	}
	if store.AuthKey == "" && (store.State == StateAdopting || store.State == StateConnected || store.UseAESGCM) {
		return errors.New("adopted state requires AUTHKEY")
	}
	if store.State == StateFactory && store.AuthKey != "" {
		return errors.New("factory state must not contain AUTHKEY")
	}
	return nil
}

func validateIndividualAuthKeyText(value string, allowRaw bool) error {
	var key []byte
	switch {
	case allowRaw && len(value) == 16:
		key = []byte(value)
	case len(value) == 32:
		decoded, err := hex.DecodeString(value)
		if err != nil {
			return errors.New("authkey must be 16 raw bytes or 32 hexadecimal characters")
		}
		key = decoded
	default:
		return errors.New("authkey must be 16 raw bytes or 32 hexadecimal characters")
	}
	if inform.IsDefaultAuthKey(key) {
		return errors.New("individual authkey must not equal the public default key")
	}
	return nil
}

// SaveEnv writes adoption state to a key-value environment file.
func SaveEnv(path string, store Store) error {
	if path == "" {
		return errors.New("state path is required")
	}
	if err := validateStoreForPersistence(store); err != nil {
		return fmt.Errorf("invalid adoption state for persistence: %w", err)
	}
	var b strings.Builder
	for _, field := range storeFields {
		if value := field.get(store); value != "" {
			b.WriteString(field.envKey + "=" + value + "\n")
		}
	}
	if err := atomicfile.WriteFile(path, []byte(b.String()), 0o600); err != nil {
		return fmt.Errorf("write adoption state %s: %w", path, err)
	}
	return nil
}

func validateStoreForPersistence(store Store) error {
	fields := 0
	for _, field := range storeFields {
		value := field.get(store)
		if value == "" {
			continue
		}
		fields++
		if err := validatePersistedValue(value); err != nil {
			return fmt.Errorf("field %s: %w", field.envKey, err)
		}
		if err := validatePersistedField(field, value); err != nil {
			return fmt.Errorf("field %s: %w", field.envKey, err)
		}
	}
	if fields == 0 {
		return errors.New("no supported fields")
	}
	return validatePersistedStore(store)
}

// ResetEnv clears all persisted adoption data and persists factory state.
// VERSION is cleared too because legacy releases could learn it from a
// controller upgrade response; operator configuration remains outside this
// controller-owned store.
func ResetEnv(path string) (Store, error) {
	store := Store{State: StateFactory}
	if err := SaveEnv(path, store); err != nil {
		return store, err
	}
	return store, nil
}

func authKeysEqual(left, right string) bool {
	if left == right {
		return true
	}
	leftKey, leftOK := normalizedAuthKey(left)
	rightKey, rightOK := normalizedAuthKey(right)
	return leftOK && rightOK && leftKey == rightKey
}

func normalizedAuthKey(value string) (string, bool) {
	switch len(value) {
	case 16:
		return hex.EncodeToString([]byte(value)), true
	case 32:
		decoded, err := hex.DecodeString(value)
		if err != nil {
			return "", false
		}
		return hex.EncodeToString(decoded), true
	default:
		return "", false
	}
}

// Merge applies non-empty fields from update to base and reports changes.
func Merge(base, update Store) (Store, bool) {
	// Firmware returns to wire state 0 whenever controller setparam carries an
	// authkey, and when the Inform URL changes. Local SSH updates retain their
	// same-key behavior through byte-normalized comparison. The Stub's separate
	// safety boundary keeps keyless URL/noop paths pending.
	controllerAuthKeyPresent := update.State == StateConnected && update.AuthKey != ""
	keyChanged := update.AuthKey != "" && !authKeysEqual(update.AuthKey, base.AuthKey)
	urlChanged := update.InformURL != "" && update.InformURL != base.InformURL
	switch {
	case controllerAuthKeyPresent || keyChanged || urlChanged:
		resultingKey := base.AuthKey
		if update.AuthKey != "" {
			resultingKey = update.AuthKey
		}
		if resultingKey != "" {
			update.State = StateAdopting
		} else {
			update.State = StateProvisioning
		}
	case update.State == StateConnected && base.AuthKey == "" && update.AuthKey == "":
		// The response remains successful for scheduling/reporting, but it cannot
		// create an invalid persisted Connected state without an individual key.
		update.State = ""
	case update.State == StateAdopting || (base.AuthKey != "" && update.State == StateProvisioning):
		// Repeating identical identity values from local SSH must not regress a
		// keyed session. Only a real local key/URL change reopens adoption.
		update.State = ""
	}
	changed := false
	for _, field := range storeFields {
		if value := field.get(update); value != "" && field.get(base) != value {
			field.set(&base, value)
			changed = true
		}
	}
	return base, changed
}

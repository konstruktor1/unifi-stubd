package adoption

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func LoadEnv(path string) (Store, error) {
	var store Store
	data, err := os.ReadFile(path)
	if err != nil {
		return store, err
	}
	for _, line := range strings.Split(string(data), "\n") {
		key, value, ok := strings.Cut(strings.TrimSpace(line), "=")
		if !ok {
			continue
		}
		switch key {
		case "STATE":
			store.State = State(value)
		case "INFORM_URL":
			store.InformURL = value
		case "AUTHKEY":
			store.AuthKey = value
		case "CFGVERSION":
			store.CFGVersion = value
		case "USE_AES_GCM":
			store.UseAESGCM, _ = strconv.ParseBool(value)
		case "VERSION":
			store.Version = value
		}
	}
	return store, nil
}

func SaveEnv(path string, store Store) error {
	if path == "" {
		return errors.New("state path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var b strings.Builder
	if store.State != "" {
		b.WriteString("STATE=" + string(store.State) + "\n")
	}
	if store.InformURL != "" {
		b.WriteString("INFORM_URL=" + store.InformURL + "\n")
	}
	if store.AuthKey != "" {
		b.WriteString("AUTHKEY=" + store.AuthKey + "\n")
	}
	if store.CFGVersion != "" {
		b.WriteString("CFGVERSION=" + store.CFGVersion + "\n")
	}
	if store.UseAESGCM {
		b.WriteString("USE_AES_GCM=true\n")
	}
	if store.Version != "" {
		b.WriteString("VERSION=" + store.Version + "\n")
	}
	return os.WriteFile(path, []byte(b.String()), 0o600)
}

func Merge(base, update Store) (Store, bool) {
	changed := false
	setString := func(dst *string, src string) {
		if src != "" && *dst != src {
			*dst = src
			changed = true
		}
	}
	setString(&base.InformURL, update.InformURL)
	setString(&base.AuthKey, update.AuthKey)
	setString(&base.CFGVersion, update.CFGVersion)
	setString(&base.Version, update.Version)
	if update.UseAESGCM && !base.UseAESGCM {
		base.UseAESGCM = true
		changed = true
	}
	if update.State != "" && base.State != update.State {
		base.State = update.State
		changed = true
	}
	return base, changed
}

func ParseSetParamResponse(data []byte) (Store, bool, error) {
	store, kind, ok, err := ParseControllerResponse(data)
	return store, ok && kind == "setparam", err
}

func ParseControllerResponse(data []byte) (Store, string, bool, error) {
	var raw struct {
		Type    string `json:"_type"`
		MgmtCFG string `json:"mgmt_cfg"`
		Version string `json:"version"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return Store{}, "", false, err
	}
	switch raw.Type {
	case "setparam":
		if raw.MgmtCFG == "" {
			return Store{}, raw.Type, false, nil
		}
		store := Store{State: StateProvisioning}
		for _, line := range strings.Split(raw.MgmtCFG, "\n") {
			key, value, ok := strings.Cut(strings.TrimSpace(line), "=")
			if !ok {
				continue
			}
			switch key {
			case "inform_url":
				store.InformURL = value
			case "authkey":
				store.AuthKey = value
			case "cfgversion":
				store.CFGVersion = value
			case "use_aes_gcm":
				store.UseAESGCM, _ = strconv.ParseBool(value)
			}
		}
		return store, raw.Type, true, nil
	case "upgrade":
		if raw.Version == "" {
			return Store{}, raw.Type, false, nil
		}
		return Store{
			State:   StateProvisioning,
			Version: raw.Version,
		}, raw.Type, true, nil
	case "noop":
		return Store{State: StateConnected}, raw.Type, true, nil
	default:
		return Store{}, raw.Type, false, nil
	}
}

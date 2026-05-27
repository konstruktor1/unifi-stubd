package adoption

import "strconv"

// storeField maps one whitelisted adoption value between env files and
// controller mgmt_cfg keys.
type storeField struct {
	envKey  string
	mgmtKey string
	get     func(Store) string
	set     func(*Store, string)
}

// storeFields is the adoption-state allowlist accepted from controller
// responses.
var storeFields = []storeField{
	{
		envKey: "STATE",
		get:    func(store Store) string { return string(store.State) },
		set:    func(store *Store, value string) { store.State = State(value) },
	},
	{
		envKey:  "INFORM_URL",
		mgmtKey: "inform_url",
		get:     func(store Store) string { return store.InformURL },
		set:     func(store *Store, value string) { store.InformURL = value },
	},
	{
		envKey:  "AUTHKEY",
		mgmtKey: "authkey",
		get:     func(store Store) string { return store.AuthKey },
		set:     func(store *Store, value string) { store.AuthKey = value },
	},
	{
		envKey:  "CFGVERSION",
		mgmtKey: "cfgversion",
		get:     func(store Store) string { return store.CFGVersion },
		set:     func(store *Store, value string) { store.CFGVersion = value },
	},
	{
		envKey:  "USE_AES_GCM",
		mgmtKey: "use_aes_gcm",
		get: func(store Store) string {
			if store.UseAESGCM {
				return "true"
			}
			return ""
		},
		set: func(store *Store, value string) {
			store.UseAESGCM, _ = strconv.ParseBool(value)
		},
	},
	{
		envKey: "VERSION",
		get:    func(store Store) string { return store.Version },
		set:    func(store *Store, value string) { store.Version = value },
	},
}

// storeHasStateUpdate reports whether parsing found any durable adoption field
// worth persisting.
func storeHasStateUpdate(store Store) bool {
	for _, field := range storeFields {
		if field.get(store) != "" {
			return true
		}
	}
	return false
}

// storeFieldByEnvKey resolves persisted environment keys through the same field
// table used for saving adoption state.
func storeFieldByEnvKey(key string) (storeField, bool) {
	for _, field := range storeFields {
		if field.envKey == key {
			return field, true
		}
	}
	return storeField{}, false
}

// storeFieldByMgmtKey is the allowlist boundary for controller-provided
// mgmt_cfg fields.
func storeFieldByMgmtKey(key string) (storeField, bool) {
	for _, field := range storeFields {
		if field.mgmtKey != "" && field.mgmtKey == key {
			return field, true
		}
	}
	return storeField{}, false
}

package adoption

import "strings"

// parseMgmtCFG accepts only whitelisted mgmt_cfg keys that affect future
// inform identity; other controller provisioning keys are ignored.
func parseMgmtCFG(mgmtCFG string) Store {
	store := Store{State: StateProvisioning}
	for _, line := range strings.Split(mgmtCFG, "\n") {
		key, value, ok := strings.Cut(strings.TrimSpace(line), "=")
		if !ok {
			continue
		}
		// Unknown controller keys are ignored by design. The storeFields table is
		// the policy boundary for adoption data accepted from the controller.
		if field, ok := storeFieldByMgmtKey(key); ok {
			field.set(&store, value)
		}
	}
	return store
}

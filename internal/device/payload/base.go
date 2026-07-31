package payload

import (
	"fmt"

	"github.com/konstruktor1/unifi-stubd/internal/device"
)

type basePayload struct {
	MAC               string `json:"mac"`
	IP                string `json:"ip"`
	Hostname          string `json:"hostname"`
	Model             string `json:"model"`
	ModelDisplay      string `json:"model_display"`
	Type              string `json:"type"`
	Version           string `json:"version"`
	Serial            string `json:"serial"`
	NumPort           int    `json:"num_port"`
	State             int    `json:"state"`
	Adopted           bool   `json:"adopted"`
	Default           bool   `json:"default"`
	DiscoveryResponse bool   `json:"discovery_response"`
	RequiredVersion   string `json:"required_version"`
	CFGVersion        string `json:"cfgversion"`
	Uptime            int    `json:"uptime"`
	Time              int64  `json:"time"`
	InformURL         string `json:"inform_url"`
	ManagementVLAN    int    `json:"management_vlan,omitempty"`
	InformIP          string `json:"inform_ip,omitempty"`
}

// identityUptime clamps reported uptime to a positive value because controller
// freshness checks treat zero-like uptime as suspicious.
func identityUptime(uptime int) int {
	if uptime < 1 {
		return 1
	}
	return uptime
}

// informState resolves the typed lifecycle while preserving the historical
// adopted/default fallback for callers that do not select an explicit state.
func informState(identity device.Identity) (device.InformState, error) {
	if identity.InformStateSet {
		if !identity.InformState.Valid() {
			return 0, fmt.Errorf("invalid inform state %d", identity.InformState)
		}
		return identity.InformState, nil
	}
	if identity.Adopted {
		return device.InformStateManaged, nil
	}
	return device.InformStateDefault, nil
}

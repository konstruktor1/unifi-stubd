package adoption

// State describes the local adoption lifecycle state.
type State string

const (
	// StateFactory means the stub has not been adopted yet.
	StateFactory State = "factory"
	// StateDiscovered means the controller has seen the stub.
	StateDiscovered State = "discovered"
	// StateAdopting means the controller is attempting adoption.
	StateAdopting State = "adopting"
	// StateProvisioning means controller settings are being applied.
	StateProvisioning State = "provisioning"
	// StateConnected means the controller considers the stub connected.
	StateConnected State = "connected"
	// StateFailed means adoption or provisioning failed.
	StateFailed State = "failed"
)

// Store is the persisted adoption data learned from the controller.
type Store struct {
	// InformURL is the controller inform URL assigned to the device.
	InformURL string `json:"inform_url"`
	// AuthKey is the adoption key assigned by the controller.
	AuthKey string `json:"authkey"`
	// CFGVersion is the controller configuration version.
	CFGVersion string `json:"cfgversion"`
	// UseAESGCM reports whether inform traffic should use AES-GCM.
	UseAESGCM bool `json:"use_aes_gcm"`
	// Version is a controller-requested firmware version.
	Version string `json:"version"`
	// State is the current local adoption lifecycle state.
	State State `json:"state"`
}

// ControllerResponse is a sanitized summary of one decoded controller response.
type ControllerResponse struct {
	// Type is the controller response type, such as setparam, noop, or upgrade.
	Type string `json:"type,omitempty"`
	// Store contains safe adoption state updates extracted from the response.
	Store Store `json:"store,omitempty"`
	// HasStateUpdate reports whether Store contains data worth merging.
	HasStateUpdate bool `json:"has_state_update,omitempty"`
	// HasMgmtCFG reports whether the response carried a management config block.
	HasMgmtCFG bool `json:"has_mgmt_cfg,omitempty"`
	// HasSystemCFG reports whether the response carried gateway provisioning data.
	HasSystemCFG bool `json:"has_system_cfg,omitempty"`
	// SystemCFGBytes is the byte length of the provisioning block without content.
	SystemCFGBytes int `json:"system_cfg_bytes,omitempty"`
	// SystemCFGKeys contains top-level provisioning keys only, never raw values.
	SystemCFGKeys []string `json:"system_cfg_keys,omitempty"`
	// IntervalSeconds is the inform interval requested by a noop response.
	IntervalSeconds int `json:"interval_seconds,omitempty"`
	// IncludeBlocks lists controller-requested status blocks when present.
	IncludeBlocks []string `json:"include_blocks,omitempty"`
	// Ignored reports that a controller command was intentionally not applied.
	Ignored bool `json:"ignored,omitempty"`
	// IgnoredReason explains why a command was not applied to the host.
	IgnoredReason string `json:"ignored_reason,omitempty"`
}

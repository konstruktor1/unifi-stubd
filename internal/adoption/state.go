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

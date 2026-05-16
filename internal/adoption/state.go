package adoption

type State string

const (
	StateFactory      State = "factory"
	StateDiscovered   State = "discovered"
	StateAdopting     State = "adopting"
	StateProvisioning State = "provisioning"
	StateConnected    State = "connected"
	StateFailed       State = "failed"
)

type Store struct {
	InformURL  string `json:"inform_url"`
	AuthKey    string `json:"authkey"`
	CFGVersion string `json:"cfgversion"`
	UseAESGCM  bool   `json:"use_aes_gcm"`
	Version    string `json:"version"`
	State      State  `json:"state"`
}

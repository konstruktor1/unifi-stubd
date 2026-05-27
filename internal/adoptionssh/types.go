package adoptionssh

import (
	"net"
	"sync"
)

// Adoption SSH defaults and command tokens match the minimal UniFi command
// surface that controllers probe.
const (
	defaultSSHUser     = "ubnt"
	defaultSSHPassword = "ubnt"
	defaultStatePath   = "/var/lib/unifi-stubd/adoption.env"
	commandInfo        = "info"
	commandSetInform   = "set-inform"
	okOutput           = "OK\n"
)

// Identity describes the device data exposed through adoption SSH commands.
type Identity struct {
	// MAC is the fake device MAC address.
	MAC string
	// IP is the fake device management IP address.
	IP string
	// Hostname is the fake device hostname.
	Hostname string
	// Model is the UniFi model identifier.
	Model string
	// Version is the firmware version reported by the SSH shim.
	Version string
	// InformURL is the controller inform URL currently known by the device.
	InformURL string
}

// Config controls the built-in adoption SSH server.
type Config struct {
	// Listen is the TCP listen address, such as 0.0.0.0:22.
	Listen string
	// User is the accepted SSH username.
	User string
	// Password is the accepted SSH password.
	Password string
	// HostKeyPath stores the persistent SSH host key.
	HostKeyPath string
	// StatePath stores adoption state learned through SSH commands.
	StatePath string
	// Identity is the fake device identity exposed to the controller.
	Identity Identity
}

// Server is a running adoption SSH listener.
type Server struct {
	listener net.Listener
	handler  *Handler
}

// Handler executes the small command subset used during UniFi adoption.
type Handler struct {
	config Config
	mu     sync.Mutex
}

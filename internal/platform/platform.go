package platform

import "runtime"

// hostPlatform is the concrete read-only adapter selected for the current OS or
// tests.
type hostPlatform struct {
	goos string
	cfg  Config
}

// New returns the default read-only platform adapter for the current host.
func New(cfg Config) Platform {
	cfg = normalizeConfig(cfg)
	return hostPlatform{goos: runtime.GOOS, cfg: cfg}
}

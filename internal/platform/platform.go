// Package platform hides optional host integrations behind read-only adapter
// interfaces. It may inspect interfaces, bridges, logs, procfs, D-Bus, and
// LLDP, but it must not mutate host networking or execute controller-provided
// commands.
package platform

import (
	"runtime"
	"strings"
)

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

// NewForOS returns a read-only platform adapter for tests and fixtures.
func NewForOS(goos string, cfg Config) Platform {
	cfg = normalizeConfig(cfg)
	if strings.TrimSpace(goos) == "" {
		goos = runtime.GOOS
	}
	return hostPlatform{goos: strings.ToLower(strings.TrimSpace(goos)), cfg: cfg}
}

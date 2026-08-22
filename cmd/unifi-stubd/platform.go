// Platform wiring builds the read-only OS adapter used by operation modes and
// status output. It deliberately keeps host observation behind one boundary so
// controller-facing payload logic cannot call OS integrations directly.
package main

import (
	"context"
	"log"
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/device"
	"github.com/konstruktor1/unifi-stubd/internal/observe/ifsource"
	"github.com/konstruktor1/unifi-stubd/internal/platform"
)

// runtimePlatform builds the read-only host adapter from resolved runtime
// flags.
func runtimePlatform(flags runtimeFlags) platform.Platform {
	return platform.New(runtimePlatformConfig(flags))
}

// runtimePlatformConfig trims optional source settings before they reach the
// platform adapter.
func runtimePlatformConfig(flags runtimeFlags) platform.Config {
	return platform.Config{
		LLDPSource: strings.TrimSpace(flags.lldpSource),
		ProcSource: strings.TrimSpace(flags.procSource),
	}
}

// enrichOverrides reads local interface facts for explicit operator overrides.
func enrichOverrides(ctx context.Context, plt platform.Platform, overrides []device.PortOverride) []device.PortOverride {
	out, warnings := ifsource.RefreshOverrides(ctx, plt, overrides)
	for _, warning := range warnings {
		log.Printf("port %d interface source %s warning: %v", warning.Port, warning.Interface, warning.Err)
	}
	return out
}

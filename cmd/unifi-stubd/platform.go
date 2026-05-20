// Platform wiring builds the read-only OS adapter used by operation modes and
// status output. It deliberately keeps host observation behind one boundary so
// controller-facing payload logic cannot start calling Linux, FreeBSD, D-Bus,
// LLDP, or log readers directly.
package main

import (
	"context"
	"log"
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/device"
	"github.com/konstruktor1/unifi-stubd/internal/observe/ifsource"
	"github.com/konstruktor1/unifi-stubd/internal/platform"
)

func runtimePlatform(flags runtimeFlags) platform.Platform {
	return platform.New(runtimePlatformConfig(flags))
}

func runtimePlatformConfig(flags runtimeFlags) platform.Config {
	return platform.Config{
		LLDPSource:    strings.TrimSpace(flags.lldpSource),
		TrafficSource: strings.TrimSpace(flags.trafficSource),
		LogSource:     strings.TrimSpace(flags.logSource),
		ProcSource:    strings.TrimSpace(flags.procSource),
		DBusEnabled:   flags.dbusEnabled,
		DBusBus:       strings.TrimSpace(flags.dbusBus),
		SyslogPath:    strings.TrimSpace(flags.syslogPath),
	}
}

func enrichPortOverridesWithPlatform(ctx context.Context, plt platform.Platform, overrides []device.PortOverride) []device.PortOverride {
	if len(overrides) == 0 {
		return overrides
	}
	out := device.ClonePortOverrides(overrides)
	for index := range out {
		iface := strings.TrimSpace(out[index].Interface)
		if iface == "" {
			continue
		}
		out[index].Interface = iface
		observation, errs := plt.Interface(ctx, iface)
		for _, err := range errs {
			log.Printf("port %d interface source %s warning: %v", out[index].Port, iface, err)
		}
		ifsource.ApplyObservation(&out[index], observation)
	}
	return out
}

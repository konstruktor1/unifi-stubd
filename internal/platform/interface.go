package platform

import (
	"context"
	"fmt"
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/observe"
	"github.com/konstruktor1/unifi-stubd/internal/observe/ifsource"
)

// Ports resolves explicit port-map entries into portable observations, using
// interface reads only for mappings that request a real host interface.
func (p hostPlatform) Ports(ctx context.Context, cfg observe.PortMapConfig) (observe.PortMapObservation, []error) {
	out := observe.PortMapObservation{Ports: map[int]observe.PortObservation{}}
	var errs []error
	for _, mapping := range cfg.Mappings {
		switch {
		case strings.TrimSpace(mapping.Interface) != "":
			observation, warnings := p.Interface(ctx, mapping.Interface)
			observation.Port = mapping.Port
			out.Ports[mapping.Port] = observation
			for _, warning := range warnings {
				errs = append(errs, fmt.Errorf("port %d interface %s: %w", mapping.Port, strings.TrimSpace(mapping.Interface), warning))
			}
		case mapping.Disabled:
			up := false
			out.Ports[mapping.Port] = observe.PortObservation{Port: mapping.Port, Up: &up, Disabled: true}
		case mapping.Unmapped:
			out.Ports[mapping.Port] = observe.PortObservation{Port: mapping.Port, Unmapped: true}
		}
	}
	return out, errs
}

// Interface reads one host interface and optionally enriches counters from
// procfs when that source is enabled.
func (p hostPlatform) Interface(ctx context.Context, iface string) (observe.PortObservation, []error) {
	observation, errs := ifsource.ObserveInterface(iface)
	if p.goos != goosLinux || p.cfg.ProcSource != ProcSourceProcFS {
		return observation, errs
	}
	// procfs can provide counters when sysfs/interface probing found link
	// metadata but incomplete traffic values. It is merged as read-only fallback
	// data for status and payload rendering.
	procSnapshot, procErrs := p.Proc(ctx, ProcConfig{Source: p.cfg.ProcSource})
	if stats, ok := procSnapshot.Interfaces[strings.TrimSpace(iface)]; ok {
		observation.Stats = mergeInterfaceStats(observation.Stats, stats)
	}
	return observation, append(errs, procErrs...)
}

// mergeInterfaceStats fills missing fields from a fallback counter source while
// preserving values already returned by the primary interface reader.
func mergeInterfaceStats(primary, fallback observe.InterfaceStats) observe.InterfaceStats {
	if primary.RXBytes == 0 {
		primary.RXBytes = fallback.RXBytes
	}
	if primary.TXBytes == 0 {
		primary.TXBytes = fallback.TXBytes
	}
	if primary.RXPackets == 0 {
		primary.RXPackets = fallback.RXPackets
	}
	if primary.TXPackets == 0 {
		primary.TXPackets = fallback.TXPackets
	}
	if primary.RXErrors == 0 {
		primary.RXErrors = fallback.RXErrors
	}
	if primary.TXErrors == 0 {
		primary.TXErrors = fallback.TXErrors
	}
	if primary.SpeedMbps == 0 {
		primary.SpeedMbps = fallback.SpeedMbps
	}
	return primary
}

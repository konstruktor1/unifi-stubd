package ifsource

import (
	"context"
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/device"
	"github.com/konstruktor1/unifi-stubd/internal/observe"
)

// InterfaceReader is the read-only interface subset needed to refresh
// interface-backed port overrides.
type InterfaceReader interface {
	Interface(context.Context, string) (observe.PortObservation, []error)
}

// OverrideWarning associates one interface read warning with its configured
// port without embedding it in the refreshed override.
type OverrideWarning struct {
	// Port is the one-based configured port index.
	Port int
	// Interface is the configured host interface name.
	Interface string
	// Err is the source warning returned by the platform reader.
	Err error
}

// RefreshOverrides reads one new snapshot for every interface-backed override.
// It always starts from configured so prior observations cannot become implicit
// operator values on later heartbeats.
func RefreshOverrides(ctx context.Context, reader InterfaceReader, configured []device.PortOverride) ([]device.PortOverride, []OverrideWarning) {
	out := device.ClonePortOverrides(configured)
	if len(out) == 0 || reader == nil {
		return out, nil
	}

	var warnings []OverrideWarning
	for index := range out {
		iface := strings.TrimSpace(out[index].Interface)
		if iface == "" {
			continue
		}
		out[index].Interface = iface
		observation, errs := reader.Interface(ctx, iface)
		for _, err := range errs {
			warnings = append(warnings, OverrideWarning{Port: out[index].Port, Interface: iface, Err: err})
		}
		ApplyObservation(&out[index], observation)
	}
	return out, warnings
}

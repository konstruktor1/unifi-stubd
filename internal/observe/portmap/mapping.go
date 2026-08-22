// Package portmap converts explicit port-map config into payload overrides.
package portmap

import (
	"context"
	"errors"
	"strings"

	appconfig "github.com/konstruktor1/unifi-stubd/internal/config"
	"github.com/konstruktor1/unifi-stubd/internal/device"
	"github.com/konstruktor1/unifi-stubd/internal/observe"
	"github.com/konstruktor1/unifi-stubd/internal/observe/ifsource"
)

// OverridesFromSource returns payload overrides from a portable observation source.
func OverridesFromSource(ctx context.Context, source observe.ObservationSource, mappings []appconfig.PortMapping) ([]device.PortOverride, []error) {
	if len(mappings) == 0 {
		return nil, nil
	}
	if source == nil {
		return nil, []error{errors.New("observation source is required")}
	}
	observation, errs := source.Ports(ctx, observe.PortMapConfig{Mappings: observePortMappings(mappings)})
	return OverridesFromObservation(mappings, observation), errs
}

// OverridesFromObservation returns payload overrides from already-read port observations.
func OverridesFromObservation(mappings []appconfig.PortMapping, observation observe.PortMapObservation) []device.PortOverride {
	out := make([]device.PortOverride, 0, len(mappings))
	for _, mapping := range mappings {
		switch {
		case strings.TrimSpace(mapping.Interface) != "":
			out = append(out, OverrideFromObservation(mapping.Port, observation.Ports[mapping.Port]))
		case mapping.Disabled:
			out = append(out, disabledOverride(mapping.Port))
		case mapping.Unmapped:
			continue
		default:
			continue
		}
	}
	return out
}

// OverrideFromObservation maps one portable source observation to a port override.
func OverrideFromObservation(port int, observation observe.PortObservation) device.PortOverride {
	override := device.PortOverride{
		Port:      port,
		Interface: strings.TrimSpace(observation.Interface),
	}
	if observation.Port > 0 {
		override.Port = observation.Port
	}
	ifsource.ApplyObservation(&override, observation)
	return override
}

// disabledOverride renders a mapped port as explicitly disabled and down.
func disabledOverride(port int) device.PortOverride {
	up := false
	return device.PortOverride{Port: port, Up: &up, Disabled: true}
}

// observePortMappings converts config mappings into the portable observation
// request shape.
func observePortMappings(mappings []appconfig.PortMapping) []observe.PortMapping {
	out := make([]observe.PortMapping, 0, len(mappings))
	for _, mapping := range mappings {
		out = append(out, observe.PortMapping{
			Port:      mapping.Port,
			Interface: strings.TrimSpace(mapping.Interface),
			Disabled:  mapping.Disabled,
			Unmapped:  mapping.Unmapped,
		})
	}
	return out
}

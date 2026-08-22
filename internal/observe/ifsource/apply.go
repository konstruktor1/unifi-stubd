package ifsource

import (
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/device"
	"github.com/konstruktor1/unifi-stubd/internal/observe"
)

// ApplyObservation overlays one portable interface observation onto a port
// override without replacing operator-specified values.
func ApplyObservation(override *device.PortOverride, observation observe.PortObservation) {
	if strings.TrimSpace(override.Interface) == "" {
		override.Interface = strings.TrimSpace(observation.Interface)
	}
	if strings.TrimSpace(override.MAC) == "" {
		override.MAC = observation.MAC
	}
	if override.Up == nil {
		up := observation.Up != nil && *observation.Up
		override.Up = &up
	}
	if strings.TrimSpace(override.IP) == "" {
		override.IP = observation.IP
	}
	if strings.TrimSpace(override.Netmask) == "" {
		override.Netmask = observation.Netmask
	}
	if len(override.IPv6) == 0 && len(observation.IPv6) > 0 {
		override.IPv6 = append([]string(nil), observation.IPv6...)
	}
	if override.Speed <= 0 && *override.Up && observation.SpeedMbps > 0 {
		override.Speed = observation.SpeedMbps
	}
	if strings.TrimSpace(override.Media) == "" {
		override.Media = observation.Media
	}
	for _, field := range interfaceCounterFields {
		field.setOverride(override, field.get(observation.Stats))
	}
	override.CountersSet = observation.Stats.CountersSet || observationHasCounters(observation.Stats)
}

func observationHasCounters(stats observe.InterfaceStats) bool {
	for _, field := range interfaceCounterFields {
		if field.get(stats) != 0 {
			return true
		}
	}
	return false
}

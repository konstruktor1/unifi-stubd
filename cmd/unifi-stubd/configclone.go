package main

import (
	appconfig "github.com/konstruktor1/unifi-stubd/internal/config"
	"github.com/konstruktor1/unifi-stubd/internal/device"
)

// configPortOverrides detaches YAML overrides before later platform enrichment
// and payload merging can mutate runtime copies.
func configPortOverrides(overrides []appconfig.PortOverride) []device.PortOverride {
	return device.ClonePortOverrides(overrides)
}

// cloneWANHealth detaches active probe targets from loaded config.
func cloneWANHealth(value appconfig.WANHealthConfig) appconfig.WANHealthConfig {
	value.Targets = cloneNonEmptySlice(value.Targets)
	return value
}

// cloneStrings returns a detached copy for config slices that runtime code may
// normalize later.
func cloneStrings(values []string) []string {
	return cloneNonEmptySlice(values)
}

// cloneBridgeObserve detaches nested bridge-observe slices from loaded config.
func cloneBridgeObserve(value appconfig.BridgeObserve) appconfig.BridgeObserve {
	value.IgnoredMembers = cloneStrings(value.IgnoredMembers)
	value.MemberPortMap = cloneBridgeMemberPortMaps(value.MemberPortMap)
	return value
}

// cloneBridgeMemberPortMaps detaches bridge member pinning entries from config
// input.
func cloneBridgeMemberPortMaps(values []appconfig.BridgeMemberPortMap) []appconfig.BridgeMemberPortMap {
	return cloneNonEmptySlice(values)
}

// clonePortMappings detaches explicit port-map entries before observation code
// can normalize them.
func clonePortMappings(values []appconfig.PortMapping) []appconfig.PortMapping {
	return cloneNonEmptySlice(values)
}

// cloneNonEmptySlice copies slices while preserving nil for empty config fields.
func cloneNonEmptySlice[T any](values []T) []T {
	if len(values) == 0 {
		return nil
	}
	out := make([]T, len(values))
	copy(out, values)
	return out
}

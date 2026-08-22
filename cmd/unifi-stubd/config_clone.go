package main

import (
	appconfig "github.com/konstruktor1/unifi-stubd/internal/config"
)

// cloneWANHealth detaches active probe targets from loaded config.
func cloneWANHealth(value appconfig.WANHealthConfig) appconfig.WANHealthConfig {
	value.Targets = cloneNonEmptySlice(value.Targets)
	return value
}

// cloneBridgeObserve detaches nested bridge-observe slices from loaded config.
func cloneBridgeObserve(value appconfig.BridgeObserve) appconfig.BridgeObserve {
	value.IgnoredMembers = cloneNonEmptySlice(value.IgnoredMembers)
	value.MemberPortMap = cloneNonEmptySlice(value.MemberPortMap)
	return value
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

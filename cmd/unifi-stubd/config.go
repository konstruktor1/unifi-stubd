// Runtime configuration is layered as defaults, optional YAML, then explicit
// CLI flags. YAML values are copied only for flags the operator did not set,
// preserving the command-line override contract.
package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	appconfig "github.com/konstruktor1/unifi-stubd/internal/config"
	"github.com/konstruktor1/unifi-stubd/internal/device"
)

// loadConfig treats the default config path as optional, but reports missing
// files when the operator explicitly supplied a path.
func loadConfig(path string, explicit bool) (appconfig.Config, error) {
	if strings.TrimSpace(path) == "" {
		return appconfig.Default(), nil
	}
	cfg, err := appconfig.Load(path)
	if err == nil {
		log.Printf("loaded config from %s", path)
		return cfg, nil
	}
	if errors.Is(err, os.ErrNotExist) && !explicit {
		return appconfig.Default(), nil
	}
	return appconfig.Config{}, fmt.Errorf("load config %s: %w", path, err)
}

// applyConfig copies YAML settings into runtime flags only where the equivalent
// CLI flag was not explicitly provided.
func applyConfig(cfg appconfig.Config, changed map[string]bool, flags *runtimeFlags) {
	for _, setting := range runtimeSettings {
		if !changed[setting.flagName] {
			setting.apply(cfg, flags)
		}
	}
	if !changed["bridge-member-port"] {
		flags.bridgeObserve.MemberPortMap = cloneBridgeMemberPortMaps(cfg.BridgeObserve.MemberPortMap)
	}
	if !changed["bridge-ignore-member"] {
		flags.bridgeObserve.IgnoredMembers = cloneStrings(cfg.BridgeObserve.IgnoredMembers)
	}
	if !changed["port-map"] {
		flags.portMappings = clonePortMappings(cfg.PortMappings)
	}
	flags.uplinkNeighbor = configUplinkNeighbor(cfg.UplinkNeighbor)
	flags.portNeighbors = configPortNeighbors(cfg.PortNeighbors)
	flags.portOverrides = configPortOverrides(cfg.PortOverrides)
	flags.discoveryTargets = cloneStrings(cfg.DiscoveryTargets)
}

// applyConfigInterval adapts YAML seconds into the runtime duration used by the
// heartbeat loop.
func applyConfigInterval(cfg appconfig.Config, flags *runtimeFlags) {
	if cfg.IntervalSeconds > 0 {
		flags.interval = time.Duration(cfg.IntervalSeconds) * time.Second
	}
}

// configUplinkNeighbor converts YAML neighbor metadata into the payload-facing
// MAC-table entry used on the represented uplink port.
func configUplinkNeighbor(neighbor *appconfig.UplinkNeighbor) *device.MacTableEntry {
	if neighbor == nil || strings.TrimSpace(neighbor.MAC) == "" {
		return nil
	}
	return &device.MacTableEntry{
		MAC:      strings.TrimSpace(neighbor.MAC),
		Hostname: defaultNeighborHostname(neighbor.Hostname, neighbor.Name),
		IP:       strings.TrimSpace(neighbor.IP),
		Age:      defaultNeighborAge(neighbor.Age),
		Uptime:   defaultNeighborUptime(neighbor.Uptime),
		VLAN:     neighbor.VLAN,
		Static:   neighbor.Static,
		Type:     defaultNeighborType(neighbor.Type),
	}
}

// configPortNeighbors converts YAML per-port neighbor metadata into payload
// MAC-table entries without requiring live observation.
func configPortNeighbors(neighbors []appconfig.PortNeighbor) []device.PortNeighbor {
	out := make([]device.PortNeighbor, 0, len(neighbors))
	for _, neighbor := range neighbors {
		if strings.TrimSpace(neighbor.MAC) == "" {
			continue
		}
		out = append(out, device.PortNeighbor{
			Port: neighbor.Port,
			Entry: device.MacTableEntry{
				MAC:      strings.TrimSpace(neighbor.MAC),
				Hostname: defaultNeighborHostname(neighbor.Hostname, neighbor.Name),
				IP:       strings.TrimSpace(neighbor.IP),
				Age:      defaultNeighborAge(neighbor.Age),
				Uptime:   defaultNeighborUptime(neighbor.Uptime),
				VLAN:     neighbor.VLAN,
				Static:   neighbor.Static,
				Type:     defaultPortNeighborType(neighbor.Type),
			},
		})
	}
	return out
}

// configPortOverrides detaches YAML overrides before later platform enrichment
// and payload merging can mutate runtime copies.
func configPortOverrides(overrides []appconfig.PortOverride) []device.PortOverride {
	return device.ClonePortOverrides(overrides)
}

// defaultNeighborAge supplies a fresh-looking MAC-table age for synthetic
// configured neighbors.
func defaultNeighborAge(age int) int {
	if age == 0 {
		return 4
	}
	return age
}

// defaultNeighborUptime supplies a stable non-zero uptime for configured
// neighbors shown in controller topology views.
func defaultNeighborUptime(uptime int) int {
	if uptime == 0 {
		return 1200
	}
	return uptime
}

// defaultNeighborType marks the configured uplink neighbor as switch-like when
// the operator did not provide a more specific topology type.
func defaultNeighborType(neighborType string) string {
	neighborType = strings.TrimSpace(neighborType)
	if neighborType == "" {
		return "usw"
	}
	return neighborType
}

// defaultPortNeighborType marks per-port configured neighbors as clients unless
// the operator supplied a specific topology type.
func defaultPortNeighborType(neighborType string) string {
	neighborType = strings.TrimSpace(neighborType)
	if neighborType == "" {
		return "client"
	}
	return neighborType
}

// defaultNeighborHostname accepts both the new hostname field and the legacy
// name alias used by older config examples.
func defaultNeighborHostname(hostname, name string) string {
	if hostname = strings.TrimSpace(hostname); hostname != "" {
		return hostname
	}
	return strings.TrimSpace(name)
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

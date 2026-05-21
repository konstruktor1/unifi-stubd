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

func applyConfigInterval(cfg appconfig.Config, flags *runtimeFlags) {
	if cfg.IntervalSeconds > 0 {
		flags.interval = time.Duration(cfg.IntervalSeconds) * time.Second
	}
}

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

func configPortOverrides(overrides []appconfig.PortOverride) []device.PortOverride {
	return device.ClonePortOverrides(overrides)
}

func defaultNeighborAge(age int) int {
	if age == 0 {
		return 4
	}
	return age
}

func defaultNeighborUptime(uptime int) int {
	if uptime == 0 {
		return 1200
	}
	return uptime
}

func defaultNeighborType(neighborType string) string {
	neighborType = strings.TrimSpace(neighborType)
	if neighborType == "" {
		return "usw"
	}
	return neighborType
}

func defaultPortNeighborType(neighborType string) string {
	neighborType = strings.TrimSpace(neighborType)
	if neighborType == "" {
		return "client"
	}
	return neighborType
}

func defaultNeighborHostname(hostname, name string) string {
	if hostname = strings.TrimSpace(hostname); hostname != "" {
		return hostname
	}
	return strings.TrimSpace(name)
}

func cloneStrings(values []string) []string {
	return cloneNonEmptySlice(values)
}

func cloneBridgeObserve(value appconfig.BridgeObserve) appconfig.BridgeObserve {
	value.IgnoredMembers = cloneStrings(value.IgnoredMembers)
	value.MemberPortMap = cloneBridgeMemberPortMaps(value.MemberPortMap)
	return value
}

func cloneBridgeMemberPortMaps(values []appconfig.BridgeMemberPortMap) []appconfig.BridgeMemberPortMap {
	return cloneNonEmptySlice(values)
}

func clonePortMappings(values []appconfig.PortMapping) []appconfig.PortMapping {
	return cloneNonEmptySlice(values)
}

func cloneNonEmptySlice[T any](values []T) []T {
	if len(values) == 0 {
		return nil
	}
	out := make([]T, len(values))
	copy(out, values)
	return out
}

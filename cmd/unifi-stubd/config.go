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
	flags.wanHealth = cloneWANHealth(cfg.WANHealth)
	flags.discoveryTargets = cloneStrings(cfg.DiscoveryTargets)
}

// applyConfigInterval adapts YAML seconds into the runtime duration used by the
// heartbeat loop.
func applyConfigInterval(cfg appconfig.Config, flags *runtimeFlags) {
	if cfg.IntervalSeconds > 0 {
		flags.interval = time.Duration(cfg.IntervalSeconds) * time.Second
	}
}

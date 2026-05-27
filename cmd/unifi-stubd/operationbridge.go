package main

import (
	"fmt"
	"strings"

	appconfig "github.com/konstruktor1/unifi-stubd/internal/config"
)

// effectiveBridgeObserve merges structured bridge_observe config with the older
// observe_bridge/observe_interface flags.
func effectiveBridgeObserve(flags runtimeFlags) appconfig.BridgeObserve {
	cfg := cloneBridgeObserve(flags.bridgeObserve)
	if strings.TrimSpace(cfg.Bridge) == "" {
		cfg.Bridge = strings.TrimSpace(flags.observeBridge)
	}
	if strings.TrimSpace(cfg.UplinkInterface) == "" {
		cfg.UplinkInterface = strings.TrimSpace(flags.observeInterface)
	}
	return cfg
}

// bridgeMemberPortMap converts operator pinning into the observe package shape
// used for deterministic bridge-member assignment.
func bridgeMemberPortMap(mappings []appconfig.BridgeMemberPortMap) map[string]int {
	if len(mappings) == 0 {
		return nil
	}
	out := make(map[string]int, len(mappings))
	for _, mapping := range mappings {
		member := strings.TrimSpace(mapping.Member)
		if member != "" {
			out[member] = mapping.Port
		}
	}
	return out
}

// validateBridgeMemberPortMap rejects ambiguous bridge-member pinning before
// passive observation can assign represented UniFi ports.
func validateBridgeMemberPortMap(mappings []appconfig.BridgeMemberPortMap) []error {
	var errs []error
	seenMembers := map[string]bool{}
	seenPorts := map[int]string{}
	for _, mapping := range mappings {
		member := strings.TrimSpace(mapping.Member)
		if member == "" {
			continue
		}
		if seenMembers[member] {
			errs = append(errs, fmt.Errorf("duplicate bridge_observe.member_port_map member %q", member))
		}
		seenMembers[member] = true
		if previous := seenPorts[mapping.Port]; previous != "" {
			errs = append(errs, fmt.Errorf("duplicate bridge_observe.member_port_map port %d for %q and %q", mapping.Port, previous, member))
		}
		seenPorts[mapping.Port] = member
	}
	return errs
}

// validateBridgeIgnoredMembers catches configuration that both pins and ignores
// the same bridge member, because ignored members must never consume ports.
func validateBridgeIgnoredMembers(cfg appconfig.BridgeObserve) []error {
	var errs []error
	seen := map[string]bool{}
	pinned := map[string]bool{}
	for _, mapping := range cfg.MemberPortMap {
		if key := bridgeMemberKey(mapping.Member); key != "" {
			pinned[key] = true
		}
	}
	bridgeKey := bridgeMemberKey(cfg.Bridge)
	uplinkKey := bridgeMemberKey(cfg.UplinkInterface)
	for _, member := range cfg.IgnoredMembers {
		key := bridgeMemberKey(member)
		if key == "" {
			continue
		}
		switch {
		case seen[key]:
			errs = append(errs, fmt.Errorf("duplicate bridge_observe.ignored_members member %q", strings.TrimSpace(member)))
		case key == bridgeKey:
			errs = append(errs, fmt.Errorf("bridge_observe.ignored_members cannot ignore bridge %q", strings.TrimSpace(member)))
		case key == uplinkKey:
			errs = append(errs, fmt.Errorf("bridge_observe.ignored_members cannot ignore uplink_interface %q", strings.TrimSpace(member)))
		case pinned[key]:
			errs = append(errs, fmt.Errorf("bridge_observe.ignored_members member %q also appears in member_port_map", strings.TrimSpace(member)))
		}
		seen[key] = true
	}
	return errs
}

// bridgeMemberKey normalizes bridge member names for duplicate and ignore-list
// policy checks.
func bridgeMemberKey(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

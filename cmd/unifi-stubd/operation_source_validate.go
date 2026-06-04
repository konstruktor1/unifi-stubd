package main

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

// validateSourceMappings validates bridge-observe and port-map inputs in two
// phases: structural checks first, optional live interface existence checks when
// requested by -validate or runtime startup.
func validateSourceMappings(flags runtimeFlags, live bool) error {
	var errs []error
	mode := normalizeMode(flags.operationMode)
	switch mode {
	case operationModeBridgeObserve:
		errs = append(errs, validateObserveSources(flags, live)...)
	case operationModePortMap:
		errs = append(errs, validatePortMapSources(flags, live)...)
	}
	if err := errors.Join(errs...); err != nil {
		return fmt.Errorf("validate source mappings: %w", err)
	}
	return nil
}

func validateObserveSources(flags runtimeFlags, live bool) []error {
	var errs []error
	cfg := effectiveBridgeObserve(flags)
	if err := validateInterfaceName("bridge_observe.bridge", cfg.Bridge, live); err != nil {
		errs = append(errs, err)
	}
	if err := validateInterfaceName("bridge_observe.uplink_interface", cfg.UplinkInterface, live); err != nil {
		errs = append(errs, err)
	}
	for _, member := range cfg.IgnoredMembers {
		if err := validateInterfaceName("bridge_observe.ignored_members", member, live); err != nil {
			errs = append(errs, err)
		}
	}
	for _, mapping := range cfg.MemberPortMap {
		member := strings.TrimSpace(mapping.Member)
		if member == "" || strings.Contains(member, "/") {
			errs = append(errs, fmt.Errorf("invalid bridge_observe.member_port_map member %q", mapping.Member))
		}
		if mapping.Port < 1 || mapping.Port > flags.portCount {
			errs = append(errs, fmt.Errorf("invalid bridge_observe.member_port_map port %d for %q; use 1..%d", mapping.Port, member, flags.portCount))
		}
	}
	errs = append(errs, validateBridgeMemberPortMap(cfg.MemberPortMap)...)
	errs = append(errs, validateIgnoredMembers(cfg)...)
	return errs
}

func validatePortMapSources(flags runtimeFlags, live bool) []error {
	var errs []error
	seenPorts := map[int]bool{}
	for _, mapping := range flags.portMappings {
		if mapping.Port < 1 || mapping.Port > flags.portCount {
			errs = append(errs, fmt.Errorf("invalid port_mappings port %d; use 1..%d", mapping.Port, flags.portCount))
			continue
		}
		if seenPorts[mapping.Port] {
			errs = append(errs, fmt.Errorf("duplicate port_mappings entry for port %d", mapping.Port))
		}
		seenPorts[mapping.Port] = true
		sources := 0
		if strings.TrimSpace(mapping.Interface) != "" {
			sources++
			if err := validateInterfaceName("port_mappings.interface", mapping.Interface, live); err != nil {
				errs = append(errs, fmt.Errorf("port %d: %w", mapping.Port, err))
			}
		}
		if mapping.Disabled {
			sources++
		}
		if mapping.Unmapped {
			sources++
		}
		if sources != 1 {
			errs = append(errs, fmt.Errorf("invalid port_mappings entry on port %d; set exactly one of interface, disabled, or unmapped", mapping.Port))
		}
	}
	for port := 1; port <= flags.portCount; port++ {
		if !seenPorts[port] {
			errs = append(errs, fmt.Errorf("missing port_mappings entry for port %d", port))
		}
	}
	return errs
}

// validateInterfaceName rejects path-like names and optionally checks
// local existence for modes that will read host interface data.
func validateInterfaceName(field, value string, live bool) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	if strings.Contains(value, "/") {
		return fmt.Errorf("invalid %s %q", field, value)
	}
	if live {
		if _, err := net.InterfaceByName(value); err != nil {
			return fmt.Errorf("%s %q not found: %w", field, value, err)
		}
	}
	return nil
}

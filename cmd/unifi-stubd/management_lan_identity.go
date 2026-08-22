package main

import (
	"context"
	"fmt"
	"net"
	"strings"

	appconfig "github.com/konstruktor1/unifi-stubd/internal/config"
	"github.com/konstruktor1/unifi-stubd/internal/platform"
)

// managementLANSourceIP opts into source binding only for a preexisting
// management interface; metadata-only and planned modes never bind traffic.
func managementLANSourceIP(flags runtimeFlags, ip net.IP) net.IP {
	cfg := effectiveManagementLAN(flags)
	if !cfg.Enabled || cfg.Mode != managementLANModePreexistingInterface {
		return nil
	}
	return ip.To4()
}

// informSourceIP chooses a safe local source for inform traffic, preferring the
// management-LAN address and otherwise using the identity IP only when it is
// already assigned on the host.
func informSourceIP(flags runtimeFlags, ip net.IP, addresses platform.AddressReader) net.IP {
	if source := managementLANSourceIP(flags, ip); source != nil {
		return source
	}
	candidate := ip.To4()
	if candidate == nil || !hostHasIPv4(candidate, addresses) {
		return nil
	}
	return candidate
}

// effectiveDiscoveryInterface binds discovery to the operator-selected
// interface, or to a preexisting management-LAN interface when configured.
func effectiveDiscoveryInterface(flags runtimeFlags) string {
	if iface := strings.TrimSpace(flags.discoveryInterface); iface != "" {
		return iface
	}
	cfg := effectiveManagementLAN(flags)
	if cfg.Enabled && cfg.Mode == managementLANModePreexistingInterface {
		return cfg.Interface
	}
	return ""
}

// resolveManagementIP replaces the fake default when management-LAN mode binds
// discovery and inform traffic to a preexisting interface.
func resolveManagementIP(flags runtimeFlags, fallback net.IP, addresses platform.AddressReader) (net.IP, error) {
	cfg := effectiveManagementLAN(flags)
	if !cfg.Enabled || cfg.Mode != managementLANModePreexistingInterface {
		return fallback, nil
	}
	if cfg.IP != "" {
		ip := net.ParseIP(cfg.IP).To4()
		if ip == nil {
			return nil, fmt.Errorf("invalid management_lan.ip %q", cfg.IP)
		}
		return ip, nil
	}
	ip, err := managementLANInterfaceIP(cfg, addresses)
	if err != nil {
		return nil, err
	}
	return ip, nil
}

// managementLANInterfaceIP returns the first IPv4 address on the selected
// management interface for validation-time source binding checks.
func managementLANInterfaceIP(cfg appconfig.ManagementLAN, addresses platform.AddressReader) (net.IP, error) {
	if addresses == nil {
		return nil, fmt.Errorf("management_lan.interface %s cannot be checked without platform", cfg.Interface)
	}
	identity, err := addresses.InterfaceIdentity(context.Background(), cfg.Interface)
	if err != nil {
		return nil, fmt.Errorf("management_lan.interface %s not found: %w", cfg.Interface, err)
	}
	if len(identity.IPv4) > 0 {
		return identity.IPv4[0], nil
	}
	return nil, fmt.Errorf("management_lan.interface %s has no IPv4 address", cfg.Interface)
}

// interfaceHasIPv4 checks whether a configured management IP is actually
// assigned to the selected local interface.
func interfaceHasIPv4(ifaceName string, ip net.IP, addresses platform.AddressReader) bool {
	if addresses == nil {
		return false
	}
	assigned, err := addresses.IPv4Assigned(context.Background(), ifaceName, ip)
	return err == nil && assigned
}

// hostHasIPv4 prevents inform source binding to synthetic or unassigned
// identity addresses.
func hostHasIPv4(ip net.IP, addresses platform.AddressReader) bool {
	if addresses == nil {
		return false
	}
	assigned, err := addresses.IPv4Assigned(context.Background(), "", ip)
	return err == nil && assigned
}

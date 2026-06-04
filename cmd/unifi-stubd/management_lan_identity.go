package main

import (
	"context"
	"fmt"
	"log"
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
func informSourceIP(flags runtimeFlags, ip net.IP) net.IP {
	if source := managementLANSourceIP(flags, ip); source != nil {
		return source
	}
	candidate := ip.To4()
	if candidate == nil || !hostHasIPv4(candidate) {
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
func resolveManagementIP(flags runtimeFlags, fallback net.IP, plt platform.Platform) (net.IP, error) {
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
	if plt == nil {
		plt = runtimePlatform(flags)
	}
	ctx, cancel := context.WithTimeout(context.Background(), observeTimeout)
	defer cancel()
	observation, errs := plt.Interface(ctx, cfg.Interface)
	for _, err := range errs {
		log.Printf("management LAN interface %s warning: %v", cfg.Interface, err)
	}
	ip := net.ParseIP(observation.IP).To4()
	if ip == nil {
		return nil, fmt.Errorf("management_lan.interface %s has no IPv4 address; set management_lan.ip or -ip", cfg.Interface)
	}
	return ip, nil
}

// managementLANInterfaceIP returns the first IPv4 address on the selected
// management interface for validation-time source binding checks.
func managementLANInterfaceIP(cfg appconfig.ManagementLAN) (net.IP, error) {
	iface, err := net.InterfaceByName(cfg.Interface)
	if err != nil {
		return nil, fmt.Errorf("management_lan.interface %s not found: %w", cfg.Interface, err)
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return nil, fmt.Errorf("read management_lan.interface %s addresses: %w", cfg.Interface, err)
	}
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}
		if ip := ipNet.IP.To4(); ip != nil {
			return ip, nil
		}
	}
	return nil, fmt.Errorf("management_lan.interface %s has no IPv4 address", cfg.Interface)
}

// interfaceHasIPv4 checks whether a configured management IP is actually
// assigned to the selected local interface.
func interfaceHasIPv4(ifaceName string, ip net.IP) bool {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return false
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return false
	}
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}
		if current := ipNet.IP.To4(); current != nil && current.Equal(ip) {
			return true
		}
	}
	return false
}

// hostHasIPv4 prevents inform source binding to synthetic or unassigned
// identity addresses.
func hostHasIPv4(ip net.IP) bool {
	ifaces, err := net.Interfaces()
	if err != nil {
		return false
	}
	for _, iface := range ifaces {
		if interfaceHasIPv4(iface.Name, ip) {
			return true
		}
	}
	return false
}

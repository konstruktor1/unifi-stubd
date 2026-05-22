// Management LAN handling keeps switch management VLAN behavior explicit. The
// daemon may report metadata or bind to a preexisting VLAN interface, but it
// never creates host VLAN devices or applies controller provisioning locally.
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/url"
	"strings"
	"time"

	appconfig "github.com/konstruktor1/unifi-stubd/internal/config"
	"github.com/konstruktor1/unifi-stubd/internal/device"
	"github.com/konstruktor1/unifi-stubd/internal/platform"
)

// Management-LAN modes and policies are explicit so the daemon can report VLAN
// intent without creating host VLAN interfaces.
const (
	managementLANModeMetadataOnly         = "metadata-only"
	managementLANModePreexistingInterface = "preexisting-interface"
	managementLANModePlannedHostVLAN      = "planned-host-vlan"

	managementLANReachOff      = "off"
	managementLANReachWarn     = "warn"
	managementLANReachRequired = "required"

	managementLANAdoptUntaggedFirst = "untagged-first"
	managementLANAdoptTaggedOnly    = "tagged-only"
)

// effectiveManagementLAN normalizes the structured management-LAN config and
// marks it enabled whenever any meaningful field asks for management metadata.
func effectiveManagementLAN(flags runtimeFlags) appconfig.ManagementLAN {
	cfg := flags.managementLAN
	cfg.Mode = strings.ToLower(strings.TrimSpace(cfg.Mode))
	cfg.Interface = strings.TrimSpace(cfg.Interface)
	cfg.IP = strings.TrimSpace(cfg.IP)
	cfg.NetworkName = strings.TrimSpace(cfg.NetworkName)
	cfg.ControllerReachable = strings.ToLower(strings.TrimSpace(cfg.ControllerReachable))
	cfg.AdoptionStrategy = strings.ToLower(strings.TrimSpace(cfg.AdoptionStrategy))
	if cfg.Mode == "" {
		cfg.Mode = managementLANModeMetadataOnly
	}
	if cfg.ControllerReachable == "" {
		cfg.ControllerReachable = managementLANReachOff
	}
	if cfg.AdoptionStrategy == "" {
		cfg.AdoptionStrategy = managementLANAdoptUntaggedFirst
	}
	if cfg.VLAN > 0 || cfg.Mode != managementLANModeMetadataOnly || cfg.Interface != "" || cfg.IP != "" || cfg.NetworkName != "" {
		cfg.Enabled = true
	}
	return cfg
}

// structuredManagementLANRequested detects whether the operator used any
// management-LAN field, even when effective normalization would otherwise hide
// an implicit default.
func structuredManagementLANRequested(flags runtimeFlags) bool {
	cfg := flags.managementLAN
	return cfg.Enabled ||
		cfg.VLAN != 0 ||
		(strings.TrimSpace(cfg.Mode) != "" && !strings.EqualFold(strings.TrimSpace(cfg.Mode), managementLANModeMetadataOnly)) ||
		strings.TrimSpace(cfg.Interface) != "" ||
		strings.TrimSpace(cfg.IP) != "" ||
		strings.TrimSpace(cfg.NetworkName) != "" ||
		(strings.TrimSpace(cfg.ControllerReachable) != "" && !strings.EqualFold(strings.TrimSpace(cfg.ControllerReachable), managementLANReachOff)) ||
		(strings.TrimSpace(cfg.AdoptionStrategy) != "" && !strings.EqualFold(strings.TrimSpace(cfg.AdoptionStrategy), managementLANAdoptUntaggedFirst))
}

// effectiveManagementVLAN exposes the normalized management VLAN to payload
// identity construction.
func effectiveManagementVLAN(flags runtimeFlags) int {
	return effectiveManagementLAN(flags).VLAN
}

// statusManagementLAN returns management-LAN metadata only when the feature is
// actually active, keeping status output quiet for default stub mode.
func statusManagementLAN(flags runtimeFlags) *appconfig.ManagementLAN {
	cfg := effectiveManagementLAN(flags)
	if !cfg.Enabled {
		return nil
	}
	return &cfg
}

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

// validateManagementLAN enforces that management VLAN handling remains payload
// metadata or binding to a preexisting interface; planned host VLAN creation is
// dry-run-only.
func validateManagementLAN(flags runtimeFlags, profile device.Profile, live bool) error {
	cfg := effectiveManagementLAN(flags)
	if cfg.VLAN < 0 || cfg.VLAN > 4094 {
		return fmt.Errorf("invalid management_lan.vlan %d; use 0..4094", cfg.VLAN)
	}
	if !cfg.Enabled {
		return nil
	}
	if structuredManagementLANRequested(flags) && strings.ToLower(strings.TrimSpace(profile.Payload.Kind)) != "switch" {
		return fmt.Errorf("management_lan is supported for switch profiles only in this release")
	}
	switch cfg.Mode {
	case managementLANModeMetadataOnly, managementLANModePreexistingInterface, managementLANModePlannedHostVLAN:
	default:
		return fmt.Errorf("invalid management_lan.mode %q; use metadata-only, preexisting-interface, or planned-host-vlan", cfg.Mode)
	}
	switch cfg.ControllerReachable {
	case managementLANReachOff, managementLANReachWarn, managementLANReachRequired:
	default:
		return fmt.Errorf("invalid management_lan.controller_reachable %q; use off, warn, or required", cfg.ControllerReachable)
	}
	switch cfg.AdoptionStrategy {
	case managementLANAdoptUntaggedFirst, managementLANAdoptTaggedOnly:
	default:
		return fmt.Errorf("invalid management_lan.adoption_strategy %q; use untagged-first or tagged-only", cfg.AdoptionStrategy)
	}
	if cfg.Mode == managementLANModePlannedHostVLAN && !flags.dryRunPlan {
		return fmt.Errorf("management_lan.mode planned-host-vlan is dry-run-plan only")
	}
	if cfg.Mode == managementLANModeMetadataOnly {
		return nil
	}
	if cfg.Interface == "" {
		return fmt.Errorf("management_lan.interface is required for mode %s", cfg.Mode)
	}
	if strings.Contains(cfg.Interface, "/") {
		return fmt.Errorf("invalid management_lan.interface %q", cfg.Interface)
	}
	if cfg.Mode == managementLANModePreexistingInterface && live {
		return validatePreexistingManagementLAN(flags, cfg)
	}
	return nil
}

// validatePreexistingManagementLAN checks the local interface and optional
// controller reachability before the daemon binds discovery or inform traffic
// to that management address.
func validatePreexistingManagementLAN(flags runtimeFlags, cfg appconfig.ManagementLAN) error {
	sourceIP, err := managementLANInterfaceIP(cfg)
	if err != nil {
		return err
	}
	if cfg.IP != "" {
		configured := net.ParseIP(cfg.IP).To4()
		if configured == nil {
			return fmt.Errorf("invalid management_lan.ip %q", cfg.IP)
		}
		if !configured.Equal(sourceIP) && !interfaceHasIPv4(cfg.Interface, configured) {
			return fmt.Errorf("management_lan.ip %s is not assigned to interface %s", configured, cfg.Interface)
		}
		sourceIP = configured
	}
	if err := validateManagementLANReachability(flags, cfg, sourceIP); err != nil {
		if cfg.ControllerReachable == managementLANReachRequired {
			return err
		}
		log.Printf("management LAN reachability warning: %v", err)
	}
	return nil
}

// resolveManagementLANIdentityIP replaces the default fake management IP with
// the selected preexisting interface address when management-LAN mode asks for
// source-bound discovery and inform traffic.
func resolveManagementLANIdentityIP(flags runtimeFlags, fallback net.IP, plt platform.Platform) (net.IP, error) {
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

// validateManagementLANReachability optionally proves that the chosen local
// source address can reach the controller, depending on warn/required policy.
func validateManagementLANReachability(flags runtimeFlags, cfg appconfig.ManagementLAN, sourceIP net.IP) error {
	if cfg.ControllerReachable == managementLANReachOff {
		return nil
	}
	hostPort, err := controllerHostPort(flags.controller)
	if err != nil {
		return err
	}
	dialer := net.Dialer{
		Timeout:   2 * time.Second,
		LocalAddr: &net.TCPAddr{IP: sourceIP},
	}
	conn, err := dialer.Dial("tcp4", hostPort)
	if err != nil {
		return fmt.Errorf("management LAN cannot reach controller %s from %s: %w", hostPort, sourceIP, err)
	}
	_ = conn.Close()
	return nil
}

// controllerHostPort turns the configured controller URL into the TCP endpoint
// used for management-LAN reachability checks.
func controllerHostPort(rawURL string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed.Hostname() == "" {
		return "", fmt.Errorf("management_lan.controller_reachable requires controller_url with host")
	}
	port := parsed.Port()
	if port == "" {
		switch parsed.Scheme {
		case "https":
			port = "443"
		default:
			port = "80"
		}
	}
	return net.JoinHostPort(parsed.Hostname(), port), nil
}

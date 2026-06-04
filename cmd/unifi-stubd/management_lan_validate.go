package main

import (
	"fmt"
	"log"
	"net"
	"strings"

	appconfig "github.com/konstruktor1/unifi-stubd/internal/config"
	"github.com/konstruktor1/unifi-stubd/internal/device"
)

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
	if managementRequested(flags) && strings.ToLower(strings.TrimSpace(profile.Payload.Kind)) != "switch" {
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
		return validatePreexistingManagement(flags, cfg)
	}
	return nil
}

// validatePreexistingManagement checks the local interface before source-bound
// discovery or inform traffic can use it.
func validatePreexistingManagement(flags runtimeFlags, cfg appconfig.ManagementLAN) error {
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
	if err := checkControllerReachability(flags, cfg, sourceIP); err != nil {
		if cfg.ControllerReachable == managementLANReachRequired {
			return err
		}
		log.Printf("management LAN reachability warning: %v", err)
	}
	return nil
}

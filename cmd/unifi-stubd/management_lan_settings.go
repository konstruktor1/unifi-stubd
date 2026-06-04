package main

import (
	"strings"

	appconfig "github.com/konstruktor1/unifi-stubd/internal/config"
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

func managementRequested(flags runtimeFlags) bool {
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

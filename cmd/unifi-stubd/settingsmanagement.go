package main

import appconfig "github.com/konstruktor1/unifi-stubd/internal/config"

func managementRuntimeSettings() []runtimeSetting {
	return []runtimeSetting{
		boolSetting("management-lan-enabled", "enable structured switch management LAN handling",
			func(flags *runtimeFlags) *bool { return &flags.managementLAN.Enabled },
			func(cfg *appconfig.Config) *bool { return &cfg.ManagementLAN.Enabled },
		),
		intSetting("management-lan-vlan", "structured management LAN VLAN ID; 0 leaves it unset",
			func(flags *runtimeFlags) *int { return &flags.managementLAN.VLAN },
			func(cfg *appconfig.Config) *int { return &cfg.ManagementLAN.VLAN },
		),
		stringSetting("management-lan-mode", "management LAN mode: metadata-only, preexisting-interface, or planned-host-vlan",
			func(flags *runtimeFlags) *string { return &flags.managementLAN.Mode },
			func(cfg *appconfig.Config) *string { return &cfg.ManagementLAN.Mode },
		),
		stringSetting("management-lan-interface", "existing VLAN interface used by -management-lan-mode preexisting-interface",
			func(flags *runtimeFlags) *string { return &flags.managementLAN.Interface },
			func(cfg *appconfig.Config) *string { return &cfg.ManagementLAN.Interface },
		),
		stringSetting("management-lan-ip", "optional management IP for -management-lan-interface",
			func(flags *runtimeFlags) *string { return &flags.managementLAN.IP },
			func(cfg *appconfig.Config) *string { return &cfg.ManagementLAN.IP },
		),
		stringSetting("management-lan-network", "optional controller-facing management network label",
			func(flags *runtimeFlags) *string { return &flags.managementLAN.NetworkName },
			func(cfg *appconfig.Config) *string { return &cfg.ManagementLAN.NetworkName },
		),
		stringSetting("management-lan-controller-reachable", "management LAN controller reachability validation: off, warn, or required",
			func(flags *runtimeFlags) *string { return &flags.managementLAN.ControllerReachable },
			func(cfg *appconfig.Config) *string { return &cfg.ManagementLAN.ControllerReachable },
		),
		stringSetting("management-lan-adoption-strategy", "management LAN adoption strategy: untagged-first or tagged-only",
			func(flags *runtimeFlags) *string { return &flags.managementLAN.AdoptionStrategy },
			func(cfg *appconfig.Config) *string { return &cfg.ManagementLAN.AdoptionStrategy },
		),
	}
}

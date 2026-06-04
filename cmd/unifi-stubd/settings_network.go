package main

import appconfig "github.com/konstruktor1/unifi-stubd/internal/config"

func networkRuntimeSettings() []runtimeSetting {
	return []runtimeSetting{
		stringSetting("controller", "optional UniFi inform URL, for example http://192.168.1.10:8080/inform",
			func(flags *runtimeFlags) *string { return &flags.controller },
			func(cfg *appconfig.Config) *string { return &cfg.ControllerURL },
		),
		intervalSetting(),
		boolSetting("no-discovery", "skip UDP discovery and only send inform when -controller is set",
			func(flags *runtimeFlags) *bool { return &flags.noDiscovery },
			func(cfg *appconfig.Config) *bool { return &cfg.NoDiscovery },
		),
		stringSetting("discovery-interface", "optional local interface name used for UDP discovery sends",
			func(flags *runtimeFlags) *string { return &flags.discoveryInterface },
			func(cfg *appconfig.Config) *string { return &cfg.DiscoveryInterface },
		),
	}
}

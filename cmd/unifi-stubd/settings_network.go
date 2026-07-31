package main

import (
	"time"

	appconfig "github.com/konstruktor1/unifi-stubd/internal/config"
)

func networkRuntimeSettings() []runtimeSetting {
	return []runtimeSetting{
		stringSetting("controller", "optional UniFi inform URL, for example http://192.168.1.10:8080/inform",
			func(flags *runtimeFlags) *string { return &flags.controller },
			func(cfg *appconfig.Config) *string { return &cfg.ControllerURL },
		),
		intervalSetting(),
		secondsDurationSetting("controller-interval-min", "minimum accepted controller heartbeat interval",
			func(flags *runtimeFlags) *time.Duration { return &flags.controllerIntervalMin },
			func(cfg *appconfig.Config) *int { return &cfg.ControllerIntervalMinSeconds },
		),
		secondsDurationSetting("controller-interval-max", "maximum accepted controller heartbeat interval",
			func(flags *runtimeFlags) *time.Duration { return &flags.controllerIntervalMax },
			func(cfg *appconfig.Config) *int { return &cfg.ControllerIntervalMaxSeconds },
		),
		secondsDurationSetting("inform-backoff-max", "maximum failed-heartbeat retry delay",
			func(flags *runtimeFlags) *time.Duration { return &flags.informBackoffMax },
			func(cfg *appconfig.Config) *int { return &cfg.InformBackoffMaxSeconds },
		),
		intSetting("inform-backoff-jitter-percent", "inform retry jitter percentage",
			func(flags *runtimeFlags) *int { return &flags.informBackoffJitterPercent },
			func(cfg *appconfig.Config) *int { return &cfg.InformBackoffJitterPercent },
		),
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

package main

import appconfig "github.com/konstruktor1/unifi-stubd/internal/config"

func observationRuntimeSettings() []runtimeSetting {
	return []runtimeSetting{
		stringSetting("observe-interface", "host interface used for passive link counters and speed",
			func(flags *runtimeFlags) *string { return &flags.observeInterface },
			func(cfg *appconfig.Config) *string { return &cfg.ObserveInterface },
		),
		stringSetting("observe-bridge", "Linux bridge used for passive FDB MAC table data",
			func(flags *runtimeFlags) *string { return &flags.observeBridge },
			func(cfg *appconfig.Config) *string { return &cfg.ObserveBridge },
		),
		stringSetting("bridge-observe-bridge", "host bridge represented as virtual UniFi switch ports",
			func(flags *runtimeFlags) *string { return &flags.bridgeObserve.Bridge },
			func(cfg *appconfig.Config) *string { return &cfg.BridgeObserve.Bridge },
		),
		stringSetting("bridge-observe-uplink-interface", "bridge member that represents the upstream link",
			func(flags *runtimeFlags) *string { return &flags.bridgeObserve.UplinkInterface },
			func(cfg *appconfig.Config) *string { return &cfg.BridgeObserve.UplinkInterface },
		),
		stringSetting("lldp-source", "passive LLDP source: off or lldpd",
			func(flags *runtimeFlags) *string { return &flags.lldpSource },
			func(cfg *appconfig.Config) *string { return &cfg.LLDPSource },
		),
		boolSetting("traffic-rates-enabled", "report read-only interface traffic rates in inform payloads",
			func(flags *runtimeFlags) *bool { return &flags.trafficRatesEnabled },
			func(cfg *appconfig.Config) *bool { return &cfg.TrafficRatesEnabled },
		),
		stringSetting("proc-source", "read-only proc source: off or procfs",
			func(flags *runtimeFlags) *string { return &flags.procSource },
			func(cfg *appconfig.Config) *string { return &cfg.ProcSource },
		),
	}
}

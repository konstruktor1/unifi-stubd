// runtimeSettings is the shared registry for flag registration and YAML config
// application. Keeping both directions together prevents the CLI and packaged
// config surface from drifting apart.
package main

import (
	"flag"
	"time"

	appconfig "github.com/konstruktor1/unifi-stubd/internal/config"
)

// runtimeSetting ties one CLI flag to its matching YAML config field.
type runtimeSetting struct {
	flagName string
	register func(*runtimeFlags, appconfig.Config)
	apply    func(appconfig.Config, *runtimeFlags)
}

// runtimeSettings enumerates the runtime surface that participates in
// CLI-over-YAML precedence.
var runtimeSettings = []runtimeSetting{
	stringSetting("operation-mode", "runtime mode: stub, bridge-observe, observe, port-map, host-direct, or macvlan", func(flags *runtimeFlags) *string { return &flags.operationMode }, func(cfg appconfig.Config) string { return cfg.OperationMode }),
	stringSetting("profile", "device profile to emulate; use -list-profiles to show options", func(flags *runtimeFlags) *string { return &flags.profileName }, func(cfg appconfig.Config) string { return cfg.Profile }),
	stringSetting("profile-file", "optional external device profile YAML file", func(flags *runtimeFlags) *string { return &flags.profileFile }, func(cfg appconfig.Config) string { return cfg.ProfileFile }),
	stringSetting("profile-dir", "optional directory with external device profile YAML files", func(flags *runtimeFlags) *string { return &flags.profileDir }, func(cfg appconfig.Config) string { return cfg.ProfileDir }),
	stringSetting("mac", "fake device MAC address, or auto to derive one from hostname and profile", func(flags *runtimeFlags) *string { return &flags.macText }, func(cfg appconfig.Config) string { return cfg.MAC }),
	stringSetting("ip", "fake device IPv4 address", func(flags *runtimeFlags) *string { return &flags.ipText }, func(cfg appconfig.Config) string { return cfg.IP }),
	stringSetting("hostname", "fake device hostname, or auto to use the OS hostname", func(flags *runtimeFlags) *string { return &flags.hostname }, func(cfg appconfig.Config) string { return cfg.Hostname }),
	stringSetting("model", "override UniFi model identifier from the selected profile", func(flags *runtimeFlags) *string { return &flags.model }, func(cfg appconfig.Config) string { return cfg.Model }),
	stringSetting("model-display", "override display name from the selected profile", func(flags *runtimeFlags) *string { return &flags.modelDisplay }, func(cfg appconfig.Config) string { return cfg.ModelDisplay }),
	stringSetting("firmware-version", "override firmware version from the selected profile", func(flags *runtimeFlags) *string { return &flags.version }, func(cfg appconfig.Config) string { return cfg.Version }),
	intSetting("ports", "override number of switch ports from the selected profile", func(flags *runtimeFlags) *int { return &flags.portCount }, func(cfg appconfig.Config) int { return cfg.Ports }),
	intSetting("link-speed", "override default switch port speed in Mbps; 0 uses selected profile", func(flags *runtimeFlags) *int { return &flags.linkSpeed }, func(cfg appconfig.Config) int { return cfg.LinkSpeed }),
	stringSetting("uplink-speed", "uplink speed in Mbps, auto, or profile", func(flags *runtimeFlags) *string { return &flags.uplinkSpeed }, func(cfg appconfig.Config) string { return cfg.UplinkSpeed }),
	intSetting("uplink-port", "override uplink port index; 0 uses selected profile", func(flags *runtimeFlags) *int { return &flags.uplinkPort }, func(cfg appconfig.Config) int { return cfg.UplinkPort }),
	stringSetting("observe-interface", "host interface used for passive link counters and speed", func(flags *runtimeFlags) *string { return &flags.observeInterface }, func(cfg appconfig.Config) string { return cfg.ObserveInterface }),
	stringSetting("observe-bridge", "Linux bridge used for passive FDB MAC table data", func(flags *runtimeFlags) *string { return &flags.observeBridge }, func(cfg appconfig.Config) string { return cfg.ObserveBridge }),
	stringSetting("bridge-observe-bridge", "host bridge represented as virtual UniFi switch ports", func(flags *runtimeFlags) *string { return &flags.bridgeObserve.Bridge }, func(cfg appconfig.Config) string { return cfg.BridgeObserve.Bridge }),
	stringSetting("bridge-observe-uplink-interface", "bridge member that represents the upstream link", func(flags *runtimeFlags) *string { return &flags.bridgeObserve.UplinkInterface }, func(cfg appconfig.Config) string { return cfg.BridgeObserve.UplinkInterface }),
	stringSetting("lldp-source", "passive LLDP source: off or lldpd", func(flags *runtimeFlags) *string { return &flags.lldpSource }, func(cfg appconfig.Config) string { return cfg.LLDPSource }),
	stringSetting("traffic-source", "traffic metadata source: off", func(flags *runtimeFlags) *string { return &flags.trafficSource }, func(cfg appconfig.Config) string { return cfg.TrafficSource }),
	boolSetting("traffic-rates-enabled", "report read-only interface traffic rates in inform payloads", func(flags *runtimeFlags) *bool { return &flags.trafficRatesEnabled }, func(cfg appconfig.Config) bool { return cfg.TrafficRatesEnabled }),
	stringSetting("log-source", "read-only log source: off, journalctl, or syslog", func(flags *runtimeFlags) *string { return &flags.logSource }, func(cfg appconfig.Config) string { return cfg.LogSource }),
	stringSetting("proc-source", "read-only proc source: off or procfs", func(flags *runtimeFlags) *string { return &flags.procSource }, func(cfg appconfig.Config) string { return cfg.ProcSource }),
	boolSetting("dbus-enabled", "enable optional D-Bus availability checks", func(flags *runtimeFlags) *bool { return &flags.dbusEnabled }, func(cfg appconfig.Config) bool { return cfg.DBusEnabled }),
	stringSetting("dbus-bus", "D-Bus bus for optional checks: system or session", func(flags *runtimeFlags) *string { return &flags.dbusBus }, func(cfg appconfig.Config) string { return cfg.DBusBus }),
	stringSetting("syslog-path", "syslog file path used by -log-source syslog", func(flags *runtimeFlags) *string { return &flags.syslogPath }, func(cfg appconfig.Config) string { return cfg.SyslogPath }),
	stringSetting("controller", "optional UniFi inform URL, for example http://192.168.1.10:8080/inform", func(flags *runtimeFlags) *string { return &flags.controller }, func(cfg appconfig.Config) string { return cfg.ControllerURL }),
	intervalSetting(),
	boolSetting("no-discovery", "skip UDP discovery and only send inform when -controller is set", func(flags *runtimeFlags) *bool { return &flags.noDiscovery }, func(cfg appconfig.Config) bool { return cfg.NoDiscovery }),
	stringSetting("discovery-interface", "optional local interface name used for UDP discovery sends", func(flags *runtimeFlags) *string { return &flags.discoveryInterface }, func(cfg appconfig.Config) string { return cfg.DiscoveryInterface }),
	boolSetting("management-lan-enabled", "enable structured switch management LAN handling", func(flags *runtimeFlags) *bool { return &flags.managementLAN.Enabled }, func(cfg appconfig.Config) bool { return cfg.ManagementLAN.Enabled }),
	intSetting("management-lan-vlan", "structured management LAN VLAN ID; 0 leaves it unset", func(flags *runtimeFlags) *int { return &flags.managementLAN.VLAN }, func(cfg appconfig.Config) int { return cfg.ManagementLAN.VLAN }),
	stringSetting("management-lan-mode", "management LAN mode: metadata-only, preexisting-interface, or planned-host-vlan", func(flags *runtimeFlags) *string { return &flags.managementLAN.Mode }, func(cfg appconfig.Config) string { return cfg.ManagementLAN.Mode }),
	stringSetting("management-lan-interface", "existing VLAN interface used by -management-lan-mode preexisting-interface", func(flags *runtimeFlags) *string { return &flags.managementLAN.Interface }, func(cfg appconfig.Config) string { return cfg.ManagementLAN.Interface }),
	stringSetting("management-lan-ip", "optional management IP for -management-lan-interface", func(flags *runtimeFlags) *string { return &flags.managementLAN.IP }, func(cfg appconfig.Config) string { return cfg.ManagementLAN.IP }),
	stringSetting("management-lan-network", "optional controller-facing management network label", func(flags *runtimeFlags) *string { return &flags.managementLAN.NetworkName }, func(cfg appconfig.Config) string { return cfg.ManagementLAN.NetworkName }),
	stringSetting("management-lan-controller-reachable", "management LAN controller reachability validation: off, warn, or required", func(flags *runtimeFlags) *string { return &flags.managementLAN.ControllerReachable }, func(cfg appconfig.Config) string { return cfg.ManagementLAN.ControllerReachable }),
	stringSetting("management-lan-adoption-strategy", "management LAN adoption strategy: untagged-first or tagged-only", func(flags *runtimeFlags) *string { return &flags.managementLAN.AdoptionStrategy }, func(cfg appconfig.Config) string { return cfg.ManagementLAN.AdoptionStrategy }),
	stringSetting("ssh-listen", "optional built-in adoption SSH listen address, for example 0.0.0.0:22", func(flags *runtimeFlags) *string { return &flags.sshListen }, func(cfg appconfig.Config) string { return cfg.SSHListen }),
	stringSetting("ssh-user", "built-in adoption SSH username", func(flags *runtimeFlags) *string { return &flags.sshUser }, func(cfg appconfig.Config) string { return cfg.SSHUser }),
	stringSetting("ssh-password", "built-in adoption SSH password", func(flags *runtimeFlags) *string { return &flags.sshPassword }, func(cfg appconfig.Config) string { return cfg.SSHPassword }),
	stringSetting("ssh-host-key", "built-in adoption SSH host key path", func(flags *runtimeFlags) *string { return &flags.sshHostKey }, func(cfg appconfig.Config) string { return cfg.SSHHostKeyPath }),
	stringSetting("ssh-state", "built-in adoption SSH state file path", func(flags *runtimeFlags) *string { return &flags.sshState }, func(cfg appconfig.Config) string { return cfg.StatePath }),
	stringSetting("status-path", "non-sensitive runtime status file path", func(flags *runtimeFlags) *string { return &flags.statusPath }, func(cfg appconfig.Config) string { return cfg.StatusPath }),
}

// registerRuntimeSettings binds flags to the same fields that YAML config can
// later populate, preserving the CLI-over-YAML precedence model.
func registerRuntimeSettings(flags *runtimeFlags, defaults appconfig.Config) {
	for _, setting := range runtimeSettings {
		setting.register(flags, defaults)
	}
}

// stringSetting describes one string-valued setting for both flag registration
// and deferred config application.
func stringSetting(name string, usage string, target func(*runtimeFlags) *string, value func(appconfig.Config) string) runtimeSetting {
	return runtimeSetting{
		flagName: name,
		register: func(flags *runtimeFlags, defaults appconfig.Config) {
			flag.StringVar(target(flags), name, value(defaults), usage)
		},
		apply: func(cfg appconfig.Config, flags *runtimeFlags) {
			*target(flags) = value(cfg)
		},
	}
}

// intSetting describes one integer setting for both flag registration and YAML
// fallback application.
func intSetting(name string, usage string, target func(*runtimeFlags) *int, value func(appconfig.Config) int) runtimeSetting {
	return runtimeSetting{
		flagName: name,
		register: func(flags *runtimeFlags, defaults appconfig.Config) {
			flag.IntVar(target(flags), name, value(defaults), usage)
		},
		apply: func(cfg appconfig.Config, flags *runtimeFlags) {
			*target(flags) = value(cfg)
		},
	}
}

// boolSetting describes one boolean setting while preserving explicit false CLI
// overrides.
func boolSetting(name string, usage string, target func(*runtimeFlags) *bool, value func(appconfig.Config) bool) runtimeSetting {
	return runtimeSetting{
		flagName: name,
		register: func(flags *runtimeFlags, defaults appconfig.Config) {
			flag.BoolVar(target(flags), name, value(defaults), usage)
		},
		apply: func(cfg appconfig.Config, flags *runtimeFlags) {
			*target(flags) = value(cfg)
		},
	}
}

// intervalSetting accepts CLI durations but reads YAML intervals in seconds to
// match the packaged config schema.
func intervalSetting() runtimeSetting {
	return runtimeSetting{
		flagName: "interval",
		register: func(flags *runtimeFlags, defaults appconfig.Config) {
			flag.DurationVar(&flags.interval, "interval", time.Duration(defaults.IntervalSeconds)*time.Second, "announcement interval")
		},
		apply: applyConfigInterval,
	}
}

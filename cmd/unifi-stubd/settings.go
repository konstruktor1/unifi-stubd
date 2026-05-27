// runtimeSettings is the shared registry for flag registration and YAML config
// application. Keeping both directions together prevents the CLI and packaged
// config surface from drifting apart.
package main

import (
	"flag"
	"time"

	appconfig "github.com/konstruktor1/unifi-stubd/internal/config"
)

type boolConfigField func(*appconfig.Config) *bool

type boolRuntimeField func(*runtimeFlags) *bool

type intConfigField func(*appconfig.Config) *int

type intRuntimeField func(*runtimeFlags) *int

// runtimeSetting ties one CLI flag to its matching YAML config field.
type runtimeSetting struct {
	flagName string
	register func(*runtimeFlags, appconfig.Config)
	apply    func(appconfig.Config, *runtimeFlags)
}

type stringConfigField func(*appconfig.Config) *string

type stringRuntimeField func(*runtimeFlags) *string

// runtimeSettings is assembled from focused groups so the flag/config surface
// stays readable while preserving one shared CLI-over-YAML registry.
var runtimeSettings = collectRuntimeSettings(
	deviceRuntimeSettings(),
	observationRuntimeSettings(),
	networkRuntimeSettings(),
	managementRuntimeSettings(),
	serviceRuntimeSettings(),
)

func collectRuntimeSettings(groups ...[]runtimeSetting) []runtimeSetting {
	total := 0
	for _, group := range groups {
		total += len(group)
	}
	out := make([]runtimeSetting, 0, total)
	for _, group := range groups {
		out = append(out, group...)
	}
	return out
}

// registerRuntimeSettings binds flags to the same fields that YAML config can
// later populate, preserving the CLI-over-YAML precedence model.
func registerRuntimeSettings(flags *runtimeFlags, defaults appconfig.Config) {
	for _, setting := range runtimeSettings {
		setting.register(flags, defaults)
	}
}

// stringSetting describes one string setting by direct runtime and config field
// accessors.
func stringSetting(name string, usage string, target stringRuntimeField, source stringConfigField) runtimeSetting {
	return runtimeSetting{
		flagName: name,
		register: func(flags *runtimeFlags, defaults appconfig.Config) {
			flag.StringVar(target(flags), name, *source(&defaults), usage)
		},
		apply: func(cfg appconfig.Config, flags *runtimeFlags) {
			*target(flags) = *source(&cfg)
		},
	}
}

// intSetting describes one integer setting by direct runtime and config field
// accessors.
func intSetting(name string, usage string, target intRuntimeField, source intConfigField) runtimeSetting {
	return runtimeSetting{
		flagName: name,
		register: func(flags *runtimeFlags, defaults appconfig.Config) {
			flag.IntVar(target(flags), name, *source(&defaults), usage)
		},
		apply: func(cfg appconfig.Config, flags *runtimeFlags) {
			*target(flags) = *source(&cfg)
		},
	}
}

// boolSetting describes one boolean setting while preserving explicit false CLI
// overrides.
func boolSetting(name string, usage string, target boolRuntimeField, source boolConfigField) runtimeSetting {
	return runtimeSetting{
		flagName: name,
		register: func(flags *runtimeFlags, defaults appconfig.Config) {
			flag.BoolVar(target(flags), name, *source(&defaults), usage)
		},
		apply: func(cfg appconfig.Config, flags *runtimeFlags) {
			*target(flags) = *source(&cfg)
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

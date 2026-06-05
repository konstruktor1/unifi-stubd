package main

import (
	"context"
	"fmt"
	"strings"

	appconfig "github.com/konstruktor1/unifi-stubd/internal/config"
	"github.com/konstruktor1/unifi-stubd/internal/device"
	"github.com/konstruktor1/unifi-stubd/internal/platform"
)

type serveRuntime struct {
	profile  device.Profile
	platform platform.Platform
}

func parseServeFlags() (runtimeFlags, map[string]bool) {
	return parseRuntimeFlags(appconfig.Default())
}

// handleServeEarlyExit handles commands that must not read host state or start
// controller-facing traffic.
func handleServeEarlyExit(flags runtimeFlags) (bool, error) {
	switch {
	case flags.binaryVersion:
		fmt.Println(version)
		return true, nil
	case flags.configMigrate || flags.configMigrateDryRun:
		return true, runConfigMigration(flags)
	case strings.TrimSpace(flags.profileTemplate) != "":
		return true, printProfileTemplate(flags.profileTemplate)
	case strings.TrimSpace(flags.profileValidate) != "":
		return true, validateProfilePath(flags.profileValidate)
	default:
		return false, nil
	}
}

func prepareServeRuntime(flags *runtimeFlags, changed map[string]bool) (serveRuntime, bool, error) {
	cfg, err := loadConfig(flags.configPath, changed["config"])
	if err != nil {
		if flags.validate {
			return serveRuntime{}, false, withExitCode(2, err)
		}
		return serveRuntime{}, false, err
	}
	applyConfig(cfg, changed, flags)

	registry, err := loadProfileRegistry(*flags)
	if err != nil {
		if flags.validate {
			return serveRuntime{}, false, withExitCode(profileErrorExitCode(err), err)
		}
		return serveRuntime{}, false, err
	}
	if flags.listProfiles {
		fmt.Print(registry.FormatProfiles())
		return serveRuntime{}, true, nil
	}
	if strings.TrimSpace(flags.profileExport) != "" {
		return serveRuntime{}, true, printProfileExport(registry, flags.profileExport)
	}
	if err := validateOperationFlags(flags); err != nil {
		if flags.validate {
			return serveRuntime{}, false, withExitCode(1, err)
		}
		return serveRuntime{}, false, err
	}

	profile, ok := registry.LookupProfile(flags.profileName)
	if !ok {
		err := fmt.Errorf("unknown profile %q; known profiles: %s", flags.profileName, registry.ProfileNames())
		if flags.validate {
			return serveRuntime{}, false, withExitCode(1, err)
		}
		return serveRuntime{}, false, err
	}
	applyProfile(profile, flags)
	plt := runtimePlatform(*flags)
	if err := validateServeRuntime(*flags, profile); err != nil {
		return serveRuntime{}, false, err
	}
	if flags.validate {
		return serveRuntime{profile: profile, platform: plt}, true, validateRuntime(*flags, profile)
	}
	if err := validateLiveServeSources(*flags, profile); err != nil {
		return serveRuntime{}, false, err
	}

	enrichCtx, enrichCancel := context.WithTimeout(context.Background(), observeTimeout)
	flags.portOverrides = enrichOverrides(enrichCtx, plt, flags.portOverrides)
	enrichCancel()
	if err := validatePortOverrides(*flags); err != nil {
		return serveRuntime{}, false, err
	}
	if err := validateWANHealthTargets(*flags, profile); err != nil {
		return serveRuntime{}, false, err
	}
	return serveRuntime{profile: profile, platform: plt}, false, nil
}

func validateServeRuntime(flags runtimeFlags, profile device.Profile) error {
	if err := validateWANHealthConfig(flags); err != nil {
		if flags.validate {
			return withExitCode(1, err)
		}
		return err
	}
	// Non-live validation catches schema and policy mistakes without touching
	// host interfaces. Live checks are intentionally delayed until -validate or
	// actual runtime, where missing local interfaces should be reported.
	if err := validateManagementLAN(flags, profile, false); err != nil {
		if flags.validate {
			return withExitCode(1, err)
		}
		return err
	}
	if err := validateSourceMappings(flags, false); err != nil {
		if flags.validate {
			return withExitCode(1, err)
		}
		return err
	}
	return nil
}

func validateLiveServeSources(flags runtimeFlags, profile device.Profile) error {
	if flags.dryRunPlan {
		return nil
	}
	// Dry-run plans may describe unsupported host-network actions. All other
	// paths require live source checks before any controller-visible payload is
	// built.
	if err := validateSourceMappings(flags, true); err != nil {
		return err
	}
	return validateManagementLAN(flags, profile, true)
}

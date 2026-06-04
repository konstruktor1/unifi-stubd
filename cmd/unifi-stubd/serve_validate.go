package main

import (
	"fmt"

	"github.com/konstruktor1/unifi-stubd/internal/device"
)

func validateRuntime(flags runtimeFlags, profile device.Profile) error {
	if err := validateIdentityFlags(flags); err != nil {
		return withExitCode(1, err)
	}
	if err := validatePortOverrides(flags); err != nil {
		return withExitCode(1, err)
	}
	if err := validateWANHealthTargets(flags, profile); err != nil {
		return withExitCode(1, err)
	}
	if err := validateSourceMappings(flags, true); err != nil {
		return withExitCode(1, err)
	}
	if err := validateManagementLAN(flags, profile, true); err != nil {
		return withExitCode(1, err)
	}
	fmt.Printf("configuration valid: profile=%s source=%s payload=%s\n", profile.Name, profile.Source, profile.Payload.Kind)
	return nil
}

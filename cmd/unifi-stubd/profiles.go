// Profile CLI actions run before any network traffic can start.
// External profile loading, validation, export, and template generation share
// the same registry path as the daemon runtime.
package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/device"
	"github.com/konstruktor1/unifi-stubd/internal/device/profiledata"
)

func loadProfileRegistry(flags runtimeFlags) (device.ProfileRegistry, error) {
	registry := device.NewProfileRegistry()
	for _, path := range []string{strings.TrimSpace(flags.profileDir), strings.TrimSpace(flags.profileFile)} {
		if path == "" {
			continue
		}
		if err := registry.LoadProfilePath(path); err != nil {
			return registry, fmt.Errorf("load profile registry: %w", err)
		}
	}
	return registry, nil
}

func printProfileTemplate(kind string) error {
	data, err := device.ProfileTemplateYAML(kind)
	if err != nil {
		return withExitCode(1, err)
	}
	fmt.Print(string(data))
	return nil
}

func validateProfilePath(path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return withExitCode(1, errors.New("-profile-validate requires a file or directory path"))
	}
	registry := device.NewProfileRegistry()
	if err := registry.LoadProfilePath(path); err != nil {
		return withExitCode(profileErrorExitCode(err), err)
	}
	fmt.Printf("profiles valid: %s\n", path)
	return nil
}

func printProfileExport(registry device.ProfileRegistry, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return withExitCode(1, errors.New("-profile-export requires a profile name"))
	}
	data, err := registry.ExportProfileYAML(name)
	if err != nil {
		return withExitCode(1, err)
	}
	fmt.Print(string(data))
	return nil
}

func profileErrorExitCode(err error) int {
	var pathErr *profiledata.PathError
	if errors.As(err, &pathErr) {
		switch pathErr.Kind {
		case profiledata.ErrorKindIO, profiledata.ErrorKindParse:
			return 2
		default:
			return 1
		}
	}
	return 1
}

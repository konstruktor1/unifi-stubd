// Package device provides the public facade over built-in and external profile
// data. CLI and payload code use this API instead of importing the
// loader/registry implementation directly.
package device

import (
	"fmt"

	"github.com/konstruktor1/unifi-stubd/internal/device/profiledata"
)

// ProfileRegistry contains built-in and caller-loaded device profiles.
type ProfileRegistry struct {
	data profiledata.Registry
}

// NewProfileRegistry returns a registry initialized with built-in profiles.
func NewProfileRegistry() ProfileRegistry {
	return ProfileRegistry{data: profiledata.BuiltinRegistry()}
}

// LoadProfilePath loads one external profile file or profile directory.
func (r *ProfileRegistry) LoadProfilePath(path string) error {
	if err := r.data.LoadPath(path); err != nil {
		return fmt.Errorf("load profile path: %w", err)
	}
	return nil
}

// Profiles returns a copy of the built-in device profiles.
func Profiles() []Profile {
	return NewProfileRegistry().Profiles()
}

// Profiles returns a copy of the registry profiles.
func (r ProfileRegistry) Profiles() []Profile {
	return r.data.Profiles()
}

// LookupProfile returns a built-in profile by profile name or model identifier.
func LookupProfile(name string) (Profile, bool) {
	return NewProfileRegistry().LookupProfile(name)
}

// LookupProfile returns a profile by profile name or model identifier.
func (r ProfileRegistry) LookupProfile(name string) (Profile, bool) {
	dataProfile, ok := r.data.Lookup(name)
	if !ok {
		return Profile{}, false
	}
	return dataProfile, true
}

// ProfileNames returns the known profile names as a comma-separated list.
func ProfileNames() string {
	return NewProfileRegistry().ProfileNames()
}

// ProfileNames returns the known profile names as a comma-separated list.
func (r ProfileRegistry) ProfileNames() string {
	return r.data.Names()
}

// FormatProfiles returns a human-readable table of built-in profiles.
func FormatProfiles() string {
	return NewProfileRegistry().FormatProfiles()
}

// FormatProfiles returns a human-readable table of registry profiles.
func (r ProfileRegistry) FormatProfiles() string {
	return r.data.Format()
}

// ExportProfileYAML returns a profile as canonical YAML.
func (r ProfileRegistry) ExportProfileYAML(name string) ([]byte, error) {
	data, err := r.data.ExportYAML(name)
	if err != nil {
		return nil, fmt.Errorf("export profile YAML: %w", err)
	}
	return data, nil
}

// ProfileTemplateYAML returns a starter profile template for kind.
func ProfileTemplateYAML(kind string) ([]byte, error) {
	data, err := profiledata.TemplateYAML(kind)
	if err != nil {
		return nil, fmt.Errorf("profile template YAML: %w", err)
	}
	return data, nil
}

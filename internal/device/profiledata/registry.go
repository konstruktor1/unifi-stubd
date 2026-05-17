package profiledata

import (
	"fmt"
	"sort"
	"strings"
)

const defaultDeviceType = "usw"

type record struct {
	source  string
	order   int
	profile Profile
}

var registry []record

// Register adds one decoded profile to the global built-in profile registry.
func Register(source string, order int, profile Profile) {
	for _, record := range registry {
		if record.profile.Name == profile.Name {
			panic(fmt.Sprintf("duplicate profile name %q in %s and %s", profile.Name, record.source, source))
		}
		if strings.EqualFold(record.profile.Model, profile.Model) {
			panic(fmt.Sprintf("duplicate profile model %q in %s and %s", profile.Model, record.source, source))
		}
	}
	registry = append(registry, record{
		source:  source,
		order:   order,
		profile: cloneProfile(profile),
	})
}

// Profiles returns a copy of the built-in device profiles.
func Profiles() []Profile {
	records := append([]record{}, registry...)
	sort.SliceStable(records, func(i, j int) bool {
		if records[i].order != records[j].order {
			return records[i].order < records[j].order
		}
		return records[i].profile.Name < records[j].profile.Name
	})
	out := make([]Profile, 0, len(records))
	for _, record := range records {
		out = append(out, cloneProfile(record.profile))
	}
	return out
}

// Lookup returns a built-in profile by profile name or model identifier.
func Lookup(name string) (Profile, bool) {
	name = strings.ToLower(strings.TrimSpace(name))
	for _, profile := range Profiles() {
		if profile.Name == name || strings.ToLower(profile.Model) == name {
			return profile, true
		}
	}
	return Profile{}, false
}

// Names returns the known profile names as a comma-separated list.
func Names() string {
	profiles := Profiles()
	names := make([]string, 0, len(profiles))
	for _, profile := range profiles {
		names = append(names, profile.Name)
	}
	return strings.Join(names, ", ")
}

// Format returns a human-readable table of built-in profiles.
func Format() string {
	var b strings.Builder
	for _, profile := range Profiles() {
		fmt.Fprintf(&b, "%-15s %-6s %-15s ports=%-2d speed=%-5d version=%s  %s\n",
			profile.Name,
			deviceTypeOrDefault(profile.DeviceType),
			profile.Model,
			profile.Ports,
			firstNonZero(profile.PortSpeed, 1000),
			profile.Version,
			profile.Description,
		)
	}
	return b.String()
}

func cloneProfile(profile Profile) Profile {
	profile.PortGroups = clonePortGroups(profile.PortGroups)
	profile.PortNames = cloneStrings(profile.PortNames)
	profile.PortRoles = cloneStrings(profile.PortRoles)
	profile.PortNetworkGroups = cloneStrings(profile.PortNetworkGroups)
	return profile
}

func clonePortGroups(groups []PortGroup) []PortGroup {
	if len(groups) == 0 {
		return nil
	}
	out := make([]PortGroup, len(groups))
	copy(out, groups)
	return out
}

func cloneStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, len(values))
	copy(out, values)
	return out
}

func deviceTypeOrDefault(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return defaultDeviceType
	}
	return value
}

func firstNonZero(values ...int) int {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}

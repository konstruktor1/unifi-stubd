package device

import (
	"fmt"
	"strings"
)

// ProfileNames returns the known profile names as a comma-separated list.
func ProfileNames() string {
	return NewProfileRegistry().ProfileNames()
}

// ProfileNames returns the known profile names as a comma-separated list.
func (r ProfileRegistry) ProfileNames() string {
	profiles := r.Profiles()
	names := make([]string, 0, len(profiles))
	for _, profile := range profiles {
		names = append(names, profile.Name)
	}
	return strings.Join(names, ", ")
}

// FormatProfiles returns a human-readable table of built-in profiles.
func FormatProfiles() string {
	return NewProfileRegistry().FormatProfiles()
}

// FormatProfiles returns a human-readable table of profiles.
func (r ProfileRegistry) FormatProfiles() string {
	var b strings.Builder
	for _, profile := range r.Profiles() {
		recommended := ""
		if profile.Recommended {
			recommended = " recommended"
		}
		fmt.Fprintf(&b, "%-15s %-6s %-15s kind=%-7s source=%-8s stability=%-12s ports=%-2d speed=%-5d version=%s%s  %s\n",
			profile.Name,
			deviceTypeOrDefault(profile.DeviceType),
			profile.Model,
			profile.Payload.Kind,
			profile.SourceType,
			profile.Stability,
			profile.Ports,
			firstNonZero(profile.PortSpeed, 1000),
			profile.Version,
			recommended,
			profile.Description,
		)
	}
	return b.String()
}

// deviceTypeOrDefault keeps older profiles that predate explicit device_type
// usable as switch profiles.
func deviceTypeOrDefault(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return defaultDeviceType
	}
	return value
}

// firstNonZero returns the first configured numeric value from a fallback list.
func firstNonZero(values ...int) int {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}

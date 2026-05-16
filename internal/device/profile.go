package device

import (
	"crypto/sha256"
	"fmt"
	"net"
	"strings"
)

// Profile defines a built-in UniFi switch profile.
type Profile struct {
	// Name is the short CLI and config name.
	Name string
	// Model is the UniFi model identifier.
	Model string
	// ModelDisplay is the human-readable UniFi model name.
	ModelDisplay string
	// Version is the firmware version reported by this profile.
	Version string
	// Ports is the number of switch ports.
	Ports int
	// PortGroups describe non-uniform physical port layouts.
	PortGroups []PortGroup
	// PortSpeed is the default access port speed in Mbps.
	PortSpeed int
	// UplinkSpeed is the uplink port speed in Mbps.
	UplinkSpeed int
	// PortMedia is the default access port media label.
	PortMedia string
	// UplinkMedia is the uplink port media label.
	UplinkMedia string
	// Description is the short label shown in profile listings.
	Description string
}

var profiles = []Profile{
	{
		Name:         "us8",
		Model:        "US8",
		ModelDisplay: "UniFi Switch 8",
		Version:      "7.4.1.16850",
		Ports:        8,
		PortSpeed:    1000,
		UplinkSpeed:  1000,
		PortMedia:    "GE",
		UplinkMedia:  "GE",
		Description:  "8-port UniFi Switch",
	},
	{
		Name:         "us8p60",
		Model:        "US8P60",
		ModelDisplay: "UniFi Switch 8 60W",
		Version:      "7.4.1.16850",
		Ports:        8,
		PortSpeed:    1000,
		UplinkSpeed:  1000,
		PortMedia:    "GE",
		UplinkMedia:  "GE",
		Description:  "8-port UniFi Switch with PoE",
	},
	{
		Name:         "us16p150",
		Model:        "US16P150",
		ModelDisplay: "UniFi Switch 16 POE-150W",
		Version:      "7.4.1.16850",
		Ports:        16,
		PortSpeed:    1000,
		UplinkSpeed:  1000,
		PortMedia:    "GE",
		UplinkMedia:  "GE",
		Description:  "16-port UniFi Switch with PoE",
	},
	{
		Name:         "us16xg",
		Model:        "US16XG",
		ModelDisplay: "UniFi Switch 16 XG",
		Version:      "7.4.1.16850",
		Ports:        16,
		PortSpeed:    10000,
		UplinkSpeed:  10000,
		PortMedia:    "SFP+",
		UplinkMedia:  "SFP+",
		Description:  "16-port 10G aggregation switch",
	},
	{
		Name:         "usw-pro-xg-48",
		Model:        "USWProXG48",
		ModelDisplay: "UniFi Pro XG 48",
		Version:      "7.4.1.16850",
		Ports:        52,
		PortGroups: []PortGroup{
			{Count: 16, Speed: 2500, Media: "GE"},
			{Count: 32, Speed: 10000, Media: "GE"},
			{Count: 4, Speed: 25000, Media: "SFP28", Uplink: true},
		},
		PortSpeed:   10000,
		UplinkSpeed: 25000,
		PortMedia:   "GE",
		UplinkMedia: "SFP28",
		Description: "Pro XG 48 with 10G RJ45 and 25G SFP28 uplinks",
	},
	{
		Name:         "us24p250",
		Model:        "US24P250",
		ModelDisplay: "UniFi Switch 24 POE-250W",
		Version:      "7.4.1.16850",
		Ports:        24,
		PortSpeed:    1000,
		UplinkSpeed:  1000,
		PortMedia:    "GE",
		UplinkMedia:  "GE",
		Description:  "24-port UniFi Switch with PoE",
	},
	{
		Name:         "us48p500",
		Model:        "US48P500",
		ModelDisplay: "UniFi Switch 48 POE-500W",
		Version:      "7.4.1.16850",
		Ports:        48,
		PortSpeed:    1000,
		UplinkSpeed:  1000,
		PortMedia:    "GE",
		UplinkMedia:  "GE",
		Description:  "48-port UniFi Switch with PoE",
	},
}

// Profiles returns a copy of the built-in device profiles.
func Profiles() []Profile {
	out := make([]Profile, len(profiles))
	copy(out, profiles)
	for i := range out {
		out[i].PortGroups = clonePortGroups(out[i].PortGroups)
	}
	return out
}

// LookupProfile returns a built-in profile by profile name or model identifier.
func LookupProfile(name string) (Profile, bool) {
	name = strings.ToLower(strings.TrimSpace(name))
	for _, profile := range profiles {
		if profile.Name == name || strings.ToLower(profile.Model) == name {
			return profile, true
		}
	}
	return Profile{}, false
}

// PortOptions converts p to generated switch port options.
func (p Profile) PortOptions() PortOptions {
	return PortOptions{
		Speed:       p.PortSpeed,
		UplinkSpeed: p.UplinkSpeed,
		Media:       p.PortMedia,
		UplinkMedia: p.UplinkMedia,
		PortGroups:  clonePortGroups(p.PortGroups),
	}
}

func clonePortGroups(groups []PortGroup) []PortGroup {
	if len(groups) == 0 {
		return nil
	}
	out := make([]PortGroup, len(groups))
	copy(out, groups)
	return out
}

// AutoMAC derives a stable locally administered MAC address from seed.
func AutoMAC(seed string) net.HardwareAddr {
	sum := sha256.Sum256([]byte(strings.TrimSpace(seed)))
	mac := net.HardwareAddr{sum[0], sum[1], sum[2], sum[3], sum[4], sum[5]}
	mac[0] = (mac[0] | 0x02) & 0xfe
	return mac
}

// ProfileNames returns the known profile names as a comma-separated list.
func ProfileNames() string {
	names := make([]string, 0, len(profiles))
	for _, profile := range profiles {
		names = append(names, profile.Name)
	}
	return strings.Join(names, ", ")
}

// FormatProfiles returns a human-readable table of built-in profiles.
func FormatProfiles() string {
	var b strings.Builder
	for _, profile := range profiles {
		fmt.Fprintf(&b, "%-15s %-15s ports=%-2d speed=%-5d version=%s  %s\n",
			profile.Name,
			profile.Model,
			profile.Ports,
			firstNonZero(profile.PortSpeed, 1000),
			profile.Version,
			profile.Description,
		)
	}
	return b.String()
}

func firstNonZero(values ...int) int {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}

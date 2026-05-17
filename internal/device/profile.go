package device

import (
	"crypto/sha256"
	"fmt"
	"net"
	"strings"
)

// Profile defines a built-in UniFi device profile.
type Profile struct {
	// Name is the short CLI and config name.
	Name string
	// Model is the UniFi model identifier.
	Model string
	// ModelDisplay is the human-readable UniFi model name.
	ModelDisplay string
	// DeviceType is the controller-facing UniFi device family.
	DeviceType string
	// Version is the firmware version reported by this profile.
	Version string
	// Ports is the number of reported Ethernet ports.
	Ports int
	// PortGroups describe non-uniform physical port layouts.
	PortGroups []PortGroup
	// PortNames optionally override one-based port display labels.
	PortNames []string
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

const (
	defaultFirmwareVersion = "7.4.1.16850"
	defaultGatewayVersion  = "4.4.57.5578372"
	defaultUXGVersion      = "5.0.16"
	deviceTypeUSW          = "usw"
	deviceTypeUGW          = "ugw"
	deviceTypeUXG          = "uxg"
	mediaSFPPlus           = "SFP+"
	mediaSFP28             = "SFP28"
)

var profiles = []Profile{
	{
		Name:         "us8",
		Model:        "US8",
		ModelDisplay: "UniFi Switch 8",
		DeviceType:   deviceTypeUSW,
		Version:      defaultFirmwareVersion,
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
		DeviceType:   deviceTypeUSW,
		Version:      defaultFirmwareVersion,
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
		DeviceType:   deviceTypeUSW,
		Version:      defaultFirmwareVersion,
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
		DeviceType:   deviceTypeUSW,
		Version:      defaultFirmwareVersion,
		Ports:        16,
		PortSpeed:    10000,
		UplinkSpeed:  10000,
		PortMedia:    mediaSFPPlus,
		UplinkMedia:  mediaSFPPlus,
		Description:  "16-port 10G aggregation switch",
	},
	{
		Name:         "usaggpro",
		Model:        "USAGGPRO",
		ModelDisplay: "UniFi Switch Aggregation PRO",
		DeviceType:   deviceTypeUSW,
		Version:      defaultFirmwareVersion,
		Ports:        32,
		PortGroups: []PortGroup{
			{Count: 28, Speed: 10000, Media: mediaSFPPlus},
			{Count: 4, Speed: 25000, Media: mediaSFP28, Uplink: true},
		},
		PortSpeed:   10000,
		UplinkSpeed: 25000,
		PortMedia:   mediaSFPPlus,
		UplinkMedia: mediaSFP28,
		Description: "32-port Pro Aggregation switch with 10G SFP+ and 25G SFP28",
	},
	{
		Name:         "usw-pro-xg-48",
		Model:        "USWProXG48",
		ModelDisplay: "UniFi Pro XG 48",
		DeviceType:   deviceTypeUSW,
		Version:      defaultFirmwareVersion,
		Ports:        52,
		PortGroups: []PortGroup{
			{Count: 16, Speed: 2500, Media: "GE"},
			{Count: 32, Speed: 10000, Media: "GE"},
			{Count: 4, Speed: 25000, Media: mediaSFP28, Uplink: true},
		},
		PortSpeed:   10000,
		UplinkSpeed: 25000,
		PortMedia:   "GE",
		UplinkMedia: mediaSFP28,
		Description: "Pro XG 48 with 10G RJ45 and 25G SFP28 uplinks",
	},
	{
		Name:         "us24p250",
		Model:        "US24P250",
		ModelDisplay: "UniFi Switch 24 POE-250W",
		DeviceType:   deviceTypeUSW,
		Version:      defaultFirmwareVersion,
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
		DeviceType:   deviceTypeUSW,
		Version:      defaultFirmwareVersion,
		Ports:        48,
		PortSpeed:    1000,
		UplinkSpeed:  1000,
		PortMedia:    "GE",
		UplinkMedia:  "GE",
		Description:  "48-port UniFi Switch with PoE",
	},
	{
		Name:         "ugw3",
		Model:        "UGW3",
		ModelDisplay: "UniFi Security Gateway 3P",
		DeviceType:   deviceTypeUGW,
		Version:      defaultGatewayVersion,
		Ports:        3,
		PortNames:    []string{"WAN 1", "LAN 1", "WAN 2 / LAN 2"},
		PortSpeed:    1000,
		UplinkSpeed:  1000,
		PortMedia:    "GE",
		UplinkMedia:  "GE",
		Description:  "experimental 3-port UniFi Security Gateway stub",
	},
	{
		Name:         "uxgpro",
		Model:        "UXGPRO",
		ModelDisplay: "UniFi Next-Generation Gateway Pro",
		DeviceType:   deviceTypeUXG,
		Version:      defaultUXGVersion,
		Ports:        4,
		PortGroups: []PortGroup{
			{Count: 2, Speed: 1000, Media: "GE"},
			{Count: 2, Speed: 10000, Media: mediaSFPPlus, Uplink: true},
		},
		PortNames:   []string{"WAN", "LAN", "WAN2", "LAN2"},
		PortSpeed:   1000,
		UplinkSpeed: 10000,
		PortMedia:   "GE",
		UplinkMedia: mediaSFPPlus,
		Description: "experimental UXG-Pro gateway with 1G RJ45 and 10G SFP+ ports",
	},
}

// Profiles returns a copy of the built-in device profiles.
func Profiles() []Profile {
	out := make([]Profile, len(profiles))
	copy(out, profiles)
	for i := range out {
		out[i].PortGroups = clonePortGroups(out[i].PortGroups)
		out[i].PortNames = cloneStrings(out[i].PortNames)
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
		UplinkPort:  0,
		PortGroups:  clonePortGroups(p.PortGroups),
		PortNames:   cloneStrings(p.PortNames),
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

func cloneStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, len(values))
	copy(out, values)
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

func deviceTypeOrDefault(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return deviceTypeUSW
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

package device

import (
	"crypto/sha256"
	"fmt"
	"net"
	"strings"
)

type Profile struct {
	Name         string
	Model        string
	ModelDisplay string
	Version      string
	Ports        int
	PortSpeed    int
	UplinkSpeed  int
	PortMedia    string
	UplinkMedia  string
	Description  string
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

func Profiles() []Profile {
	out := make([]Profile, len(profiles))
	copy(out, profiles)
	return out
}

func LookupProfile(name string) (Profile, bool) {
	name = strings.ToLower(strings.TrimSpace(name))
	for _, profile := range profiles {
		if profile.Name == name || strings.ToLower(profile.Model) == name {
			return profile, true
		}
	}
	return Profile{}, false
}

func (p Profile) PortOptions() PortOptions {
	return PortOptions{
		Speed:       p.PortSpeed,
		UplinkSpeed: p.UplinkSpeed,
		Media:       p.PortMedia,
		UplinkMedia: p.UplinkMedia,
	}
}

func AutoMAC(seed string) net.HardwareAddr {
	sum := sha256.Sum256([]byte(strings.TrimSpace(seed)))
	mac := net.HardwareAddr{sum[0], sum[1], sum[2], sum[3], sum[4], sum[5]}
	mac[0] = (mac[0] | 0x02) & 0xfe
	return mac
}

func ProfileNames() string {
	names := make([]string, 0, len(profiles))
	for _, profile := range profiles {
		names = append(names, profile.Name)
	}
	return strings.Join(names, ", ")
}

func FormatProfiles() string {
	var b strings.Builder
	for _, profile := range profiles {
		fmt.Fprintf(&b, "%-10s %-10s ports=%-2d speed=%-5d version=%s  %s\n",
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

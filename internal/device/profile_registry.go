package device

import (
	"github.com/konstruktor1/unifi-stubd/internal/device/profiledata"
)

// Profiles returns a copy of the built-in device profiles.
func Profiles() []Profile {
	dataProfiles := profiledata.Profiles()
	out := make([]Profile, 0, len(dataProfiles))
	for _, dataProfile := range dataProfiles {
		out = append(out, profileFromData(dataProfile))
	}
	return out
}

// LookupProfile returns a built-in profile by profile name or model identifier.
func LookupProfile(name string) (Profile, bool) {
	dataProfile, ok := profiledata.Lookup(name)
	if !ok {
		return Profile{}, false
	}
	return profileFromData(dataProfile), true
}

// ProfileNames returns the known profile names as a comma-separated list.
func ProfileNames() string {
	return profiledata.Names()
}

// FormatProfiles returns a human-readable table of built-in profiles.
func FormatProfiles() string {
	return profiledata.Format()
}

func profileFromData(dataProfile profiledata.Profile) Profile {
	return Profile{
		Name:              dataProfile.Name,
		Model:             dataProfile.Model,
		ModelDisplay:      dataProfile.ModelDisplay,
		DeviceType:        dataProfile.DeviceType,
		Version:           dataProfile.Version,
		Ports:             dataProfile.Ports,
		PortGroups:        portGroupsFromData(dataProfile.PortGroups),
		PortNames:         cloneStrings(dataProfile.PortNames),
		PortRoles:         cloneStrings(dataProfile.PortRoles),
		PortNetworkGroups: cloneStrings(dataProfile.PortNetworkGroups),
		PortSpeed:         dataProfile.PortSpeed,
		UplinkSpeed:       dataProfile.UplinkSpeed,
		PortMedia:         dataProfile.PortMedia,
		UplinkMedia:       dataProfile.UplinkMedia,
		Description:       dataProfile.Description,
	}
}

func portGroupsFromData(groups []profiledata.PortGroup) []PortGroup {
	if len(groups) == 0 {
		return nil
	}
	out := make([]PortGroup, 0, len(groups))
	for _, group := range groups {
		out = append(out, PortGroup{
			Count:  group.Count,
			Speed:  group.Speed,
			Media:  group.Media,
			Uplink: group.Uplink,
		})
	}
	return out
}

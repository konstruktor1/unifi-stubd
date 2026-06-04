package device

// portLayout is the internal resolved profile layout used while building ports.
type portLayout struct {
	Speed             int
	UplinkSpeed       int
	Media             string
	UplinkMedia       string
	UplinkPort        int
	PortGroups        []PortGroup
	PortNames         []string
	PortRoles         []string
	PortNetworkGroups []string
}

// profilePortLayout resolves profile layout plus runtime-only overrides.
func profilePortLayout(profile Profile, options PortBuildOptions) portLayout {
	layout := portLayout{
		Speed:             profile.PortSpeed,
		UplinkSpeed:       profile.UplinkSpeed,
		Media:             profile.PortMedia,
		UplinkMedia:       profile.UplinkMedia,
		UplinkPort:        options.UplinkPort,
		PortGroups:        cloneNonEmptySlice(profile.PortGroups),
		PortNames:         cloneNonEmptySlice(profile.PortNames),
		PortRoles:         cloneNonEmptySlice(profile.PortRoles),
		PortNetworkGroups: cloneNonEmptySlice(profile.PortNetworkGroups),
	}
	if options.LinkSpeed > 0 {
		layout.Speed = options.LinkSpeed
		layout.UplinkSpeed = options.LinkSpeed
		layout.Media = ""
		layout.UplinkMedia = ""
		layout.PortGroups = nil
	}
	if options.UplinkSpeed > 0 {
		layout.UplinkSpeed = options.UplinkSpeed
		if layout.UplinkMedia == "" || layout.UplinkMedia == layout.Media {
			layout.UplinkMedia = ""
		}
	}
	return layout
}

// normalizePortLayout applies profile-neutral defaults used by generated ports.
func normalizePortLayout(layout portLayout) portLayout {
	if layout.Speed <= 0 {
		layout.Speed = 1000
	}
	if layout.UplinkSpeed <= 0 {
		layout.UplinkSpeed = layout.Speed
	}
	if layout.Media == "" {
		layout.Media = mediaForSpeed(layout.Speed)
	}
	if layout.UplinkMedia == "" {
		layout.UplinkMedia = mediaForSpeed(layout.UplinkSpeed)
	}
	return layout
}

// mediaForSpeed returns the UniFi media label implied by a link speed.
func mediaForSpeed(speed int) string {
	if speed >= 10000 {
		return mediaSFPPlus
	}
	return "GE"
}

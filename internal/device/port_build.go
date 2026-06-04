package device

// switchPortsWithLayout returns count generated switch ports using layout.
func switchPortsWithLayout(count int, layout portLayout) []Port {
	if count < 1 {
		count = 1
	}
	layout = normalizePortLayout(layout)
	if ports := groupedSwitchPorts(count, layout); len(ports) > 0 {
		return applyUplinkPort(ports, layout.UplinkPort)
	}

	ports := make([]Port, 0, count)
	for i := 1; i <= count; i++ {
		speed := layout.Speed
		media := layout.Media
		if i == 1 {
			speed = layout.UplinkSpeed
			media = layout.UplinkMedia
		}
		ports = append(ports, generatedPort(i, speed, media, i == 1, i == 1, layout.PortNames, layout.PortRoles, layout.PortNetworkGroups))
	}
	return applyUplinkPort(ports, layout.UplinkPort)
}

// groupedSwitchPorts generates physical layouts with non-uniform speed or media blocks.
func groupedSwitchPorts(count int, layout portLayout) []Port {
	if len(layout.PortGroups) == 0 {
		return nil
	}
	total := 0
	uplinkIndex := 0
	for _, group := range layout.PortGroups {
		if group.Count < 1 {
			return nil
		}
		if group.Uplink && uplinkIndex == 0 {
			uplinkIndex = total + 1
		}
		total += group.Count
	}
	if total != count {
		return nil
	}
	if uplinkIndex == 0 {
		uplinkIndex = 1
	}

	ports := make([]Port, 0, count)
	index := 0
	for _, group := range layout.PortGroups {
		speed := group.Speed
		if speed <= 0 {
			speed = layout.Speed
		}
		media := group.Media
		if media == "" {
			media = mediaForSpeed(speed)
		}
		for range group.Count {
			index++
			isUplink := index == uplinkIndex
			portSpeed := speed
			portMedia := media
			if isUplink {
				portSpeed = layout.UplinkSpeed
				portMedia = layout.UplinkMedia
			}
			ports = append(ports, generatedPort(
				index,
				portSpeed,
				portMedia,
				isUplink,
				group.Uplink,
				layout.PortNames,
				layout.PortRoles,
				layout.PortNetworkGroups,
			))
		}
	}
	return ports
}

// generatedPort builds one deterministic port entry before runtime overrides.
func generatedPort(index, speed int, media string, uplink, profileUplink bool, names, roles, networkGroups []string) Port {
	profileRole := normalizeGatewayRole(oneBasedString(index, roles))
	port := Port{
		Index:         index,
		Name:          portName(index, names),
		Role:          profileRole,
		ProfileRole:   profileRole,
		NetworkGroup:  normalizeNetworkGroup(oneBasedString(index, networkGroups)),
		Media:         media,
		Uplink:        uplink,
		ProfileUplink: profileUplink,
		Up:            true,
		Speed:         speed,
		RXBytes:       int64(1000 * index),
		TXBytes:       int64(900 * index),
		RXPackets:     1,
		TXPackets:     1,
	}
	if uplink {
		// Give the active uplink one plausible locally administered neighbor so
		// factory-default payloads are useful before any bridge or LLDP source is
		// configured.
		port.MACs = []MacTableEntry{
			{MAC: "02:aa:bb:cc:dd:01", Age: 4, Uptime: 1200, VLAN: 1, Type: deviceTypeUSW},
		}
	}
	return port
}

// applyUplinkPort moves the active upstream marker to a caller-selected
// one-based port. ProfileUplink remains on the profile-defined uplink group.
func applyUplinkPort(ports []Port, uplinkPort int) []Port {
	if uplinkPort <= 0 {
		return ports
	}
	if uplinkPort > len(ports) {
		return ports
	}
	targetIndex := uplinkPort - 1
	var uplinkMACs []MacTableEntry
	for index := range ports {
		if ports[index].Uplink && len(ports[index].MACs) > 0 {
			uplinkMACs = append([]MacTableEntry{}, ports[index].MACs...)
		}
		ports[index].Uplink = false
		if index != targetIndex {
			// Moving the active uplink also moves the synthetic topology hint so
			// the controller does not see two upstream neighbors.
			ports[index].MACs = nil
		}
	}
	ports[targetIndex].Uplink = true
	if len(ports[targetIndex].MACs) == 0 {
		ports[targetIndex].MACs = uplinkMACs
	}
	return ports
}

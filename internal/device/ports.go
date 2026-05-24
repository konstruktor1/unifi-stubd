// Package device turns profile layout data into deterministic UniFi ports
// before observations and overrides are merged. The generator preserves profile
// media, speed groups, names, roles, and uplink selection.
package device

import (
	"strconv"
	"strings"
)

const (
	deviceTypeUSW       = "usw"
	gatewayPortRoleLAN  = "lan"
	gatewayPortRoleLAN2 = "lan2"
	gatewayPortRoleWAN  = "wan"
	gatewayPortRoleWAN2 = "wan2"
	mediaSFPPlus        = "SFP+"
)

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

// BuildPorts returns generated switch ports from profile plus runtime options.
func BuildPorts(profile Profile, options PortBuildOptions) []Port {
	count := profile.Ports
	if options.Count > 0 {
		count = options.Count
	}
	return switchPortsWithLayout(count, profilePortLayout(profile, options))
}

// SwitchPorts returns count generated switch ports with profile-neutral defaults.
func SwitchPorts(count int) []Port {
	return BuildPorts(Profile{Ports: count}, PortBuildOptions{})
}

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

// ApplyUplinkNeighbor adds a configured neighbor entry to the uplink port.
func ApplyUplinkNeighbor(ports []Port, neighbor *MacTableEntry) []Port {
	if neighbor == nil || strings.TrimSpace(neighbor.MAC) == "" {
		return ports
	}
	entry := normalizeMacTableEntry(*neighbor, deviceTypeUSW)
	for index := range ports {
		if !ports[index].Uplink {
			continue
		}
		for macIndex := range ports[index].MACs {
			if strings.EqualFold(ports[index].MACs[macIndex].MAC, entry.MAC) {
				ports[index].MACs[macIndex] = entry
				return ports
			}
		}
		ports[index].MACs = append([]MacTableEntry{entry}, ports[index].MACs...)
		return ports
	}
	return ports
}

// ApplyPortNeighbors adds configured MAC-table entries to their target ports.
func ApplyPortNeighbors(ports []Port, neighbors []PortNeighbor) []Port {
	if len(neighbors) == 0 || len(ports) == 0 {
		return ports
	}
	for _, neighbor := range neighbors {
		if neighbor.Port < 1 || neighbor.Port > len(ports) || strings.TrimSpace(neighbor.Entry.MAC) == "" {
			continue
		}
		entry := normalizeMacTableEntry(neighbor.Entry, "client")
		port := &ports[neighbor.Port-1]
		replaced := false
		for index := range port.MACs {
			if strings.EqualFold(port.MACs[index].MAC, entry.MAC) {
				port.MACs[index] = mergeMacTableEntry(port.MACs[index], entry)
				replaced = true
				break
			}
		}
		if !replaced {
			port.MACs = append(port.MACs, entry)
		}
	}
	return ports
}

// mergeMacTableEntry lets configured neighbor metadata win while keeping
// observed fields as fallbacks.
func mergeMacTableEntry(observed, configured MacTableEntry) MacTableEntry {
	out := configured
	if strings.TrimSpace(out.Hostname) == "" {
		out.Hostname = observed.Hostname
	}
	if strings.TrimSpace(out.IP) == "" {
		out.IP = observed.IP
	}
	if out.VLAN == 0 {
		out.VLAN = observed.VLAN
	}
	out.Static = observed.Static || configured.Static
	return out
}

// normalizeMacTableEntry fills controller-facing defaults for configured neighbors.
func normalizeMacTableEntry(entry MacTableEntry, defaultType string) MacTableEntry {
	entry.MAC = strings.ToLower(strings.TrimSpace(entry.MAC))
	entry.Hostname = strings.TrimSpace(entry.Hostname)
	entry.IP = strings.TrimSpace(entry.IP)
	if entry.Age == 0 {
		entry.Age = 4
	}
	if entry.Uptime == 0 {
		entry.Uptime = 1200
	}
	if strings.TrimSpace(entry.Type) == "" {
		entry.Type = defaultType
	}
	return entry
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
	port := Port{
		Index:         index,
		Name:          portName(index, names),
		Role:          normalizeGatewayRole(oneBasedString(index, roles)),
		NetworkGroup:  normalizeGatewayNetworkGroup(oneBasedString(index, networkGroups)),
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

// portName returns a configured one-based port label or a deterministic default.
func portName(index int, names []string) string {
	if index < 1 {
		index = 1
	}
	if index <= len(names) {
		if name := strings.TrimSpace(names[index-1]); name != "" {
			return name
		}
	}
	return "Port " + strconv.Itoa(index)
}

// oneBasedString returns a trimmed one-based list value.
func oneBasedString(index int, values []string) string {
	if index < 1 || index > len(values) {
		return ""
	}
	return strings.TrimSpace(values[index-1])
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

// normalizeGatewayRole normalizes configured gateway role labels.
func normalizeGatewayRole(role string) string {
	return strings.ToLower(strings.TrimSpace(role))
}

// normalizeGatewayNetworkGroup normalizes configured network group labels.
func normalizeGatewayNetworkGroup(networkGroup string) string {
	return strings.TrimSpace(networkGroup)
}

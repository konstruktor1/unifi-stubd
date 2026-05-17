package payload

// This file generates deterministic port layouts and applies runtime overrides.

import (
	"strconv"
	"strings"
)

// SwitchPorts returns count generated switch ports with profile-neutral defaults.
func SwitchPorts(count int) []Port {
	return SwitchPortsWithOptions(count, PortOptions{})
}

// SwitchPortsWithOptions returns count generated switch ports using options.
func SwitchPortsWithOptions(count int, options PortOptions) []Port {
	if count < 1 {
		count = 1
	}
	options = normalizePortOptions(options)
	if ports := groupedSwitchPorts(count, options); len(ports) > 0 {
		return applyUplinkPort(ports, options.UplinkPort)
	}

	ports := make([]Port, 0, count)
	for i := 1; i <= count; i++ {
		speed := options.Speed
		media := options.Media
		if i == 1 {
			speed = options.UplinkSpeed
			media = options.UplinkMedia
		}
		ports = append(ports, generatedPort(i, speed, media, i == 1, options.PortNames, options.PortRoles, options.PortNetworkGroups))
	}
	return applyUplinkPort(ports, options.UplinkPort)
}

// ApplyPortOverrides applies per-port overrides to ports.
func ApplyPortOverrides(ports []Port, overrides []PortOverride) []Port {
	if len(overrides) == 0 || len(ports) == 0 {
		return ports
	}
	for _, override := range overrides {
		if override.Port < 1 || override.Port > len(ports) {
			continue
		}
		port := &ports[override.Port-1]
		if name := strings.TrimSpace(override.Name); name != "" {
			port.Name = name
		}
		if iface := strings.TrimSpace(override.Interface); iface != "" {
			port.Interface = iface
		}
		if mac := strings.TrimSpace(override.MAC); mac != "" {
			port.MAC = strings.ToLower(mac)
		}
		if ip := strings.TrimSpace(override.IP); ip != "" {
			port.IP = ip
		}
		if netmask := strings.TrimSpace(override.Netmask); netmask != "" {
			port.Netmask = netmask
		}
		if role := normalizeGatewayRole(override.Role); role != "" {
			port.Role = role
		}
		if networkGroup := normalizeGatewayNetworkGroup(override.NetworkGroup); networkGroup != "" {
			port.NetworkGroup = networkGroup
		}
		if override.Speed > 0 {
			port.Speed = override.Speed
			if strings.TrimSpace(override.Media) == "" {
				port.Media = mediaForSpeed(override.Speed)
			}
		}
		if override.RXBytes != 0 {
			port.RXBytes = override.RXBytes
		}
		if override.TXBytes != 0 {
			port.TXBytes = override.TXBytes
		}
		if override.RXPackets != 0 {
			port.RXPackets = override.RXPackets
		}
		if override.TXPackets != 0 {
			port.TXPackets = override.TXPackets
		}
		if override.RXErrors != 0 {
			port.RXErrors = override.RXErrors
		}
		if override.TXErrors != 0 {
			port.TXErrors = override.TXErrors
		}
		if media := strings.TrimSpace(override.Media); media != "" {
			port.Media = media
		}
		if override.Up != nil {
			port.Up = *override.Up
			if !*override.Up && override.Speed <= 0 {
				port.Speed = 0
			}
		}
	}
	return ports
}

// ApplyUplinkNeighbor adds a configured neighbor entry to the uplink port.
func ApplyUplinkNeighbor(ports []Port, neighbor *MacTableEntry) []Port {
	if neighbor == nil || strings.TrimSpace(neighbor.MAC) == "" {
		return ports
	}
	entry := normalizeMacTableEntry(*neighbor)
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
		entry := normalizeMacTableEntry(neighbor.Entry)
		port := &ports[neighbor.Port-1]
		replaced := false
		for index := range port.MACs {
			if strings.EqualFold(port.MACs[index].MAC, entry.MAC) {
				port.MACs[index] = entry
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

// normalizeMacTableEntry fills controller-facing defaults for configured neighbors.
func normalizeMacTableEntry(entry MacTableEntry) MacTableEntry {
	entry.MAC = strings.ToLower(strings.TrimSpace(entry.MAC))
	if entry.Age == 0 {
		entry.Age = 4
	}
	if entry.Uptime == 0 {
		entry.Uptime = 1200
	}
	if strings.TrimSpace(entry.Type) == "" {
		entry.Type = deviceTypeUSW
	}
	return entry
}

// groupedSwitchPorts generates physical layouts with non-uniform speed or media blocks.
func groupedSwitchPorts(count int, options PortOptions) []Port {
	if len(options.PortGroups) == 0 {
		return nil
	}
	total := 0
	uplinkIndex := 0
	for _, group := range options.PortGroups {
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
	for _, group := range options.PortGroups {
		speed := group.Speed
		if speed <= 0 {
			speed = options.Speed
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
				portSpeed = options.UplinkSpeed
				portMedia = options.UplinkMedia
			}
			ports = append(ports, generatedPort(
				index,
				portSpeed,
				portMedia,
				isUplink,
				options.PortNames,
				options.PortRoles,
				options.PortNetworkGroups,
			))
		}
	}
	return ports
}

// generatedPort builds one deterministic port entry before runtime overrides.
func generatedPort(index, speed int, media string, uplink bool, names, roles, networkGroups []string) Port {
	port := Port{
		Index:        index,
		Name:         portName(index, names),
		Role:         normalizeGatewayRole(oneBasedString(index, roles)),
		NetworkGroup: normalizeGatewayNetworkGroup(oneBasedString(index, networkGroups)),
		Media:        media,
		Uplink:       uplink,
		Up:           true,
		Speed:        speed,
		RXBytes:      int64(1000 * index),
		TXBytes:      int64(900 * index),
		RXPackets:    1,
		TXPackets:    1,
	}
	if uplink {
		port.MACs = []MacTableEntry{
			{MAC: "02:aa:bb:cc:dd:01", Age: 4, Uptime: 1200, VLAN: 1, Type: deviceTypeUSW},
		}
	}
	return port
}

// applyUplinkPort moves uplink metadata to a caller-selected one-based port.
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

// normalizePortOptions applies profile-neutral defaults used by generated ports.
func normalizePortOptions(options PortOptions) PortOptions {
	if options.Speed <= 0 {
		options.Speed = 1000
	}
	if options.UplinkSpeed <= 0 {
		options.UplinkSpeed = options.Speed
	}
	if options.Media == "" {
		options.Media = mediaForSpeed(options.Speed)
	}
	if options.UplinkMedia == "" {
		options.UplinkMedia = mediaForSpeed(options.UplinkSpeed)
	}
	return options
}

// mediaForSpeed returns the UniFi media label implied by a link speed.
func mediaForSpeed(speed int) string {
	if speed >= 10000 {
		return mediaSFPPlus
	}
	return "GE"
}

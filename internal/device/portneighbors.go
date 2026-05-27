package device

import "strings"

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

package device

import (
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
)

// ParseRemotePortID extracts numeric LLDP port identifiers such as "Port 50".
// Non-numeric identifiers are preserved separately and return zero here.
func ParseRemotePortID(value string) int {
	value = strings.TrimSpace(value)
	lower := strings.ToLower(value)
	if strings.HasPrefix(lower, "port ") {
		value = strings.TrimSpace(value[len("port "):])
	}
	port, err := strconv.Atoi(value)
	if err != nil || port < 1 || port > 65535 {
		return 0
	}
	return port
}

// TopologyDirection is the direction of a direct infrastructure edge.
type TopologyDirection string

const (
	// DirectionUpstream points through the active uplink port.
	DirectionUpstream TopologyDirection = "upstream"
	// DirectionDownstream points through any other represented port.
	DirectionDownstream TopologyDirection = "downstream"
)

// TopologyEdge is one resolved direct infrastructure connection.
type TopologyEdge struct {
	LocalPort int
	Direction TopologyDirection
	Neighbor  TopologyNeighbor
}

// TopologyResolution contains deterministic edges plus non-fatal source conflicts.
type TopologyResolution struct {
	Edges    []TopologyEdge
	Warnings []string
}

// ResolveTopology selects at most one direct infrastructure neighbor per port.
// Explicit configuration wins over LLDP. FDB/client MAC entries are deliberately
// absent from this resolver because reachability does not establish direction.
func ResolveTopology(ports []Port) (TopologyResolution, error) {
	resolution := TopologyResolution{}
	for _, port := range ports {
		neighbor, warnings, ok, err := resolvePortNeighbor(port)
		resolution.Warnings = append(resolution.Warnings, warnings...)
		if err != nil {
			return TopologyResolution{}, err
		}
		if !ok {
			continue
		}
		direction := DirectionDownstream
		if port.Uplink {
			direction = DirectionUpstream
		}
		resolution.Edges = append(resolution.Edges, TopologyEdge{
			LocalPort: port.Index,
			Direction: direction,
			Neighbor:  neighbor,
		})
	}
	return resolveDirectionConflicts(resolution)
}

func resolvePortNeighbor(port Port) (TopologyNeighbor, []string, bool, error) {
	configured := map[string]TopologyNeighbor{}
	lldp := map[string]TopologyNeighbor{}
	for _, raw := range port.Neighbors {
		neighbor, ok := normalizeTopologyNeighbor(raw)
		if !ok {
			continue
		}
		key := normalizedTopologyMAC(neighbor.MAC)
		target := lldp
		if neighbor.Source != NeighborSourceLLDP {
			target = configured
		}
		if existing, found := target[key]; found {
			target[key] = mergeTopologyNeighbor(existing, neighbor)
		} else {
			target[key] = neighbor
		}
	}

	if len(configured) > 1 {
		return TopologyNeighbor{}, nil, false, fmt.Errorf(
			"port %d has conflicting configured infrastructure neighbors: %s",
			port.Index,
			strings.Join(sortedTopologyKeys(configured), ", "),
		)
	}
	if len(configured) == 1 {
		neighbor := configured[sortedTopologyKeys(configured)[0]]
		warnings := make([]string, 0, len(lldp))
		for mac := range lldp {
			if mac == normalizedTopologyMAC(neighbor.MAC) {
				neighbor = mergeTopologyNeighbor(lldp[mac], neighbor)
				continue
			}
			warnings = append(warnings, fmt.Sprintf(
				"topology port %d: configured neighbor %s overrides LLDP neighbor %s",
				port.Index,
				neighbor.MAC,
				lldp[mac].MAC,
			))
		}
		sort.Strings(warnings)
		return neighbor, warnings, true, nil
	}

	if len(lldp) > 1 {
		return TopologyNeighbor{}, []string{fmt.Sprintf(
			"topology port %d: multiple LLDP neighbors (%s); no directed edge selected",
			port.Index,
			strings.Join(sortedTopologyKeys(lldp), ", "),
		)}, false, nil
	}
	if len(lldp) == 1 {
		return lldp[sortedTopologyKeys(lldp)[0]], nil, true, nil
	}
	return TopologyNeighbor{}, nil, false, nil
}

func resolveDirectionConflicts(resolution TopologyResolution) (TopologyResolution, error) {
	byMAC := map[string]int{}
	keep := make([]bool, len(resolution.Edges))
	for index := range keep {
		keep[index] = true
	}
	for index, edge := range resolution.Edges {
		key := normalizedTopologyMAC(edge.Neighbor.MAC)
		previousIndex, found := byMAC[key]
		if !found {
			byMAC[key] = index
			continue
		}
		previous := resolution.Edges[previousIndex]
		if previous.Direction == edge.Direction {
			continue
		}
		previousPriority := topologySourcePriority(previous.Neighbor.Source)
		currentPriority := topologySourcePriority(edge.Neighbor.Source)
		switch {
		case previousPriority == currentPriority && previousPriority > 1:
			return TopologyResolution{}, fmt.Errorf(
				"neighbor %s is configured as both %s on port %d and %s on port %d",
				edge.Neighbor.MAC,
				previous.Direction,
				previous.LocalPort,
				edge.Direction,
				edge.LocalPort,
			)
		case previousPriority == currentPriority:
			keep[previousIndex] = false
			keep[index] = false
			resolution.Warnings = append(resolution.Warnings, fmt.Sprintf(
				"topology neighbor %s appears in both directions through LLDP; both edges omitted",
				edge.Neighbor.MAC,
			))
		case previousPriority > currentPriority:
			keep[index] = false
			resolution.Warnings = append(resolution.Warnings, fmt.Sprintf(
				"topology neighbor %s: %s edge from %s overrides %s edge from %s",
				edge.Neighbor.MAC,
				previous.Direction,
				previous.Neighbor.Source,
				edge.Direction,
				edge.Neighbor.Source,
			))
		default:
			keep[previousIndex] = false
			byMAC[key] = index
			resolution.Warnings = append(resolution.Warnings, fmt.Sprintf(
				"topology neighbor %s: %s edge from %s overrides %s edge from %s",
				edge.Neighbor.MAC,
				edge.Direction,
				edge.Neighbor.Source,
				previous.Direction,
				previous.Neighbor.Source,
			))
		}
	}
	filtered := make([]TopologyEdge, 0, len(resolution.Edges))
	for index, edge := range resolution.Edges {
		if keep[index] {
			filtered = append(filtered, edge)
		}
	}
	resolution.Edges = filtered
	sort.Strings(resolution.Warnings)
	return resolution, nil
}

func normalizeTopologyNeighbor(neighbor TopologyNeighbor) (TopologyNeighbor, bool) {
	mac, err := net.ParseMAC(strings.TrimSpace(neighbor.MAC))
	if err != nil {
		return TopologyNeighbor{}, false
	}
	neighbor.MAC = strings.ToLower(mac.String())
	neighbor.Hostname = strings.TrimSpace(neighbor.Hostname)
	neighbor.IP = strings.TrimSpace(neighbor.IP)
	neighbor.DeviceType = strings.ToLower(strings.TrimSpace(neighbor.DeviceType))
	neighbor.RemotePortID = strings.TrimSpace(neighbor.RemotePortID)
	return neighbor, true
}

func mergeTopologyNeighbor(fallback, preferred TopologyNeighbor) TopologyNeighbor {
	out := preferred
	if strings.TrimSpace(out.Hostname) == "" {
		out.Hostname = fallback.Hostname
	}
	if strings.TrimSpace(out.IP) == "" {
		out.IP = fallback.IP
	}
	if strings.TrimSpace(out.DeviceType) == "" {
		out.DeviceType = fallback.DeviceType
	}
	if out.VLAN == 0 {
		out.VLAN = fallback.VLAN
	}
	if out.Age == 0 {
		out.Age = fallback.Age
	}
	if out.Uptime == 0 {
		out.Uptime = fallback.Uptime
	}
	if out.RemotePort == 0 {
		out.RemotePort = fallback.RemotePort
	}
	if strings.TrimSpace(out.RemotePortID) == "" {
		out.RemotePortID = fallback.RemotePortID
	}
	out.Static = out.Static || fallback.Static
	return out
}

func normalizedTopologyMAC(value string) string {
	mac, err := net.ParseMAC(strings.TrimSpace(value))
	if err != nil {
		return strings.ToLower(strings.TrimSpace(value))
	}
	return strings.ToLower(mac.String())
}

func sortedTopologyKeys(values map[string]TopologyNeighbor) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func topologySourcePriority(source NeighborSource) int {
	switch source {
	case NeighborSourceConfigUplink, NeighborSourceConfigPort:
		return 2
	case NeighborSourceLLDP:
		return 1
	default:
		return 0
	}
}

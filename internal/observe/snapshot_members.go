package observe

import (
	"sort"
	"strconv"
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/device"
)

// linuxMemberPortObservations reads per-member sysfs counters and speed when
// Linux bridge members correspond to visible host interfaces.
func linuxMemberPortObservations(sysfsRoot string, memberMACs map[string][]device.MacTableEntry, roles map[string]BridgeMemberRole) map[string]PortObservation {
	if len(memberMACs) == 0 {
		return nil
	}
	out := map[string]PortObservation{}
	for member := range memberMACs {
		role := memberRole(roles, member)
		if role == BridgeMemberRoleBridge || role == BridgeMemberRoleIgnored {
			continue
		}
		stats, err := ReadInterfaceStats(sysfsRoot, member)
		if err != nil && !hasCounters(stats) && stats.SpeedMbps <= 0 {
			continue
		}
		out[member] = PortObservation{
			Interface: strings.TrimSpace(member),
			SpeedMbps: stats.SpeedMbps,
			Stats:     stats,
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// mapBridgeMemberInterfaces records member interface names on platforms where
// counters are not available through the bridge observation path.
func mapBridgeMemberInterfaces(memberMACs map[string][]device.MacTableEntry, roles map[string]BridgeMemberRole) map[string]PortObservation {
	if len(memberMACs) == 0 {
		return nil
	}
	out := map[string]PortObservation{}
	for member := range memberMACs {
		role := memberRole(roles, member)
		if role == BridgeMemberRoleBridge || role == BridgeMemberRoleIgnored {
			continue
		}
		out[member] = PortObservation{Interface: strings.TrimSpace(member)}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// sortedDeviceNames provides stable member-to-port assignment independent of
// Go map iteration order.
func sortedDeviceNames(deviceMACs map[string][]device.MacTableEntry, iface, bridge string) []string {
	names := make([]string, 0, len(deviceMACs))
	for deviceName := range deviceMACs {
		names = append(names, deviceName)
	}
	sort.Slice(names, func(i, j int) bool {
		left := deviceSortKey(names[i], iface, bridge)
		right := deviceSortKey(names[j], iface, bridge)
		if left.rank != right.rank {
			return left.rank < right.rank
		}
		if left.number != right.number {
			return left.number < right.number
		}
		return left.name < right.name
	})
	return names
}

// sortKey makes bridge-member ordering deterministic across map iteration.
type sortKey struct {
	rank   int
	number int
	name   string
}

// deviceSortKey ranks uplink, virtual access devices, and bridge metadata so
// deterministic mapping follows the same topology assumptions every run.
func deviceSortKey(deviceName, iface, bridge string) sortKey {
	name := strings.ToLower(strings.TrimSpace(deviceName))
	rank := 50
	switch {
	case isUplinkDevice(name, iface, bridge):
		rank = 0
	case isBridgeDevice(name, bridge):
		rank = 90
	case strings.HasPrefix(name, "tap"):
		rank = 10
	case strings.HasPrefix(name, "veth"):
		rank = 20
	case strings.HasPrefix(name, "fwln"), strings.HasPrefix(name, "fwpr"), strings.HasPrefix(name, "fwbr"):
		rank = 30
	}
	return sortKey{rank: rank, number: firstNumber(name), name: name}
}

// isUplinkDevice recognizes the configured physical uplink member.
func isUplinkDevice(deviceName, iface, _ string) bool {
	name := strings.ToLower(strings.TrimSpace(deviceName))
	if name == "" {
		return false
	}
	return name == strings.ToLower(strings.TrimSpace(iface))
}

// isBridgeDevice recognizes the bridge device itself so it is not rendered as a
// client-facing port.
func isBridgeDevice(deviceName, bridge string) bool {
	name := strings.ToLower(strings.TrimSpace(deviceName))
	return name != "" && name == strings.ToLower(strings.TrimSpace(bridge))
}

// firstNumber gives deterministic ordering to similarly named bridge members
// such as tap2 and tap10.
func firstNumber(value string) int {
	start := -1
	for i, r := range value {
		if r >= '0' && r <= '9' {
			start = i
			break
		}
	}
	if start < 0 {
		return 0
	}
	end := start
	for end < len(value) && value[end] >= '0' && value[end] <= '9' {
		end++
	}
	number, err := strconv.Atoi(value[start:end])
	if err != nil {
		return 0
	}
	return number
}

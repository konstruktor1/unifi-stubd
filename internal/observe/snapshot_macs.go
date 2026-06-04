package observe

import (
	"net"
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/device"
)

// flattenDeviceMACs returns a deterministic fallback MAC list when per-member
// port assignment is unavailable.
func flattenDeviceMACs(deviceMACs map[string][]device.MacTableEntry, iface, bridge string) []device.MacTableEntry {
	return flattenDeviceMACsExcept(deviceMACs, iface, bridge, nil)
}

// flattenDeviceMACsExcept excludes remote upstream MACs from the fallback
// client list.
func flattenDeviceMACsExcept(deviceMACs map[string][]device.MacTableEntry, iface, bridge string, remoteMACs map[string]bool) []device.MacTableEntry {
	count := 0
	for _, macs := range deviceMACs {
		count += len(macs)
	}
	out := make([]device.MacTableEntry, 0, count)
	for _, deviceName := range sortedDeviceNames(deviceMACs, iface, bridge) {
		out = append(out, filterRemoteMACEntries(deviceMACs[deviceName], remoteMACs)...)
	}
	return out
}

// flattenDeviceMACsByRole produces the fallback uplink MAC list while excluding
// bridge metadata, ignored members, and remote upstream MACs.
func flattenDeviceMACsByRole(deviceMACs map[string][]device.MacTableEntry, roles map[string]BridgeMemberRole, iface, bridge string, remoteMACs map[string]bool) []device.MacTableEntry {
	count := 0
	for _, macs := range deviceMACs {
		count += len(macs)
	}
	out := make([]device.MacTableEntry, 0, count)
	for _, deviceName := range sortedDeviceNames(deviceMACs, iface, bridge) {
		role := memberRole(roles, deviceName)
		if role == BridgeMemberRoleBridge || role == BridgeMemberRoleIgnored {
			continue
		}
		out = append(out, filterRemoteMACEntries(deviceMACs[deviceName], remoteMACs)...)
	}
	return out
}

// RemoteMACsByBridgeMember returns MACs learned on the physical uplink member.
// These entries describe devices behind the real neighbor switch, not local
// participants of the represented virtual switch.
func RemoteMACsByBridgeMember(memberMACs map[string][]device.MacTableEntry, roles map[string]BridgeMemberRole, iface, bridge string) map[string]bool {
	if len(memberMACs) == 0 {
		return nil
	}
	out := map[string]bool{}
	for member, macs := range memberMACs {
		role := memberRole(roles, member)
		if role != BridgeMemberRoleUplink && !isUplinkDevice(member, iface, bridge) {
			continue
		}
		for _, entry := range macs {
			if key := normalizedMACKey(entry.MAC); key != "" {
				out[key] = true
			}
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// filterRemoteMACEntries removes MACs known to live behind the physical uplink.
func filterRemoteMACEntries(entries []device.MacTableEntry, remoteMACs map[string]bool) []device.MacTableEntry {
	if len(entries) == 0 || len(remoteMACs) == 0 {
		return entries
	}
	out := make([]device.MacTableEntry, 0, len(entries))
	for _, entry := range entries {
		if key := normalizedMACKey(entry.MAC); key != "" && remoteMACs[key] {
			continue
		}
		out = append(out, entry)
	}
	return out
}

// normalizeRemoteMACSet canonicalizes upstream MACs before filtering local
// bridge-member clients.
func normalizeRemoteMACSet(values map[string]bool) map[string]bool {
	if len(values) == 0 {
		return nil
	}
	out := make(map[string]bool, len(values))
	for value, enabled := range values {
		if !enabled {
			continue
		}
		if key := normalizedMACKey(value); key != "" {
			out[key] = true
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// normalizedMACKey parses MAC strings into a stable lowercase comparison key.
func normalizedMACKey(value string) string {
	mac, err := net.ParseMAC(strings.TrimSpace(value))
	if err != nil {
		return ""
	}
	return strings.ToLower(mac.String())
}

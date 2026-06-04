// Package observe classifies bridge-member forwarding-database rows into
// payload roles before any MAC table is rendered. The split matters for
// bridge-observe: access ports are local VM/container participants, the uplink
// is remote infrastructure, and the bridge device itself is only backplane
// metadata.
package observe

import (
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/device"
)

// ClassifyMembers assigns bridge-member roles for payload port mapping.
func ClassifyMembers(memberMACs map[string][]device.MacTableEntry, bridge, uplinkInterface string) map[string]BridgeMemberRole {
	if len(memberMACs) == 0 {
		return nil
	}
	roles := make(map[string]BridgeMemberRole, len(memberMACs))
	var physicalCandidates []string
	for member := range memberMACs {
		role := ClassifyMember(member, bridge, uplinkInterface)
		roles[member] = role
		if strings.TrimSpace(uplinkInterface) == "" && role == BridgeMemberRoleUnknown && isPhysicalMember(member) {
			physicalCandidates = append(physicalCandidates, member)
		}
	}
	if strings.TrimSpace(uplinkInterface) == "" && len(physicalCandidates) == 1 {
		// With exactly one physical-looking unknown member, treating it as the
		// uplink is safer than mapping upstream infrastructure as an access port.
		roles[physicalCandidates[0]] = BridgeMemberRoleUplink
	}
	return roles
}

// ClassifyMembersWithIgnores assigns roles and then marks explicitly
// ignored bridge members so they cannot consume a UniFi port.
func ClassifyMembersWithIgnores(memberMACs map[string][]device.MacTableEntry, bridge, uplinkInterface string, ignoredMembers []string) map[string]BridgeMemberRole {
	return ApplyIgnoredMembers(ClassifyMembers(memberMACs, bridge, uplinkInterface), ignoredMembers)
}

// ApplyIgnoredMembers marks configured bridge members as ignored.
func ApplyIgnoredMembers(roles map[string]BridgeMemberRole, ignoredMembers []string) map[string]BridgeMemberRole {
	ignored := ignoredMemberSet(ignoredMembers)
	if len(ignored) == 0 {
		return roles
	}
	out := make(map[string]BridgeMemberRole, len(roles)+len(ignored))
	for member, role := range roles {
		out[member] = role
		if ignored[memberNameKey(member)] {
			out[member] = BridgeMemberRoleIgnored
		}
	}
	for member := range ignored {
		if _, ok := roleByMember(out, member); !ok {
			out[member] = BridgeMemberRoleIgnored
		}
	}
	return out
}

// ClassifyMember classifies one Linux or FreeBSD bridge member.
func ClassifyMember(member, bridge, uplinkInterface string) BridgeMemberRole {
	name := strings.ToLower(strings.TrimSpace(member))
	if name == "" {
		return BridgeMemberRoleUnknown
	}
	if bridgeName := strings.ToLower(strings.TrimSpace(bridge)); bridgeName != "" && name == bridgeName {
		return BridgeMemberRoleBridge
	}
	if uplinkName := strings.ToLower(strings.TrimSpace(uplinkInterface)); uplinkName != "" && name == uplinkName {
		return BridgeMemberRoleUplink
	}
	if isVirtualAccessMember(name) {
		return BridgeMemberRoleAccess
	}
	return BridgeMemberRoleUnknown
}

func memberRole(roles map[string]BridgeMemberRole, member string) BridgeMemberRole {
	if len(roles) == 0 {
		return BridgeMemberRoleUnknown
	}
	if role, ok := roles[strings.TrimSpace(member)]; ok {
		return role
	}
	lower := strings.ToLower(strings.TrimSpace(member))
	for key, role := range roles {
		if strings.ToLower(strings.TrimSpace(key)) == lower {
			return role
		}
	}
	return BridgeMemberRoleUnknown
}

func roleByMember(roles map[string]BridgeMemberRole, member string) (BridgeMemberRole, bool) {
	member = memberNameKey(member)
	for key, role := range roles {
		if memberNameKey(key) == member {
			return role, true
		}
	}
	return BridgeMemberRoleUnknown, false
}

func ignoredMemberSet(values []string) map[string]bool {
	if len(values) == 0 {
		return nil
	}
	out := map[string]bool{}
	for _, value := range values {
		if key := memberNameKey(value); key != "" {
			out[key] = true
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func memberNameKey(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func isVirtualAccessMember(name string) bool {
	name = strings.ToLower(strings.TrimSpace(name))
	return strings.HasPrefix(name, "tap") ||
		strings.HasPrefix(name, "veth") ||
		strings.HasPrefix(name, "fwpr") ||
		strings.HasPrefix(name, "fwln") ||
		strings.HasPrefix(name, "fwbr") ||
		strings.HasPrefix(name, "epair") ||
		strings.HasPrefix(name, "vnet")
}

func isPhysicalMember(name string) bool {
	name = strings.ToLower(strings.TrimSpace(name))
	prefixes := []string{
		"eth", "eno", "ens", "enp", "enx",
		"bond", "team", "lagg",
		"em", "igb", "ix", "ixl", "ice", "bnxt", "bge", "re", "vtnet",
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

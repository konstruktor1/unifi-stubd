package observe

import "strings"

// normalizeMemberPortMap trims bridge-member pinning input before assignment.
func normalizeMemberPortMap(values map[string]int) map[string]int {
	if len(values) == 0 {
		return nil
	}
	out := make(map[string]int, len(values))
	for member, port := range values {
		member = strings.TrimSpace(member)
		if member == "" {
			continue
		}
		out[member] = port
	}
	return out
}

// normalizeMemberPorts trims source-provided member observations before merging
// them into the legacy snapshot shape.
func normalizeMemberPorts(values map[string]PortObservation) map[string]PortObservation {
	if len(values) == 0 {
		return nil
	}
	out := make(map[string]PortObservation, len(values))
	for member, observation := range values {
		member = strings.TrimSpace(member)
		if member == "" {
			continue
		}
		out[member] = observation
	}
	return out
}

// normalizeMemberRoles trims source-provided role keys before ignored-member
// policy is applied.
func normalizeMemberRoles(values map[string]BridgeMemberRole) map[string]BridgeMemberRole {
	if len(values) == 0 {
		return nil
	}
	out := make(map[string]BridgeMemberRole, len(values))
	for member, role := range values {
		member = strings.TrimSpace(member)
		if member == "" {
			continue
		}
		out[member] = role
	}
	return out
}

// cloneStrings detaches ignored-member lists passed to observation sources.
func cloneStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, len(values))
	copy(out, values)
	return out
}

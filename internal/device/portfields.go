package device

import (
	"strconv"
	"strings"
)

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

// normalizeGatewayRole normalizes configured gateway role labels.
func normalizeGatewayRole(role string) string {
	return strings.ToLower(strings.TrimSpace(role))
}

// normalizeGatewayNetworkGroup normalizes configured network group labels.
func normalizeGatewayNetworkGroup(networkGroup string) string {
	return strings.TrimSpace(networkGroup)
}

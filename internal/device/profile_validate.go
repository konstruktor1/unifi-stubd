package device

import (
	"fmt"
	"strings"
)

// validateProfile checks the semantic profile contract used by profile
// generation and payload rendering after YAML has been strictly decoded.
func validateProfile(profile Profile) error {
	if profile.SchemaVersion != schemaVersion {
		return fmt.Errorf("schema_version must be %d", schemaVersion)
	}
	for _, field := range []struct {
		name  string
		value string
	}{
		{name: "name", value: profile.Name},
		{name: "model", value: profile.Model},
	} {
		if field.value == "" {
			if field.name == "name" {
				return fmt.Errorf("name is required")
			}
			return fmt.Errorf("%s is required for %q", field.name, profile.Name)
		}
	}
	for _, field := range []struct {
		name     string
		value    int
		positive bool
	}{
		{name: "ports", value: profile.Ports, positive: true},
		{name: "port_speed", value: profile.PortSpeed},
		{name: "uplink_speed", value: profile.UplinkSpeed},
	} {
		if field.positive && field.value < 1 {
			return fmt.Errorf("%s must be positive for %q", field.name, profile.Name)
		}
		if !field.positive && field.value < 0 {
			return fmt.Errorf("%s must not be negative for %q", field.name, profile.Name)
		}
	}
	if err := validatePortGroups(profile); err != nil {
		return err
	}
	if err := validateOneBasedStrings("port_names", profile.Name, profile.Ports, profile.PortNames); err != nil {
		return err
	}
	if err := validatePortRoles(profile); err != nil {
		return err
	}
	if err := validateOneBasedStrings("port_network_groups", profile.Name, profile.Ports, profile.PortNetworkGroups); err != nil {
		return err
	}
	if err := validatePayload(profile); err != nil {
		return err
	}
	return nil
}

// validatePortGroups ensures grouped hardware layouts exactly cover the profile
// port count and declare at most one profile-defined uplink group.
func validatePortGroups(profile Profile) error {
	total := 0
	uplinkGroups := 0
	for index, group := range profile.PortGroups {
		if group.Count < 1 {
			return fmt.Errorf("port_groups[%d].count must be positive for %q", index, profile.Name)
		}
		if group.Speed < 0 {
			return fmt.Errorf("port_groups[%d].speed must not be negative for %q", index, profile.Name)
		}
		if group.Uplink {
			uplinkGroups++
		}
		total += group.Count
	}
	if len(profile.PortGroups) > 0 && total != profile.Ports {
		return fmt.Errorf("port_groups total %d != ports %d", total, profile.Ports)
	}
	if uplinkGroups > 1 {
		return fmt.Errorf("only one port_groups entry may set uplink for %q", profile.Name)
	}
	return nil
}

// validateOneBasedStrings checks profile arrays that map directly to one-based
// port indexes.
func validateOneBasedStrings(field, name string, ports int, values []string) error {
	if len(values) > ports {
		return fmt.Errorf("%s length %d exceeds ports %d for %q", field, len(values), ports, name)
	}
	for index, value := range values {
		if strings.ContainsAny(value, "\r\n\t") {
			return fmt.Errorf("%s[%d] contains unsupported whitespace for %q", field, index, name)
		}
	}
	return nil
}

// validatePortRoles keeps gateway role labels constrained to the renderer's
// known WAN/LAN role model.
func validatePortRoles(profile Profile) error {
	if err := validateOneBasedStrings("port_roles", profile.Name, profile.Ports, profile.PortRoles); err != nil {
		return err
	}
	for index, role := range profile.PortRoles {
		role = strings.ToLower(strings.TrimSpace(role))
		if role == "" {
			continue
		}
		switch role {
		case "wan", "lan", "wan2", "lan2":
		default:
			return fmt.Errorf("port_roles[%d] has invalid role %q; use wan, lan, wan2, or lan2", index, role)
		}
	}
	return nil
}

// validatePayload checks profile-driven renderer settings that are shared by
// switch and gateway payload generation.
func validatePayload(profile Profile) error {
	switch strings.ToLower(strings.TrimSpace(profile.Payload.Kind)) {
	case payloadKindSwitch, payloadKindGateway:
	default:
		return fmt.Errorf("payload.kind %q is invalid; use switch or gateway", profile.Payload.Kind)
	}
	if strings.TrimSpace(profile.Payload.ManagementInterface) == "" {
		return fmt.Errorf("payload.management_interface is required for %q", profile.Name)
	}
	if strings.Contains(profile.Payload.ManagementInterface, "/") {
		return fmt.Errorf("payload.management_interface %q is invalid for %q", profile.Payload.ManagementInterface, profile.Name)
	}
	if strings.TrimSpace(profile.Payload.GatewayInterfacePrefix) == "" {
		return fmt.Errorf("payload.gateway_interface_prefix is required for %q", profile.Name)
	}
	if strings.ContainsAny(profile.Payload.GatewayInterfacePrefix, "/ \t\r\n") {
		return fmt.Errorf("payload.gateway_interface_prefix %q is invalid for %q", profile.Payload.GatewayInterfacePrefix, profile.Name)
	}
	return nil
}

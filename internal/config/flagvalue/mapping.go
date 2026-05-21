// Package flagvalue contains flag.Value adapters for runtime config structures.
package flagvalue

// Repeated mapping flags accept compact operator syntax and produce typed config
// records for later validation.

import (
	"fmt"
	"strconv"
	"strings"

	appconfig "github.com/konstruktor1/unifi-stubd/internal/config"
)

// StringList parses repeated string flags into a trimmed list.
type StringList []string

func (f *StringList) String() string {
	if f == nil {
		return ""
	}
	return strings.Join(*f, ",")
}

// Set appends one non-empty string value.
func (f *StringList) Set(value string) error {
	value = strings.TrimSpace(value)
	if value == "" || strings.Contains(value, "/") {
		return fmt.Errorf("invalid value %q", value)
	}
	*f = append(*f, value)
	return nil
}

// BridgeMemberPortMap parses repeated bridge member pin flags.
type BridgeMemberPortMap []appconfig.BridgeMemberPortMap

func (f *BridgeMemberPortMap) String() string {
	if f == nil {
		return ""
	}
	values := make([]string, 0, len(*f))
	for _, mapping := range *f {
		values = append(values, fmt.Sprintf("%s=%d", mapping.Member, mapping.Port))
	}
	return strings.Join(values, ",")
}

// Set appends one bridge member to port mapping.
func (f *BridgeMemberPortMap) Set(value string) error {
	member, portText, ok := strings.Cut(strings.TrimSpace(value), "=")
	if !ok {
		return fmt.Errorf("invalid bridge member port map %q; use member=PORT", value)
	}
	port, err := strconv.Atoi(strings.TrimSpace(portText))
	if err != nil || port < 1 {
		return fmt.Errorf("invalid bridge member port %q; use a positive port", portText)
	}
	member = strings.TrimSpace(member)
	if member == "" || strings.Contains(member, "/") {
		return fmt.Errorf("invalid bridge member %q", member)
	}
	*f = append(*f, appconfig.BridgeMemberPortMap{Member: member, Port: port})
	return nil
}

// PortMapping parses repeated port source mapping flags.
type PortMapping []appconfig.PortMapping

func (f *PortMapping) String() string {
	if f == nil {
		return ""
	}
	values := make([]string, 0, len(*f))
	for _, mapping := range *f {
		values = append(values, fmt.Sprintf("port=%d,interface=%s,disabled=%t,unmapped=%t",
			mapping.Port,
			mapping.Interface,
			mapping.Disabled,
			mapping.Unmapped,
		))
	}
	return strings.Join(values, ";")
}

// Set appends one explicit port mapping.
func (f *PortMapping) Set(value string) error {
	fields := parseCommaKeyValues(value)
	port, err := strconv.Atoi(fields["port"])
	if err != nil || port < 1 {
		return fmt.Errorf("invalid port mapping %q; port must be positive", value)
	}
	mapping := appconfig.PortMapping{
		Port:      port,
		Interface: strings.TrimSpace(fields["interface"]),
		Disabled:  parseBoolField(fields["disabled"]),
		Unmapped:  parseBoolField(fields["unmapped"]),
	}
	*f = append(*f, mapping)
	return nil
}

func parseCommaKeyValues(value string) map[string]string {
	out := map[string]string{}
	for _, field := range strings.Split(value, ",") {
		key, fieldValue, ok := strings.Cut(strings.TrimSpace(field), "=")
		if !ok {
			continue
		}
		out[strings.ToLower(strings.TrimSpace(key))] = strings.TrimSpace(fieldValue)
	}
	return out
}

func parseBoolField(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

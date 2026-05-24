// Package device loads and stores embedded device profile data.
package device

// Embedded profile config decoding turns checked-in YAML documents into
// validated built-in profiles during init registration.

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Profile schema defaults describe the current YAML version and renderer
// fallback values.
const (
	schemaVersion          = 1
	payloadKindSwitch      = "switch"
	payloadKindGateway     = "gateway"
	defaultRequiredVersion = "5.0.0"
	defaultMgmtInterface   = "eth0"
	defaultGatewayPrefix   = "eth"
)

// LoadPath loads one profile YAML file or all profile YAML files in a directory.
func (r *ProfileRegistry) LoadPath(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return &ProfilePathError{Path: path, Kind: ErrorKindIO, Err: err}
	}
	if !info.IsDir() {
		return r.LoadFile(path)
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return &ProfilePathError{Path: path, Kind: ErrorKindIO, Err: err}
	}
	var paths []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if ext == ".yaml" || ext == ".yml" {
			paths = append(paths, filepath.Join(path, name))
		}
	}
	sort.Strings(paths)
	for _, profilePath := range paths {
		if err := r.LoadFile(profilePath); err != nil {
			return err
		}
	}
	return nil
}

// LoadFile loads one external profile YAML file into r.
func (r *ProfileRegistry) LoadFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return &ProfilePathError{Path: path, Kind: ErrorKindIO, Err: err}
	}
	decoded, err := r.decodeExternalConfigRecord(data)
	if err != nil {
		return &ProfilePathError{Path: path, Kind: classifyProfileError(err), Err: err}
	}
	if err := r.register(path, decoded.order, decoded.profile, decoded.document, false); err != nil {
		return &ProfilePathError{Path: path, Kind: ErrorKindValidation, Err: err}
	}
	return nil
}

// DecodeConfig decodes and validates one standalone profile YAML document.
func DecodeConfig(data []byte) (Profile, int, error) {
	decoded, err := decodeConfigRecord(data)
	if err != nil {
		return Profile{}, 0, err
	}
	return decoded.profile, decoded.order, nil
}

// DecodeExternalConfig decodes a profile YAML document with optional inheritance.
func (r *ProfileRegistry) DecodeExternalConfig(data []byte) (Profile, int, error) {
	decoded, err := r.decodeExternalConfigRecord(data)
	if err != nil {
		return Profile{}, 0, err
	}
	return decoded.profile, decoded.order, nil
}

// decodedProfile carries both the typed profile and original YAML document
// needed for inheritance and registry storage.
type decodedProfile struct {
	profile  Profile
	order    int
	document *yaml.Node
}

// decodeConfigRecord handles built-in profiles, which must be standalone YAML
// documents with defaults and validation applied during init registration.
func decodeConfigRecord(data []byte) (decodedProfile, error) {
	document, err := decodeProfileDocument(data)
	if err != nil {
		return decodedProfile{}, err
	}
	profile, err := decodeProfileYAML(document)
	if err != nil {
		return decodedProfile{}, err
	}
	applyProfileDefaults(&profile)
	if err := validateProfile(profile); err != nil {
		return decodedProfile{}, err
	}
	return decodedProfile{profile: profile, order: profile.Order, document: document}, nil
}

// decodeExternalConfigRecord applies optional YAML-level inheritance before the
// final strict typed decode, preserving explicit zero-value overrides.
func (r *ProfileRegistry) decodeExternalConfigRecord(data []byte) (decodedProfile, error) {
	document, err := decodeProfileDocument(data)
	if err != nil {
		return decodedProfile{}, err
	}
	profile, err := decodeProfileYAML(document)
	if err != nil {
		return decodedProfile{}, err
	}
	order := profile.Order
	mergedDocument := document
	if strings.TrimSpace(profile.Extends) != "" {
		base, ok := r.lookupRecord(profile.Extends)
		if !ok {
			return decodedProfile{}, fmt.Errorf("extends %q not found", profile.Extends)
		}
		mergedDocument = mergeProfileDocuments(base.document, document)
		profile, err = decodeProfileYAML(mergedDocument)
		if err != nil {
			return decodedProfile{}, err
		}
	}
	applyProfileDefaults(&profile)
	if err := validateProfile(profile); err != nil {
		return decodedProfile{}, err
	}
	return decodedProfile{profile: profile, order: order, document: mergedDocument}, nil
}

// decodeProfileDocument keeps a YAML node tree so external profiles can merge
// inheritance before strict typed decoding.
func decodeProfileDocument(data []byte) (*yaml.Node, error) {
	var doc yaml.Node
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&doc); err != nil {
		return nil, fmt.Errorf("decode profile YAML: %w", err)
	}
	if yamlDocumentContent(&doc) == nil {
		return nil, fmt.Errorf("decode profile YAML: empty document")
	}
	return cloneYAMLNode(&doc), nil
}

// decodeProfileYAML performs the strict KnownFields decode that rejects
// misspelled profile keys.
func decodeProfileYAML(document *yaml.Node) (Profile, error) {
	var profile Profile
	data, err := yaml.Marshal(yamlDocumentContent(document))
	if err != nil {
		return Profile{}, fmt.Errorf("decode profile YAML: marshal merged YAML: %w", err)
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&profile); err != nil {
		return Profile{}, fmt.Errorf("decode profile YAML: %w", err)
	}
	normalizeProfile(&profile)
	return profile, nil
}

// normalizeProfile delegates model-level normalization shared by built-in and
// external profiles.
func normalizeProfile(profile *Profile) {
	NormalizeProfile(profile)
}

// applyProfileDefaults fills renderer metadata that older or minimal profiles
// may omit, without changing fields explicitly set in YAML.
func applyProfileDefaults(profile *Profile) {
	setDefaultInt(&profile.SchemaVersion, schemaVersion)
	for _, field := range []struct {
		target *string
		value  string
	}{
		{target: &profile.Stability, value: "tested"},
		{target: &profile.Payload.Kind, value: defaultPayloadKind(profile.DeviceType)},
		{target: &profile.Payload.RequiredVersion, value: defaultRequiredVersion},
		{target: &profile.Payload.ManagementInterface, value: defaultMgmtInterface},
		{target: &profile.Payload.GatewayInterfacePrefix, value: defaultGatewayPrefix},
	} {
		setDefaultString(field.target, field.value)
	}
}

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

// defaultPayloadKind selects gateway-shaped payloads only for gateway device
// families; switches remain the conservative default.
func defaultPayloadKind(deviceType string) string {
	switch strings.TrimSpace(deviceType) {
	case "ugw", "uxg", "udm":
		return payloadKindGateway
	default:
		return payloadKindSwitch
	}
}

// setDefaultString fills profile defaults after decode without overwriting YAML
// values.
func setDefaultString(target *string, value string) {
	if *target == "" {
		*target = value
	}
}

// setDefaultInt fills numeric profile defaults after decode without overwriting
// YAML values.
func setDefaultInt(target *int, value int) {
	if *target == 0 {
		*target = value
	}
}

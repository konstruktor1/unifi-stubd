// Package profiledata combines built-in and external profile records and
// enforces duplicate and override rules before profiles reach CLI validation or
// payload rendering.
package profiledata

import (
	"fmt"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const defaultDeviceType = "usw"

const (
	// ErrorKindIO marks filesystem access errors.
	ErrorKindIO = "io"
	// ErrorKindParse marks YAML decoding errors.
	ErrorKindParse = "parse"
	// ErrorKindValidation marks semantic profile validation errors.
	ErrorKindValidation = "validation"
	sourceTypeBuiltIn   = "built-in"
	sourceTypeExternal  = "external"
	stabilityExternal   = "external"
)

// PathError wraps a profile path error with a broad error class for CLI exit codes.
type PathError struct {
	// Path is the profile file or directory path.
	Path string
	// Kind is one of ErrorKindIO, ErrorKindParse, or ErrorKindValidation.
	Kind string
	// Err is the underlying error.
	Err error
}

func (e *PathError) Error() string {
	if e.Path == "" {
		return e.Err.Error()
	}
	return fmt.Sprintf("%s: %v", e.Path, e.Err)
}

func (e *PathError) Unwrap() error {
	return e.Err
}

type record struct {
	source   string
	builtin  bool
	order    int
	profile  Profile
	document *yaml.Node
}

var registry []record

// Registry contains built-in and caller-loaded profile records.
type Registry struct {
	records []record
}

// Register adds one decoded profile to the global built-in profile registry.
func Register(source string, order int, profile Profile, document *yaml.Node) {
	if err := registerRecord(&registry, source, order, profile, document, true); err != nil {
		panic(err)
	}
}

// BuiltinRegistry returns a new registry initialized with built-in profiles.
func BuiltinRegistry() Registry {
	return Registry{records: cloneRecords(registry)}
}

// Register adds one profile to r.
func (r *Registry) register(source string, order int, profile Profile, document *yaml.Node, builtin bool) error {
	return registerRecord(&r.records, source, order, profile, document, builtin)
}

func registerRecord(records *[]record, source string, order int, profile Profile, document *yaml.Node, builtin bool) error {
	for index, record := range *records {
		nameDuplicate := record.profile.Name == profile.Name
		modelDuplicate := strings.EqualFold(record.profile.Model, profile.Model)
		if !nameDuplicate && !modelDuplicate {
			continue
		}
		if profile.AllowBuiltinOverride && record.builtin {
			(*records)[index] = recordEntry(source, order, profile, document, builtin)
			return nil
		}
		switch {
		case nameDuplicate:
			return fmt.Errorf("duplicate profile name %q in %s and %s", profile.Name, record.source, source)
		default:
			return fmt.Errorf("duplicate profile model %q in %s and %s", profile.Model, record.source, source)
		}
	}
	*records = append(*records, recordEntry(source, order, profile, document, builtin))
	return nil
}

func recordEntry(source string, order int, profile Profile, document *yaml.Node, builtin bool) record {
	return record{
		source:   source,
		builtin:  builtin,
		order:    order,
		profile:  cloneProfile(profile),
		document: cloneYAMLNode(document),
	}
}

// Profiles returns a copy of the built-in device profiles.
func Profiles() []Profile {
	return BuiltinRegistry().Profiles()
}

// Profiles returns a copy of all profiles in r.
func (r Registry) Profiles() []Profile {
	records := cloneRecords(r.records)
	sort.SliceStable(records, func(i, j int) bool {
		if records[i].order != records[j].order {
			return records[i].order < records[j].order
		}
		return records[i].profile.Name < records[j].profile.Name
	})
	out := make([]Profile, 0, len(records))
	for _, record := range records {
		out = append(out, profileWithSource(record))
	}
	return out
}

// Lookup returns a built-in profile by profile name or model identifier.
func Lookup(name string) (Profile, bool) {
	return BuiltinRegistry().Lookup(name)
}

// Lookup returns a profile by profile name or model identifier.
func (r Registry) Lookup(name string) (Profile, bool) {
	record, ok := r.lookupRecord(name)
	if !ok {
		return Profile{}, false
	}
	return profileWithSource(record), true
}

func (r Registry) lookupRecord(name string) (record, bool) {
	name = strings.ToLower(strings.TrimSpace(name))
	for _, record := range r.records {
		profile := record.profile
		if strings.ToLower(profile.Name) == name || strings.ToLower(profile.Model) == name {
			return recordEntry(record.source, record.order, profile, record.document, record.builtin), true
		}
	}
	return record{}, false
}

// Names returns the known profile names as a comma-separated list.
func Names() string {
	return BuiltinRegistry().Names()
}

// Names returns the known profile names as a comma-separated list.
func (r Registry) Names() string {
	profiles := r.Profiles()
	names := make([]string, 0, len(profiles))
	for _, profile := range profiles {
		names = append(names, profile.Name)
	}
	return strings.Join(names, ", ")
}

// Format returns a human-readable table of built-in profiles.
func Format() string {
	return BuiltinRegistry().Format()
}

// Format returns a human-readable table of profiles.
func (r Registry) Format() string {
	var b strings.Builder
	for _, profile := range r.Profiles() {
		recommended := ""
		if profile.Recommended {
			recommended = " recommended"
		}
		fmt.Fprintf(&b, "%-15s %-6s %-15s kind=%-7s source=%-8s stability=%-12s ports=%-2d speed=%-5d version=%s%s  %s\n",
			profile.Name,
			deviceTypeOrDefault(profile.DeviceType),
			profile.Model,
			profile.Payload.Kind,
			profile.SourceType,
			profile.Stability,
			profile.Ports,
			firstNonZero(profile.PortSpeed, 1000),
			profile.Version,
			recommended,
			profile.Description,
		)
	}
	return b.String()
}

func cloneProfile(profile Profile) Profile {
	profile.PortGroups = clonePortGroups(profile.PortGroups)
	profile.PortNames = cloneStrings(profile.PortNames)
	profile.PortRoles = cloneStrings(profile.PortRoles)
	profile.PortNetworkGroups = cloneStrings(profile.PortNetworkGroups)
	profile.ValidatedControllerVersions = cloneStrings(profile.ValidatedControllerVersions)
	return profile
}

func profileWithSource(record record) Profile {
	profile := cloneProfile(record.profile)
	profile.Source = record.source
	if record.builtin {
		profile.SourceType = sourceTypeBuiltIn
	} else {
		profile.SourceType = sourceTypeExternal
	}
	return profile
}

func cloneRecords(records []record) []record {
	if len(records) == 0 {
		return nil
	}
	out := make([]record, len(records))
	for index, record := range records {
		out[index] = record
		out[index].profile = cloneProfile(record.profile)
		out[index].document = cloneYAMLNode(record.document)
	}
	return out
}

func clonePortGroups(groups []PortGroup) []PortGroup {
	return cloneNonEmptySlice(groups)
}

func cloneStrings(values []string) []string {
	return cloneNonEmptySlice(values)
}

func cloneNonEmptySlice[T any](values []T) []T {
	if len(values) == 0 {
		return nil
	}
	out := make([]T, len(values))
	copy(out, values)
	return out
}

func deviceTypeOrDefault(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return defaultDeviceType
	}
	return value
}

func firstNonZero(values ...int) int {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}

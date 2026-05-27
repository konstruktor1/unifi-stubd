// Package device combines built-in and external profile records and
// enforces duplicate and override rules before profiles reach CLI validation or
// payload rendering.
package device

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// defaultDeviceType keeps legacy profiles switch-shaped when device_type is
// omitted.
const defaultDeviceType = "usw"

// Registry metadata constants separate user-facing error classes from internal
// source labels.
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

// ProfilePathError wraps a profile path error with a broad error class for CLI exit codes.
type ProfilePathError struct {
	// Path is the profile file or directory path.
	Path string
	// Kind is one of ErrorKindIO, ErrorKindParse, or ErrorKindValidation.
	Kind string
	// Err is the underlying error.
	Err error
}

// Error includes the profile path when one is available for CLI diagnostics.
func (e *ProfilePathError) Error() string {
	if e.Path == "" {
		return e.Err.Error()
	}
	return fmt.Sprintf("%s: %v", e.Path, e.Err)
}

// Unwrap exposes the underlying profile loading error.
func (e *ProfilePathError) Unwrap() error {
	return e.Err
}

// record stores one registered profile plus the YAML document used for external
// inheritance.
type record struct {
	source   string
	builtin  bool
	order    int
	profile  Profile
	document *yaml.Node
}

// ProfileRegistry contains built-in and caller-loaded profile records.
type ProfileRegistry struct {
	records []record
}

// NewProfileRegistry returns a registry initialized with built-in profiles.
func NewProfileRegistry() ProfileRegistry {
	return BuiltinRegistry()
}

// BuiltinRegistry returns a new registry initialized with built-in profiles.
func BuiltinRegistry() ProfileRegistry {
	return ProfileRegistry{records: cloneRecords(builtinProfileRecords())}
}

// LoadProfilePath loads one external profile file or profile directory.
func (r *ProfileRegistry) LoadProfilePath(path string) error {
	return r.LoadPath(path)
}

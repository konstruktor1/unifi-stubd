// Package config defines the user-facing YAML schema shared by packaged config
// files, CLI validation, and runtime startup. Only non-zero runtime defaults
// are set here; package-specific examples live in packaging and lab configs.
package config

// DefaultPath is the system-wide YAML configuration path.
const DefaultPath = "/etc/unifi-stubd/config.yaml"

// YAML sentinel values mirror CLI defaults for automatic identity resolution
// and disabled optional sources.
const (
	automaticValue = "auto"
	sourceOffValue = "off"
)

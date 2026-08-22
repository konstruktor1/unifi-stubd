package config

// DefaultPath is the system-wide YAML configuration path.
const DefaultPath = "/etc/unifi-stubd/config.yaml"

// YAML sentinel values mirror CLI defaults for automatic identity resolution
// and disabled optional sources.
const (
	automaticValue = "auto"
	sourceOffValue = "off"
)

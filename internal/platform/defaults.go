package platform

import (
	"strings"
	"time"
)

// Default platform settings keep optional host integrations bounded and
// predictable when config omits them.
const defaultCommandTimeout = 2 * time.Second

// normalizeConfig applies platform defaults for optional read-only sources.
func normalizeConfig(cfg Config) Config {
	cfg.LLDPSource = normalizedSource(cfg.LLDPSource)
	cfg.ProcSource = normalizedSource(cfg.ProcSource)
	if cfg.CommandTimeout <= 0 {
		cfg.CommandTimeout = defaultCommandTimeout
	}
	return cfg
}

// normalizedSource treats an empty optional integration as explicitly off.
func normalizedSource(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return SourceOff
	}
	return value
}

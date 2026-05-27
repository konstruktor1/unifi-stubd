package platform

import (
	"strings"
	"time"
)

// Default platform settings keep optional host integrations bounded and
// predictable when config omits them.
const (
	defaultCommandTimeout = 2 * time.Second
	defaultSyslogPath     = "/var/log/messages"
	defaultJournalUnit    = "unifi-stubd"
)

// normalizeConfig applies platform defaults for optional read-only sources.
func normalizeConfig(cfg Config) Config {
	cfg.LLDPSource = normalizedSource(cfg.LLDPSource)
	cfg.TrafficSource = normalizedSource(cfg.TrafficSource)
	cfg.LogSource = normalizedSource(cfg.LogSource)
	cfg.ProcSource = normalizedSource(cfg.ProcSource)
	cfg.DBusBus = normalizedDBusBus(cfg.DBusBus)
	if strings.TrimSpace(cfg.SyslogPath) == "" {
		cfg.SyslogPath = defaultSyslogPath
	}
	if strings.TrimSpace(cfg.JournalUnit) == "" {
		cfg.JournalUnit = defaultJournalUnit
	}
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

// normalizedDBusBus defaults to the system bus because service deployments do
// not usually have a session bus.
func normalizedDBusBus(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case DBusBusSession:
		return DBusBusSession
	default:
		return DBusBusSystem
	}
}

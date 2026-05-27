package wanhealth

import (
	"context"
	"time"
)

// Config describes one active WAN health measurement batch.
type Config struct {
	// Source selects off, static, or ping.
	Source string
	// Interval is the configured probe interval.
	Interval time.Duration
	// Timeout is the per-target timeout.
	Timeout time.Duration
	// Targets lists the represented WAN ports to probe.
	Targets []Target
}

// Target maps one probe host to one represented WAN port.
type Target struct {
	// Port is the one-based UniFi port index.
	Port int
	// Host is the IP or hostname passed to ping.
	Host string
}

// Result is one sanitized WAN health sample.
type Result struct {
	// Port is the one-based UniFi port index.
	Port int
	// Host is the probed host.
	Host string
	// Connected reports whether ping succeeded.
	Connected bool
	// LatencyMS is the measured or approximated latency.
	LatencyMS int
	// DowntimeSeconds is the stateless downtime approximation for a failed sample.
	DowntimeSeconds int
	// UptimePercent is the stateless availability sample.
	UptimePercent float64
	// LastError describes the failed sample without raw command output.
	LastError string
}

// Runner executes one probe. Tests can inject a fake runner.
type Runner interface {
	Ping(ctx context.Context, host string, timeout time.Duration) ([]byte, time.Duration, error)
}

// CommandRunner executes the local OS ping command.
type CommandRunner struct{}

// Package wanhealth performs opt-in, read-only gateway WAN health probes.
package wanhealth

import (
	"context"
	"strings"
	"time"
)

// Supported WAN health source values.
const (
	SourceOff    = "off"
	SourceStatic = "static"
	SourcePing   = "ping"
)

const (
	defaultInterval = 10 * time.Second
	defaultTimeout  = time.Second
	maxErrorLength  = 180
)

// NormalizeSource normalizes a WAN health source value.
func NormalizeSource(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return SourceOff
	}
	return value
}

// Measure executes active WAN health probes using the local OS ping command.
func Measure(ctx context.Context, cfg Config) []Result {
	return MeasureWithRunner(ctx, cfg, CommandRunner{})
}

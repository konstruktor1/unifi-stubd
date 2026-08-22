package wanhealth

import (
	"context"
	"strings"
	"time"
)

// Measure executes active WAN health probes using the local OS ping command.
func Measure(ctx context.Context, cfg Config) []Result {
	if NormalizeSource(cfg.Source) != SourcePing || len(cfg.Targets) == 0 {
		return nil
	}
	timeout := positiveDuration(cfg.Timeout, defaultTimeout)
	interval := positiveDuration(cfg.Interval, defaultInterval)
	out := make([]Result, 0, len(cfg.Targets))
	for _, target := range cfg.Targets {
		host := strings.TrimSpace(target.Host)
		result := Result{Port: target.Port, Host: host}
		if target.Port < 1 {
			result.LastError = "invalid port"
			result.DowntimeSeconds = downtimeSeconds(interval, timeout)
			out = append(out, result)
			continue
		}
		if host == "" {
			result.LastError = "host is required"
			result.DowntimeSeconds = downtimeSeconds(interval, timeout)
			out = append(out, result)
			continue
		}
		targetCtx, cancel := context.WithTimeout(ctx, timeout)
		output, elapsed, err := ping(targetCtx, host)
		cancel()
		if err != nil {
			result.LastError = sanitizeError(err)
			result.DowntimeSeconds = downtimeSeconds(interval, timeout)
			out = append(out, result)
			continue
		}
		latency, ok := parseLatencyMS(output)
		if !ok {
			latency = durationMS(elapsed)
		}
		result.Connected = true
		result.LatencyMS = latency
		result.UptimePercent = 100
		out = append(out, result)
	}
	return out
}

func positiveDuration(value, fallback time.Duration) time.Duration {
	if value > 0 {
		return value
	}
	return fallback
}

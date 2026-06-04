package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/device"
	"github.com/konstruktor1/unifi-stubd/internal/observe/wanhealth"
)

func buildStatusWANHealth(flags runtimeFlags, profile device.Profile) statusWANHealthConfig {
	cfg := flags.wanHealth
	out := statusWANHealthConfig{
		Source:          wanhealth.NormalizeSource(cfg.Source),
		IntervalSeconds: cfg.IntervalSeconds,
		TimeoutMS:       cfg.TimeoutMS,
	}
	for _, target := range cfg.Targets {
		out.Targets = append(out.Targets, statusWANHealthTarget{
			Port: target.Port,
			Host: strings.TrimSpace(target.Host),
		})
	}
	if out.Source != wanhealth.SourcePing || !isGatewayProfile(profile) {
		return out
	}
	ports := healthPreviewPorts(flags, profile)
	targets := probeTargets(cfg, ports, explicitWANPorts(flags.portOverrides))
	if len(targets) == 0 {
		return out
	}
	results := wanhealth.Measure(context.Background(), probeConfig(cfg, targets))
	for _, result := range results {
		out.Results = append(out.Results, statusWANHealthResult{
			Port:            result.Port,
			Host:            result.Host,
			Connected:       result.Connected,
			LatencyMS:       result.LatencyMS,
			DowntimeSeconds: result.DowntimeSeconds,
			UptimePercent:   result.UptimePercent,
			LastError:       result.LastError,
		})
	}
	return out
}

func printWANHealthStatus(status statusWANHealthConfig) {
	fmt.Printf("wan_health_source: %s\n", status.Source)
	if status.IntervalSeconds > 0 {
		fmt.Printf("wan_health_interval_seconds: %d\n", status.IntervalSeconds)
	}
	if status.TimeoutMS > 0 {
		fmt.Printf("wan_health_timeout_ms: %d\n", status.TimeoutMS)
	}
	for _, target := range status.Targets {
		fmt.Printf("wan_health_target: port=%d host=%s\n", target.Port, valueOrDash(target.Host))
	}
	for _, result := range status.Results {
		fmt.Printf("wan_health_result: port=%d host=%s connected=%t latency_ms=%d downtime_seconds=%d uptime_percent=%.1f last_error=%s\n",
			result.Port,
			valueOrDash(result.Host),
			result.Connected,
			result.LatencyMS,
			result.DowntimeSeconds,
			result.UptimePercent,
			valueOrDash(result.LastError),
		)
	}
}

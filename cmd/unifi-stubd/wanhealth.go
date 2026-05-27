package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	appconfig "github.com/konstruktor1/unifi-stubd/internal/config"
	"github.com/konstruktor1/unifi-stubd/internal/device"
	"github.com/konstruktor1/unifi-stubd/internal/observe/wanhealth"
)

// validateWANHealthConfig checks the YAML-only WAN health surface before any
// active probes can run.
func validateWANHealthConfig(flags runtimeFlags) error {
	cfg := flags.wanHealth
	source := wanhealth.NormalizeSource(cfg.Source)
	switch source {
	case wanhealth.SourceOff, wanhealth.SourceStatic, wanhealth.SourcePing:
	default:
		return fmt.Errorf("invalid wan_health.source %q; use off, static, or ping", cfg.Source)
	}
	if cfg.IntervalSeconds <= 0 {
		return fmt.Errorf("invalid wan_health.interval_seconds %d; use a positive seconds value", cfg.IntervalSeconds)
	}
	if cfg.TimeoutMS <= 0 {
		return fmt.Errorf("invalid wan_health.timeout_ms %d; use a positive millisecond value", cfg.TimeoutMS)
	}
	if source == wanhealth.SourcePing && len(cfg.Targets) == 0 {
		return fmt.Errorf("wan_health.targets is required when wan_health.source is ping")
	}
	for _, target := range cfg.Targets {
		if target.Port < 1 || target.Port > flags.portCount {
			return fmt.Errorf("invalid wan_health target port %d; use 1..%d", target.Port, flags.portCount)
		}
		if strings.TrimSpace(target.Host) == "" {
			return fmt.Errorf("invalid wan_health target host on port %d; host is required", target.Port)
		}
	}
	return nil
}

// normalizeWANHealthConfig stores the canonical source spelling on runtime
// flags. Probe timings are kept as config integers so status can echo them.
func normalizeWANHealthConfig(flags *runtimeFlags) {
	flags.wanHealth.Source = wanhealth.NormalizeSource(flags.wanHealth.Source)
}

// validateWANHealthTargetRoles verifies active probes only target gateway WAN
// ports after profile defaults and operator port overrides have been applied.
func validateWANHealthTargetRoles(flags runtimeFlags, profile device.Profile) error {
	if wanhealth.NormalizeSource(flags.wanHealth.Source) != wanhealth.SourcePing {
		return nil
	}
	if !gatewayProfile(profile) {
		return fmt.Errorf("wan_health.source ping requires a gateway profile")
	}
	ports := wanHealthPreviewPorts(flags, profile)
	for _, target := range flags.wanHealth.Targets {
		if target.Port < 1 || target.Port > len(ports) {
			continue
		}
		port := ports[target.Port-1]
		if !wanHealthPort(port) {
			return fmt.Errorf("wan_health target port %d has role %q; use a port with role wan or wan2", target.Port, valueOrDash(port.Role))
		}
	}
	return nil
}

// applyWANHealth overlays active probe results onto the already resolved port
// list. Only WAN health fields are set by the generated overrides.
func applyWANHealth(ports []device.Port, flags runtimeFlags, profile device.Profile) []device.Port {
	if wanhealth.NormalizeSource(flags.wanHealth.Source) != wanhealth.SourcePing || !gatewayProfile(profile) {
		return ports
	}
	targets := wanHealthRuntimeTargets(flags.wanHealth, ports)
	if len(targets) == 0 {
		return ports
	}
	cfg := wanHealthRuntimeConfig(flags.wanHealth, targets)
	results := wanhealth.Measure(context.Background(), cfg)
	for _, result := range results {
		if result.LastError != "" {
			log.Printf("wan_health ping warning: port=%d host=%s error=%s", result.Port, result.Host, result.LastError)
		}
	}
	// Reuse the normal override path so active samples can only touch the same
	// WAN telemetry fields as static YAML hints. Role, VLAN, assignment IDs,
	// addresses, and source interfaces are intentionally left untouched.
	return device.ApplyPortOverrides(ports, device.PortOverridesFromWANHealthResults(deviceWANHealthResults(results)))
}

func wanHealthRuntimeConfig(cfg appconfig.WANHealthConfig, targets []wanhealth.Target) wanhealth.Config {
	return wanhealth.Config{
		Source:   wanhealth.NormalizeSource(cfg.Source),
		Interval: time.Duration(cfg.IntervalSeconds) * time.Second,
		Timeout:  time.Duration(cfg.TimeoutMS) * time.Millisecond,
		Targets:  targets,
	}
}

func wanHealthRuntimeTargets(cfg appconfig.WANHealthConfig, ports []device.Port) []wanhealth.Target {
	out := make([]wanhealth.Target, 0, len(cfg.Targets))
	for _, target := range cfg.Targets {
		if target.Port < 1 || target.Port > len(ports) {
			continue
		}
		port := ports[target.Port-1]
		if !wanHealthPort(port) {
			log.Printf("wan_health target skipped: port=%d role=%q is not wan or wan2", target.Port, port.Role)
			continue
		}
		out = append(out, wanhealth.Target{
			Port: target.Port,
			Host: strings.TrimSpace(target.Host),
		})
	}
	return out
}

func deviceWANHealthResults(results []wanhealth.Result) []device.WANHealthResult {
	out := make([]device.WANHealthResult, 0, len(results))
	for _, result := range results {
		out = append(out, device.WANHealthResult{
			Port:            result.Port,
			Connected:       result.Connected,
			LatencyMS:       result.LatencyMS,
			DowntimeSeconds: result.DowntimeSeconds,
			UptimePercent:   result.UptimePercent,
		})
	}
	return out
}

func wanHealthPreviewPorts(flags runtimeFlags, profile device.Profile) []device.Port {
	ports := device.BuildPorts(profile, device.PortBuildOptions{
		Count:      flags.portCount,
		LinkSpeed:  flags.linkSpeed,
		UplinkPort: effectiveUplinkPort(profile, flags),
	})
	return device.ApplyPortOverrides(ports, flags.portOverrides)
}

func gatewayProfile(profile device.Profile) bool {
	return strings.EqualFold(strings.TrimSpace(profile.Payload.Kind), "gateway")
}

func wanHealthPort(port device.Port) bool {
	switch strings.ToLower(strings.TrimSpace(port.Role)) {
	case "wan", "wan2":
		return true
	default:
		return false
	}
}

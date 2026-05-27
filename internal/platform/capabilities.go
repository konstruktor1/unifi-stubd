package platform

import (
	"context"
	"os"
	"os/exec"
)

// Capabilities reports which optional platform sources are disabled, available,
// missing, or unsupported for status output.
func (p hostPlatform) Capabilities(ctx context.Context, cfg Config) CapabilityReport {
	cfg = normalizeConfig(cfg)
	report := CapabilityReport{GOOS: p.goos}
	// Capability reporting is diagnostic only. Missing optional tools become
	// status output, not install attempts or host changes.
	report.Capabilities = append(report.Capabilities,
		p.commandCapability(capabilityLLDP, cfg.LLDPSource, "lldpcli"),
		p.logCapability(cfg),
		p.procCapability(cfg),
		p.dbusCapability(ctx, cfg),
		Capability{Name: capabilityTraffic, Source: cfg.TrafficSource, State: trafficState(cfg.TrafficSource), Detail: trafficDetail(cfg.TrafficSource)},
	)
	return report
}

// commandCapability reports whether an optional command-backed source is
// usable, without installing or invoking it.
func (p hostPlatform) commandCapability(name, source, command string) Capability {
	source = normalizedSource(source)
	if source == SourceOff {
		return Capability{Name: name, Source: source, State: capabilityDisabled}
	}
	if _, err := exec.LookPath(command); err != nil {
		return Capability{Name: name, Source: source, State: capabilityMissing, Detail: err.Error()}
	}
	return Capability{Name: name, Source: source, State: capabilityAvailable}
}

// logCapability reports the configured log reader state for status output.
func (p hostPlatform) logCapability(cfg Config) Capability {
	source := normalizedSource(cfg.LogSource)
	if source == SourceOff {
		return Capability{Name: capabilityLogs, Source: source, State: capabilityDisabled}
	}
	switch source {
	case LogSourceJournalctl:
		return p.commandCapability(capabilityLogs, source, "journalctl")
	case LogSourceSyslog:
		if _, err := os.Stat(cfg.SyslogPath); err != nil {
			return Capability{Name: capabilityLogs, Source: source, State: capabilityMissing, Detail: err.Error()}
		}
		return Capability{Name: capabilityLogs, Source: source, State: capabilityAvailable, Detail: cfg.SyslogPath}
	default:
		return Capability{Name: capabilityLogs, Source: source, State: capabilityUnsupported}
	}
}

// procCapability reports whether Linux procfs counters can be used as optional
// fallback telemetry.
func (p hostPlatform) procCapability(cfg Config) Capability {
	source := normalizedSource(cfg.ProcSource)
	if source == SourceOff {
		return Capability{Name: capabilityProc, Source: source, State: capabilityDisabled}
	}
	if source != ProcSourceProcFS {
		return Capability{Name: capabilityProc, Source: source, State: capabilityUnsupported}
	}
	if p.goos != goosLinux {
		return Capability{Name: capabilityProc, Source: source, State: capabilityUnsupported, Detail: "procfs source is Linux-only"}
	}
	if _, err := os.Stat("/proc/net/dev"); err != nil {
		return Capability{Name: capabilityProc, Source: source, State: capabilityMissing, Detail: err.Error()}
	}
	return Capability{Name: capabilityProc, Source: source, State: capabilityAvailable}
}

// trafficState currently marks traffic metadata sources as disabled or
// unsupported; rate calculations use interface counters instead.
func trafficState(source string) string {
	if normalizedSource(source) == SourceOff {
		return capabilityDisabled
	}
	return capabilityUnsupported
}

// trafficDetail explains unsupported traffic metadata sources in status output.
func trafficDetail(source string) string {
	if normalizedSource(source) == SourceOff {
		return ""
	}
	return "traffic metadata sources are not implemented"
}

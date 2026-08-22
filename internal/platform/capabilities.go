package platform

import (
	"os"
	"os/exec"
)

// Capabilities reports which optional platform sources are disabled, available,
// missing, or unsupported for status output.
func (p hostPlatform) Capabilities(cfg Config) CapabilityReport {
	cfg = normalizeConfig(cfg)
	report := CapabilityReport{GOOS: p.goos}
	// Capability reporting is diagnostic only. Missing optional tools become
	// status output, not install attempts or host changes.
	report.Capabilities = append(report.Capabilities,
		p.commandCapability(capabilityLLDP, cfg.LLDPSource, "lldpcli"),
		p.procCapability(cfg),
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

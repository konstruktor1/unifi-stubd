// Package platform hides optional host integrations behind read-only adapter
// interfaces. It may inspect interfaces, bridges, logs, procfs, D-Bus, and
// LLDP, but it must not mutate host networking or execute controller-provided
// commands.
package platform

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/konstruktor1/unifi-stubd/internal/adapters/freebsdifconfig"
	"github.com/konstruktor1/unifi-stubd/internal/device"
	"github.com/konstruktor1/unifi-stubd/internal/observe"
	"github.com/konstruktor1/unifi-stubd/internal/observe/ifsource"
	"github.com/prometheus/procfs"
)

const (
	defaultCommandTimeout = 2 * time.Second
	defaultSyslogPath     = "/var/log/messages"
	defaultJournalUnit    = "unifi-stubd"
)

type hostPlatform struct {
	goos string
	cfg  Config
}

// New returns the default read-only platform adapter for the current host.
func New(cfg Config) Platform {
	cfg = normalizeConfig(cfg)
	return hostPlatform{goos: runtime.GOOS, cfg: cfg}
}

// NewForOS returns a read-only platform adapter for tests and fixtures.
func NewForOS(goos string, cfg Config) Platform {
	cfg = normalizeConfig(cfg)
	if strings.TrimSpace(goos) == "" {
		goos = runtime.GOOS
	}
	return hostPlatform{goos: strings.ToLower(strings.TrimSpace(goos)), cfg: cfg}
}

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

func normalizedSource(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return SourceOff
	}
	return value
}

func normalizedDBusBus(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case DBusBusSession:
		return DBusBusSession
	default:
		return DBusBusSystem
	}
}

func (p hostPlatform) Bridge(ctx context.Context, cfg observe.BridgeConfig) (observe.BridgeObservation, []error) {
	switch p.goos {
	case goosLinux:
		return p.linuxBridge(ctx, cfg)
	case goosFreeBSD:
		return p.freebsdBridge(ctx, cfg)
	default:
		return observe.BridgeObservation{
			Bridge:          strings.TrimSpace(cfg.Bridge),
			UplinkInterface: strings.TrimSpace(cfg.UplinkInterface),
			MemberPortMap:   cloneMemberPortMap(cfg.MemberPortMap),
		}, []error{fmt.Errorf("bridge observation is not implemented on %s", p.goos)}
	}
}

func (p hostPlatform) Ports(ctx context.Context, cfg observe.PortMapConfig) (observe.PortMapObservation, []error) {
	out := observe.PortMapObservation{Ports: map[int]observe.PortObservation{}}
	var errs []error
	for _, mapping := range cfg.Mappings {
		switch {
		case strings.TrimSpace(mapping.Interface) != "":
			observation, warnings := p.Interface(ctx, mapping.Interface)
			observation.Port = mapping.Port
			out.Ports[mapping.Port] = observation
			for _, warning := range warnings {
				errs = append(errs, fmt.Errorf("port %d interface %s: %w", mapping.Port, strings.TrimSpace(mapping.Interface), warning))
			}
		case mapping.Disabled:
			up := false
			out.Ports[mapping.Port] = observe.PortObservation{Port: mapping.Port, Up: &up, Disabled: true}
		case mapping.Unmapped:
			out.Ports[mapping.Port] = observe.PortObservation{Port: mapping.Port, Unmapped: true}
		}
	}
	return out, errs
}

func (p hostPlatform) Interface(ctx context.Context, iface string) (observe.PortObservation, []error) {
	observation, errs := ifsource.ObserveInterface(iface)
	if p.goos != goosLinux || p.cfg.ProcSource != ProcSourceProcFS {
		return observation, errs
	}
	procSnapshot, procErrs := p.Proc(ctx, ProcConfig{Source: p.cfg.ProcSource})
	if stats, ok := procSnapshot.Interfaces[strings.TrimSpace(iface)]; ok {
		observation.Stats = mergeInterfaceStats(observation.Stats, stats)
	}
	return observation, append(errs, procErrs...)
}

func (p hostPlatform) linuxBridge(ctx context.Context, cfg observe.BridgeConfig) (observe.BridgeObservation, []error) {
	observation := observe.BridgeObservation{
		Bridge:          strings.TrimSpace(cfg.Bridge),
		UplinkInterface: strings.TrimSpace(cfg.UplinkInterface),
		MemberPortMap:   cloneMemberPortMap(cfg.MemberPortMap),
	}
	var errs []error
	if observation.UplinkInterface != "" {
		uplink, warnings := p.Interface(ctx, observation.UplinkInterface)
		observation.Uplink = uplink
		errs = append(errs, warnings...)
	}
	if observation.Bridge != "" {
		entries, err := observe.BridgeFDB(ctx, observation.Bridge)
		if err != nil {
			errs = append(errs, err)
		} else {
			observation.MemberMACs = observe.MACEntriesByDevice(entries)
			if err := observe.EnrichMACEntriesWithLocalARP(observation.MemberMACs); err != nil {
				errs = append(errs, err)
			}
			observation.MemberRoles = observe.ClassifyBridgeMembers(observation.MemberMACs, observation.Bridge, observation.UplinkInterface)
			observation.RemoteMACs = observe.RemoteMACsByBridgeMember(observation.MemberMACs, observation.MemberRoles, observation.UplinkInterface, observation.Bridge)
			observation.MemberPorts, errs = p.bridgeMemberObservations(ctx, observation.MemberMACs, observation.MemberRoles, errs)
		}
	}
	return observation, errs
}

func (p hostPlatform) freebsdBridge(ctx context.Context, cfg observe.BridgeConfig) (observe.BridgeObservation, []error) {
	observation := observe.BridgeObservation{
		Bridge:          strings.TrimSpace(cfg.Bridge),
		UplinkInterface: strings.TrimSpace(cfg.UplinkInterface),
		MemberPortMap:   cloneMemberPortMap(cfg.MemberPortMap),
	}
	var errs []error
	if observation.Bridge != "" {
		entries, err := observe.FreeBSDBridgeAddr(ctx, observation.Bridge)
		if err != nil {
			errs = append(errs, err)
		} else {
			observation.MemberMACs = freeBSDMACEntriesByInterface(entries)
			observation.MemberRoles = observe.ClassifyBridgeMembers(observation.MemberMACs, observation.Bridge, observation.UplinkInterface)
			observation.RemoteMACs = observe.RemoteMACsByBridgeMember(observation.MemberMACs, observation.MemberRoles, observation.UplinkInterface, observation.Bridge)
			observation.MemberPorts, errs = p.bridgeMemberObservations(ctx, observation.MemberMACs, observation.MemberRoles, errs)
		}
	}
	if observation.UplinkInterface != "" {
		uplink, warnings := p.Interface(ctx, observation.UplinkInterface)
		observation.Uplink = uplink
		errs = append(errs, warnings...)
	}
	return observation, errs
}

func (p hostPlatform) bridgeMemberObservations(ctx context.Context, memberMACs map[string][]device.MacTableEntry, roles map[string]observe.BridgeMemberRole, errs []error) (map[string]observe.PortObservation, []error) {
	if len(memberMACs) == 0 {
		return nil, errs
	}
	out := map[string]observe.PortObservation{}
	for member := range memberMACs {
		role := roleForMember(roles, member)
		if role == observe.BridgeMemberRoleBridge || role == observe.BridgeMemberRoleIgnored {
			continue
		}
		observation, warnings := p.Interface(ctx, member)
		if strings.TrimSpace(observation.Interface) != "" || observation.SpeedMbps > 0 || observation.Up != nil {
			out[member] = observation
		}
		for _, warning := range warnings {
			errs = append(errs, fmt.Errorf("bridge member %s: %w", member, warning))
		}
	}
	if len(out) == 0 {
		return nil, errs
	}
	return out, errs
}

func roleForMember(roles map[string]observe.BridgeMemberRole, member string) observe.BridgeMemberRole {
	if len(roles) == 0 {
		return observe.BridgeMemberRoleUnknown
	}
	if role, ok := roles[strings.TrimSpace(member)]; ok {
		return role
	}
	lower := strings.ToLower(strings.TrimSpace(member))
	for name, role := range roles {
		if strings.ToLower(strings.TrimSpace(name)) == lower {
			return role
		}
	}
	return observe.BridgeMemberRoleUnknown
}

func (p hostPlatform) Proc(_ context.Context, cfg ProcConfig) (ProcSnapshot, []error) {
	source := normalizedSource(cfg.Source)
	if source == SourceOff {
		return ProcSnapshot{}, nil
	}
	if source != ProcSourceProcFS {
		return ProcSnapshot{}, []error{fmt.Errorf("unsupported proc source %q", source)}
	}
	root := strings.TrimSpace(cfg.Root)
	if root == "" {
		root = "/proc"
	}
	fs, err := procfs.NewFS(root)
	if err != nil {
		return ProcSnapshot{}, []error{fmt.Errorf("open procfs %s: %w", root, err)}
	}
	netdev, err := fs.NetDev()
	if err != nil {
		return ProcSnapshot{}, []error{fmt.Errorf("read procfs netdev: %w", err)}
	}
	out := ProcSnapshot{Interfaces: map[string]observe.InterfaceStats{}}
	for name, line := range netdev {
		out.Interfaces[name] = observe.InterfaceStats{
			RXBytes:   int64(line.RxBytes),
			TXBytes:   int64(line.TxBytes),
			RXPackets: int64(line.RxPackets),
			TXPackets: int64(line.TxPackets),
			RXErrors:  int64(line.RxErrors),
			TXErrors:  int64(line.TxErrors),
		}
	}
	return out, nil
}

func (p hostPlatform) Capabilities(ctx context.Context, cfg Config) CapabilityReport {
	cfg = normalizeConfig(cfg)
	report := CapabilityReport{GOOS: p.goos}
	report.Capabilities = append(report.Capabilities,
		p.commandCapability(capabilityLLDP, cfg.LLDPSource, "lldpcli"),
		p.logCapability(cfg),
		p.procCapability(cfg),
		p.dbusCapability(ctx, cfg),
		Capability{Name: capabilityTraffic, Source: cfg.TrafficSource, State: trafficState(cfg.TrafficSource), Detail: trafficDetail(cfg.TrafficSource)},
	)
	return report
}

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

func trafficState(source string) string {
	if normalizedSource(source) == SourceOff {
		return capabilityDisabled
	}
	return capabilityUnsupported
}

func trafficDetail(source string) string {
	if normalizedSource(source) == SourceOff {
		return ""
	}
	return "traffic metadata sources are not implemented"
}

func cloneMemberPortMap(values map[string]int) map[string]int {
	if len(values) == 0 {
		return nil
	}
	out := make(map[string]int, len(values))
	for member, port := range values {
		member = strings.TrimSpace(member)
		if member == "" {
			continue
		}
		out[member] = port
	}
	return out
}

func freeBSDMACEntriesByInterface(entries []freebsdifconfig.BridgeAddress) map[string][]device.MacTableEntry {
	return observe.FreeBSDMACEntriesByInterface(entries)
}

func mergeInterfaceStats(primary, fallback observe.InterfaceStats) observe.InterfaceStats {
	if primary.RXBytes == 0 {
		primary.RXBytes = fallback.RXBytes
	}
	if primary.TXBytes == 0 {
		primary.TXBytes = fallback.TXBytes
	}
	if primary.RXPackets == 0 {
		primary.RXPackets = fallback.RXPackets
	}
	if primary.TXPackets == 0 {
		primary.TXPackets = fallback.TXPackets
	}
	if primary.RXErrors == 0 {
		primary.RXErrors = fallback.RXErrors
	}
	if primary.TXErrors == 0 {
		primary.TXErrors = fallback.TXErrors
	}
	if primary.SpeedMbps == 0 {
		primary.SpeedMbps = fallback.SpeedMbps
	}
	return primary
}

func commandContext(ctx context.Context, timeout time.Duration, name string, args ...string) ([]byte, error) {
	if timeout <= 0 {
		timeout = defaultCommandTimeout
	}
	commandCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	out, err := exec.CommandContext(commandCtx, name, args...).Output()
	if errors.Is(commandCtx.Err(), context.DeadlineExceeded) {
		return out, fmt.Errorf("%s timed out after %s", name, timeout)
	}
	if err != nil {
		return out, fmt.Errorf("run %s: %w", name, err)
	}
	return out, nil
}

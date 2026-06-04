// Package observe builds passive host-network observations for switch payloads.
package observe

// Linux snapshots collect bridge FDB and sysfs observations for payload merging.
// They own command execution, counter reads, and deterministic member-to-port
// assignment.

import (
	"context"
	"fmt"
	"runtime"
	"strings"
)

// LinuxSnapshot reads passive Linux bridge and sysfs data.
func LinuxSnapshot(ctx context.Context, cfg Config, uplinkPortIndex int) (Snapshot, []error) {
	var errs []error
	snapshot := Snapshot{
		UplinkPortIndex: uplinkPortIndex,
		Interface:       strings.TrimSpace(cfg.Interface),
		Bridge:          strings.TrimSpace(cfg.Bridge),
		MemberPortMap:   normalizeMemberPortMap(cfg.MemberPortMap),
	}
	if strings.TrimSpace(cfg.SysfsRoot) == "" {
		cfg.SysfsRoot = "/sys"
	}
	if runtime.GOOS != "linux" {
		return snapshot, []error{fmt.Errorf("passive observation is not implemented on %s", runtime.GOOS)}
	}

	if cfg.Interface != "" {
		stats, err := ReadInterfaceStats(cfg.SysfsRoot, cfg.Interface)
		if err != nil {
			errs = append(errs, err)
		}
		if err == nil || hasCounters(stats) || stats.SpeedMbps > 0 {
			snapshot.Stats = stats
		}
	}
	if cfg.Bridge != "" {
		entries, err := BridgeFDB(ctx, cfg.Bridge)
		if err != nil {
			errs = append(errs, err)
		} else {
			snapshot.DeviceMACs = MACEntriesByDevice(entries)
			if err := EnrichMACsFromLocalARP(snapshot.DeviceMACs); err != nil {
				errs = append(errs, err)
			}
			snapshot.MemberRoles = ClassifyMembersWithIgnores(snapshot.DeviceMACs, snapshot.Bridge, snapshot.Interface, cfg.IgnoredMembers)
			snapshot.RemoteMACs = RemoteMACsByBridgeMember(snapshot.DeviceMACs, snapshot.MemberRoles, snapshot.Interface, snapshot.Bridge)
			snapshot.MemberPorts = linuxMemberPortObservations(cfg.SysfsRoot, snapshot.DeviceMACs, snapshot.MemberRoles)
			snapshot.MACs = flattenDeviceMACsByRole(snapshot.DeviceMACs, snapshot.MemberRoles, snapshot.Interface, snapshot.Bridge, snapshot.RemoteMACs)
		}
	}
	return snapshot, errs
}

// HostSnapshotFromSource reads a bridge observation through source and converts
// it to the legacy snapshot shape consumed by payload merge and status code.
func HostSnapshotFromSource(ctx context.Context, source ObservationSource, cfg Config, uplinkPortIndex int) (Snapshot, []error) {
	if source == nil {
		return HostSnapshot(ctx, cfg, uplinkPortIndex)
	}
	bridge, errs := source.Bridge(ctx, BridgeConfig{
		Bridge:          strings.TrimSpace(cfg.Bridge),
		UplinkInterface: strings.TrimSpace(cfg.Interface),
		IgnoredMembers:  cloneStrings(cfg.IgnoredMembers),
		MemberPortMap:   normalizeMemberPortMap(cfg.MemberPortMap),
	})
	snapshot := Snapshot{
		UplinkPortIndex: uplinkPortIndex,
		Interface:       strings.TrimSpace(bridge.UplinkInterface),
		Bridge:          strings.TrimSpace(bridge.Bridge),
		Stats:           bridge.Uplink.Stats,
		DeviceMACs:      bridge.MemberMACs,
		RemoteMACs:      normalizeRemoteMACSet(bridge.RemoteMACs),
		MemberPorts:     normalizeMemberPorts(bridge.MemberPorts),
		MemberPortMap:   normalizeMemberPortMap(bridge.MemberPortMap),
		MemberRoles:     normalizeMemberRoles(bridge.MemberRoles),
	}
	if snapshot.Stats.SpeedMbps == 0 {
		snapshot.Stats.SpeedMbps = bridge.Uplink.SpeedMbps
	}
	if len(snapshot.MemberRoles) == 0 {
		snapshot.MemberRoles = ClassifyMembers(snapshot.DeviceMACs, snapshot.Bridge, snapshot.Interface)
	}
	snapshot.MemberRoles = ApplyIgnoredMembers(snapshot.MemberRoles, cfg.IgnoredMembers)
	if len(snapshot.RemoteMACs) == 0 {
		snapshot.RemoteMACs = RemoteMACsByBridgeMember(snapshot.DeviceMACs, snapshot.MemberRoles, snapshot.Interface, snapshot.Bridge)
	}
	snapshot.MACs = flattenDeviceMACsByRole(snapshot.DeviceMACs, snapshot.MemberRoles, snapshot.Interface, snapshot.Bridge, snapshot.RemoteMACs)
	return snapshot, errs
}

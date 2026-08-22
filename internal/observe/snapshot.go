// Package observe builds passive host-network observations for switch payloads.
package observe

import (
	"context"
	"errors"
	"strings"
)

// HostSnapshotFromSource reads a bridge observation through source and converts
// it to the snapshot shape consumed by payload merge and status code.
func HostSnapshotFromSource(ctx context.Context, source ObservationSource, cfg Config, uplinkPortIndex int) (Snapshot, []error) {
	if source == nil {
		return Snapshot{
			UplinkPortIndex: uplinkPortIndex,
			Interface:       strings.TrimSpace(cfg.Interface),
			Bridge:          strings.TrimSpace(cfg.Bridge),
			MemberPortMap:   normalizeMemberPortMap(cfg.MemberPortMap),
		}, []error{errors.New("observation source is required")}
	}
	bridge, errs := source.Bridge(ctx, BridgeConfig{
		Bridge:          strings.TrimSpace(cfg.Bridge),
		UplinkInterface: strings.TrimSpace(cfg.Interface),
		IgnoredMembers:  append([]string(nil), cfg.IgnoredMembers...),
		MemberPortMap:   normalizeMemberPortMap(cfg.MemberPortMap),
	})
	snapshot := Snapshot{
		UplinkPortIndex: uplinkPortIndex,
		Interface:       strings.TrimSpace(bridge.UplinkInterface),
		Bridge:          strings.TrimSpace(bridge.Bridge),
		Up:              bridge.Uplink.Up,
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

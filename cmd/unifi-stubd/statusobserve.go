package main

import (
	"context"
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/device"
	"github.com/konstruktor1/unifi-stubd/internal/observe"
	"github.com/konstruktor1/unifi-stubd/internal/platform"
)

// shouldObserveStatus limits live observation during status to modes where the
// daemon would also read host data at runtime.
func shouldObserveStatus(mode string) bool {
	mode = normalizeMode(mode)
	return mode == operationModeBridgeObserve || mode == operationModePortMap || mode == operationModeHostDirect
}

// buildObservationStatus reads the same passive source model used by runtime
// payload rendering so --status can explain what the daemon would observe.
func buildObservationStatus(flags runtimeFlags, ports []device.Port, plt platform.Platform) statusObservation {
	ctx, cancel := context.WithTimeout(context.Background(), observeTimeout)
	defer cancel()

	bridgeObserve := effectiveBridgeObserve(flags)
	snapshot, errs := observe.HostSnapshotFromSource(ctx, plt, observe.Config{
		Interface:      strings.TrimSpace(bridgeObserve.UplinkInterface),
		Bridge:         strings.TrimSpace(bridgeObserve.Bridge),
		IgnoredMembers: cloneStrings(bridgeObserve.IgnoredMembers),
		MemberPortMap:  bridgeMemberPortMap(bridgeObserve.MemberPortMap),
	}, uplinkPortIndex(ports))

	out := statusObservation{
		Interface:      strings.TrimSpace(bridgeObserve.UplinkInterface),
		Bridge:         strings.TrimSpace(bridgeObserve.Bridge),
		InterfaceStats: snapshot.Stats,
		BridgeDevices:  len(snapshot.DeviceMACs),
		LearnedMACs:    len(snapshot.MACs),
	}
	for _, err := range errs {
		out.SourceWarnings = append(out.SourceWarnings, err.Error())
	}
	return out
}

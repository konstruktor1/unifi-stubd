// Package observe parses FreeBSD bridge forwarding rows through ifconfig and
// reports them in the shared observation model. Interface counters are
// deliberately marked planned instead of emulating Linux sysfs semantics.
package observe

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/adapters/freebsdifconfig"
	"github.com/konstruktor1/unifi-stubd/internal/device"
)

// HostSnapshot reads passive host-network data for the current OS.
func HostSnapshot(ctx context.Context, cfg Config, uplinkPortIndex int) (Snapshot, []error) {
	switch runtime.GOOS {
	case "linux":
		return LinuxSnapshot(ctx, cfg, uplinkPortIndex)
	case "freebsd":
		return FreeBSDSnapshot(ctx, cfg, uplinkPortIndex)
	default:
		snapshot := Snapshot{
			UplinkPortIndex: uplinkPortIndex,
			Interface:       strings.TrimSpace(cfg.Interface),
			Bridge:          strings.TrimSpace(cfg.Bridge),
			MemberPortMap:   normalizeMemberPortMap(cfg.MemberPortMap),
		}
		return snapshot, []error{fmt.Errorf("passive observation is not implemented on %s", runtime.GOOS)}
	}
}

// FreeBSDSnapshot reads passive FreeBSD bridge data into the shared snapshot
// model and reports unavailable counter support as a warning.
func FreeBSDSnapshot(ctx context.Context, cfg Config, uplinkPortIndex int) (Snapshot, []error) {
	snapshot := Snapshot{
		UplinkPortIndex: uplinkPortIndex,
		Interface:       strings.TrimSpace(cfg.Interface),
		Bridge:          strings.TrimSpace(cfg.Bridge),
		MemberPortMap:   normalizeMemberPortMap(cfg.MemberPortMap),
	}
	if runtime.GOOS != "freebsd" {
		return snapshot, []error{fmt.Errorf("freebsd observation is not implemented on %s", runtime.GOOS)}
	}

	var errs []error
	if snapshot.Bridge != "" {
		entries, err := FreeBSDBridgeAddr(ctx, snapshot.Bridge)
		if err != nil {
			errs = append(errs, err)
		} else {
			snapshot.DeviceMACs = FreeBSDMACsByInterface(entries)
			snapshot.MemberRoles = ClassifyMembersWithIgnores(snapshot.DeviceMACs, snapshot.Bridge, snapshot.Interface, cfg.IgnoredMembers)
			snapshot.RemoteMACs = RemoteMACsByBridgeMember(snapshot.DeviceMACs, snapshot.MemberRoles, snapshot.Interface, snapshot.Bridge)
			snapshot.MemberPorts = mapBridgeMemberInterfaces(snapshot.DeviceMACs, snapshot.MemberRoles)
			snapshot.MACs = flattenDeviceMACsByRole(snapshot.DeviceMACs, snapshot.MemberRoles, snapshot.Interface, snapshot.Bridge, snapshot.RemoteMACs)
		}
	}
	if snapshot.Interface != "" {
		errs = append(errs, fmt.Errorf("freebsd bridge-observe interface counters are planned; use port-map interface sources for per-port state"))
	}
	return snapshot, errs
}

// FreeBSDBridgeAddr reads FreeBSD bridge forwarding rows.
func FreeBSDBridgeAddr(ctx context.Context, bridge string) ([]freebsdifconfig.BridgeAddress, error) {
	bridge = strings.TrimSpace(bridge)
	if bridge == "" || strings.Contains(bridge, "/") {
		return nil, fmt.Errorf("invalid bridge name %q", bridge)
	}
	cmd := exec.CommandContext(ctx, "ifconfig", bridge, "addr")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("run ifconfig bridge addr command for %s: %w", bridge, err)
	}
	return freebsdifconfig.ParseBridgeAddr(strings.NewReader(string(out))), nil
}

// FreeBSDMACsByInterface converts learned FreeBSD bridge rows into
// per-member UniFi MAC entries, filtering local and multicast rows.
func FreeBSDMACsByInterface(entries []freebsdifconfig.BridgeAddress) map[string][]device.MacTableEntry {
	out := map[string][]device.MacTableEntry{}
	seen := map[string]bool{}
	for _, entry := range entries {
		if !learnedFreeBSDBridgeEntry(entry) {
			continue
		}
		iface := strings.TrimSpace(entry.Interface)
		key := iface + "|" + entry.MAC
		if seen[key] {
			continue
		}
		seen[key] = true
		out[iface] = append(out[iface], device.MacTableEntry{
			MAC:    entry.MAC,
			Age:    defaultBridgeAge(entry.Age),
			Uptime: 1200,
			VLAN:   entry.VLAN,
			Type:   "client",
		})
	}
	return out
}

// learnedFreeBSDBridgeEntry accepts only non-local unicast rows that can
// represent downstream clients.
func learnedFreeBSDBridgeEntry(entry freebsdifconfig.BridgeAddress) bool {
	mac, err := net.ParseMAC(entry.MAC)
	if err != nil || len(mac) == 0 {
		return false
	}
	if mac[0]&0x01 != 0 {
		return false
	}
	if entry.Local {
		return false
	}
	return strings.TrimSpace(entry.Interface) != ""
}

// defaultBridgeAge supplies a fresh MAC-table age when FreeBSD output omits
// one.
func defaultBridgeAge(value int) int {
	if value > 0 {
		return value
	}
	return 4
}

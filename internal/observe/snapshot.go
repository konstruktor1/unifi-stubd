// Package observe builds passive host-network observations for switch payloads.
package observe

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/adapters/linuxbridge"
	"github.com/konstruktor1/unifi-stubd/internal/device"
)

// Config selects passive Linux observation sources.
type Config struct {
	// Interface is the host interface used for counters and link speed.
	Interface string
	// Bridge is the Linux bridge used for FDB MAC table data.
	Bridge string
	// SysfsRoot is the sysfs root, usually /sys.
	SysfsRoot string
}

// InterfaceStats contains passive counters and link speed for one interface.
type InterfaceStats struct {
	RXBytes   int64
	TXBytes   int64
	RXPackets int64
	TXPackets int64
	RXErrors  int64
	TXErrors  int64
	SpeedMbps int
}

// Snapshot contains passive data that can be merged into generated switch ports.
type Snapshot struct {
	UplinkPortIndex int
	Stats           InterfaceStats
	MACs            []device.MacTableEntry
}

// LinuxSnapshot reads passive Linux bridge and sysfs data.
func LinuxSnapshot(ctx context.Context, cfg Config, uplinkPortIndex int) (Snapshot, []error) {
	var errs []error
	snapshot := Snapshot{UplinkPortIndex: uplinkPortIndex}
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
			snapshot.MACs = MACEntries(entries)
		}
	}
	return snapshot, errs
}

// Apply merges a passive snapshot into generated switch ports.
func Apply(ports []device.Port, snapshot Snapshot) []device.Port {
	if len(ports) == 0 {
		return ports
	}
	out := make([]device.Port, len(ports))
	copy(out, ports)
	index := snapshot.UplinkPortIndex
	if index < 1 || index > len(out) {
		index = uplinkPortIndex(out)
	}
	port := &out[index-1]
	if snapshot.Stats.SpeedMbps > 0 {
		port.Speed = snapshot.Stats.SpeedMbps
	}
	if hasCounters(snapshot.Stats) {
		port.RXBytes = snapshot.Stats.RXBytes
		port.TXBytes = snapshot.Stats.TXBytes
		port.RXPackets = snapshot.Stats.RXPackets
		port.TXPackets = snapshot.Stats.TXPackets
		port.RXErrors = snapshot.Stats.RXErrors
		port.TXErrors = snapshot.Stats.TXErrors
	}
	if len(snapshot.MACs) > 0 {
		port.MACs = snapshot.MACs
	}
	return out
}

// ReadInterfaceStats reads interface counters and speed from a sysfs tree.
func ReadInterfaceStats(sysfsRoot, iface string) (InterfaceStats, error) {
	iface = strings.TrimSpace(iface)
	if iface == "" || strings.Contains(iface, "/") {
		return InterfaceStats{}, fmt.Errorf("invalid interface name %q", iface)
	}
	base := filepath.Join(sysfsRoot, "class", "net", iface)
	stats := filepath.Join(base, "statistics")
	out := InterfaceStats{}
	var errs []error
	read := func(name string) int64 {
		value, err := readInt64(filepath.Join(stats, name))
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", name, err))
		}
		return value
	}
	out.RXBytes = read("rx_bytes")
	out.TXBytes = read("tx_bytes")
	out.RXPackets = read("rx_packets")
	out.TXPackets = read("tx_packets")
	out.RXErrors = read("rx_errors")
	out.TXErrors = read("tx_errors")
	speed, err := readInt64(filepath.Join(base, "speed"))
	if err != nil {
		errs = append(errs, fmt.Errorf("speed: %w", err))
	} else if speed > 0 {
		out.SpeedMbps = int(speed)
	}
	return out, errors.Join(errs...)
}

// BridgeFDB reads bridge forwarding database rows for bridge.
func BridgeFDB(ctx context.Context, bridge string) ([]linuxbridge.FDBEntry, error) {
	bridge = strings.TrimSpace(bridge)
	if bridge == "" || strings.Contains(bridge, "/") {
		return nil, fmt.Errorf("invalid bridge name %q", bridge)
	}
	cmd := exec.CommandContext(ctx, "bridge", "fdb", "show", "br", bridge)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return linuxbridge.ParseFDB(strings.NewReader(string(out))), nil
}

// MACEntries converts Linux bridge FDB rows into UniFi MAC table entries.
func MACEntries(entries []linuxbridge.FDBEntry) []device.MacTableEntry {
	out := make([]device.MacTableEntry, 0, len(entries))
	seen := map[string]bool{}
	for _, entry := range entries {
		if !learnedFDBEntry(entry) || seen[entry.MAC] {
			continue
		}
		seen[entry.MAC] = true
		mac := device.MacTableEntry{
			MAC:    entry.MAC,
			Age:    4,
			Uptime: 1200,
			VLAN:   entry.VLAN,
			Type:   "client",
		}
		out = append(out, mac)
	}
	return out
}

func learnedFDBEntry(entry linuxbridge.FDBEntry) bool {
	mac, err := net.ParseMAC(entry.MAC)
	if err != nil || len(mac) == 0 {
		return false
	}
	if mac[0]&0x01 != 0 {
		return false
	}
	if entry.Local || entry.Permanent || entry.Self {
		return false
	}
	return entry.Dynamic || entry.Static || (!entry.Local && !entry.Permanent)
}

func uplinkPortIndex(ports []device.Port) int {
	for _, port := range ports {
		if port.Uplink {
			return port.Index
		}
	}
	return 1
}

func hasCounters(stats InterfaceStats) bool {
	return stats.RXBytes != 0 ||
		stats.TXBytes != 0 ||
		stats.RXPackets != 0 ||
		stats.TXPackets != 0 ||
		stats.RXErrors != 0 ||
		stats.TXErrors != 0
}

func readInt64(path string) (int64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	value, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		return 0, err
	}
	return value, nil
}

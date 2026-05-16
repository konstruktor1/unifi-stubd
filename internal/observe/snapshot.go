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
	"sort"
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
	Interface       string
	Bridge          string
	Stats           InterfaceStats
	MACs            []device.MacTableEntry
	DeviceMACs      map[string][]device.MacTableEntry
}

// LinuxSnapshot reads passive Linux bridge and sysfs data.
func LinuxSnapshot(ctx context.Context, cfg Config, uplinkPortIndex int) (Snapshot, []error) {
	var errs []error
	snapshot := Snapshot{
		UplinkPortIndex: uplinkPortIndex,
		Interface:       strings.TrimSpace(cfg.Interface),
		Bridge:          strings.TrimSpace(cfg.Bridge),
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
			snapshot.MACs = flattenDeviceMACs(snapshot.DeviceMACs, snapshot.Interface, snapshot.Bridge)
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
	if len(snapshot.DeviceMACs) > 0 {
		applyDeviceMACs(out, snapshot, index)
	} else if len(snapshot.MACs) > 0 {
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
	return flattenDeviceMACs(MACEntriesByDevice(entries), "", "")
}

// MACEntriesByDevice converts Linux bridge FDB rows into MAC entries grouped by bridge member.
func MACEntriesByDevice(entries []linuxbridge.FDBEntry) map[string][]device.MacTableEntry {
	out := map[string][]device.MacTableEntry{}
	seen := map[string]bool{}
	for _, entry := range entries {
		if !learnedFDBEntry(entry) {
			continue
		}
		deviceName := strings.TrimSpace(entry.Device)
		key := deviceName + "|" + entry.MAC
		if seen[key] {
			continue
		}
		seen[key] = true
		mac := device.MacTableEntry{
			MAC:    entry.MAC,
			Age:    4,
			Uptime: 1200,
			VLAN:   entry.VLAN,
			Type:   "client",
		}
		out[deviceName] = append(out[deviceName], mac)
	}
	return out
}

func flattenDeviceMACs(deviceMACs map[string][]device.MacTableEntry, iface, bridge string) []device.MacTableEntry {
	count := 0
	for _, macs := range deviceMACs {
		count += len(macs)
	}
	out := make([]device.MacTableEntry, 0, count)
	for _, deviceName := range sortedDeviceNames(deviceMACs, iface, bridge) {
		out = append(out, deviceMACs[deviceName]...)
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

func applyDeviceMACs(ports []device.Port, snapshot Snapshot, uplinkIndex int) {
	for index := range ports {
		ports[index].MACs = nil
	}

	accessIndexes := make([]int, 0, len(ports)-1)
	for _, port := range ports {
		if port.Index != uplinkIndex {
			accessIndexes = append(accessIndexes, port.Index)
		}
	}
	nextAccess := 0

	for _, deviceName := range sortedDeviceNames(snapshot.DeviceMACs, snapshot.Interface, snapshot.Bridge) {
		macs := snapshot.DeviceMACs[deviceName]
		if len(macs) == 0 {
			continue
		}

		portIndex := uplinkIndex
		isUplink := isUplinkDevice(deviceName, snapshot.Interface, snapshot.Bridge)
		if !isUplink && nextAccess < len(accessIndexes) {
			portIndex = accessIndexes[nextAccess]
			nextAccess++
		}
		if portIndex < 1 || portIndex > len(ports) {
			portIndex = uplinkIndex
		}

		port := &ports[portIndex-1]
		port.MACs = append(port.MACs, macs...)
		if deviceName != "" && (isUplink || portIndex != uplinkIndex) {
			port.Name = deviceName
		}
	}
}

func sortedDeviceNames(deviceMACs map[string][]device.MacTableEntry, iface, bridge string) []string {
	names := make([]string, 0, len(deviceMACs))
	for deviceName := range deviceMACs {
		names = append(names, deviceName)
	}
	sort.Slice(names, func(i, j int) bool {
		left := deviceSortKey(names[i], iface, bridge)
		right := deviceSortKey(names[j], iface, bridge)
		if left.rank != right.rank {
			return left.rank < right.rank
		}
		if left.number != right.number {
			return left.number < right.number
		}
		return left.name < right.name
	})
	return names
}

type sortKey struct {
	rank   int
	number int
	name   string
}

func deviceSortKey(deviceName, iface, bridge string) sortKey {
	name := strings.ToLower(strings.TrimSpace(deviceName))
	rank := 50
	switch {
	case isUplinkDevice(name, iface, bridge):
		rank = 0
	case strings.HasPrefix(name, "tap"):
		rank = 10
	case strings.HasPrefix(name, "veth"):
		rank = 20
	case strings.HasPrefix(name, "fwln"), strings.HasPrefix(name, "fwpr"), strings.HasPrefix(name, "fwbr"):
		rank = 30
	}
	return sortKey{rank: rank, number: firstNumber(name), name: name}
}

func isUplinkDevice(deviceName, iface, bridge string) bool {
	name := strings.ToLower(strings.TrimSpace(deviceName))
	if name == "" {
		return false
	}
	return name == strings.ToLower(strings.TrimSpace(iface)) ||
		name == strings.ToLower(strings.TrimSpace(bridge))
}

func firstNumber(value string) int {
	start := -1
	for i, r := range value {
		if r >= '0' && r <= '9' {
			start = i
			break
		}
	}
	if start < 0 {
		return 0
	}
	end := start
	for end < len(value) && value[end] >= '0' && value[end] <= '9' {
		end++
	}
	number, err := strconv.Atoi(value[start:end])
	if err != nil {
		return 0
	}
	return number
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

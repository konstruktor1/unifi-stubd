// Package observe builds passive host-network observations for switch payloads.
package observe

// Linux snapshots collect bridge FDB and sysfs observations for payload merging.
// They own command execution, counter reads, and deterministic member-to-port
// assignment.

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
	// MemberPortMap pins bridge member interfaces to one-based UniFi ports.
	MemberPortMap map[string]int
	// SysfsRoot is the sysfs root, usually /sys.
	SysfsRoot string
}

// InterfaceStats contains passive counters and link speed for one interface.
type InterfaceStats struct {
	// RXBytes is the received byte counter.
	RXBytes int64 `json:"rx_bytes,omitempty"`
	// TXBytes is the transmitted byte counter.
	TXBytes int64 `json:"tx_bytes,omitempty"`
	// RXPackets is the received packet counter.
	RXPackets int64 `json:"rx_packets,omitempty"`
	// TXPackets is the transmitted packet counter.
	TXPackets int64 `json:"tx_packets,omitempty"`
	// RXErrors is the receive error counter.
	RXErrors int64 `json:"rx_errors,omitempty"`
	// TXErrors is the transmit error counter.
	TXErrors int64 `json:"tx_errors,omitempty"`
	// SpeedMbps is the reported link speed in Mbps.
	SpeedMbps int `json:"speed_mbps,omitempty"`
}

// Snapshot contains passive data that can be merged into generated switch ports.
type Snapshot struct {
	// UplinkPortIndex is the one-based target port for uplink observations.
	UplinkPortIndex int
	// Interface is the observed host interface name.
	Interface string
	// Bridge is the observed Linux bridge name.
	Bridge string
	// Stats contains counters and link speed from the observed interface.
	Stats InterfaceStats
	// MACs contains learned MAC entries flattened for the uplink fallback.
	MACs []device.MacTableEntry
	// DeviceMACs contains learned MAC entries grouped by bridge member.
	DeviceMACs map[string][]device.MacTableEntry
	// RemoteMACs contains MACs learned behind the physical uplink neighbor.
	RemoteMACs map[string]bool
	// MemberPorts contains observed interface state grouped by bridge member.
	MemberPorts map[string]PortObservation
	// MemberPortMap pins bridge member interfaces to one-based UniFi ports.
	MemberPortMap map[string]int
	// MemberRoles classifies bridge members before they are mapped to ports.
	MemberRoles map[string]BridgeMemberRole
}

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
			if err := EnrichMACEntriesWithLocalARP(snapshot.DeviceMACs); err != nil {
				errs = append(errs, err)
			}
			snapshot.MemberRoles = ClassifyBridgeMembers(snapshot.DeviceMACs, snapshot.Bridge, snapshot.Interface)
			snapshot.RemoteMACs = RemoteMACsByBridgeMember(snapshot.DeviceMACs, snapshot.MemberRoles, snapshot.Interface, snapshot.Bridge)
			snapshot.MemberPorts = linuxMemberPortObservations(cfg.SysfsRoot, snapshot.DeviceMACs, snapshot.MemberRoles)
			snapshot.MACs = flattenDeviceMACsExcept(snapshot.DeviceMACs, snapshot.Interface, snapshot.Bridge, snapshot.RemoteMACs)
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
		snapshot.MemberRoles = ClassifyBridgeMembers(snapshot.DeviceMACs, snapshot.Bridge, snapshot.Interface)
	}
	if len(snapshot.RemoteMACs) == 0 {
		snapshot.RemoteMACs = RemoteMACsByBridgeMember(snapshot.DeviceMACs, snapshot.MemberRoles, snapshot.Interface, snapshot.Bridge)
	}
	snapshot.MACs = flattenDeviceMACsExcept(snapshot.DeviceMACs, snapshot.Interface, snapshot.Bridge, snapshot.RemoteMACs)
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
		applyInterfaceStatsToPort(port, snapshot.Stats)
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
	for _, field := range interfaceStatsFields {
		field.setStats(&out, read(field.sysfsName))
	}
	speed, err := readInt64(filepath.Join(base, "speed"))
	if err != nil {
		errs = append(errs, fmt.Errorf("speed: %w", err))
	} else if speed > 0 {
		out.SpeedMbps = int(speed)
	}
	if err := errors.Join(errs...); err != nil {
		return out, fmt.Errorf("read interface stats for %s: %w", iface, err)
	}
	return out, nil
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
		return nil, fmt.Errorf("run bridge FDB command for %s: %w", bridge, err)
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
	return flattenDeviceMACsExcept(deviceMACs, iface, bridge, nil)
}

func flattenDeviceMACsExcept(deviceMACs map[string][]device.MacTableEntry, iface, bridge string, remoteMACs map[string]bool) []device.MacTableEntry {
	count := 0
	for _, macs := range deviceMACs {
		count += len(macs)
	}
	out := make([]device.MacTableEntry, 0, count)
	for _, deviceName := range sortedDeviceNames(deviceMACs, iface, bridge) {
		out = append(out, filterRemoteMACEntries(deviceMACs[deviceName], remoteMACs)...)
	}
	return out
}

// RemoteMACsByBridgeMember returns MACs learned on the physical uplink member.
// These entries describe devices behind the real neighbor switch, not local
// participants of the represented virtual switch.
func RemoteMACsByBridgeMember(memberMACs map[string][]device.MacTableEntry, roles map[string]BridgeMemberRole, iface, bridge string) map[string]bool {
	if len(memberMACs) == 0 {
		return nil
	}
	out := map[string]bool{}
	for member, macs := range memberMACs {
		role := bridgeMemberRole(roles, member)
		if role != BridgeMemberRoleUplink && !isUplinkDevice(member, iface, bridge) {
			continue
		}
		for _, entry := range macs {
			if key := normalizedMACKey(entry.MAC); key != "" {
				out[key] = true
			}
		}
	}
	if len(out) == 0 {
		return nil
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
	for _, field := range interfaceStatsFields {
		if field.get(stats) != 0 {
			return true
		}
	}
	return false
}

type interfaceStatsField struct {
	sysfsName string
	get       func(InterfaceStats) int64
	setStats  func(*InterfaceStats, int64)
	setPort   func(*device.Port, int64)
}

var interfaceStatsFields = []interfaceStatsField{
	{
		sysfsName: "rx_bytes",
		get:       func(stats InterfaceStats) int64 { return stats.RXBytes },
		setStats:  func(stats *InterfaceStats, value int64) { stats.RXBytes = value },
		setPort:   func(port *device.Port, value int64) { port.RXBytes = value },
	},
	{
		sysfsName: "tx_bytes",
		get:       func(stats InterfaceStats) int64 { return stats.TXBytes },
		setStats:  func(stats *InterfaceStats, value int64) { stats.TXBytes = value },
		setPort:   func(port *device.Port, value int64) { port.TXBytes = value },
	},
	{
		sysfsName: "rx_packets",
		get:       func(stats InterfaceStats) int64 { return stats.RXPackets },
		setStats:  func(stats *InterfaceStats, value int64) { stats.RXPackets = value },
		setPort:   func(port *device.Port, value int64) { port.RXPackets = value },
	},
	{
		sysfsName: "tx_packets",
		get:       func(stats InterfaceStats) int64 { return stats.TXPackets },
		setStats:  func(stats *InterfaceStats, value int64) { stats.TXPackets = value },
		setPort:   func(port *device.Port, value int64) { port.TXPackets = value },
	},
	{
		sysfsName: "rx_errors",
		get:       func(stats InterfaceStats) int64 { return stats.RXErrors },
		setStats:  func(stats *InterfaceStats, value int64) { stats.RXErrors = value },
		setPort:   func(port *device.Port, value int64) { port.RXErrors = value },
	},
	{
		sysfsName: "tx_errors",
		get:       func(stats InterfaceStats) int64 { return stats.TXErrors },
		setStats:  func(stats *InterfaceStats, value int64) { stats.TXErrors = value },
		setPort:   func(port *device.Port, value int64) { port.TXErrors = value },
	},
}

func applyInterfaceStatsToPort(port *device.Port, stats InterfaceStats) {
	for _, field := range interfaceStatsFields {
		field.setPort(port, field.get(stats))
	}
}

func applyDeviceMACs(ports []device.Port, snapshot Snapshot, uplinkIndex int) {
	for index := range ports {
		ports[index].MACs = nil
	}

	remoteMACs := normalizeRemoteMACSet(snapshot.RemoteMACs)
	if len(remoteMACs) == 0 {
		remoteMACs = RemoteMACsByBridgeMember(snapshot.DeviceMACs, snapshot.MemberRoles, snapshot.Interface, snapshot.Bridge)
	}
	usedPorts := map[int]bool{uplinkIndex: true}
	accessIndexes := make([]int, 0, len(ports)-1)
	pinned := validPinnedPortSet(snapshot.MemberPortMap, len(ports))
	for _, port := range ports {
		if port.Index != uplinkIndex && !pinned[port.Index] {
			accessIndexes = append(accessIndexes, port.Index)
		}
	}
	nextAccess := 0

	for _, deviceName := range sortedDeviceNames(snapshot.DeviceMACs, snapshot.Interface, snapshot.Bridge) {
		macs := snapshot.DeviceMACs[deviceName]
		if len(macs) == 0 {
			continue
		}
		role := bridgeMemberRole(snapshot.MemberRoles, deviceName)
		if role == BridgeMemberRoleBridge || role == BridgeMemberRoleIgnored {
			continue
		}

		portIndex := uplinkIndex
		isUplink := role == BridgeMemberRoleUplink || isUplinkDevice(deviceName, snapshot.Interface, snapshot.Bridge)
		if isUplink {
			port := &ports[portIndex-1]
			applyMemberPortObservation(port, snapshot.MemberPorts, deviceName)
			usedPorts[portIndex] = true
			if deviceName != "" {
				port.Name = deviceName
			}
			continue
		}
		macs = filterRemoteMACEntries(macs, remoteMACs)
		if len(macs) == 0 {
			continue
		}
		if pinnedPort := snapshot.MemberPortMap[strings.TrimSpace(deviceName)]; pinnedPort >= 1 && pinnedPort <= len(ports) {
			portIndex = pinnedPort
		} else if nextAccess < len(accessIndexes) {
			portIndex = accessIndexes[nextAccess]
			nextAccess++
		}
		if portIndex < 1 || portIndex > len(ports) {
			portIndex = uplinkIndex
		}

		port := &ports[portIndex-1]
		applyMemberPortObservation(port, snapshot.MemberPorts, deviceName)
		port.MACs = append(port.MACs, macs...)
		usedPorts[portIndex] = true
		if deviceName != "" && (isUplink || portIndex != uplinkIndex) {
			port.Name = deviceName
		}
	}
	for index := range ports {
		if !usedPorts[ports[index].Index] {
			markBridgePortDisconnected(&ports[index])
		}
	}
}

func filterRemoteMACEntries(entries []device.MacTableEntry, remoteMACs map[string]bool) []device.MacTableEntry {
	if len(entries) == 0 || len(remoteMACs) == 0 {
		return entries
	}
	out := make([]device.MacTableEntry, 0, len(entries))
	for _, entry := range entries {
		if key := normalizedMACKey(entry.MAC); key != "" && remoteMACs[key] {
			continue
		}
		out = append(out, entry)
	}
	return out
}

func markBridgePortDisconnected(port *device.Port) {
	port.Up = false
	port.Speed = 0
	port.MACs = nil
	for _, field := range interfaceStatsFields {
		field.setPort(port, 0)
	}
}

func normalizeMemberPortMap(values map[string]int) map[string]int {
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

func normalizeMemberPorts(values map[string]PortObservation) map[string]PortObservation {
	if len(values) == 0 {
		return nil
	}
	out := make(map[string]PortObservation, len(values))
	for member, observation := range values {
		member = strings.TrimSpace(member)
		if member == "" {
			continue
		}
		out[member] = observation
	}
	return out
}

func normalizeMemberRoles(values map[string]BridgeMemberRole) map[string]BridgeMemberRole {
	if len(values) == 0 {
		return nil
	}
	out := make(map[string]BridgeMemberRole, len(values))
	for member, role := range values {
		member = strings.TrimSpace(member)
		if member == "" {
			continue
		}
		out[member] = role
	}
	return out
}

func normalizeRemoteMACSet(values map[string]bool) map[string]bool {
	if len(values) == 0 {
		return nil
	}
	out := make(map[string]bool, len(values))
	for value, enabled := range values {
		if !enabled {
			continue
		}
		if key := normalizedMACKey(value); key != "" {
			out[key] = true
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func normalizedMACKey(value string) string {
	mac, err := net.ParseMAC(strings.TrimSpace(value))
	if err != nil {
		return ""
	}
	return strings.ToLower(mac.String())
}

func applyMemberPortObservation(port *device.Port, observations map[string]PortObservation, member string) {
	observation, ok := memberPortObservation(observations, member)
	if !ok {
		return
	}
	if iface := strings.TrimSpace(observation.Interface); iface != "" {
		port.Interface = iface
	}
	if observation.Up != nil {
		port.Up = *observation.Up
	}
	if observation.SpeedMbps > 0 {
		port.Speed = observation.SpeedMbps
	}
	if media := strings.TrimSpace(observation.Media); media != "" {
		port.Media = media
	}
	if hasCounters(observation.Stats) {
		applyInterfaceStatsToPort(port, observation.Stats)
	}
	if !port.Up && observation.SpeedMbps <= 0 {
		port.Speed = 0
	}
}

func memberPortObservation(observations map[string]PortObservation, member string) (PortObservation, bool) {
	if len(observations) == 0 {
		return PortObservation{}, false
	}
	if observation, ok := observations[strings.TrimSpace(member)]; ok {
		return observation, true
	}
	lower := strings.ToLower(strings.TrimSpace(member))
	for key, observation := range observations {
		if strings.ToLower(strings.TrimSpace(key)) == lower {
			return observation, true
		}
	}
	return PortObservation{}, false
}

func validPinnedPortSet(values map[string]int, portCount int) map[int]bool {
	out := map[int]bool{}
	for _, port := range values {
		if port >= 1 && port <= portCount {
			out[port] = true
		}
	}
	return out
}

func linuxMemberPortObservations(sysfsRoot string, memberMACs map[string][]device.MacTableEntry, roles map[string]BridgeMemberRole) map[string]PortObservation {
	if len(memberMACs) == 0 {
		return nil
	}
	out := map[string]PortObservation{}
	for member := range memberMACs {
		role := bridgeMemberRole(roles, member)
		if role == BridgeMemberRoleBridge || role == BridgeMemberRoleIgnored {
			continue
		}
		stats, err := ReadInterfaceStats(sysfsRoot, member)
		if err != nil && !hasCounters(stats) && stats.SpeedMbps <= 0 {
			continue
		}
		out[member] = PortObservation{
			Interface: strings.TrimSpace(member),
			SpeedMbps: stats.SpeedMbps,
			Stats:     stats,
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func mapBridgeMemberInterfaces(memberMACs map[string][]device.MacTableEntry, roles map[string]BridgeMemberRole) map[string]PortObservation {
	if len(memberMACs) == 0 {
		return nil
	}
	out := map[string]PortObservation{}
	for member := range memberMACs {
		role := bridgeMemberRole(roles, member)
		if role == BridgeMemberRoleBridge || role == BridgeMemberRoleIgnored {
			continue
		}
		out[member] = PortObservation{Interface: strings.TrimSpace(member)}
	}
	if len(out) == 0 {
		return nil
	}
	return out
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
	case isBridgeDevice(name, bridge):
		rank = 90
	case strings.HasPrefix(name, "tap"):
		rank = 10
	case strings.HasPrefix(name, "veth"):
		rank = 20
	case strings.HasPrefix(name, "fwln"), strings.HasPrefix(name, "fwpr"), strings.HasPrefix(name, "fwbr"):
		rank = 30
	}
	return sortKey{rank: rank, number: firstNumber(name), name: name}
}

func isUplinkDevice(deviceName, iface, _ string) bool {
	name := strings.ToLower(strings.TrimSpace(deviceName))
	if name == "" {
		return false
	}
	return name == strings.ToLower(strings.TrimSpace(iface))
}

func isBridgeDevice(deviceName, bridge string) bool {
	name := strings.ToLower(strings.TrimSpace(deviceName))
	return name != "" && name == strings.ToLower(strings.TrimSpace(bridge))
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
		return 0, fmt.Errorf("read integer file %s: %w", path, err)
	}
	value, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse integer file %s: %w", path, err)
	}
	return value, nil
}

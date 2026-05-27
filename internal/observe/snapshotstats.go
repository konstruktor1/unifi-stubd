package observe

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/device"
)

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

// hasCounters distinguishes a usable zero-speed observation from a snapshot
// with no traffic data at all.
func hasCounters(stats InterfaceStats) bool {
	for _, field := range interfaceStatsFields {
		if field.get(stats) != 0 {
			return true
		}
	}
	return false
}

// interfaceStatsField maps one Linux sysfs statistic into observation and
// payload port fields.
type interfaceStatsField struct {
	sysfsName string
	get       func(InterfaceStats) int64
	setStats  func(*InterfaceStats, int64)
	setPort   func(*device.Port, int64)
}

// interfaceStatsFields enumerates the sysfs counters copied into observed port
// state.
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

// applyInterfaceStatsToPort copies the observed counter set into the rendered
// port so payload tables and status use the same values.
func applyInterfaceStatsToPort(port *device.Port, stats InterfaceStats) {
	for _, field := range interfaceStatsFields {
		field.setPort(port, field.get(stats))
	}
}

// readInt64 parses Linux sysfs counter and speed files.
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

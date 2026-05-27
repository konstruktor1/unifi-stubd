package ifsource

import (
	"os/exec"
	"strconv"
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/device"
	"github.com/konstruktor1/unifi-stubd/internal/observe"
)

// readHostInterfaceCounters uses netstat as a portable fallback when sysfs did
// not provide complete counters.
func readHostInterfaceCounters(ifaceName string) (observe.InterfaceStats, bool) {
	out, err := exec.Command("netstat", "-ibn", "-I", ifaceName).Output()
	if err != nil {
		return observe.InterfaceStats{}, false
	}
	return parseNetstatCounters(string(out), ifaceName)
}

// mergeInterfaceStats uses fallback counters only for fields the primary source
// did not provide, keeping more specific interface reads authoritative.
func mergeInterfaceStats(primary, fallback observe.InterfaceStats) observe.InterfaceStats {
	for _, field := range interfaceCounterFields {
		if field.get(primary) == 0 {
			field.setCounter(&primary, field.get(fallback))
		}
	}
	if primary.SpeedMbps == 0 {
		primary.SpeedMbps = fallback.SpeedMbps
	}
	return primary
}

// interfaceCounterField maps netstat columns into portable counters and port
// overrides.
type interfaceCounterField struct {
	netstatIndex int
	get          func(observe.InterfaceStats) int64
	setCounter   func(*observe.InterfaceStats, int64)
	setOverride  func(*device.PortOverride, int64)
}

// interfaceCounterFields lists the counter columns read from netstat fallback
// output.
var interfaceCounterFields = []interfaceCounterField{
	{
		netstatIndex: 4,
		get:          func(counters observe.InterfaceStats) int64 { return counters.RXPackets },
		setCounter:   func(counters *observe.InterfaceStats, value int64) { counters.RXPackets = value },
		setOverride:  func(override *device.PortOverride, value int64) { override.RXPackets = value },
	},
	{
		netstatIndex: 5,
		get:          func(counters observe.InterfaceStats) int64 { return counters.RXErrors },
		setCounter:   func(counters *observe.InterfaceStats, value int64) { counters.RXErrors = value },
		setOverride:  func(override *device.PortOverride, value int64) { override.RXErrors = value },
	},
	{
		netstatIndex: 7,
		get:          func(counters observe.InterfaceStats) int64 { return counters.RXBytes },
		setCounter:   func(counters *observe.InterfaceStats, value int64) { counters.RXBytes = value },
		setOverride:  func(override *device.PortOverride, value int64) { override.RXBytes = value },
	},
	{
		netstatIndex: 8,
		get:          func(counters observe.InterfaceStats) int64 { return counters.TXPackets },
		setCounter:   func(counters *observe.InterfaceStats, value int64) { counters.TXPackets = value },
		setOverride:  func(override *device.PortOverride, value int64) { override.TXPackets = value },
	},
	{
		netstatIndex: 9,
		get:          func(counters observe.InterfaceStats) int64 { return counters.TXErrors },
		setCounter:   func(counters *observe.InterfaceStats, value int64) { counters.TXErrors = value },
		setOverride:  func(override *device.PortOverride, value int64) { override.TXErrors = value },
	},
	{
		netstatIndex: 10,
		get:          func(counters observe.InterfaceStats) int64 { return counters.TXBytes },
		setCounter:   func(counters *observe.InterfaceStats, value int64) { counters.TXBytes = value },
		setOverride:  func(override *device.PortOverride, value int64) { override.TXBytes = value },
	},
}

// parseNetstatCounters extracts counters for one link-layer interface row.
func parseNetstatCounters(output, ifaceName string) (observe.InterfaceStats, bool) {
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 11 || fields[0] != ifaceName || !strings.HasPrefix(fields[2], "<Link#") {
			continue
		}
		var counters observe.InterfaceStats
		for _, field := range interfaceCounterFields {
			value, err := strconv.ParseInt(fields[field.netstatIndex], 10, 64)
			if err != nil {
				return observe.InterfaceStats{}, false
			}
			field.setCounter(&counters, value)
		}
		return counters, true
	}
	return observe.InterfaceStats{}, false
}

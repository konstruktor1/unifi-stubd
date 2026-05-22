// Package ifsource maps host interface metadata into payload port overrides.
package ifsource

// Interface sources convert host NIC state into controller-facing port
// overrides for port-map and explicit port_overrides[].interface entries.

import (
	"errors"
	"log"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/device"
	"github.com/konstruktor1/unifi-stubd/internal/observe"
)

// hostInterfaceDetails carries link state parsed from platform network tools.
type hostInterfaceDetails struct {
	// Up is the optional parsed carrier state.
	Up *bool
	// Speed is the parsed link speed in Mbps.
	Speed int
	// Media is the UniFi media label inferred from platform output.
	Media string
}

// EnrichPortOverrides overlays configured ports with host interface data.
func EnrichPortOverrides(overrides []device.PortOverride) []device.PortOverride {
	if len(overrides) == 0 {
		return overrides
	}
	out := make([]device.PortOverride, len(overrides))
	copy(out, overrides)
	for index := range out {
		ifaceName := strings.TrimSpace(out[index].Interface)
		if ifaceName == "" {
			continue
		}
		out[index].Interface = ifaceName
		EnrichPortOverride(&out[index], ifaceName)
	}
	return out
}

// EnrichPortOverride applies one host interface snapshot to an override.
func EnrichPortOverride(override *device.PortOverride, ifaceName string) {
	observation, errs := ObserveInterface(ifaceName)
	for _, err := range errs {
		log.Printf("port %d interface source %s warning: %v", override.Port, ifaceName, err)
	}
	if len(errs) > 0 && strings.TrimSpace(observation.Interface) == "" {
		return
	}
	ApplyObservation(override, observation)
}

// ObserveInterface reads one host interface and returns a portable observation
// assembled from net.Interface, sysfs on Linux, ifconfig media details, and
// netstat counters when available.
func ObserveInterface(ifaceName string) (observe.PortObservation, []error) {
	ifaceName = strings.TrimSpace(ifaceName)
	out := observe.PortObservation{Interface: ifaceName}
	if ifaceName == "" {
		return out, []error{errors.New("interface name is required")}
	}
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return out, []error{err}
	}
	if len(iface.HardwareAddr) > 0 {
		out.MAC = iface.HardwareAddr.String()
	}
	up := iface.Flags&net.FlagUp != 0
	out.Up = &up
	out.IP, out.Netmask = firstInterfaceIPv4(iface)

	var errs []error
	if runtime.GOOS == "linux" {
		stats, err := observe.ReadInterfaceStats("/sys", ifaceName)
		if err != nil {
			errs = append(errs, err)
		}
		out.Stats = stats
		out.SpeedMbps = stats.SpeedMbps
	}
	details := readHostInterfaceDetails(ifaceName)
	if details.Up != nil {
		out.Up = cloneBoolPointer(details.Up)
	}
	if out.SpeedMbps <= 0 && details.Speed > 0 {
		out.SpeedMbps = details.Speed
	}
	if strings.TrimSpace(out.Media) == "" {
		out.Media = details.Media
	}
	counters, ok := readHostInterfaceCounters(ifaceName)
	if ok {
		out.Stats = mergeInterfaceStats(out.Stats, counters)
	}
	return out, errs
}

// ApplyObservation overlays one portable interface observation onto a port
// override without replacing operator-specified values.
func ApplyObservation(override *device.PortOverride, observation observe.PortObservation) {
	if strings.TrimSpace(override.Interface) == "" {
		override.Interface = strings.TrimSpace(observation.Interface)
	}
	if strings.TrimSpace(override.MAC) == "" {
		override.MAC = observation.MAC
	}
	if override.Up == nil {
		override.Up = cloneBoolPointer(observation.Up)
	}
	if strings.TrimSpace(override.IP) == "" {
		override.IP = observation.IP
	}
	if strings.TrimSpace(override.Netmask) == "" {
		override.Netmask = observation.Netmask
	}
	if override.Speed <= 0 && observation.SpeedMbps > 0 {
		override.Speed = observation.SpeedMbps
	}
	if strings.TrimSpace(override.Media) == "" {
		override.Media = observation.Media
	}
	for _, field := range interfaceCounterFields {
		field.setOverride(override, field.get(observation.Stats))
	}
}

// cloneBoolPointer preserves the difference between unknown and explicit link
// state.
func cloneBoolPointer(value *bool) *bool {
	if value == nil {
		return nil
	}
	clone := *value
	return &clone
}

// firstInterfaceIPv4 returns the first IPv4 address and netmask for iface.
func firstInterfaceIPv4(iface *net.Interface) (string, string) {
	addrs, err := iface.Addrs()
	if err != nil {
		log.Printf("read addresses for interface %s: %v", iface.Name, err)
		return "", ""
	}
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}
		ip := ipNet.IP.To4()
		if ip == nil {
			continue
		}
		return ip.String(), net.IP(ipNet.Mask).String()
	}
	return "", ""
}

// readHostInterfaceDetails reads platform link details through ifconfig.
func readHostInterfaceDetails(ifaceName string) hostInterfaceDetails {
	out, err := exec.Command("ifconfig", ifaceName).Output()
	if err != nil {
		return hostInterfaceDetails{}
	}
	return parseIfconfigDetails(string(out))
}

// parseIfconfigDetails extracts link state, speed, and media from ifconfig output.
func parseIfconfigDetails(output string) hostInterfaceDetails {
	var details hostInterfaceDetails
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		lower := strings.ToLower(line)
		switch {
		case strings.HasPrefix(lower, "status:"):
			up := strings.Contains(lower, "active") || strings.Contains(lower, "up")
			details.Up = &up
		case strings.HasPrefix(lower, "media:"):
			details.Speed = speedFromMediaLine(lower)
			details.Media = mediaFromMediaLine(lower)
		}
	}
	return details
}

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

// speedFromMediaLine maps ifconfig media text to Mbps.
func speedFromMediaLine(line string) int {
	switch {
	case strings.Contains(line, "25g"):
		return 25000
	case strings.Contains(line, "10g"):
		return 10000
	case strings.Contains(line, "5g"):
		return 5000
	case strings.Contains(line, "2.5g"), strings.Contains(line, "2500base"):
		return 2500
	case strings.Contains(line, "1000base"), strings.Contains(line, "1g"):
		return 1000
	case strings.Contains(line, "100base"):
		return 100
	case strings.Contains(line, "10base"):
		return 10
	default:
		return 0
	}
}

// mediaFromMediaLine maps ifconfig media text to UniFi media labels.
func mediaFromMediaLine(line string) string {
	switch {
	case strings.Contains(line, "sfp28"), strings.Contains(line, "25gbase"):
		return "SFP28"
	case strings.Contains(line, "sfp+"),
		strings.Contains(line, "twinax"),
		strings.Contains(line, "10gbase-sr"),
		strings.Contains(line, "10gbase-lr"),
		strings.Contains(line, "10gbase-lrm"),
		strings.Contains(line, "10gbase-er"),
		strings.Contains(line, "10gbase-cr"),
		strings.Contains(line, "10gbase-cu"):
		return "SFP+"
	case speedFromMediaLine(line) > 0:
		return "GE"
	default:
		return ""
	}
}

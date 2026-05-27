// Package ifsource maps host interface metadata into payload port overrides.
package ifsource

// Interface sources convert host NIC state into controller-facing port
// overrides for port-map and explicit port_overrides[].interface entries.

import (
	"errors"
	"log"
	"net"
	"runtime"
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/device"
	"github.com/konstruktor1/unifi-stubd/internal/observe"
)

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

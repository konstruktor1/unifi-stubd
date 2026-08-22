// Package ifsource maps host interface metadata into payload port overrides.
package ifsource

import (
	"context"
	"errors"
	"net"
	"runtime"
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/observe"
)

// ObserveInterface reads one host interface and returns a portable observation
// assembled from net.Interface, sysfs on Linux, ifconfig media details, and
// netstat counters when available.
func ObserveInterface(ctx context.Context, ifaceName string) (observe.PortObservation, []error) {
	ifaceName = strings.TrimSpace(ifaceName)
	out := observe.PortObservation{Interface: ifaceName}
	if ifaceName == "" {
		return out, []error{errors.New("interface name is required")}
	}
	if err := ctx.Err(); err != nil {
		return out, []error{err}
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
	out.IP, out.Netmask, out.IPv6 = interfaceAddresses(iface)

	var errs []error
	if runtime.GOOS == "linux" {
		stats, err := observe.ReadInterfaceStats("/sys", ifaceName)
		if err != nil {
			errs = append(errs, err)
		}
		out.Stats = stats
		out.Up = stats.Up
		out.SpeedMbps = stats.SpeedMbps
	}
	details := readHostInterfaceDetails(ctx, ifaceName)
	if err := ctx.Err(); err != nil {
		return out, append(errs, err)
	}
	if details.Up != nil {
		out.Up = details.Up
	}
	if out.SpeedMbps <= 0 && details.Speed > 0 {
		out.SpeedMbps = details.Speed
	}
	if strings.TrimSpace(out.Media) == "" {
		out.Media = details.Media
	}
	counters, ok := readHostInterfaceCounters(ctx, ifaceName)
	if err := ctx.Err(); err != nil {
		return out, append(errs, err)
	}
	if ok {
		out.Stats = mergeInterfaceStats(out.Stats, counters)
	}
	if out.Up == nil || !*out.Up {
		out.SpeedMbps = 0
		out.Stats.SpeedMbps = 0
	}
	return out, errs
}

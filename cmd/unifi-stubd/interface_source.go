package main

// This file maps host interface data into controller-facing port overrides.

import (
	"log"
	"net"
	"os/exec"
	"strconv"
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/device"
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

// enrichPortOverridesFromInterfaces overlays configured ports with host interface data.
func enrichPortOverridesFromInterfaces(overrides []device.PortOverride) []device.PortOverride {
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
		enrichPortOverrideFromInterface(&out[index], ifaceName)
	}
	return out
}

// enrichPortOverrideFromInterface applies one host interface snapshot to an override.
func enrichPortOverrideFromInterface(override *device.PortOverride, ifaceName string) {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		log.Printf("port %d interface source %s unavailable: %v", override.Port, ifaceName, err)
		return
	}
	if strings.TrimSpace(override.MAC) == "" && len(iface.HardwareAddr) > 0 {
		override.MAC = iface.HardwareAddr.String()
	}
	if override.Up == nil {
		up := iface.Flags&net.FlagUp != 0
		override.Up = &up
	}
	if strings.TrimSpace(override.IP) == "" || strings.TrimSpace(override.Netmask) == "" {
		ip, netmask := firstInterfaceIPv4(iface)
		if strings.TrimSpace(override.IP) == "" {
			override.IP = ip
		}
		if strings.TrimSpace(override.Netmask) == "" {
			override.Netmask = netmask
		}
	}
	details := readHostInterfaceDetails(ifaceName)
	if details.Up != nil {
		override.Up = cloneBoolPointer(details.Up)
	}
	if override.Speed <= 0 && details.Speed > 0 {
		override.Speed = details.Speed
	}
	if strings.TrimSpace(override.Media) == "" {
		override.Media = details.Media
	}
	enrichPortOverrideCounters(override, ifaceName)
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

// enrichPortOverrideCounters overlays counters from netstat when available.
func enrichPortOverrideCounters(override *device.PortOverride, ifaceName string) {
	out, err := exec.Command("netstat", "-ibn", "-I", ifaceName).Output()
	if err != nil {
		return
	}
	counters, ok := parseNetstatCounters(string(out), ifaceName)
	if !ok {
		return
	}
	override.RXPackets = counters.RXPackets
	override.TXPackets = counters.TXPackets
	override.RXBytes = counters.RXBytes
	override.TXBytes = counters.TXBytes
	override.RXErrors = counters.RXErrors
	override.TXErrors = counters.TXErrors
}

// interfaceCounters contains packet and byte counters parsed from netstat.
type interfaceCounters struct {
	// RXBytes is the received byte counter.
	RXBytes int64
	// TXBytes is the transmitted byte counter.
	TXBytes int64
	// RXPackets is the received packet counter.
	RXPackets int64
	// TXPackets is the transmitted packet counter.
	TXPackets int64
	// RXErrors is the receive error counter.
	RXErrors int64
	// TXErrors is the transmit error counter.
	TXErrors int64
}

// parseNetstatCounters extracts counters for one link-layer interface row.
func parseNetstatCounters(output, ifaceName string) (interfaceCounters, bool) {
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 11 || fields[0] != ifaceName || !strings.HasPrefix(fields[2], "<Link#") {
			continue
		}
		values := make([]int64, 6)
		for index, fieldIndex := range []int{4, 5, 7, 8, 9, 10} {
			value, err := strconv.ParseInt(fields[fieldIndex], 10, 64)
			if err != nil {
				return interfaceCounters{}, false
			}
			values[index] = value
		}
		return interfaceCounters{
			RXPackets: values[0],
			RXErrors:  values[1],
			RXBytes:   values[2],
			TXPackets: values[3],
			TXErrors:  values[4],
			TXBytes:   values[5],
		}, true
	}
	return interfaceCounters{}, false
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

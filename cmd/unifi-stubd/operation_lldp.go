package main

import (
	"context"
	"log"
	"net"
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/device"
	"github.com/konstruktor1/unifi-stubd/internal/platform"
)

// applyLLDPNeighbors adds passive LLDP neighbors as MAC-table hints on the
// matching represented UniFi ports.
func applyLLDPNeighbors(ports []device.Port, flags runtimeFlags, plt platform.Platform) []device.Port {
	if strings.TrimSpace(flags.lldpSource) == "" || strings.EqualFold(strings.TrimSpace(flags.lldpSource), platform.SourceOff) {
		return ports
	}
	ctx, cancel := context.WithTimeout(context.Background(), observeTimeout)
	defer cancel()
	neighbors, errs := plt.LLDP(ctx, platform.LLDPConfig{Source: flags.lldpSource, Timeout: observeTimeout})
	for _, err := range errs {
		log.Printf("lldp observation warning: %v", err)
	}
	if len(neighbors) == 0 {
		return ports
	}
	out := make([]device.Port, len(ports))
	copy(out, ports)
	portByInterface := lldpInterfacePortMap(flags, out)
	for _, neighbor := range neighbors {
		portIndex := portByInterface[strings.ToLower(strings.TrimSpace(neighbor.Interface))]
		if portIndex < 1 || portIndex > len(out) {
			continue
		}
		entry := lldpNeighborMACEntry(neighbor)
		if strings.TrimSpace(entry.MAC) == "" {
			continue
		}
		// LLDP neighbors are represented only as controller-facing MAC-table
		// hints. They are never used to configure host networking.
		out[portIndex-1].MACs = append(out[portIndex-1].MACs, entry)
	}
	return out
}

// lldpInterfacePortMap maps observed interface names back to represented port
// indexes using explicit bridge, port-map, and override bindings.
func lldpInterfacePortMap(flags runtimeFlags, ports []device.Port) map[string]int {
	out := map[string]int{}
	bridgeObserve := effectiveBridgeObserve(flags)
	if iface := strings.ToLower(strings.TrimSpace(bridgeObserve.UplinkInterface)); iface != "" {
		out[iface] = uplinkPortIndex(ports)
	}
	for _, mapping := range bridgeObserve.MemberPortMap {
		if iface := strings.ToLower(strings.TrimSpace(mapping.Member)); iface != "" {
			out[iface] = mapping.Port
		}
	}
	for _, mapping := range flags.portMappings {
		if iface := strings.ToLower(strings.TrimSpace(mapping.Interface)); iface != "" {
			out[iface] = mapping.Port
		}
	}
	for _, override := range flags.portOverrides {
		if iface := strings.ToLower(strings.TrimSpace(override.Interface)); iface != "" {
			out[iface] = override.Port
		}
	}
	return out
}

// lldpNeighborMACEntry turns one LLDP neighbor into the same MAC-table metadata
// shape used for configured neighbors.
func lldpNeighborMACEntry(neighbor platform.LLDPNeighbor) device.MacTableEntry {
	mac := strings.TrimSpace(neighbor.ChassisMAC)
	if mac == "" {
		mac = strings.TrimSpace(neighbor.ChassisID)
	}
	if parsed, err := net.ParseMAC(mac); err == nil {
		mac = parsed.String()
	} else {
		return device.MacTableEntry{}
	}
	return device.MacTableEntry{
		MAC:      mac,
		Hostname: strings.TrimSpace(neighbor.SystemName),
		IP:       ipv4Text(neighbor.ManagementIP),
		Age:      4,
		Uptime:   1200,
		Type:     "lldp",
	}
}

// ipv4Text keeps LLDP management addresses limited to IPv4 strings accepted by
// the UniFi MAC-table payload.
func ipv4Text(value string) string {
	ip := net.ParseIP(strings.TrimSpace(value))
	if ip == nil || ip.To4() == nil {
		return ""
	}
	return ip.String()
}

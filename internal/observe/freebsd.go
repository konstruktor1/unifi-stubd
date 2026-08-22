package observe

import (
	"net"
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/adapters/freebsdifconfig"
	"github.com/konstruktor1/unifi-stubd/internal/device"
)

// FreeBSDMACsByInterface converts learned FreeBSD bridge rows into
// per-member UniFi MAC entries, filtering local and multicast rows.
func FreeBSDMACsByInterface(entries []freebsdifconfig.BridgeAddress) map[string][]device.MacTableEntry {
	out := map[string][]device.MacTableEntry{}
	seen := map[string]bool{}
	for _, entry := range entries {
		if !learnedFreeBSDBridgeEntry(entry) {
			continue
		}
		iface := strings.TrimSpace(entry.Interface)
		key := iface + "|" + entry.MAC
		if seen[key] {
			continue
		}
		seen[key] = true
		out[iface] = append(out[iface], device.MacTableEntry{
			MAC:    entry.MAC,
			Age:    defaultBridgeAge(entry.Age),
			Uptime: 1200,
			VLAN:   entry.VLAN,
			Type:   "client",
		})
	}
	return out
}

// learnedFreeBSDBridgeEntry accepts only non-local unicast rows that can
// represent downstream clients.
func learnedFreeBSDBridgeEntry(entry freebsdifconfig.BridgeAddress) bool {
	mac, err := net.ParseMAC(entry.MAC)
	if err != nil || len(mac) == 0 {
		return false
	}
	if mac[0]&0x01 != 0 {
		return false
	}
	if entry.Local {
		return false
	}
	return strings.TrimSpace(entry.Interface) != ""
}

// defaultBridgeAge supplies a fresh MAC-table age when FreeBSD output omits
// one.
func defaultBridgeAge(value int) int {
	if value > 0 {
		return value
	}
	return 4
}

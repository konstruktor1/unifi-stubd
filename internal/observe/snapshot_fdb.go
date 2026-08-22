package observe

import (
	"net"
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/adapters/linuxbridge"
	"github.com/konstruktor1/unifi-stubd/internal/device"
)

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

// learnedFDBEntry accepts only non-local unicast FDB rows that can represent
// downstream clients.
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

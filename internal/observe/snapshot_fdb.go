package observe

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/adapters/linuxbridge"
	"github.com/konstruktor1/unifi-stubd/internal/device"
)

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

package observe

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/device"
)

// ARPEntry is one read-only row from a local ARP/neighbor cache.
type ARPEntry struct {
	IP     string
	MAC    string
	Device string
}

// ReadARPTable reads Linux procfs ARP rows from path. ARP is only used to
// enrich observed MAC-table metadata with client IPs.
func ReadARPTable(path string) ([]ARPEntry, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		path = "/proc/net/arp"
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("read ARP table %s: %w", path, err)
	}
	defer func() {
		_ = file.Close()
	}()
	return ParseARPTable(file), nil
}

// ParseARPTable converts Linux /proc/net/arp content into normalized unicast
// IPv4 rows, filtering headers, invalid MACs, multicast, and zero addresses.
func ParseARPTable(reader io.Reader) []ARPEntry {
	scanner := bufio.NewScanner(reader)
	var out []ARPEntry
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 6 || strings.EqualFold(fields[0], "IP") {
			continue
		}
		ip := net.ParseIP(fields[0])
		if ip == nil || ip.To4() == nil {
			continue
		}
		mac, err := net.ParseMAC(fields[3])
		if err != nil || len(mac) == 0 || mac[0]&0x01 != 0 || zeroMAC(mac) {
			continue
		}
		out = append(out, ARPEntry{
			IP:     ip.String(),
			MAC:    strings.ToLower(mac.String()),
			Device: strings.TrimSpace(fields[5]),
		})
	}
	return out
}

// zeroMAC filters placeholder ARP rows that cannot represent real clients.
func zeroMAC(mac net.HardwareAddr) bool {
	for _, part := range mac {
		if part != 0 {
			return false
		}
	}
	return true
}

// EnrichMACEntriesWithARP fills missing client IPs from local ARP rows while
// preserving any IPs already supplied by configuration or another observation.
func EnrichMACEntriesWithARP(memberMACs map[string][]device.MacTableEntry, arpEntries []ARPEntry) {
	if len(memberMACs) == 0 || len(arpEntries) == 0 {
		return
	}
	byMAC := map[string][]ARPEntry{}
	for _, entry := range arpEntries {
		key := normalizedMACKey(entry.MAC)
		if key == "" || strings.TrimSpace(entry.IP) == "" {
			continue
		}
		byMAC[key] = append(byMAC[key], entry)
	}
	for member, macs := range memberMACs {
		for index := range macs {
			if strings.TrimSpace(macs[index].IP) != "" {
				continue
			}
			arp, ok := arpEntryForMAC(member, macs[index].MAC, byMAC)
			if !ok {
				continue
			}
			macs[index].IP = arp.IP
		}
		memberMACs[member] = macs
	}
}

// EnrichMACEntriesWithLocalARP fills missing client IPs from the host ARP cache.
func EnrichMACEntriesWithLocalARP(memberMACs map[string][]device.MacTableEntry) error {
	if len(memberMACs) == 0 {
		return nil
	}
	entries, err := ReadARPTable("")
	if err != nil {
		return err
	}
	EnrichMACEntriesWithARP(memberMACs, entries)
	return nil
}

// arpEntryForMAC prefers an ARP row learned on the same bridge member, then
// falls back to any row for that MAC.
func arpEntryForMAC(member, mac string, byMAC map[string][]ARPEntry) (ARPEntry, bool) {
	entries := byMAC[normalizedMACKey(mac)]
	if len(entries) == 0 {
		return ARPEntry{}, false
	}
	member = strings.ToLower(strings.TrimSpace(member))
	for _, entry := range entries {
		if strings.ToLower(strings.TrimSpace(entry.Device)) == member {
			return entry, true
		}
	}
	return entries[0], true
}

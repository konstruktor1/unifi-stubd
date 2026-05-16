package linuxbridge

import (
	"bufio"
	"io"
	"strconv"
	"strings"
)

// FDBEntry represents one Linux bridge forwarding database row.
type FDBEntry struct {
	// MAC is the learned forwarding database MAC address.
	MAC string
	// Device is the bridge member interface that owns the entry.
	Device string
	// VLAN is the optional VLAN identifier parsed from the row.
	VLAN int
}

// ParseFDB parses bridge fdb output into forwarding database entries.
func ParseFDB(r io.Reader) []FDBEntry {
	var entries []FDBEntry
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 0 || !strings.Contains(fields[0], ":") {
			continue
		}
		entry := FDBEntry{MAC: strings.ToLower(fields[0])}
		for i := 1; i < len(fields)-1; i++ {
			switch fields[i] {
			case "dev":
				entry.Device = fields[i+1]
			case "vlan":
				if vlan, err := strconv.Atoi(fields[i+1]); err == nil {
					entry.VLAN = vlan
				}
			}
		}
		entries = append(entries, entry)
	}
	return entries
}

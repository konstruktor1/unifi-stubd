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
	// Dynamic reports whether the row is a dynamic learned entry.
	Dynamic bool
	// Static reports whether the row is a static entry.
	Static bool
	// Local reports whether the row terminates locally on the host.
	Local bool
	// Permanent reports whether the row is permanent.
	Permanent bool
	// Self reports whether the row belongs to the device rather than the bridge master.
	Self bool
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
		for _, field := range fields[1:] {
			switch field {
			case "dynamic":
				entry.Dynamic = true
			case "static":
				entry.Static = true
			case "local":
				entry.Local = true
			case "permanent":
				entry.Permanent = true
			case "self":
				entry.Self = true
			}
		}
		entries = append(entries, entry)
	}
	return entries
}

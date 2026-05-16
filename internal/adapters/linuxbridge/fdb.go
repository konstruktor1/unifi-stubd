package linuxbridge

import (
	"bufio"
	"io"
	"strconv"
	"strings"
)

type FDBEntry struct {
	MAC    string
	Device string
	VLAN   int
}

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

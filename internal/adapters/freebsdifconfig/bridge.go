// Package freebsdifconfig parses FreeBSD ifconfig output used by observation adapters.
package freebsdifconfig

import (
	"bufio"
	"io"
	"net"
	"strconv"
	"strings"
)

// BridgeAddress is one learned FreeBSD bridge forwarding entry.
type BridgeAddress struct {
	// MAC is the learned client MAC address.
	MAC string
	// VLAN is the optional VLAN ID reported as VlanN.
	VLAN int
	// Interface is the bridge member interface that learned the MAC.
	Interface string
	// Age is the optional FreeBSD bridge age/expires value.
	Age int
	// Static reports whether ifconfig marked the entry static.
	Static bool
	// Local reports whether ifconfig marked the entry local.
	Local bool
}

// ParseBridgeAddr parses `ifconfig <bridge> addr` output.
func ParseBridgeAddr(r io.Reader) []BridgeAddress {
	var entries []BridgeAddress
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}
		mac := strings.ToLower(strings.TrimSpace(fields[0]))
		if _, err := net.ParseMAC(mac); err != nil {
			continue
		}

		entry := BridgeAddress{MAC: mac}
		index := 1
		if strings.HasPrefix(strings.ToLower(fields[index]), "vlan") {
			entry.VLAN = parseVLAN(fields[index])
			index++
		}
		if index >= len(fields) {
			continue
		}
		entry.Interface = strings.TrimSpace(fields[index])
		index++
		if index < len(fields) {
			entry.Age = parseOptionalInt(fields[index])
		}
		for ; index < len(fields); index++ {
			flag := strings.ToLower(fields[index])
			if strings.Contains(flag, "static") {
				entry.Static = true
			}
			if strings.Contains(flag, "local") || strings.Contains(flag, "self") {
				entry.Local = true
			}
		}
		entries = append(entries, entry)
	}
	return entries
}

// parseVLAN extracts numeric VLAN IDs from FreeBSD VlanN bridge tokens.
func parseVLAN(value string) int {
	value = strings.TrimPrefix(strings.ToLower(strings.TrimSpace(value)), "vlan")
	return parseOptionalInt(value)
}

// parseOptionalInt tolerates absent or non-numeric FreeBSD bridge fields.
func parseOptionalInt(value string) int {
	value = strings.Trim(strings.TrimSpace(value), ",")
	number, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return number
}

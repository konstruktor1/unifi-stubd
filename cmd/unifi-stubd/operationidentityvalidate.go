package main

import (
	"fmt"
	"net"
	"strings"
)

// validateIdentityFlags checks only locally supplied identity values; controller
// adoption responses are not allowed to redefine the host-facing identity.
func validateIdentityFlags(flags runtimeFlags) error {
	if ip := net.ParseIP(strings.TrimSpace(flags.ipText)).To4(); ip == nil {
		return fmt.Errorf("invalid IPv4 address: %q", flags.ipText)
	}
	macText := strings.TrimSpace(flags.macText)
	if macText == "" || strings.EqualFold(macText, automaticText) || strings.EqualFold(macText, "host") {
		return nil
	}
	if _, err := net.ParseMAC(macText); err != nil {
		return fmt.Errorf("invalid MAC address: %w", err)
	}
	return nil
}

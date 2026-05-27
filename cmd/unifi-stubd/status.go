// Runtime status reports identity, configuration, adoption, observation, and
// last-inform state without exposing authkeys. The human and JSON outputs share
// the same sanitized status document.
package main

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/konstruktor1/unifi-stubd/internal/device"
	"github.com/konstruktor1/unifi-stubd/internal/platform"
)

// printLocalStatus renders the sanitized status document in either machine JSON
// or the line-oriented human format used by lab scripts.
func printLocalStatus(flags runtimeFlags, profile device.Profile, mac net.HardwareAddr, ip net.IP, hostname string, portBuildOptions device.PortBuildOptions, plt platform.Platform) error {
	status := buildLocalStatus(flags, profile, mac, ip, hostname, portBuildOptions, plt)
	if flags.statusJSON {
		data, err := json.MarshalIndent(status, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal local status: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}
	printHumanStatus(status)
	return nil
}

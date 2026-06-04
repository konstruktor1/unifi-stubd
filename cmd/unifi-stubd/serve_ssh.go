package main

import (
	"fmt"
	"net"

	"github.com/konstruktor1/unifi-stubd/internal/adoptionssh"
)

// startAdoptionSSH wires the optional SSH compatibility shim to the current
// fake identity and adoption store path.
func startAdoptionSSH(flags runtimeFlags, mac net.HardwareAddr, ip net.IP, hostname string) (*adoptionssh.Server, error) {
	sshServer, err := adoptionssh.Start(adoptionssh.Config{
		Listen:      flags.sshListen,
		User:        flags.sshUser,
		Password:    flags.sshPassword,
		HostKeyPath: flags.sshHostKey,
		StatePath:   flags.sshState,
		Identity: adoptionssh.Identity{
			MAC:       mac.String(),
			IP:        ip.String(),
			Hostname:  hostname,
			Model:     flags.model,
			Version:   flags.version,
			InformURL: flags.controller,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("adoption ssh failed: %w", err)
	}
	return sshServer, nil
}

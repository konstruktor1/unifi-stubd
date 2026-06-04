package main

import (
	"fmt"
	"net"

	"github.com/konstruktor1/unifi-stubd/internal/device"
	"github.com/konstruktor1/unifi-stubd/internal/platform"
)

type serveIdentity struct {
	mac              net.HardwareAddr
	ip               net.IP
	hostname         string
	portBuildOptions device.PortBuildOptions
}

func resolveServeIdentity(flags runtimeFlags, profile device.Profile, plt platform.Platform) (serveIdentity, error) {
	hostname := resolveHostname(flags.hostname)
	uplinkPort := effectiveUplinkPort(profile, flags)
	portBuildOptions := resolvePortBuildOptions(flags.portCount, flags.linkSpeed, uplinkPort, effectiveUplinkSpeedMode(flags), flags.controller)
	mac := resolveMAC(flags.macText, hostname, profile, flags.model, flags.operationMode, flags.observeInterface)
	ip := net.ParseIP(flags.ipText).To4()
	if ip == nil {
		return serveIdentity{}, fmt.Errorf("invalid IPv4 address: %q", flags.ipText)
	}
	ip, err := resolveManagementIP(flags, ip, plt)
	if err != nil {
		return serveIdentity{}, err
	}
	return serveIdentity{
		mac:              mac,
		ip:               ip,
		hostname:         hostname,
		portBuildOptions: portBuildOptions,
	}, nil
}

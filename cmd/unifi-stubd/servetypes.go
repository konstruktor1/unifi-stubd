package main

import (
	"net"
	"time"

	"github.com/konstruktor1/unifi-stubd/internal/device"
	"github.com/konstruktor1/unifi-stubd/internal/discovery"
)

// controllerPresence bundles the immutable daemon identity and mutable runtime
// trackers used by the heartbeat loop.
type controllerPresence struct {
	flags              runtimeFlags
	profile            device.Profile
	mac                net.HardwareAddr
	ip                 net.IP
	hostname           string
	portBuildOptions   device.PortBuildOptions
	announcement       discovery.Announcement
	discoveryPacket    []byte
	discoverySkipped   bool
	discoveryInterface string
	discoveryTargets   []string
	startedAt          time.Time
}

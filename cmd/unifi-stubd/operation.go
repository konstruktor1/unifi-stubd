// Operation-mode validation is the safety gate between synthetic stubbing,
// read-only host observation, and planned host-network modes. Anything capable
// of mutating the host remains rejected or dry-run-only here.
package main

import "time"

// Operation modes define the host-observation boundary for the daemon.
const (
	operationModeStub          = "stub"
	operationModeObserve       = "observe"
	operationModeBridgeObserve = "bridge-observe"
	operationModePortMap       = "port-map"
	operationModeHostDirect    = "host-direct"
	operationModeMacvlan       = "macvlan"

	trafficSourceOff = "off"
	observeTimeout   = 2 * time.Second
)

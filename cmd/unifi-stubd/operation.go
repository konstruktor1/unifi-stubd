package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/konstruktor1/unifi-stubd/internal/device"
	"github.com/konstruktor1/unifi-stubd/internal/observe"
)

const (
	operationModeStub       = "stub"
	operationModeObserve    = "observe"
	operationModeHostDirect = "host-direct"
	operationModeMacvlan    = "macvlan"

	lldpSourceOff    = "off"
	lldpSourceLLDPD  = "lldpd"
	trafficSourceOff = "off"
	observeTimeout   = 2 * time.Second
)

func validateOperationFlags(flags runtimeFlags) error {
	mode := normalizeMode(*flags.operationMode)
	*flags.operationMode = mode
	switch mode {
	case operationModeStub, operationModeObserve, operationModeHostDirect, operationModeMacvlan:
	default:
		return fmt.Errorf("invalid -operation-mode %q; use stub, observe, host-direct, or macvlan", mode)
	}

	lldpSource := strings.ToLower(strings.TrimSpace(*flags.lldpSource))
	if lldpSource == "" {
		lldpSource = lldpSourceOff
	}
	*flags.lldpSource = lldpSource
	switch lldpSource {
	case lldpSourceOff, lldpSourceLLDPD:
	default:
		return fmt.Errorf("invalid -lldp-source %q; use off or lldpd", lldpSource)
	}

	trafficSource := strings.ToLower(strings.TrimSpace(*flags.trafficSource))
	if trafficSource == "" {
		trafficSource = trafficSourceOff
	}
	*flags.trafficSource = trafficSource
	if trafficSource != trafficSourceOff {
		return fmt.Errorf("invalid -traffic-source %q; only off is implemented", trafficSource)
	}

	if strings.EqualFold(strings.TrimSpace(*flags.macText), "host") && mode != operationModeHostDirect {
		return fmt.Errorf("mac: host is only allowed with -operation-mode host-direct")
	}
	if mode == operationModeMacvlan && !*flags.dryRunPlan {
		return fmt.Errorf("operation-mode macvlan is planned only; use -dry-run-plan to inspect the non-mutating plan")
	}
	return nil
}

func normalizeMode(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return operationModeStub
	}
	return value
}

func portsForRuntime(flags runtimeFlags, portOptions device.PortOptions) []device.Port {
	ports := device.SwitchPortsWithOptions(*flags.portCount, portOptions)
	mode := normalizeMode(*flags.operationMode)
	if mode != operationModeObserve && mode != operationModeHostDirect {
		return ports
	}
	ctx, cancel := context.WithTimeout(context.Background(), observeTimeout)
	defer cancel()

	snapshot, errs := observe.LinuxSnapshot(ctx, observe.Config{
		Interface: strings.TrimSpace(*flags.observeInterface),
		Bridge:    strings.TrimSpace(*flags.observeBridge),
	}, uplinkPortIndex(ports))
	for _, err := range errs {
		log.Printf("passive observation warning: %v", err)
	}
	return observe.Apply(ports, snapshot)
}

func printRuntimePlan(flags runtimeFlags, profile device.Profile, macText, ipText, hostname string) {
	mode := normalizeMode(*flags.operationMode)
	fmt.Printf("operation_mode: %s\n", mode)
	fmt.Printf("profile: %s (%s)\n", profile.Name, profile.Model)
	fmt.Printf("mac: %s\n", macText)
	fmt.Printf("ip: %s\n", ipText)
	fmt.Printf("hostname: %s\n", hostname)
	fmt.Printf("observe_interface: %s\n", strings.TrimSpace(*flags.observeInterface))
	fmt.Printf("observe_bridge: %s\n", strings.TrimSpace(*flags.observeBridge))
	fmt.Printf("lldp_source: %s\n", strings.TrimSpace(*flags.lldpSource))
	fmt.Printf("traffic_source: %s\n", strings.TrimSpace(*flags.trafficSource))
	switch mode {
	case operationModeStub:
		fmt.Println("actions: synthetic stub only; no host network changes")
	case operationModeObserve:
		fmt.Println("actions: read-only Linux sysfs/FDB observation; no host network changes")
	case operationModeHostDirect:
		fmt.Println("actions: direct host identity mode; no host network changes")
	case operationModeMacvlan:
		parent := strings.TrimSpace(*flags.observeInterface)
		if parent == "" {
			parent = "<required-parent-interface>"
		}
		fmt.Println("actions: macvlan is not executed by this release")
		fmt.Printf("planned_command: ip link add link %s name unifi-stubd0 type macvlan mode bridge\n", parent)
		fmt.Printf("planned_command: ip link set unifi-stubd0 address %s up\n", macText)
		fmt.Printf("planned_note: assign %s to unifi-stubd0 after subnet/prefix config exists\n", ipText)
	}
}

func uplinkPortIndex(ports []device.Port) int {
	for _, port := range ports {
		if port.Uplink {
			return port.Index
		}
	}
	return 1
}

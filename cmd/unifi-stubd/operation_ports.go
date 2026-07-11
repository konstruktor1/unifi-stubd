package main

import (
	"context"
	"log"
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/device"
	"github.com/konstruktor1/unifi-stubd/internal/observe"
	"github.com/konstruktor1/unifi-stubd/internal/observe/portmap"
	"github.com/konstruktor1/unifi-stubd/internal/platform"
)

// portsForRuntime merges profile defaults, passive observations, LLDP hints,
// operator overrides, and configured neighbors into one ordered port list.
func portsForRuntime(flags runtimeFlags, profile device.Profile, portBuildOptions device.PortBuildOptions, plt platform.Platform) []device.Port {
	ctx, cancel := context.WithTimeout(context.Background(), observeTimeout)
	flags.portOverrides = enrichOverrides(ctx, plt, flags.portOverrides)
	cancel()

	ports := device.BuildPorts(profile, portBuildOptions)
	mode := normalizeMode(flags.operationMode)
	if mode == operationModePortMap {
		// Explicit port-map sources become ordinary overrides first, then user
		// overrides win. This preserves the operator's final say over observed
		// host data while keeping renderer code on one merge path.
		ctx, cancel := context.WithTimeout(context.Background(), observeTimeout)
		defer cancel()
		overrides, errs := portmap.OverridesFromSource(ctx, plt, flags.portMappings)
		for _, err := range errs {
			log.Printf("port-map observation warning: %v", err)
		}
		ports = device.ApplyPortOverrides(ports, overrides)
		ports = device.ApplyPortOverrides(ports, flags.portOverrides)
		ports = applyWANHealth(ports, flags, profile)
		ports = applyLLDPNeighbors(ports, flags, plt)
		ports = device.ApplyPortNeighbors(ports, flags.portNeighbors)
		return finalizeRuntimeTopology(device.ApplyUplinkNeighbor(ports, flags.uplinkNeighbor), profile)
	}
	if mode != operationModeBridgeObserve && mode != operationModeHostDirect {
		// Stub mode stays synthetic unless the operator explicitly supplies
		// payload metadata or enables passive LLDP. No host bridge or interface
		// data is guessed here.
		ports = device.ApplyPortOverrides(ports, flags.portOverrides)
		ports = applyWANHealth(ports, flags, profile)
		ports = applyLLDPNeighbors(ports, flags, plt)
		ports = device.ApplyPortNeighbors(ports, flags.portNeighbors)
		return finalizeRuntimeTopology(device.ApplyUplinkNeighbor(ports, flags.uplinkNeighbor), profile)
	}
	ctx, cancel = context.WithTimeout(context.Background(), observeTimeout)
	defer cancel()

	bridgeObserve := effectiveBridgeObserve(flags)
	snapshot, errs := observe.HostSnapshotFromSource(ctx, plt, observe.Config{
		Interface:      strings.TrimSpace(bridgeObserve.UplinkInterface),
		Bridge:         strings.TrimSpace(bridgeObserve.Bridge),
		IgnoredMembers: cloneStrings(bridgeObserve.IgnoredMembers),
		MemberPortMap:  bridgeMemberPortMap(bridgeObserve.MemberPortMap),
	}, uplinkPortIndex(ports))
	for _, err := range errs {
		log.Printf("passive observation warning: %v", err)
	}
	observedPorts := observe.Apply(ports, snapshot)
	// Bridge observation is read-only input. Operator overrides are applied
	// after the passive snapshot so a config file can correct or mask host facts
	// without the controller mutating the host.
	if flags.trafficRatesEnabled {
		observedPorts = markRateUplink(observedPorts, bridgeObserve.UplinkInterface)
	}
	ports = device.ApplyPortOverrides(observedPorts, flags.portOverrides)
	ports = applyWANHealth(ports, flags, profile)
	ports = applyLLDPNeighbors(ports, flags, plt)
	ports = device.ApplyPortNeighbors(ports, flags.portNeighbors)
	return finalizeRuntimeTopology(device.ApplyUplinkNeighbor(ports, flags.uplinkNeighbor), profile)
}

func finalizeRuntimeTopology(ports []device.Port, profile device.Profile) []device.Port {
	resolution, err := device.ResolveTopology(ports)
	if err != nil {
		log.Printf("topology resolution error: %v", err)
		return ports
	}
	for _, warning := range resolution.Warnings {
		log.Printf("topology warning: %s", warning)
	}
	for _, port := range ports {
		if port.Up && strings.TrimSpace(port.Interface) != "" && strings.TrimSpace(port.MAC) == "" {
			log.Printf("topology warning: active mapped port %d (%s) has no observed or configured port MAC; payload uses the profile fallback only as the local interface MAC", port.Index, port.Interface)
		}
	}
	if strings.EqualFold(strings.TrimSpace(profile.Payload.Kind), "gateway") {
		for _, port := range ports {
			if !port.Uplink {
				continue
			}
			role := strings.ToLower(strings.TrimSpace(port.Role))
			if role != "wan" && role != "wan2" {
				log.Printf("topology warning: gateway uplink port %d has role %q; direction still follows uplink_port", port.Index, role)
			}
			break
		}
	}
	return ports
}

// uplinkPortIndex finds the represented uplink and falls back to port 1 for
// sparse or synthetic profiles.
func uplinkPortIndex(ports []device.Port) int {
	for _, port := range ports {
		if port.Uplink {
			return port.Index
		}
	}
	return 1
}

// markRateUplink preserves the bridge-observe uplink interface
// name so rate tracking has a stable key on later heartbeats.
func markRateUplink(ports []device.Port, iface string) []device.Port {
	iface = strings.TrimSpace(iface)
	if iface == "" || len(ports) == 0 {
		return ports
	}
	index := uplinkPortIndex(ports)
	if index < 1 || index > len(ports) || strings.TrimSpace(ports[index-1].Interface) != "" {
		return ports
	}
	out := make([]device.Port, len(ports))
	copy(out, ports)
	out[index-1].Interface = iface
	return out
}

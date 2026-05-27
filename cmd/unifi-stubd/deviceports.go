package main

import (
	"log"
	"strconv"
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/device"
)

// resolvePortBuildOptions applies runtime port-count, link-speed, and uplink
// selections before any runtime observation is merged.
func resolvePortBuildOptions(portCount int, linkSpeed int, uplinkPort int, uplinkSpeed, controller string) device.PortBuildOptions {
	options := device.PortBuildOptions{
		Count:      portCount,
		LinkSpeed:  linkSpeed,
		UplinkPort: uplinkPort,
	}
	return resolveUplinkSpeed(options, uplinkSpeed, controller)
}

// resolveUplinkSpeed handles the operator's uplink-speed policy, including the
// optional egress-link probe used only for local observation.
func resolveUplinkSpeed(options device.PortBuildOptions, value, target string) device.PortBuildOptions {
	value = strings.TrimSpace(strings.ToLower(value))
	switch value {
	case "", "profile":
		return options
	case automaticText:
		info, err := device.DetectEgressLink(target)
		if err != nil {
			log.Printf("uplink speed auto-detect failed: %v; using profile uplink speed", err)
			return options
		}
		options.UplinkSpeed = info.SpeedMbps
		log.Printf("uplink speed auto-detected: interface=%s local_ip=%s speed=%d Mbps", info.Interface, info.LocalIP, info.SpeedMbps)
		return options
	default:
		speed, err := strconv.Atoi(value)
		if err != nil || speed <= 0 {
			log.Fatalf("invalid -uplink-speed %q; use auto, profile, or a positive Mbps value", value)
		}
		options.UplinkSpeed = speed
		return options
	}
}

// effectiveUplinkSpeedMode disables egress probing when bridge observation has
// an explicit uplink interface, because that interface is already the source of
// link metadata.
func effectiveUplinkSpeedMode(flags runtimeFlags) string {
	value := strings.TrimSpace(flags.uplinkSpeed)
	if !strings.EqualFold(value, automaticText) {
		return value
	}
	if normalizeMode(flags.operationMode) != operationModeBridgeObserve {
		return value
	}
	if strings.TrimSpace(effectiveBridgeObserve(flags).UplinkInterface) == "" {
		return value
	}
	return "profile"
}

// effectiveUplinkPort keeps bridge-observe from defaulting a represented host
// uplink onto an SFP/SFP+ profile port when a safer copper/access port exists.
func effectiveUplinkPort(profile device.Profile, flags runtimeFlags) int {
	if flags.uplinkPort > 0 {
		return flags.uplinkPort
	}
	if normalizeMode(flags.operationMode) != operationModeBridgeObserve {
		return flags.uplinkPort
	}
	if strings.TrimSpace(effectiveBridgeObserve(flags).UplinkInterface) == "" {
		return flags.uplinkPort
	}
	if !strings.EqualFold(strings.TrimSpace(profile.Payload.Kind), "switch") {
		return flags.uplinkPort
	}
	ports := device.BuildPorts(profile, device.PortBuildOptions{})
	defaultUplinkMedia := ""
	for _, port := range ports {
		if port.Uplink {
			defaultUplinkMedia = strings.TrimSpace(port.Media)
			break
		}
	}
	if defaultUplinkMedia == "" || strings.EqualFold(defaultUplinkMedia, "GE") {
		return flags.uplinkPort
	}
	candidate := 0
	fallback := 0
	for _, port := range ports {
		if port.Uplink {
			continue
		}
		fallback = port.Index
		if strings.EqualFold(strings.TrimSpace(port.Media), "GE") {
			candidate = port.Index
		}
	}
	if candidate == 0 {
		candidate = fallback
	}
	return candidate
}

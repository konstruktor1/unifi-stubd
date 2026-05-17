package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/adoption"
	"github.com/konstruktor1/unifi-stubd/internal/device"
)

const automaticText = "auto"

func payloadForIdentity(
	mac net.HardwareAddr,
	ip net.IP,
	hostname string,
	informURL string,
	store adoption.Store,
	flags runtimeFlags,
	profile device.Profile,
	ports []device.Port,
) ([]byte, error) {
	return buildPayload(device.Identity{
		MAC:          mac.String(),
		IP:           ip.String(),
		Hostname:     hostname,
		Model:        *flags.model,
		ModelDisplay: *flags.modelDisplay,
		DeviceType:   profile.DeviceType,
		Version:      *flags.version,
		Serial:       serialFromMAC(mac),
		InformURL:    informURL,
	}, store, ports)
}

func buildPayload(id device.Identity, store adoption.Store, ports []device.Port) ([]byte, error) {
	id.CFGVersion = store.CFGVersion
	id.Adopted = store.AuthKey != ""
	if store.Version != "" {
		id.Version = store.Version
	}
	payload, err := device.MinimalSwitchPayload(id, ports)
	if err != nil {
		return nil, fmt.Errorf("build device payload: %w", err)
	}
	return payload, nil
}

func resolveHostname(value string) string {
	value = strings.TrimSpace(value)
	if value != "" && strings.ToLower(value) != automaticText {
		return value
	}
	host, err := os.Hostname()
	if err == nil && strings.TrimSpace(host) != "" {
		return strings.TrimSpace(host)
	}
	return "unifi-stubd"
}

func resolveMAC(value, hostname string, profile device.Profile, model, operationMode, ifaceName string) net.HardwareAddr {
	value = strings.TrimSpace(value)
	if strings.EqualFold(value, "host") {
		if operationMode != operationModeHostDirect {
			log.Fatalf("mac: host is only allowed with -operation-mode host-direct")
		}
		mac, err := hostInterfaceMAC(ifaceName)
		if err != nil {
			log.Fatalf("host MAC resolve failed: %v", err)
		}
		log.Printf("host MAC resolved: %s interface=%s", mac, ifaceName)
		return mac
	}
	if value == "" || strings.EqualFold(value, automaticText) {
		seed := strings.Join([]string{"unifi-stubd", hostname, profile.Name, model}, "|")
		mac := device.AutoMAC(seed)
		log.Printf("auto MAC resolved: %s seed=%q", mac, seed)
		return mac
	}
	mac, err := net.ParseMAC(value)
	if err != nil {
		log.Fatalf("invalid MAC address: %v", err)
	}
	return mac
}

func hostInterfaceMAC(ifaceName string) (net.HardwareAddr, error) {
	ifaceName = strings.TrimSpace(ifaceName)
	if ifaceName == "" {
		return nil, errors.New("observe_interface is required when mac is host")
	}
	if strings.Contains(ifaceName, "/") {
		return nil, fmt.Errorf("invalid interface name %q", ifaceName)
	}
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return nil, fmt.Errorf("find interface %s: %w", ifaceName, err)
	}
	if len(iface.HardwareAddr) == 0 {
		return nil, fmt.Errorf("interface %s has no hardware address", ifaceName)
	}
	return iface.HardwareAddr, nil
}

func resolvePortOptions(profile device.Profile, linkSpeed int, uplinkPort int, uplinkSpeed, controller string) device.PortOptions {
	portOptions := profile.PortOptions()
	if linkSpeed > 0 {
		portOptions.Speed = linkSpeed
		portOptions.UplinkSpeed = linkSpeed
		portOptions.Media = ""
		portOptions.UplinkMedia = ""
		portOptions.PortGroups = nil
	}
	portOptions.UplinkPort = uplinkPort
	return resolveUplinkSpeed(portOptions, uplinkSpeed, controller)
}

func resolveUplinkSpeed(options device.PortOptions, value, target string) device.PortOptions {
	value = strings.TrimSpace(strings.ToLower(value))
	switch value {
	case "", "profile":
		return options
	case automaticText:
		info, err := device.DetectEgressLink(target)
		if err != nil {
			log.Printf("uplink speed auto-detect failed: %v; using profile speed %d Mbps", err, options.UplinkSpeed)
			return options
		}
		options.UplinkSpeed = info.SpeedMbps
		if options.UplinkMedia == "" || options.UplinkMedia == options.Media {
			options.UplinkMedia = ""
		}
		log.Printf("uplink speed auto-detected: interface=%s local_ip=%s speed=%d Mbps", info.Interface, info.LocalIP, info.SpeedMbps)
		return options
	default:
		speed, err := strconv.Atoi(value)
		if err != nil || speed <= 0 {
			log.Fatalf("invalid -uplink-speed %q; use auto, profile, or a positive Mbps value", value)
		}
		options.UplinkSpeed = speed
		if options.UplinkMedia == "" || options.UplinkMedia == options.Media {
			options.UplinkMedia = ""
		}
		return options
	}
}

func applyProfile(profile device.Profile, model, modelDisplay, version *string, portCount *int) {
	if *model == "" {
		*model = profile.Model
	}
	if *modelDisplay == "" {
		*modelDisplay = profile.ModelDisplay
	}
	if *version == "" {
		*version = profile.Version
	}
	if *portCount == 0 {
		*portCount = profile.Ports
	}
}

func serialFromMAC(mac net.HardwareAddr) string {
	out := make([]byte, hex.EncodedLen(len(mac)))
	hex.Encode(out, mac)
	for i := range out {
		if out[i] >= 'a' && out[i] <= 'f' {
			out[i] -= 'a' - 'A'
		}
	}
	return string(out)
}

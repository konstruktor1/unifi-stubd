// Device identity resolution derives the controller-facing MAC, IP, hostname,
// serial, and inform metadata. Host-derived values remain explicit so the stub
// cannot silently become a controller-provisioned host agent.
package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/adoption"
	"github.com/konstruktor1/unifi-stubd/internal/device"
	"github.com/konstruktor1/unifi-stubd/internal/device/payload"
)

// automaticText is the shared CLI value for derived identity fields.
const automaticText = "auto"

// payloadForIdentity converts resolved runtime flags and adoption state into
// the stable device identity model consumed by the payload renderer.
func payloadForIdentity(
	mac net.HardwareAddr,
	ip net.IP,
	hostname string,
	informURL string,
	store adoption.Store,
	flags runtimeFlags,
	profile device.Profile,
	ports []device.Port,
	uptimeSeconds int,
) ([]byte, error) {
	return buildPayload(device.Identity{
		MAC:            mac.String(),
		IP:             ip.String(),
		Hostname:       hostname,
		Model:          flags.model,
		ModelDisplay:   flags.modelDisplay,
		DeviceType:     profile.DeviceType,
		Version:        flags.version,
		Serial:         serialFromMAC(mac),
		InformURL:      informURL,
		InformIP:       resolveInformIP(informURL),
		ManagementVLAN: effectiveManagementVLAN(flags),
		UptimeSeconds:  uptimeSeconds,
	}, profile, store, refreshPortFreshness(ports, uptimeSeconds))
}

// buildPayload overlays adoption-derived controller state onto the identity
// before delegating switch or gateway table construction to the device package.
func buildPayload(id device.Identity, profile device.Profile, store adoption.Store, ports []device.Port) ([]byte, error) {
	id.CFGVersion = store.CFGVersion
	id.Adopted = store.AuthKey != ""
	if store.Version != "" {
		id.Version = store.Version
	}
	payload, err := payload.Build(profile, id, ports)
	if err != nil {
		return nil, fmt.Errorf("build device payload: %w", err)
	}
	return payload, nil
}

// refreshPortFreshness increments synthetic connected-port counters and learned
// MAC uptimes so repeated informs do not look like a frozen first-boot payload.
func refreshPortFreshness(ports []device.Port, uptimeSeconds int) []device.Port {
	if uptimeSeconds < 1 || len(ports) == 0 {
		return ports
	}
	out := make([]device.Port, len(ports))
	copy(out, ports)
	for index := range out {
		port := &out[index]
		if port.Up && strings.TrimSpace(port.Interface) == "" {
			port.RXBytes += int64((uptimeSeconds * 64) + port.Index)
			port.TXBytes += int64((uptimeSeconds * 48) + port.Index)
			port.RXPackets += int64(uptimeSeconds)
			port.TXPackets += int64(uptimeSeconds)
		}
		for macIndex := range port.MACs {
			if port.MACs[macIndex].Uptime < uptimeSeconds {
				port.MACs[macIndex].Uptime = uptimeSeconds
			}
		}
	}
	return out
}

// resolveInformIP reports the numeric controller address when it can be derived
// from the inform URL; failures stay empty because this is payload metadata.
func resolveInformIP(informURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(informURL))
	if err != nil {
		return ""
	}
	host := parsed.Hostname()
	if host == "" {
		return ""
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.String()
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		return ""
	}
	for _, ip := range ips {
		if v4 := ip.To4(); v4 != nil {
			return v4.String()
		}
	}
	if len(ips) > 0 {
		return ips[0].String()
	}
	return ""
}

// resolveHostname keeps explicit hostnames stable but falls back to the local
// OS hostname for automatic lab identities.
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

// resolveMAC chooses the fake device MAC. The automatic path is stable for the
// same host/profile/model, while host MAC use is allowed only in explicit
// host-direct mode.
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

// hostInterfaceMAC reads a local interface address only for explicit
// host-direct mode; controller data never selects this interface.
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

// applyProfile fills CLI defaults from the selected profile after external
// profile loading and before identity and payload construction.
func applyProfile(profile device.Profile, flags *runtimeFlags) {
	for _, field := range []struct {
		target *string
		value  string
	}{
		{target: &flags.model, value: profile.Model},
		{target: &flags.modelDisplay, value: profile.ModelDisplay},
		{target: &flags.version, value: profile.Version},
	} {
		setDefaultString(field.target, field.value)
	}
	setDefaultInt(&flags.portCount, profile.Ports)
}

// setDefaultString fills profile-derived string defaults only when the operator
// did not set a value.
func setDefaultString(target *string, value string) {
	if *target == "" {
		*target = value
	}
}

// setDefaultInt fills profile-derived numeric defaults only when the operator
// did not set a value.
func setDefaultInt(target *int, value int) {
	if *target == 0 {
		*target = value
	}
}

// serialFromMAC mirrors UniFi-style serial formatting by uppercasing the device
// MAC without separators.
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

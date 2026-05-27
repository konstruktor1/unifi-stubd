package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/device"
)

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

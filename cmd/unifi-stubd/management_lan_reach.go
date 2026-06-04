package main

import (
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	appconfig "github.com/konstruktor1/unifi-stubd/internal/config"
)

// checkControllerReachability applies the management-LAN warn/required policy.
func checkControllerReachability(flags runtimeFlags, cfg appconfig.ManagementLAN, sourceIP net.IP) error {
	if cfg.ControllerReachable == managementLANReachOff {
		return nil
	}
	hostPort, err := controllerHostPort(flags.controller)
	if err != nil {
		return err
	}
	dialer := net.Dialer{
		Timeout:   2 * time.Second,
		LocalAddr: &net.TCPAddr{IP: sourceIP},
	}
	conn, err := dialer.Dial("tcp4", hostPort)
	if err != nil {
		return fmt.Errorf("management LAN cannot reach controller %s from %s: %w", hostPort, sourceIP, err)
	}
	_ = conn.Close()
	return nil
}

// controllerHostPort turns the configured controller URL into the TCP endpoint
// used for management-LAN reachability checks.
func controllerHostPort(rawURL string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed.Hostname() == "" {
		return "", fmt.Errorf("management_lan.controller_reachable requires controller_url with host")
	}
	port := parsed.Port()
	if port == "" {
		switch parsed.Scheme {
		case "https":
			port = "443"
		default:
			port = "80"
		}
	}
	return net.JoinHostPort(parsed.Hostname(), port), nil
}

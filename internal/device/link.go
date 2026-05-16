package device

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// LinkInfo describes the host interface used to reach a target.
type LinkInfo struct {
	// Interface is the operating system interface name.
	Interface string
	// SpeedMbps is the detected interface speed in Mbps.
	SpeedMbps int
	// LocalIP is the local address selected for the target route.
	LocalIP net.IP
}

// DetectEgressLink returns the local interface, IP, and speed used for target.
func DetectEgressLink(target string) (LinkInfo, error) {
	address, err := targetAddress(target)
	if err != nil {
		return LinkInfo{}, err
	}
	conn, err := net.Dial("udp", address)
	if err != nil {
		return LinkInfo{}, err
	}
	defer func() {
		_ = conn.Close()
	}()

	local, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok || local.IP == nil {
		return LinkInfo{}, errors.New("could not determine local egress IP")
	}
	iface, err := interfaceByIP(local.IP)
	if err != nil {
		return LinkInfo{LocalIP: local.IP}, err
	}
	speed, err := InterfaceSpeedMbps(iface.Name)
	if err != nil {
		return LinkInfo{Interface: iface.Name, LocalIP: local.IP}, err
	}
	return LinkInfo{Interface: iface.Name, SpeedMbps: speed, LocalIP: local.IP}, nil
}

// InterfaceSpeedMbps reads the Linux sysfs speed value for name.
func InterfaceSpeedMbps(name string) (int, error) {
	if strings.Contains(name, "/") {
		return 0, fmt.Errorf("invalid interface name %q", name)
	}
	if runtime.GOOS != "linux" {
		return 0, fmt.Errorf("interface speed auto-detect is not implemented on %s", runtime.GOOS)
	}
	data, err := os.ReadFile(filepath.Join("/sys/class/net", name, "speed"))
	if err != nil {
		return 0, err
	}
	speed, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, err
	}
	if speed <= 0 {
		return 0, fmt.Errorf("interface %s reports unknown speed %d", name, speed)
	}
	return speed, nil
}

func targetAddress(target string) (string, error) {
	target = strings.TrimSpace(target)
	if target == "" {
		return "", errors.New("no target URL or address available")
	}
	if strings.Contains(target, "://") {
		parsed, err := url.Parse(target)
		if err != nil {
			return "", err
		}
		host := parsed.Hostname()
		if host == "" {
			return "", fmt.Errorf("target URL %q has no host", target)
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
		return net.JoinHostPort(host, port), nil
	}
	host, port, err := net.SplitHostPort(target)
	if err == nil {
		return net.JoinHostPort(host, port), nil
	}
	return net.JoinHostPort(target, "9"), nil
}

func interfaceByIP(ip net.IP) (*net.Interface, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for i := range ifaces {
		if ifaces[i].Flags&net.FlagUp == 0 {
			continue
		}
		addrs, err := ifaces[i].Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			if addrContainsIP(addr, ip) {
				return &ifaces[i], nil
			}
		}
	}
	return nil, fmt.Errorf("no interface owns local IP %s", ip)
}

func addrContainsIP(addr net.Addr, ip net.IP) bool {
	switch value := addr.(type) {
	case *net.IPNet:
		return value.IP.Equal(ip)
	case *net.IPAddr:
		return value.IP.Equal(ip)
	default:
		return false
	}
}

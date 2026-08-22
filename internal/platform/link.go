package platform

import (
	"context"
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

// EgressLink describes the host interface selected for a target route.
type EgressLink struct {
	Interface string
	SpeedMbps int
	LocalIP   net.IP
}

// EgressLink returns the local interface, address, and speed used for target.
func (p hostPlatform) EgressLink(ctx context.Context, target string) (EgressLink, error) {
	address, err := targetAddress(target)
	if err != nil {
		return EgressLink{}, fmt.Errorf("resolve target address: %w", err)
	}
	commandCtx, cancel := context.WithTimeout(ctx, p.cfg.CommandTimeout)
	defer cancel()
	conn, err := (&net.Dialer{}).DialContext(commandCtx, "udp", address)
	if err != nil {
		return EgressLink{}, fmt.Errorf("dial route probe %s: %w", address, err)
	}
	defer func() { _ = conn.Close() }()

	local, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok || local.IP == nil {
		return EgressLink{}, errors.New("could not determine local egress IP")
	}
	iface, err := interfaceByIP(local.IP)
	if err != nil {
		return EgressLink{LocalIP: local.IP}, fmt.Errorf("find interface for local IP %s: %w", local.IP, err)
	}
	speed, err := interfaceSpeedMbps(iface.Name)
	if err != nil {
		return EgressLink{Interface: iface.Name, LocalIP: local.IP}, fmt.Errorf("detect speed for interface %s: %w", iface.Name, err)
	}
	return EgressLink{Interface: iface.Name, SpeedMbps: speed, LocalIP: local.IP}, nil
}

// InterfaceIdentity reads addresses and the hardware address without invoking
// optional platform tools or counter sources.
func (p hostPlatform) InterfaceIdentity(ctx context.Context, name string) (InterfaceIdentity, error) {
	if err := ctx.Err(); err != nil {
		return InterfaceIdentity{}, fmt.Errorf("read interface identity: %w", err)
	}
	name = strings.TrimSpace(name)
	if name == "" || strings.Contains(name, "/") {
		return InterfaceIdentity{}, fmt.Errorf("invalid interface name %q", name)
	}
	iface, err := net.InterfaceByName(name)
	if err != nil {
		return InterfaceIdentity{}, fmt.Errorf("find interface %s: %w", name, err)
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return InterfaceIdentity{}, fmt.Errorf("read interface %s addresses: %w", name, err)
	}
	identity := InterfaceIdentity{MAC: append(net.HardwareAddr(nil), iface.HardwareAddr...)}
	for _, addr := range addrs {
		if ip := addressIP(addr).To4(); ip != nil {
			identity.IPv4 = append(identity.IPv4, append(net.IP(nil), ip...))
		}
	}
	return identity, nil
}

// IPv4Assigned reports whether ip belongs to iface, or to any interface when
// iface is empty.
func (p hostPlatform) IPv4Assigned(ctx context.Context, iface string, ip net.IP) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, fmt.Errorf("check IPv4 assignment: %w", err)
	}
	ip = ip.To4()
	if ip == nil {
		return false, errors.New("IPv4 address is required")
	}
	iface = strings.TrimSpace(iface)
	if iface != "" {
		selected, err := net.InterfaceByName(iface)
		if err != nil {
			return false, fmt.Errorf("find interface %s: %w", iface, err)
		}
		return interfaceOwnsIPv4(selected, ip)
	}
	ifaces, err := net.Interfaces()
	if err != nil {
		return false, fmt.Errorf("list network interfaces: %w", err)
	}
	var readErr error
	for index := range ifaces {
		if err := ctx.Err(); err != nil {
			return false, fmt.Errorf("check IPv4 assignment: %w", err)
		}
		assigned, err := interfaceOwnsIPv4(&ifaces[index], ip)
		if err != nil && readErr == nil {
			readErr = err
		}
		if assigned {
			return true, nil
		}
	}
	if readErr != nil {
		return false, fmt.Errorf("check IPv4 assignment: %w", readErr)
	}
	return false, nil
}

func targetAddress(target string) (string, error) {
	target = strings.TrimSpace(target)
	if target == "" {
		return "", errors.New("no target URL or address available")
	}
	if !strings.Contains(target, "://") {
		if host, port, err := net.SplitHostPort(target); err == nil {
			return net.JoinHostPort(host, port), nil
		}
		return net.JoinHostPort(target, "9"), nil
	}
	parsed, err := url.Parse(target)
	if err != nil {
		return "", fmt.Errorf("parse target URL %q: %w", target, err)
	}
	host := parsed.Hostname()
	if host == "" {
		return "", fmt.Errorf("target URL %q has no host", target)
	}
	port := parsed.Port()
	switch {
	case port != "":
	case parsed.Scheme == "https":
		port = "443"
	default:
		port = "80"
	}
	return net.JoinHostPort(host, port), nil
}

func interfaceByIP(ip net.IP) (*net.Interface, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("list network interfaces: %w", err)
	}
	var readErr error
	for index := range ifaces {
		if ifaces[index].Flags&net.FlagUp == 0 {
			continue
		}
		assigned, err := interfaceOwnsIPv4(&ifaces[index], ip)
		if err != nil && readErr == nil {
			readErr = err
		}
		if assigned {
			return &ifaces[index], nil
		}
	}
	if readErr != nil {
		return nil, fmt.Errorf("inspect interfaces for local IP %s: %w", ip, readErr)
	}
	return nil, fmt.Errorf("no interface owns local IP %s", ip)
}

func interfaceOwnsIPv4(iface *net.Interface, ip net.IP) (bool, error) {
	addrs, err := iface.Addrs()
	if err != nil {
		return false, fmt.Errorf("read interface %s addresses: %w", iface.Name, err)
	}
	for _, addr := range addrs {
		current := addressIP(addr)
		if current != nil && current.Equal(ip) {
			return true, nil
		}
	}
	return false, nil
}

func addressIP(addr net.Addr) net.IP {
	switch value := addr.(type) {
	case *net.IPNet:
		return value.IP
	case *net.IPAddr:
		return value.IP
	default:
		return nil
	}
}

func interfaceSpeedMbps(name string) (int, error) {
	if strings.Contains(name, "/") {
		return 0, fmt.Errorf("invalid interface name %q", name)
	}
	if runtime.GOOS != goosLinux {
		return 0, fmt.Errorf("interface speed auto-detect is not implemented on %s", runtime.GOOS)
	}
	data, err := os.ReadFile(filepath.Join("/sys/class/net", name, "speed"))
	if err != nil {
		return 0, fmt.Errorf("read interface speed for %s: %w", name, err)
	}
	speed, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, fmt.Errorf("parse interface speed for %s: %w", name, err)
	}
	if speed <= 0 {
		return 0, fmt.Errorf("interface %s reports unknown speed %d", name, speed)
	}
	return speed, nil
}

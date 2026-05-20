// Package discovery turns supplied identity data into UniFi UDP discovery TLV
// announcements. It intentionally does not choose identities or send network
// traffic.
package discovery

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strings"
)

const (
	// Port is the UDP port used by UniFi discovery.
	Port = 10001
	// BroadcastAddress is the IPv4 broadcast target for discovery packets.
	BroadcastAddress = "255.255.255.255:10001"
	// MulticastAddress is the UniFi multicast target for discovery packets.
	MulticastAddress      = "233.89.188.1:10001"
	packetVersion    byte = 0x02
	packetType       byte = 0x06
)

// Announcement describes one UniFi discovery TLV announcement.
type Announcement struct {
	// MAC is the announcing device MAC address.
	MAC net.HardwareAddr
	// IP is the announcing device IPv4 address.
	IP net.IP
	// Model is the UniFi model identifier.
	Model string
	// Version is the firmware version reported in discovery.
	Version string
	// Hostname is the optional device hostname.
	Hostname string
	// Default reports whether the device is still in factory-default state.
	Default bool
	// Uptime is the announced uptime in seconds.
	Uptime uint32
	// Sequence is the discovery packet sequence counter.
	Sequence uint32
}

// MarshalBinary encodes a as a UniFi discovery packet.
func (a Announcement) MarshalBinary() ([]byte, error) {
	if len(a.MAC) != 6 {
		return nil, fmt.Errorf("MAC must be 6 bytes")
	}
	ip := a.IP.To4()
	if ip == nil {
		return nil, fmt.Errorf("IP must be IPv4")
	}

	var payload bytes.Buffer
	writeTLV(&payload, 0x02, append(append([]byte{}, a.MAC...), ip...))
	writeTLV(&payload, 0x01, a.MAC)
	writeTLV(&payload, 0x0a, uint32Bytes(a.Uptime))
	if a.Hostname != "" {
		writeTLV(&payload, 0x0b, []byte(a.Hostname))
	}
	writeTLV(&payload, 0x16, []byte(a.Version))
	writeTLV(&payload, 0x15, []byte(a.Model))
	writeTLV(&payload, 0x13, []byte(serialFromMAC(a.MAC)))
	writeTLV(&payload, 0x12, uint32Bytes(a.Sequence))
	if a.Default {
		writeTLV(&payload, 0x17, []byte{1})
	} else {
		writeTLV(&payload, 0x17, []byte{0})
	}

	if payload.Len() > 0xffff {
		return nil, fmt.Errorf("discovery payload too large")
	}

	out := bytes.NewBuffer(make([]byte, 0, 4+payload.Len()))
	out.WriteByte(packetVersion)
	out.WriteByte(packetType)
	_ = binary.Write(out, binary.BigEndian, uint16(payload.Len()))
	out.Write(payload.Bytes())
	return out.Bytes(), nil
}

// DefaultTargets returns the standard UniFi discovery UDP targets.
func DefaultTargets() []string {
	return []string{BroadcastAddress, MulticastAddress}
}

// Send broadcasts packet to the UniFi discovery broadcast and multicast targets.
func Send(packet []byte) error {
	return SendTo(packet, nil)
}

// SendTo sends packet to explicit targets or to the default discovery targets.
func SendTo(packet []byte, targets []string) error {
	return SendToInterface(packet, targets, "")
}

// SendToInterface sends packet through ifaceName when set.
func SendToInterface(packet []byte, targets []string, ifaceName string) error {
	targets = cleanTargets(targets)
	if len(targets) == 0 {
		targets = DefaultTargets()
	}
	listenAddr := ":0"
	if ifaceName = strings.TrimSpace(ifaceName); ifaceName != "" {
		ip, err := interfaceIPv4(ifaceName)
		if err != nil {
			return err
		}
		listenAddr = ip.String() + ":0"
	}
	conn, err := net.ListenPacket("udp4", listenAddr)
	if err != nil {
		return fmt.Errorf("open discovery socket: %w", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	var errs []error
	for _, addr := range targets {
		udpAddr, err := net.ResolveUDPAddr("udp4", addr)
		if err != nil {
			errs = append(errs, fmt.Errorf("resolve discovery address %s: %w", addr, err))
			continue
		}
		if _, err := conn.WriteTo(packet, udpAddr); err != nil {
			errs = append(errs, fmt.Errorf("send discovery packet to %s: %w", addr, err))
		}
	}
	if err := errors.Join(errs...); err != nil {
		return fmt.Errorf("send discovery packets: %w", err)
	}
	return nil
}

func interfaceIPv4(name string) (net.IP, error) {
	if strings.Contains(name, "/") {
		return nil, fmt.Errorf("invalid discovery interface %q", name)
	}
	iface, err := net.InterfaceByName(name)
	if err != nil {
		return nil, fmt.Errorf("find discovery interface %s: %w", name, err)
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return nil, fmt.Errorf("read discovery interface %s addresses: %w", name, err)
	}
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}
		if ip := ipNet.IP.To4(); ip != nil {
			return ip, nil
		}
	}
	return nil, fmt.Errorf("discovery interface %s has no IPv4 address", name)
}

func cleanTargets(targets []string) []string {
	out := make([]string, 0, len(targets))
	for _, target := range targets {
		target = strings.TrimSpace(target)
		if target != "" {
			out = append(out, target)
		}
	}
	return out
}

func writeTLV(w *bytes.Buffer, typ byte, value []byte) {
	w.WriteByte(typ)
	_ = binary.Write(w, binary.BigEndian, uint16(len(value)))
	w.Write(value)
}

func uint32Bytes(v uint32) []byte {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], v)
	return b[:]
}

func serialFromMAC(mac net.HardwareAddr) string {
	const hexdigits = "0123456789ABCDEF"
	out := make([]byte, 0, len(mac)*2)
	for _, b := range mac {
		out = append(out, hexdigits[b>>4], hexdigits[b&0x0f])
	}
	return string(out)
}

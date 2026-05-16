package discovery

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
)

const (
	Port                  = 10001
	BroadcastAddress      = "255.255.255.255:10001"
	MulticastAddress      = "233.89.188.1:10001"
	packetVersion    byte = 0x02
	packetType       byte = 0x06
)

type Announcement struct {
	MAC      net.HardwareAddr
	IP       net.IP
	Model    string
	Version  string
	Hostname string
	Default  bool
	Uptime   uint32
	Sequence uint32
}

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

func Send(packet []byte) error {
	conn, err := net.ListenPacket("udp4", ":0")
	if err != nil {
		return err
	}
	defer conn.Close()

	for _, addr := range []string{BroadcastAddress, MulticastAddress} {
		udpAddr, err := net.ResolveUDPAddr("udp4", addr)
		if err != nil {
			return err
		}
		if _, err := conn.WriteTo(packet, udpAddr); err != nil {
			return err
		}
	}
	return nil
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

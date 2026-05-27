package inform

import (
	"crypto/md5"
	"encoding/binary"
	"net"
)

// TNBU constants describe the inform framing version and feature flags used on
// the wire.
const (
	// Magic is the inform packet magic marker.
	Magic = "TNBU"
	// PacketVersion is the supported inform packet header version.
	PacketVersion = uint32(0)
	// PayloadVersion is the supported inform payload version.
	PayloadVersion = uint32(1)

	// FlagEncrypted marks an encrypted inform payload.
	FlagEncrypted uint16 = 0x01
	// FlagZlib marks a zlib-compressed inform payload.
	FlagZlib uint16 = 0x02
	// FlagSnappy marks a snappy-compressed inform payload.
	FlagSnappy uint16 = 0x04
	// FlagEncryptedGCM marks an AES-GCM encrypted inform payload.
	FlagEncryptedGCM uint16 = 0x08
)

// Options controls inform packet encoding features.
type Options struct {
	// Zlib enables zlib compression before encryption.
	Zlib bool
	// GCM enables AES-GCM instead of AES-CBC.
	GCM bool
}

// Packet contains decoded inform packet metadata and encrypted payload bytes.
type Packet struct {
	// MAC is the device MAC address from the packet header.
	MAC net.HardwareAddr
	// Flags contains the inform packet feature bits.
	Flags uint16
	// IV is the initialization vector or GCM nonce from the packet header.
	IV []byte
	// Payload is the encoded payload body before decompression.
	Payload []byte
}

// DefaultAuthKey returns the default UniFi adoption key derived from ubnt.
func DefaultAuthKey() []byte {
	sum := md5.Sum([]byte("ubnt"))
	return sum[:]
}

// makeHeader builds the fixed TNBU packet header shared by CBC and GCM paths.
func makeHeader(mac net.HardwareAddr, flags uint16, iv []byte, payloadLen uint32) []byte {
	header := make([]byte, 40)
	copy(header[0:4], Magic)
	binary.BigEndian.PutUint32(header[4:8], PacketVersion)
	copy(header[8:14], mac)
	binary.BigEndian.PutUint16(header[14:16], flags)
	copy(header[16:32], iv)
	binary.BigEndian.PutUint32(header[32:36], PayloadVersion)
	binary.BigEndian.PutUint32(header[36:40], payloadLen)
	return header
}

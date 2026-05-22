// Package inform handles TNBU headers, optional zlib compression, and
// AES-CBC/AES-GCM packet bodies. JSON payload meaning is handled by higher-level
// adoption and device packages.
package inform

import (
	"bytes"
	"compress/zlib"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
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

// EncodeJSON wraps a JSON payload in a UniFi inform packet.
func EncodeJSON(mac net.HardwareAddr, key []byte, payload []byte, opts Options) ([]byte, error) {
	if len(mac) != 6 {
		return nil, fmt.Errorf("MAC must be 6 bytes")
	}
	if len(key) != 16 {
		return nil, fmt.Errorf("authkey must be 16 bytes")
	}

	body := append([]byte{}, payload...)
	flags := FlagEncrypted

	if opts.Zlib {
		// UniFi inform bodies are compressed before encryption. Keeping the order
		// here explicit prevents AES-CBC and AES-GCM paths from drifting.
		var compressed bytes.Buffer
		zw := zlib.NewWriter(&compressed)
		if _, err := zw.Write(body); err != nil {
			return nil, fmt.Errorf("compress inform payload: %w", err)
		}
		if err := zw.Close(); err != nil {
			return nil, fmt.Errorf("finish inform compression: %w", err)
		}
		body = compressed.Bytes()
		flags |= FlagZlib
	}

	iv := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("generate inform IV: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create AES cipher: %w", err)
	}

	if opts.GCM {
		flags |= FlagEncryptedGCM
		aead, err := cipher.NewGCMWithNonceSize(block, len(iv))
		if err != nil {
			return nil, fmt.Errorf("create AES-GCM cipher: %w", err)
		}
		// Newer UniFi inform responses authenticate the fixed 40-byte TNBU
		// header as associated data, so the payload length must include the GCM
		// tag before Seal runs.
		header := makeHeader(mac, flags, iv, uint32(len(body)+aead.Overhead()))
		body = aead.Seal(nil, iv, body, header)
		return append(header, body...), nil
	}

	body = pkcs7Pad(body, aes.BlockSize)
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(body, body)
	header := makeHeader(mac, flags, iv, uint32(len(body)))
	return append(header, body...), nil
}

// Decode unwraps a UniFi inform packet and returns decoded JSON payload bytes.
func Decode(data []byte, key []byte) (*Packet, []byte, error) {
	if len(data) < 40 {
		return nil, nil, fmt.Errorf("inform packet too short")
	}
	if string(data[:4]) != Magic {
		return nil, nil, fmt.Errorf("invalid magic")
	}
	if binary.BigEndian.Uint32(data[4:8]) != PacketVersion {
		return nil, nil, fmt.Errorf("unsupported packet version")
	}
	if binary.BigEndian.Uint32(data[32:36]) != PayloadVersion {
		return nil, nil, fmt.Errorf("unsupported payload version")
	}
	payloadLen := int(binary.BigEndian.Uint32(data[36:40]))
	if len(data) < 40+payloadLen {
		return nil, nil, fmt.Errorf("truncated payload")
	}

	p := &Packet{
		MAC:     append(net.HardwareAddr{}, data[8:14]...),
		Flags:   binary.BigEndian.Uint16(data[14:16]),
		IV:      append([]byte{}, data[16:32]...),
		Payload: append([]byte{}, data[40:40+payloadLen]...),
	}

	body := append([]byte{}, p.Payload...)
	body, err := decryptPayload(p, body, key, data[:40])
	if err != nil {
		return nil, nil, err
	}

	if p.Flags&FlagZlib != 0 {
		body, err = decompressPayload(body)
		if err != nil {
			return nil, nil, err
		}
	}

	return p, body, nil
}

// decryptPayload selects the cipher advertised in the TNBU flags while keeping
// compression handling outside the encryption path.
func decryptPayload(packet *Packet, body []byte, key []byte, header []byte) ([]byte, error) {
	if packet.Flags&FlagEncrypted == 0 {
		return body, nil
	}
	if len(key) != 16 {
		return nil, fmt.Errorf("authkey must be 16 bytes")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create AES cipher: %w", err)
	}
	if packet.Flags&FlagEncryptedGCM != 0 {
		// The header is associated data for AES-GCM and must match the bytes that
		// preceded the encrypted body on the wire.
		return decryptGCMPayload(block, packet.IV, body, header)
	}
	return decryptCBCPayload(block, packet.IV, body)
}

// decryptGCMPayload authenticates the TNBU header as associated data, matching
// the AES-GCM inform response shape used by newer controllers.
func decryptGCMPayload(block cipher.Block, nonce []byte, body []byte, header []byte) ([]byte, error) {
	aead, err := cipher.NewGCMWithNonceSize(block, len(nonce))
	if err != nil {
		return nil, fmt.Errorf("create AES-GCM cipher: %w", err)
	}
	out, err := aead.Open(nil, nonce, body, header)
	if err != nil {
		return nil, fmt.Errorf("decrypt AES-GCM payload: %w", err)
	}
	return out, nil
}

// decryptCBCPayload handles the legacy AES-CBC inform body format.
func decryptCBCPayload(block cipher.Block, iv []byte, body []byte) ([]byte, error) {
	if len(body)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("CBC payload is not block aligned")
	}
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(body, body)
	out, err := pkcs7Unpad(body, aes.BlockSize)
	if err != nil {
		return nil, fmt.Errorf("remove CBC padding: %w", err)
	}
	return out, nil
}

// decompressPayload runs after decryption because TNBU inform payloads compress
// plaintext before encrypting it.
func decompressPayload(body []byte) ([]byte, error) {
	zr, err := zlib.NewReader(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("open zlib payload: %w", err)
	}
	defer func() {
		_ = zr.Close()
	}()
	out, err := io.ReadAll(zr)
	if err != nil {
		return nil, fmt.Errorf("read zlib payload: %w", err)
	}
	return out, nil
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

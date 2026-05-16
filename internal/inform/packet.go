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
		var compressed bytes.Buffer
		zw := zlib.NewWriter(&compressed)
		if _, err := zw.Write(body); err != nil {
			return nil, err
		}
		if err := zw.Close(); err != nil {
			return nil, err
		}
		body = compressed.Bytes()
		flags |= FlagZlib
	}

	iv := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if opts.GCM {
		flags |= FlagEncryptedGCM
		aead, err := cipher.NewGCMWithNonceSize(block, len(iv))
		if err != nil {
			return nil, err
		}
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
	if p.Flags&FlagEncrypted != 0 {
		if len(key) != 16 {
			return nil, nil, fmt.Errorf("authkey must be 16 bytes")
		}
		block, err := aes.NewCipher(key)
		if err != nil {
			return nil, nil, err
		}
		if p.Flags&FlagEncryptedGCM != 0 {
			aead, err := cipher.NewGCMWithNonceSize(block, len(p.IV))
			if err != nil {
				return nil, nil, err
			}
			body, err = aead.Open(nil, p.IV, body, data[:40])
			if err != nil {
				return nil, nil, err
			}
		} else {
			if len(body)%aes.BlockSize != 0 {
				return nil, nil, fmt.Errorf("CBC payload is not block aligned")
			}
			cipher.NewCBCDecrypter(block, p.IV).CryptBlocks(body, body)
			var err error
			body, err = pkcs7Unpad(body, aes.BlockSize)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	if p.Flags&FlagZlib != 0 {
		zr, err := zlib.NewReader(bytes.NewReader(body))
		if err != nil {
			return nil, nil, err
		}
		defer func() {
			_ = zr.Close()
		}()
		body, err = io.ReadAll(zr)
		if err != nil {
			return nil, nil, err
		}
	}

	return p, body, nil
}

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

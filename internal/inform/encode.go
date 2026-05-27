package inform

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"net"
)

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
		var err error
		body, err = compressPayload(body)
		if err != nil {
			return nil, err
		}
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

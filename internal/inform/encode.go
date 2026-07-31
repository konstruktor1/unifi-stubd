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
	if len(payload) > MaxDecodedPayloadBytes {
		return nil, fmt.Errorf("inform plaintext exceeds %d bytes", MaxDecodedPayloadBytes)
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
	// CBC padding and the GCM tag can each add at most one AES block. Keep the
	// encoder and decoder limits coupled even if compression changes later.
	if len(body) > MaxEncodedPayloadBytes-aes.BlockSize {
		return nil, fmt.Errorf("encoded inform payload exceeds %d bytes", MaxEncodedPayloadBytes)
	}

	ivSize := aes.BlockSize
	if opts.GCM {
		ivSize = GCMNonceSize
	}
	iv := make([]byte, ivSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("generate inform IV/nonce: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create AES cipher: %w", err)
	}

	if opts.GCM {
		flags |= FlagEncryptedGCM
		aead, err := newInformGCM(block, iv)
		if err != nil {
			return nil, err
		}
		// The complete fixed TNBU header is additional authenticated data. Its
		// payload length therefore includes the tag before Seal authenticates it.
		header := makeHeader(mac, flags, iv, uint32(len(body)+GCMTagSize))
		body = aead.Seal(nil, iv, body, header)
		return append(header, body...), nil
	}

	body = pkcs7Pad(body, aes.BlockSize)
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(body, body)
	header := makeHeader(mac, flags, iv, uint32(len(body)))
	return append(header, body...), nil
}

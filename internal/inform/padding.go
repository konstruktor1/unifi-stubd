// Package inform uses PKCS#7 padding only for AES-CBC inform bodies.
// The padding helpers stay unexported so no other protocol path depends on this
// representation.
package inform

import "fmt"

// pkcs7Pad pads legacy AES-CBC plaintext before encryption.
func pkcs7Pad(data []byte, blockSize int) []byte {
	padLen := blockSize - (len(data) % blockSize)
	if padLen == 0 {
		padLen = blockSize
	}
	out := append([]byte{}, data...)
	for i := 0; i < padLen; i++ {
		out = append(out, byte(padLen))
	}
	return out
}

// pkcs7Unpad validates and removes padding from legacy AES-CBC payloads.
func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 || len(data)%blockSize != 0 {
		return nil, fmt.Errorf("invalid PKCS#7 data length")
	}
	padLen := int(data[len(data)-1])
	if padLen == 0 || padLen > blockSize || padLen > len(data) {
		return nil, fmt.Errorf("invalid PKCS#7 padding")
	}
	for _, b := range data[len(data)-padLen:] {
		if int(b) != padLen {
			return nil, fmt.Errorf("invalid PKCS#7 padding byte")
		}
	}
	return data[:len(data)-padLen], nil
}

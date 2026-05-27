package inform

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

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

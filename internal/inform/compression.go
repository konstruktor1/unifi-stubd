package inform

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
)

// compressPayload runs before encryption because TNBU inform payloads compress
// plaintext before encrypting it.
func compressPayload(body []byte) ([]byte, error) {
	var compressed bytes.Buffer
	zw := zlib.NewWriter(&compressed)
	if _, err := zw.Write(body); err != nil {
		return nil, fmt.Errorf("compress inform payload: %w", err)
	}
	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("finish inform compression: %w", err)
	}
	return compressed.Bytes(), nil
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

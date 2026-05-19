package inform

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

// DefaultMaxResponseBytes is the default controller response body limit.
const DefaultMaxResponseBytes int64 = 4 * 1024 * 1024

// Client sends UniFi inform packets to a controller.
type Client struct {
	// URL is the controller inform endpoint.
	URL string
	// MAC is the device MAC address placed in the packet header.
	MAC net.HardwareAddr
	// Key is the 16-byte inform encryption key.
	Key []byte
	// HTTPClient overrides the default HTTP client when set.
	HTTPClient *http.Client
	// Options controls inform packet encoding.
	Options Options
	// MaxResponseBytes limits the controller response body. The default is 4 MiB.
	MaxResponseBytes int64
}

// Response contains the raw and decoded controller response.
type Response struct {
	// StatusCode is the HTTP status code returned by the controller.
	StatusCode int
	// RawBody is the response body before inform decoding.
	RawBody []byte
	// JSONBody is the decoded response payload when available.
	JSONBody []byte
	// Packet is the decoded inform packet metadata when available.
	Packet *Packet
}

// Send posts payload as a UniFi inform packet and decodes the response.
func (c Client) Send(payload []byte) (*Response, error) {
	if c.URL == "" {
		return nil, fmt.Errorf("inform URL is required")
	}
	key := c.Key
	if len(key) == 0 {
		key = DefaultAuthKey()
	}
	opts := c.Options
	if !opts.Zlib && !opts.GCM {
		opts.Zlib = true
	}

	body, err := EncodeJSON(c.MAC, key, payload, opts)
	if err != nil {
		return nil, fmt.Errorf("encode inform request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.URL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create inform request: %w", err)
	}
	req.Header.Set("User-Agent", "AirControl Agent v1.0")
	req.Header.Set("Accept", "application/x-binary")
	req.Header.Set("Content-Type", "application/x-binary")

	client := c.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	httpResp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send inform request: %w", err)
	}
	defer func() {
		_ = httpResp.Body.Close()
	}()

	maxResponseBytes := c.MaxResponseBytes
	if maxResponseBytes <= 0 {
		maxResponseBytes = DefaultMaxResponseBytes
	}
	raw, err := readLimitedBody(httpResp.Body, maxResponseBytes)
	if err != nil {
		return nil, fmt.Errorf("read inform response body: %w", err)
	}
	resp := &Response{
		StatusCode: httpResp.StatusCode,
		RawBody:    raw,
	}
	if httpResp.StatusCode != http.StatusOK || len(raw) == 0 {
		return resp, nil
	}

	packet, decoded, err := Decode(raw, key)
	if err != nil {
		return resp, fmt.Errorf("decode inform response: %w", err)
	}
	resp.Packet = packet
	resp.JSONBody = decoded
	return resp, nil
}

func readLimitedBody(r io.Reader, maxBytes int64) ([]byte, error) {
	limited := &io.LimitedReader{R: r, N: maxBytes + 1}
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}
	if int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("inform response body exceeds %d bytes", maxBytes)
	}
	return data, nil
}

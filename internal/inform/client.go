// Package inform wraps JSON payloads in TNBU packets, sends them over HTTP,
// limits controller response size, and decodes inform responses when present.
package inform

import (
	"bytes"
	"context"
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
	// LocalAddr optionally binds outbound inform TCP connections to one local IP.
	LocalAddr net.IP
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
		// Legacy UniFi devices commonly send zlib-compressed AES-CBC inform
		// bodies. Use that shape unless the caller selected AES-GCM explicitly.
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
		client = defaultHTTPClient(c.LocalAddr)
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
		// Pre-adoption controllers often answer 404/empty bodies. That is a
		// valid lifecycle signal, not a packet-decoding failure.
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

// defaultHTTPClient optionally binds inform TCP connections to a chosen local
// management IP while keeping the same request timeout.
func defaultHTTPClient(localIP net.IP) *http.Client {
	if localIP == nil || localIP.To4() == nil {
		return &http.Client{Timeout: 10 * time.Second}
	}
	dialer := &net.Dialer{
		Timeout:   10 * time.Second,
		LocalAddr: &net.TCPAddr{IP: localIP.To4()},
	}
	return &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
				return dialer.DialContext(ctx, network, address)
			},
		},
	}
}

// readLimitedBody bounds controller response size before decode attempts.
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

package opnsense

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const maxResponseBytes = 8 << 20

// Client performs read-only OPNsense API requests.
type Client struct {
	baseURL     string
	credentials Credentials
	httpClient  *http.Client
}

// NewClient builds a GET-only OPNsense API client.
func NewClient(cfg SourceConfig, credentials Credentials) (*Client, error) {
	baseURL, err := normalizeBaseURL(cfg.BaseURL)
	if err != nil {
		return nil, err
	}
	transport, err := clientTransport(cfg)
	if err != nil {
		return nil, err
	}
	return &Client{
		baseURL:     baseURL,
		credentials: credentials,
		httpClient: &http.Client{
			Timeout:   cfg.Timeout(),
			Transport: transport,
		},
	}, nil
}

// InterfacesInfo reads the OPNsense interface overview endpoint.
func (client *Client) InterfacesInfo(ctx context.Context) (map[string]InterfaceStatus, error) {
	var raw any
	if err := client.getJSON(ctx, "/interfaces/overview/interfaces_info", &raw); err != nil {
		return nil, err
	}
	return DecodeInterfaces(raw), nil
}

// Interface reads one OPNsense interface detail endpoint.
func (client *Client) Interface(ctx context.Context, iface string) (InterfaceStatus, error) {
	iface = strings.TrimSpace(iface)
	if iface == "" {
		return InterfaceStatus{}, fmt.Errorf("interface name is required")
	}
	var raw any
	path := "/interfaces/overview/get_interface/" + url.PathEscape(iface)
	if err := client.getJSON(ctx, path, &raw); err != nil {
		return InterfaceStatus{}, err
	}
	status := DecodeInterface(raw, iface)
	if status.Interface == "" {
		status.Interface = iface
	}
	return status, nil
}

// GatewayStatus reads the OPNsense gateway status endpoint.
func (client *Client) GatewayStatus(ctx context.Context) (map[string]GatewayStatus, error) {
	var raw any
	if err := client.getJSON(ctx, "/routes/gateway/status", &raw); err != nil {
		return nil, err
	}
	return DecodeGatewayStatuses(raw), nil
}

func (client *Client) getJSON(ctx context.Context, path string, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, client.endpoint(path), nil)
	if err != nil {
		return fmt.Errorf("build OPNsense API request: %w", err)
	}
	req.SetBasicAuth(client.credentials.Key, client.credentials.Secret)
	req.Header.Set("Accept", "application/json")
	resp, err := client.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("OPNsense API GET %s failed: %w", redactedPath(path), err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
	if err != nil {
		return fmt.Errorf("read OPNsense API GET %s response: %w", redactedPath(path), err)
	}
	if len(body) > maxResponseBytes {
		return fmt.Errorf("OPNsense API GET %s response exceeds %d bytes", redactedPath(path), maxResponseBytes)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("OPNsense API GET %s returned HTTP %d", redactedPath(path), resp.StatusCode)
	}
	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("decode OPNsense API GET %s JSON: %w", redactedPath(path), err)
	}
	return nil
}

func (client *Client) endpoint(path string) string {
	return strings.TrimRight(client.baseURL, "/") + "/api" + path
}

func normalizeBaseURL(value string) (string, error) {
	value = strings.TrimRight(strings.TrimSpace(value), "/")
	if value == "" {
		return "", fmt.Errorf("opnsense base_url is required")
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return "", fmt.Errorf("parse opnsense base_url: %w", err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", fmt.Errorf("opnsense base_url must use http or https")
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("opnsense base_url must include a host")
	}
	return value, nil
}

func clientTransport(cfg SourceConfig) (http.RoundTripper, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: cfg.InsecureSkipVerify, //nolint:gosec // Explicit lab-only source config option.
		MinVersion:         tls.VersionTLS12,
	}
	if cfg.CAFile != "" {
		pool, err := x509.SystemCertPool()
		if err != nil {
			pool = x509.NewCertPool()
		}
		data, err := os.ReadFile(cfg.CAFile)
		if err != nil {
			return nil, fmt.Errorf("read OPNsense CA file %s: %w", cfg.CAFile, err)
		}
		if !pool.AppendCertsFromPEM(data) {
			return nil, fmt.Errorf("OPNsense CA file %s contains no PEM certificates", cfg.CAFile)
		}
		tlsConfig.RootCAs = pool
	}
	return &http.Transport{TLSClientConfig: tlsConfig}, nil
}

func redactedPath(path string) string {
	if strings.TrimSpace(path) == "" {
		return "/"
	}
	return path
}

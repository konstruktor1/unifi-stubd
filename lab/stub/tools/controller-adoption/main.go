// Command controller-adoption bootstraps and drives the disposable Docker lab
// controller without exposing controller responses or credentials in logs.
package main

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	maxControllerResponseBytes = 4 * 1024 * 1024
	commandField               = "cmd"
	actionWaitSwitchTelemetry  = "wait-switch-telemetry"
	expectedSwitchModel        = "US8"
	expectedSourceInterface    = "uplink0"
	controllerBaseURL          = "https://unifi-controller:8443"
	controllerSite             = "default"
	controllerAdminUser        = "admin"
	controllerAdminPassword    = "admin"
	controllerName             = "unifi-stubd integration"
	controllerCountryCode      = "756"
	controllerTimezone         = "UTC"
)

type options struct {
	action          string
	mac             string
	expectedVersion string
	expectedSpeed   int
	timeout         time.Duration
}

type controllerClient struct {
	baseURL *url.URL
	http    *http.Client
}

type apiMeta struct {
	RC            string `json:"rc"`
	Message       string `json:"msg"`
	ServerVersion string `json:"server_version"`
	Up            bool   `json:"up"`
}

type apiEnvelope struct {
	Meta apiMeta         `json:"meta"`
	Data json.RawMessage `json:"data"`
}

type controllerDevice struct {
	MAC               string `json:"mac"`
	Model             string `json:"model"`
	Adopted           bool   `json:"adopted"`
	State             int    `json:"state"`
	AdoptionCompleted *bool  `json:"adoption_completed"`
	controllerTraffic
	Uplink    *controllerUplink `json:"uplink"`
	PortTable []controllerPort  `json:"port_table"`
}

type controllerUplink struct {
	Speed   int  `json:"speed"`
	Up      bool `json:"up"`
	PortIdx int  `json:"port_idx"`
	controllerTraffic
}

type controllerPort struct {
	PortIdx         int    `json:"port_idx"`
	IsUplink        bool   `json:"is_uplink"`
	Up              bool   `json:"up"`
	Speed           int    `json:"speed"`
	SourceInterface string `json:"source_interface"`
	controllerTraffic
}

type controllerTraffic struct {
	RXBytes     *float64 `json:"rx_bytes"`
	TXBytes     *float64 `json:"tx_bytes"`
	RXBytesRate *float64 `json:"rx_bytes-r"`
	TXBytesRate *float64 `json:"tx_bytes-r"`
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "controller adoption failed: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	opts, err := parseOptions(args)
	if err != nil {
		return err
	}
	client, err := newControllerClient()
	if err != nil {
		return err
	}

	switch opts.action {
	case "version":
		return printVersion(client, opts.expectedVersion)
	case "bootstrap":
		return bootstrapController(client)
	}

	if opts.mac == "" {
		return fmt.Errorf("--mac is required for action %s", opts.action)
	}
	if err := client.login(); err != nil {
		return err
	}
	mac := strings.ToLower(opts.mac)
	switch opts.action {
	case "wait-visible":
		device, err := waitForDevice(client, opts, mac, "visible before adoption", func(device *controllerDevice) bool {
			return device != nil && device.Model == expectedSwitchModel && !device.Adopted
		})
		if err != nil {
			return err
		}
		fmt.Printf("controller device visible before adoption: mac=%s model=%s\n", device.MAC, device.Model)
	case "adopt":
		if _, err := client.doJSON(http.MethodPost, "/api/s/"+controllerSite+"/cmd/devmgr", map[string]any{
			commandField: "adopt", "macs": []string{mac},
		}); err != nil {
			return fmt.Errorf("controller adoption command failed: %w", err)
		}
		fmt.Printf("controller adopt requested: mac=%s\n", mac)
	case actionWaitSwitchTelemetry:
		device, err := waitForSwitchTelemetry(client, opts, mac)
		if err != nil {
			return err
		}
		port := selectedControllerPort(device)
		fmt.Printf("controller switch telemetry: model=%s traffic=active counters=advanced state=%d adopted=%t uplink_up=%t uplink_speed_mbps=%d source_interface=%s root_rx_bytes=%.0f root_tx_bytes=%.0f uplink_rx_bytes_per_second=%.0f uplink_tx_bytes_per_second=%.0f port_rx_bytes_per_second=%.0f port_tx_bytes_per_second=%.0f\n",
			device.Model, device.State, device.Adopted, device.Uplink.Up, device.Uplink.Speed,
			port.SourceInterface, *device.RXBytes, *device.TXBytes,
			*device.Uplink.RXBytesRate, *device.Uplink.TXBytesRate,
			*port.RXBytesRate, *port.TXBytesRate)
	default:
		return fmt.Errorf("unknown action %q", opts.action)
	}
	return nil
}

func parseOptions(args []string) (options, error) {
	if len(args) == 0 {
		return options{}, errors.New("action is required")
	}
	opts := options{action: args[0]}
	flags := flag.NewFlagSet("controller-adoption", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&opts.mac, "mac", "", "device MAC")
	flags.StringVar(&opts.expectedVersion, "expected-version", "", "expected controller version")
	flags.IntVar(&opts.expectedSpeed, "expected-uplink-speed", 0, "expected switch uplink speed in Mbps")
	timeoutSeconds := flags.Int("timeout", 90, "timeout in seconds")
	if err := flags.Parse(args[1:]); err != nil {
		return options{}, fmt.Errorf("parse controller options: %w", err)
	}
	if *timeoutSeconds <= 0 {
		return options{}, errors.New("timeout must be positive")
	}
	if opts.action == "version" && opts.expectedVersion == "" {
		return options{}, errors.New("--expected-version is required for version")
	}
	if opts.action == actionWaitSwitchTelemetry {
		if opts.expectedSpeed <= 0 {
			return options{}, errors.New("--expected-uplink-speed must be positive for switch telemetry")
		}
	}
	opts.timeout = time.Duration(*timeoutSeconds) * time.Second
	return opts, nil
}

func newControllerClient() (*controllerClient, error) {
	baseURL, err := url.Parse(controllerBaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse controller URL: %w", err)
	}
	if baseURL.Scheme != "https" && baseURL.Scheme != "http" {
		return nil, fmt.Errorf("controller URL must use http or https")
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("create controller cookie jar: %w", err)
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: true, //nolint:gosec // Disposable Docker lab uses a self-signed certificate.
	}
	return &controllerClient{
		baseURL: baseURL,
		http: &http.Client{
			Timeout:   15 * time.Second,
			Transport: transport,
			Jar:       jar,
		},
	}, nil
}

func bootstrapController(client *controllerClient) error {
	setupRequired, err := client.setupRequired()
	if err != nil {
		return err
	}
	if !setupRequired {
		if err := client.login(); err != nil {
			return fmt.Errorf("controller is configured but lab login failed: %w", err)
		}
		fmt.Println("controller bootstrap: already configured and login verified")
		return nil
	}

	devicePassword, err := randomHex(24)
	if err != nil {
		return err
	}
	email := controllerAdminUser + "@example.invalid"
	steps := []struct {
		path    string
		payload map[string]any
	}{
		{path: "/api/cmd/sitemgr", payload: map[string]any{commandField: "add-default-admin", "name": controllerAdminUser, "email": email, "x_password": controllerAdminPassword}},
		{path: "/api/set/setting/super_identity", payload: map[string]any{"name": controllerName}},
		{path: "/api/set/setting/country", payload: map[string]any{"code": controllerCountryCode}},
		{path: "/api/set/setting/locale", payload: map[string]any{"timezone": controllerTimezone}},
		{path: "/api/set/setting/super_mgmt", payload: map[string]any{"autobackup_enabled": false, "backup_to_cloud_enabled": false}},
		{path: "/api/set/setting/mgmt", payload: map[string]any{"x_ssh_username": controllerAdminUser, "x_ssh_password": devicePassword}},
		{path: "/api/cmd/system", payload: map[string]any{commandField: "set-installed"}},
	}
	for _, step := range steps {
		if _, err := client.doJSON(http.MethodPost, step.path, step.payload); err != nil {
			return fmt.Errorf("bootstrap controller through %s: %w", step.path, err)
		}
	}
	if err := client.login(); err != nil {
		return fmt.Errorf("verify bootstrapped controller login: %w", err)
	}
	fmt.Println("controller bootstrap: setup completed and login verified")
	return nil
}

func printVersion(client *controllerClient, expected string) error {
	envelope, err := client.doJSON(http.MethodGet, "/status", nil)
	if err != nil {
		return err
	}
	version := envelope.Meta.ServerVersion
	if version == "" {
		return errors.New("controller status did not include server_version")
	}
	if version != expected {
		return fmt.Errorf("controller version %s != expected %s", version, expected)
	}
	if !envelope.Meta.Up {
		return fmt.Errorf("controller %s is not ready", version)
	}
	fmt.Printf("controller version: server_version=%s up=true\n", version)
	return nil
}

func (c *controllerClient) setupRequired() (bool, error) {
	req, err := http.NewRequest(http.MethodGet, c.resolve("/"), nil)
	if err != nil {
		return false, fmt.Errorf("create controller setup request: %w", err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return false, fmt.Errorf("request controller setup state: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64*1024))
	return strings.HasPrefix(resp.Request.URL.Path, "/setup"), nil
}

func (c *controllerClient) login() error {
	_, err := c.doJSON(http.MethodPost, "/api/login", map[string]any{
		"username": controllerAdminUser,
		"password": controllerAdminPassword,
		"remember": true,
	})
	if err != nil {
		return fmt.Errorf("controller login failed: %w", err)
	}
	return nil
}

func (c *controllerClient) devices() ([]controllerDevice, error) {
	envelope, err := c.doJSON(http.MethodGet, "/api/s/"+controllerSite+"/stat/device", nil)
	if err != nil {
		return nil, fmt.Errorf("controller device list failed: %w", err)
	}
	var devices []controllerDevice
	if err := json.Unmarshal(envelope.Data, &devices); err != nil {
		return nil, fmt.Errorf("decode controller device list: %w", err)
	}
	return devices, nil
}

func (c *controllerClient) doJSON(method, requestPath string, payload any) (apiEnvelope, error) {
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return apiEnvelope{}, fmt.Errorf("encode controller request: %w", err)
		}
		body = strings.NewReader(string(encoded))
	}
	req, err := http.NewRequest(method, c.resolve(requestPath), body)
	if err != nil {
		return apiEnvelope{}, fmt.Errorf("create controller request: %w", err)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for _, cookie := range c.http.Jar.Cookies(c.baseURL) {
		if cookie.Name == "csrf_token" {
			req.Header.Set("X-CSRF-Token", cookie.Value)
			break
		}
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return apiEnvelope{}, fmt.Errorf("request controller API: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	limited := io.LimitReader(resp.Body, maxControllerResponseBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return apiEnvelope{}, fmt.Errorf("read controller API response: %w", err)
	}
	if len(data) > maxControllerResponseBytes {
		return apiEnvelope{}, errors.New("controller API response exceeded size limit")
	}
	var envelope apiEnvelope
	decodeErr := json.Unmarshal(data, &envelope)
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		if decodeErr == nil && envelope.Meta.Message != "" {
			return apiEnvelope{}, fmt.Errorf("controller API returned HTTP %d: %s", resp.StatusCode, envelope.Meta.Message)
		}
		return apiEnvelope{}, fmt.Errorf("controller API returned HTTP %d", resp.StatusCode)
	}
	if decodeErr != nil {
		return apiEnvelope{}, fmt.Errorf("decode controller API response: %w", decodeErr)
	}
	if envelope.Meta.RC != "ok" {
		message := envelope.Meta.Message
		if message == "" {
			message = "unspecified API error"
		}
		return apiEnvelope{}, fmt.Errorf("controller API rejected request: %s", message)
	}
	return envelope, nil
}

func (c *controllerClient) resolve(requestPath string) string {
	reference := &url.URL{Path: requestPath}
	return c.baseURL.ResolveReference(reference).String()
}

func waitForDevice(client *controllerClient, opts options, mac, description string, predicate func(*controllerDevice) bool) (*controllerDevice, error) {
	deadline := time.Now().Add(opts.timeout)
	for {
		devices, err := client.devices()
		if err != nil {
			return nil, err
		}
		device := findDevice(devices, mac)
		if predicate(device) {
			return device, nil
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timed out waiting for controller device %s (%s)", mac, description)
		}
		time.Sleep(2 * time.Second)
	}
}

func findDevice(devices []controllerDevice, mac string) *controllerDevice {
	for index := range devices {
		if strings.EqualFold(devices[index].MAC, mac) {
			return &devices[index]
		}
	}
	return nil
}

func waitForSwitchTelemetry(client *controllerClient, opts options, mac string) (*controllerDevice, error) {
	var baselineRX, baselineTX float64
	baselineSet := false
	return waitForDevice(client, opts, mac, "advancing bidirectional US8 telemetry", func(device *controllerDevice) bool {
		if !switchTelemetryReady(device, opts.expectedSpeed) {
			return false
		}
		rx, tx := *device.RXBytes, *device.TXBytes
		if !baselineSet {
			baselineRX, baselineTX, baselineSet = rx, tx, true
			return false
		}
		return rx > baselineRX && tx > baselineTX
	})
}

func switchTelemetryReady(device *controllerDevice, expectedSpeed int) bool {
	if device == nil || device.Model != expectedSwitchModel || !device.Adopted || device.State != 1 ||
		(device.AdoptionCompleted != nil && !*device.AdoptionCompleted) ||
		device.Uplink == nil || !device.Uplink.Up || device.Uplink.Speed != expectedSpeed {
		return false
	}
	port := selectedControllerPort(device)
	if port == nil || !port.Up || port.Speed != expectedSpeed || port.SourceInterface != expectedSourceInterface {
		return false
	}
	return positiveNumber(device.RXBytes) && positiveNumber(device.TXBytes) &&
		positiveNumber(device.Uplink.RXBytesRate) && positiveNumber(device.Uplink.TXBytesRate) &&
		positiveNumber(port.RXBytesRate) && positiveNumber(port.TXBytesRate)
}

func positiveNumber(value *float64) bool {
	return value != nil && *value > 0
}

func selectedControllerPort(device *controllerDevice) *controllerPort {
	if device == nil || device.Uplink == nil {
		return nil
	}
	for index := range device.PortTable {
		port := &device.PortTable[index]
		if port.IsUplink && port.PortIdx == device.Uplink.PortIdx {
			return port
		}
	}
	return nil
}

func randomHex(bytes int) (string, error) {
	value := make([]byte, bytes)
	if _, err := rand.Read(value); err != nil {
		return "", fmt.Errorf("generate lab device password: %w", err)
	}
	return hex.EncodeToString(value), nil
}

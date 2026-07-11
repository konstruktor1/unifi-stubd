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
	"strconv"
	"strings"
	"time"
)

const (
	maxControllerResponseBytes = 4 * 1024 * 1024
	commandField               = "cmd"
)

type options struct {
	action          string
	baseURL         string
	site            string
	username        string
	password        string
	mac             string
	expectedVersion string
	timeout         time.Duration
	pollInterval    time.Duration
	stableSamples   int
	insecureTLS     bool
	controllerName  string
	countryCode     string
	timezone        string
	macPrefix       string
}

type controllerClient struct {
	baseURL  *url.URL
	username string
	password string
	http     *http.Client
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
	Name              string `json:"name"`
	IP                string `json:"ip"`
	Adopted           bool   `json:"adopted"`
	State             int    `json:"state"`
	AdoptionCompleted *bool  `json:"adoption_completed"`
	LastSeen          int64  `json:"last_seen"`
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
	client, err := newControllerClient(opts)
	if err != nil {
		return err
	}

	switch opts.action {
	case "version":
		return printVersion(client, opts.expectedVersion)
	case "bootstrap":
		return bootstrapController(client, opts)
	case "reset-fixtures":
		if err := client.login(); err != nil {
			return err
		}
		return resetAdoptedFixtures(client, opts)
	}

	if opts.mac == "" {
		return fmt.Errorf("--mac is required for action %s", opts.action)
	}
	if err := client.login(); err != nil {
		return err
	}
	mac := strings.ToLower(opts.mac)
	switch opts.action {
	case "forget":
		if err := client.command(opts.site, "sitemgr", map[string]any{commandField: "delete-device", "mac": mac}); err != nil {
			return err
		}
		fmt.Printf("controller forget requested: mac=%s\n", mac)
	case "wait-clean":
		if _, err := waitForDevice(client, opts, mac, "not adopted", func(device *controllerDevice) bool {
			return device == nil || !device.Adopted
		}, 1); err != nil {
			return err
		}
		fmt.Printf("controller device not adopted: mac=%s\n", mac)
	case "wait-absent":
		if _, err := waitForDevice(client, opts, mac, "absent", func(device *controllerDevice) bool {
			return device == nil
		}, 1); err != nil {
			return err
		}
		fmt.Printf("controller device absent: mac=%s\n", mac)
	case "wait-present":
		device, err := waitForDevice(client, opts, mac, "present", func(device *controllerDevice) bool {
			return device != nil
		}, 1)
		if err != nil {
			return err
		}
		fmt.Println(deviceSummary("controller device visible", device))
	case "wait-pending":
		device, err := waitForDevice(client, opts, mac, "pending adoption", func(device *controllerDevice) bool {
			return device != nil && !device.Adopted
		}, 1)
		if err != nil {
			return err
		}
		fmt.Println(deviceSummary("controller device pending", device))
	case "adopt":
		if err := client.command(opts.site, "devmgr", map[string]any{commandField: "adopt", "macs": []string{mac}}); err != nil {
			return err
		}
		fmt.Printf("controller adopt requested: mac=%s\n", mac)
	case "wait-adopted":
		device, err := waitForDevice(client, opts, mac, "adopted", func(device *controllerDevice) bool {
			return device != nil && device.Adopted
		}, 1)
		if err != nil {
			return err
		}
		fmt.Println(deviceSummary("controller device adopted", device))
	case "wait-connected":
		device, err := waitForDevice(client, opts, mac, "stably connected", controllerDeviceConnected, opts.stableSamples)
		if err != nil {
			return err
		}
		fmt.Printf("%s stable_samples=%d\n", deviceSummary("controller device connected", device), opts.stableSamples)
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
	flags.StringVar(&opts.baseURL, "base-url", envDefault("UNIFI_STUB_LAB_API_URL", "https://127.0.0.1:8443"), "controller base URL")
	flags.StringVar(&opts.site, "site", envDefault("UNIFI_STUB_LAB_SITE", "default"), "controller site")
	flags.StringVar(&opts.username, "username", envDefault("UNIFI_STUB_LAB_ADMIN_USER", "admin"), "controller username")
	flags.StringVar(&opts.password, "password", envDefault("UNIFI_STUB_LAB_ADMIN_PASSWORD", "admin"), "controller password")
	flags.StringVar(&opts.mac, "mac", "", "device MAC")
	flags.StringVar(&opts.expectedVersion, "expected-version", "", "expected controller version")
	timeoutSeconds := flags.Int("timeout", envInt("UNIFI_STUB_DOCKER_ADOPTION_TIMEOUT", 90), "timeout in seconds")
	pollSeconds := flags.Int("poll-interval", 2, "poll interval in seconds")
	flags.IntVar(&opts.stableSamples, "stable-samples", 3, "consecutive connected samples")
	flags.BoolVar(&opts.insecureTLS, "insecure", envBool("UNIFI_STUB_LAB_INSECURE_TLS", true), "allow the disposable lab controller certificate")
	flags.StringVar(&opts.controllerName, "controller-name", "unifi-stubd integration", "bootstrap controller name")
	flags.StringVar(&opts.countryCode, "country-code", envDefault("UNIFI_STUB_LAB_COUNTRY_CODE", "756"), "bootstrap numeric country code")
	flags.StringVar(&opts.timezone, "timezone", envDefault("TZ", "Europe/Zurich"), "bootstrap timezone")
	flags.StringVar(&opts.macPrefix, "mac-prefix", envDefault("UNIFI_STUB_LAB_MAC_PREFIX", "02:15:6d"), "reserved disposable fixture MAC prefix")
	if err := flags.Parse(args[1:]); err != nil {
		return options{}, fmt.Errorf("parse controller options: %w", err)
	}
	if *timeoutSeconds <= 0 || *pollSeconds <= 0 || opts.stableSamples <= 0 {
		return options{}, errors.New("timeout, poll interval, and stable samples must be positive")
	}
	opts.timeout = time.Duration(*timeoutSeconds) * time.Second
	opts.pollInterval = time.Duration(*pollSeconds) * time.Second
	return opts, nil
}

func resetAdoptedFixtures(client *controllerClient, opts options) error {
	prefix, err := fixtureMACPrefix(opts.macPrefix)
	if err != nil {
		return err
	}
	devices, err := client.devices(opts.site)
	if err != nil {
		return err
	}
	requested := 0
	for _, device := range devices {
		if !device.Adopted || !strings.HasPrefix(strings.ToLower(device.MAC), prefix) {
			continue
		}
		if err := client.command(opts.site, "sitemgr", map[string]any{commandField: "delete-device", "mac": device.MAC}); err != nil {
			return fmt.Errorf("forget fixture %s: %w", device.MAC, err)
		}
		requested++
	}

	deadline := time.Now().Add(opts.timeout)
	for {
		devices, err = client.devices(opts.site)
		if err != nil {
			return err
		}
		adopted := 0
		for _, device := range devices {
			if device.Adopted && strings.HasPrefix(strings.ToLower(device.MAC), prefix) {
				adopted++
			}
		}
		if adopted == 0 {
			fmt.Printf("controller fixture reset: prefix=%s adopted_removed=%d\n", prefix, requested)
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out resetting %d adopted fixture devices with prefix %s", adopted, prefix)
		}
		time.Sleep(opts.pollInterval)
	}
}

func fixtureMACPrefix(value string) (string, error) {
	parts := strings.Split(strings.TrimSpace(value), ":")
	if len(parts) != 3 {
		return "", errors.New("fixture MAC prefix must contain exactly three octets")
	}
	for _, part := range parts {
		if len(part) != 2 {
			return "", errors.New("fixture MAC prefix must contain exactly three two-digit octets")
		}
	}
	octets, err := hex.DecodeString(strings.Join(parts, ""))
	if err != nil {
		return "", errors.New("fixture MAC prefix must contain hexadecimal octets")
	}
	if octets[0]&0x01 != 0 || octets[0]&0x02 == 0 {
		return "", errors.New("fixture MAC prefix must be a locally administered unicast prefix")
	}
	return fmt.Sprintf("%02x:%02x:%02x", octets[0], octets[1], octets[2]), nil
}

func newControllerClient(opts options) (*controllerClient, error) {
	baseURL, err := url.Parse(strings.TrimRight(opts.baseURL, "/"))
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
		InsecureSkipVerify: opts.insecureTLS, //nolint:gosec // Disposable loopback Docker lab uses a self-signed certificate.
	}
	return &controllerClient{
		baseURL:  baseURL,
		username: opts.username,
		password: opts.password,
		http: &http.Client{
			Timeout:   15 * time.Second,
			Transport: transport,
			Jar:       jar,
		},
	}, nil
}

func bootstrapController(client *controllerClient, opts options) error {
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
	email := opts.username + "@example.invalid"
	steps := []struct {
		path    string
		payload map[string]any
	}{
		{path: "/api/cmd/sitemgr", payload: map[string]any{commandField: "add-default-admin", "name": opts.username, "email": email, "x_password": opts.password}},
		{path: "/api/set/setting/super_identity", payload: map[string]any{"name": opts.controllerName}},
		{path: "/api/set/setting/country", payload: map[string]any{"code": opts.countryCode}},
		{path: "/api/set/setting/locale", payload: map[string]any{"timezone": opts.timezone}},
		{path: "/api/set/setting/super_mgmt", payload: map[string]any{"autobackup_enabled": false, "backup_to_cloud_enabled": false}},
		{path: "/api/set/setting/mgmt", payload: map[string]any{"x_ssh_username": opts.username, "x_ssh_password": devicePassword}},
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
	if expected != "" && version != expected {
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
		"username": c.username,
		"password": c.password,
		"remember": true,
	})
	if err != nil {
		return fmt.Errorf("controller login failed: %w", err)
	}
	return nil
}

func (c *controllerClient) command(site, manager string, payload map[string]any) error {
	_, err := c.doJSON(http.MethodPost, "/api/s/"+url.PathEscape(site)+"/cmd/"+url.PathEscape(manager), payload)
	if err != nil {
		return fmt.Errorf("controller command failed: %w", err)
	}
	return nil
}

func (c *controllerClient) devices(site string) ([]controllerDevice, error) {
	envelope, err := c.doJSON(http.MethodGet, "/api/s/"+url.PathEscape(site)+"/stat/device", nil)
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

func waitForDevice(client *controllerClient, opts options, mac, description string, predicate func(*controllerDevice) bool, stableSamples int) (*controllerDevice, error) {
	deadline := time.Now().Add(opts.timeout)
	consecutive := 0
	var last *controllerDevice
	for {
		devices, err := client.devices(opts.site)
		if err != nil {
			return nil, err
		}
		last = findDevice(devices, mac)
		if predicate(last) {
			consecutive++
			if consecutive >= stableSamples {
				return last, nil
			}
		} else {
			consecutive = 0
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timed out waiting for controller device %s (%s)", mac, description)
		}
		time.Sleep(opts.pollInterval)
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

func controllerDeviceConnected(device *controllerDevice) bool {
	if device == nil || !device.Adopted || device.State != 1 {
		return false
	}
	return device.AdoptionCompleted == nil || *device.AdoptionCompleted
}

func deviceSummary(prefix string, device *controllerDevice) string {
	if device == nil {
		return prefix + ": absent"
	}
	completed := "unset"
	if device.AdoptionCompleted != nil {
		completed = strconv.FormatBool(*device.AdoptionCompleted)
	}
	return fmt.Sprintf(
		"%s: mac=%s model=%s name=%s ip=%s adopted=%t state=%d adoption_completed=%s last_seen=%d",
		prefix,
		device.MAC,
		device.Model,
		device.Name,
		device.IP,
		device.Adopted,
		device.State,
		completed,
		device.LastSeen,
	)
}

func randomHex(bytes int) (string, error) {
	value := make([]byte, bytes)
	if _, err := rand.Read(value); err != nil {
		return "", fmt.Errorf("generate lab device password: %w", err)
	}
	return hex.EncodeToString(value), nil
}

func envDefault(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

func envInt(name string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func envBool(name string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

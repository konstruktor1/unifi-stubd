// Command controller-adoption bootstraps and drives the disposable Docker lab
// controller without exposing controller responses or credentials in logs.
package main

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	_ "embed"
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
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

//go:embed controller_modes.json
var controllerModesJSON []byte

type controllerMode struct {
	LoginPath     string `json:"login_path"`
	NetworkPrefix string `json:"network_prefix"`
}

type credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

const (
	classicMode                = "classic"
	unifiOSMode                = "unifi-os"
	maxControllerResponseBytes = 4 * 1024 * 1024
	commandField               = "cmd"
	expectedSwitchModel        = "US8"
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
	timeout         time.Duration
	mode            string
	controllerURL   string
	authFile        string
}

type controllerClient struct {
	baseURL *url.URL
	http    *http.Client
	mode    string
	routes  controllerMode
	auth    credentials
	csrf    string
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
	MAC     string `json:"mac"`
	Model   string `json:"model"`
	Adopted bool   `json:"adopted"`
	State   int    `json:"state"`
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "controller adoption failed: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	opts, err := parseOptions(args)
	if err != nil {
		return err
	}
	if opts.action == "bootstrap" && opts.mode == unifiOSMode {
		return errors.New("UniFi OS setup is owned by the companion; bootstrap is unsupported")
	}
	client, err := newControllerClient(opts)
	if err != nil {
		return err
	}

	switch opts.action {
	case "version":
		if opts.mode == unifiOSMode {
			if err := client.login(ctx); err != nil {
				return err
			}
		}
		return printVersion(ctx, client, opts.expectedVersion)
	case "bootstrap":
		return bootstrapController(ctx, client)
	case "loopback-inform":
		return setLoopbackInform(ctx, client)
	}

	if opts.mac == "" {
		return fmt.Errorf("--mac is required for action %s", opts.action)
	}
	if err := client.login(ctx); err != nil {
		return err
	}
	mac := strings.ToLower(opts.mac)
	switch opts.action {
	case "wait-visible":
		device, err := waitForUS8(ctx, client, opts, mac, false)
		if err != nil {
			return err
		}
		fmt.Printf("controller device visible before adoption: mac=%s model=%s\n", device.MAC, device.Model)
	case "adopt":
		if _, err := client.doJSON(ctx, http.MethodPost, "/api/s/"+controllerSite+"/cmd/devmgr", map[string]any{
			commandField: "adopt", "macs": []string{mac},
		}); err != nil {
			return fmt.Errorf("controller adoption command failed: %w", err)
		}
		fmt.Printf("controller adopt requested: mac=%s\n", mac)
	case "wait-connected":
		device, err := waitForUS8(ctx, client, opts, mac, true)
		if err != nil {
			return err
		}
		fmt.Printf("MAC=%s\nMODEL=%s\nADOPTED=true\nSTATE=%d\n", device.MAC, device.Model, device.State)
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
	flags.StringVar(&opts.mode, "controller-mode", classicMode, "controller authentication: classic or unifi-os")
	flags.StringVar(&opts.controllerURL, "controller-url", "", "controller origin without a path; required for unifi-os")
	flags.StringVar(&opts.authFile, "auth-file", "", "private JSON username/password file; required for unifi-os")
	timeoutSeconds := flags.Int("timeout", 90, "timeout in seconds")
	if err := flags.Parse(args[1:]); err != nil {
		return options{}, fmt.Errorf("parse controller options: %w", err)
	}
	if *timeoutSeconds <= 0 {
		return options{}, errors.New("timeout must be positive")
	}
	if flags.NArg() != 0 {
		return options{}, errors.New("unexpected positional controller options")
	}
	if opts.mode != classicMode && opts.mode != unifiOSMode {
		return options{}, errors.New("controller mode must be classic or unifi-os")
	}
	if opts.mode == unifiOSMode && (opts.controllerURL == "" || opts.authFile == "") {
		return options{}, errors.New("unifi-os requires --controller-url and --auth-file")
	}
	if opts.mode == classicMode && opts.authFile != "" {
		return options{}, errors.New("--auth-file is supported only in unifi-os mode")
	}
	if opts.action == "version" && opts.expectedVersion == "" {
		return options{}, errors.New("--expected-version is required for version")
	}
	opts.timeout = time.Duration(*timeoutSeconds) * time.Second
	return opts, nil
}

func newControllerClient(opts options) (*controllerClient, error) {
	origin := opts.controllerURL
	if origin == "" {
		origin = controllerBaseURL
	}
	baseURL, err := parseOrigin(origin)
	if err != nil {
		return nil, err
	}
	var modes map[string]controllerMode
	if err := json.Unmarshal(controllerModesJSON, &modes); err != nil {
		return nil, errors.New("invalid embedded controller modes")
	}
	auth := credentials{Username: controllerAdminUser, Password: controllerAdminPassword}
	if opts.mode == unifiOSMode {
		auth, err = readCredentials(opts.authFile)
		if err != nil {
			return nil, err
		}
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("create controller cookie jar: %w", err)
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: true, //nolint:gosec // Isolated lab controllers use self-signed certificates.
	}
	return &controllerClient{
		baseURL: baseURL,
		mode:    opts.mode,
		routes:  modes[opts.mode],
		auth:    auth,
		http: &http.Client{
			Timeout:   15 * time.Second,
			Transport: transport,
			Jar:       jar,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if req.URL.Scheme != baseURL.Scheme || !strings.EqualFold(req.URL.Host, baseURL.Host) {
					return errors.New("controller redirect changed origin")
				}
				if len(via) >= 10 {
					return errors.New("too many controller redirects")
				}
				return nil
			},
		},
	}, nil
}

func parseOrigin(value string) (*url.URL, error) {
	origin, err := url.Parse(value)
	if err != nil || origin.Opaque != "" || origin.Hostname() == "" || origin.User != nil ||
		(origin.Scheme != "http" && origin.Scheme != "https") ||
		(origin.Path != "" && origin.Path != "/") || origin.RawQuery != "" || origin.ForceQuery || origin.Fragment != "" {
		return nil, errors.New("controller URL must be an http or https origin without credentials, path, query or fragment")
	}
	if port := origin.Port(); port != "" {
		number, err := strconv.Atoi(port)
		if err != nil || number < 1 || number > 65535 {
			return nil, errors.New("controller URL has an invalid port")
		}
	} else if strings.HasSuffix(origin.Host, ":") {
		return nil, errors.New("controller URL has an empty port")
	}
	return origin, nil
}

func readCredentials(path string) (credentials, error) {
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0077 != 0 {
		return credentials{}, errors.New("auth file must be a private regular file")
	}
	file, err := os.Open(path)
	if err != nil {
		return credentials{}, errors.New("cannot open auth file")
	}
	defer func() { _ = file.Close() }()
	opened, err := file.Stat()
	if err != nil || !os.SameFile(info, opened) || opened.Mode().Perm()&0077 != 0 || opened.Size() > 64*1024 {
		return credentials{}, errors.New("auth file changed or exceeded size limit")
	}
	decoder := json.NewDecoder(io.LimitReader(file, 64*1024+1))
	decoder.DisallowUnknownFields()
	var auth credentials
	if err := decoder.Decode(&auth); err != nil || auth.Username == "" || auth.Password == "" {
		return credentials{}, errors.New("auth file must contain username and password strings")
	}
	if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		return credentials{}, errors.New("auth file contains trailing data")
	}
	return auth, nil
}

func bootstrapController(ctx context.Context, client *controllerClient) error {
	setupRequired, err := client.setupRequired(ctx)
	if err != nil {
		return err
	}
	if !setupRequired {
		if err := client.login(ctx); err != nil {
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
		if _, err := client.doJSON(ctx, http.MethodPost, step.path, step.payload); err != nil {
			return fmt.Errorf("bootstrap controller through %s: %w", step.path, err)
		}
	}
	if err := client.login(ctx); err != nil {
		return fmt.Errorf("verify bootstrapped controller login: %w", err)
	}
	fmt.Println("controller bootstrap: setup completed and login verified")
	return nil
}

func printVersion(ctx context.Context, client *controllerClient, expected string) error {
	versionPath := "/status"
	if client.mode == unifiOSMode {
		versionPath = "/api/s/" + controllerSite + "/stat/sysinfo"
	}
	envelope, err := client.doJSON(ctx, http.MethodGet, versionPath, nil)
	if err != nil {
		return err
	}
	if client.mode == unifiOSMode {
		var systems []struct {
			Version string `json:"version"`
		}
		if err := json.Unmarshal(envelope.Data, &systems); err != nil {
			return fmt.Errorf("decode controller sysinfo: %w", err)
		}
		if len(systems) != 1 {
			return errors.New("controller sysinfo must contain exactly one record")
		}
		version := systems[0].Version
		if strings.TrimSpace(version) == "" {
			return errors.New("controller sysinfo did not include version")
		}
		if version != expected {
			return fmt.Errorf("controller version %s != expected %s", version, expected)
		}
		fmt.Printf("controller version: version=%s api=ready\n", version)
		return nil
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

func (c *controllerClient) setupRequired(ctx context.Context) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.resolve("/"), nil)
	if err != nil {
		return false, fmt.Errorf("create controller setup request: %w", err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			err = urlErr.Err
		}
		return false, fmt.Errorf("request controller setup state: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64*1024))
	return strings.HasPrefix(resp.Request.URL.Path, "/setup"), nil
}

func (c *controllerClient) login(ctx context.Context) error {
	if c.mode == unifiOSMode {
		data, status, err := c.exchange(ctx, http.MethodPost, c.routes.LoginPath, map[string]any{
			"username": c.auth.Username, "password": c.auth.Password, "rememberMe": false,
		})
		if err != nil {
			return fmt.Errorf("UniFi OS login failed: %w", err)
		}
		if status != http.StatusOK {
			return fmt.Errorf("UniFi OS login requires HTTP 200; received HTTP %d", status)
		}
		var user struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(data, &user); err != nil || user.ID == "" || c.csrf == "" {
			return errors.New("UniFi OS login did not establish an authenticated user and CSRF token")
		}
		for _, cookie := range c.http.Jar.Cookies(c.baseURL) {
			if cookie.Name == "TOKEN" && cookie.Value != "" {
				return nil
			}
		}
		return errors.New("UniFi OS login did not establish a session cookie")
	}
	_, err := c.doJSON(ctx, http.MethodPost, c.routes.LoginPath, map[string]any{
		"username": c.auth.Username, "password": c.auth.Password, "remember": true,
	})
	if err != nil {
		return fmt.Errorf("controller login failed: %w", err)
	}
	return nil
}

func (c *controllerClient) devices(ctx context.Context) ([]controllerDevice, error) {
	envelope, err := c.doJSON(ctx, http.MethodGet, "/api/s/"+controllerSite+"/stat/device", nil)
	if err != nil {
		return nil, fmt.Errorf("controller device list failed: %w", err)
	}
	var devices []controllerDevice
	if err := json.Unmarshal(envelope.Data, &devices); err != nil {
		return nil, fmt.Errorf("decode controller device list: %w", err)
	}
	return devices, nil
}

func (c *controllerClient) doJSON(ctx context.Context, method, requestPath string, payload any) (apiEnvelope, error) {
	data, status, err := c.exchange(ctx, method, c.routes.NetworkPrefix+requestPath, payload)
	if err != nil {
		return apiEnvelope{}, err
	}
	if status < http.StatusOK || status >= http.StatusMultipleChoices {
		return apiEnvelope{}, fmt.Errorf("controller API returned HTTP %d", status)
	}
	var envelope apiEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return apiEnvelope{}, errors.New("controller API returned invalid JSON")
	}
	if envelope.Meta.RC != "ok" {
		return apiEnvelope{}, errors.New("controller API rejected request")
	}
	return envelope, nil
}

func (c *controllerClient) exchange(ctx context.Context, method, requestPath string, payload any) ([]byte, int, error) {
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return nil, 0, fmt.Errorf("encode controller request: %w", err)
		}
		body = strings.NewReader(string(encoded))
	}
	req, err := http.NewRequestWithContext(ctx, method, c.resolve(requestPath), body)
	if err != nil {
		return nil, 0, errors.New("cannot create controller request")
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.csrf != "" {
		req.Header.Set("X-CSRF-Token", c.csrf)
	} else if c.mode == classicMode {
		for _, cookie := range c.http.Jar.Cookies(c.baseURL) {
			if cookie.Name == "csrf_token" {
				req.Header.Set("X-CSRF-Token", cookie.Value)
				break
			}
		}
	}
	resp, err := c.http.Do(req)
	if err != nil {
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			err = urlErr.Err // Omit URLs, including redirect queries, from diagnostics.
		}
		return nil, 0, fmt.Errorf("request controller API: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	limited := io.LimitReader(resp.Body, maxControllerResponseBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, 0, fmt.Errorf("read controller API response: %w", err)
	}
	if len(data) > maxControllerResponseBytes {
		return nil, 0, errors.New("controller API response exceeded size limit")
	}
	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		if token := resp.Header.Get("X-Updated-CSRF-Token"); token != "" {
			c.csrf = token
		} else if token := resp.Header.Get("X-CSRF-Token"); token != "" {
			c.csrf = token
		}
	}
	return data, resp.StatusCode, nil
}

func (c *controllerClient) resolve(requestPath string) string {
	reference := &url.URL{Path: requestPath}
	return c.baseURL.ResolveReference(reference).String()
}

func waitForUS8(ctx context.Context, client *controllerClient, opts options, mac string, connected bool) (*controllerDevice, error) {
	ctx, cancel := context.WithTimeout(ctx, opts.timeout)
	defer cancel()
	for {
		devices, err := client.devices(ctx)
		if err != nil {
			return nil, err
		}
		device := findDevice(devices, mac)
		if device != nil && device.Model == expectedSwitchModel &&
			((!connected && !device.Adopted) || (connected && device.Adopted && device.State == 1)) {
			return device, nil
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("wait for controller device %s connected=%t: %w", mac, connected, ctx.Err())
		case <-time.After(2 * time.Second):
		}
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

func randomHex(bytes int) (string, error) {
	value := make([]byte, bytes)
	if _, err := rand.Read(value); err != nil {
		return "", fmt.Errorf("generate lab device password: %w", err)
	}
	return hex.EncodeToString(value), nil
}

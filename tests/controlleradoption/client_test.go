package controlleradoption_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const (
	unifiOSMode        = "unifi-os"
	sessionCookieName  = "TOKEN"
	sessionCookieValue = "test-session"
	uosLoginPath       = "/api/auth/login"
	firstCSRF          = "first-csrf"
	redirectFailure    = "redirect"
	controllerVersion  = "11.0.97"
)

var helperBinary string

func TestMain(m *testing.M) {
	directory, err := os.MkdirTemp("", "controller-adoption-tests-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	helperBinary = filepath.Join(directory, "controller-adoption")
	build := exec.Command("go", "build", "-o", helperBinary, "./lab/stub/tools/controller-adoption")
	build.Dir = "../.."
	if output, err := build.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "build controller helper: %v\n%s", err, output)
		if err := os.RemoveAll(directory); err != nil {
			fmt.Fprintf(os.Stderr, "clean controller helper fixture: %v\n", err)
		}
		os.Exit(1)
	}
	code := m.Run()
	if err := os.RemoveAll(directory); err != nil {
		fmt.Fprintf(os.Stderr, "clean controller helper fixture: %v\n", err)
		code = 1
	}
	os.Exit(code)
}

func authFile(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "auth.json")
	if err := os.WriteFile(path, []byte(`{"username":"lab-owner","password":"private-test-password"}`), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func runHelper(t *testing.T, args ...string) (string, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	output, err := exec.CommandContext(ctx, helperBinary, args...).CombinedOutput()
	return string(output), err
}

func uosArgs(t *testing.T, action, origin string) []string {
	t.Helper()
	return []string{action, "--controller-mode", unifiOSMode, "--controller-url", origin, "--auth-file", authFile(t)}
}

func replyJSON(w http.ResponseWriter, data string) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = fmt.Fprint(w, data)
}

func loginReply(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: sessionCookieName, Value: sessionCookieValue, Path: "/", Secure: true, HttpOnly: true})
	w.Header().Set("X-Updated-CSRF-Token", firstCSRF)
	replyJSON(w, `{"id":"lab-owner","username":"lab-owner"}`)
}

func TestLoopbackInformAuthenticationAndRotation(t *testing.T) {
	for _, mode := range []string{"classic", unifiOSMode} {
		t.Run(mode, func(t *testing.T) {
			var stage atomic.Int32
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				step := stage.Add(1)
				prefix, loginPath, cookieName := "", "/api/login", "csrf_token"
				if mode == unifiOSMode {
					prefix, loginPath, cookieName = "/proxy/network", uosLoginPath, sessionCookieName
				}
				switch step {
				case 1:
					if r.Method != http.MethodPost || r.URL.Path != loginPath {
						t.Errorf("login route = %s %s", r.Method, r.URL.Path)
					}
					var payload map[string]any
					if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
						t.Error(err)
					}
					if mode == unifiOSMode {
						if payload["username"] != "lab-owner" || payload["password"] != "private-test-password" || payload["rememberMe"] != false || payload["remember"] != nil {
							t.Error("wrong UOS login payload")
						}
						loginReply(w)
					} else {
						if payload["username"] != "admin" || payload["password"] != "admin" || payload["remember"] != true {
							t.Error("classic credentials changed")
						}
						http.SetCookie(w, &http.Cookie{Name: cookieName, Value: firstCSRF, Path: "/"})
						replyJSON(w, `{"meta":{"rc":"ok"}}`)
					}
				case 2, 3:
					if _, err := r.Cookie(cookieName); err != nil {
						t.Error("session cookie absent")
					}
					expectedToken := firstCSRF
					if step == 3 {
						expectedToken = "rotated-csrf"
					}
					if r.Header.Get("X-CSRF-Token") != expectedToken {
						t.Error("CSRF token did not follow response rotation")
					}
					if step == 2 {
						if r.Method != http.MethodPost || r.URL.Path != prefix+"/api/s/default/set/setting/super_mgmt" {
							t.Errorf("setting route = %s %s", r.Method, r.URL.Path)
						}
						var setting map[string]any
						if err := json.NewDecoder(r.Body).Decode(&setting); err != nil || setting["override_inform_host"] != true || setting["override_inform_host_location"] != "127.0.0.1" {
							t.Error("incorrect inform setting")
						}
						w.Header().Set("X-CSRF-Token", "rotated-csrf")
						replyJSON(w, `{"meta":{"rc":"ok"}}`)
					} else {
						if r.Method != http.MethodGet || r.URL.Path != prefix+"/api/s/default/get/setting" {
							t.Errorf("readback route = %s %s", r.Method, r.URL.Path)
						}
						replyJSON(w, `{"meta":{"rc":"ok"},"data":[{"key":"super_mgmt","override_inform_host":true,"override_inform_host_location":"127.0.0.1"}]}`)
					}
				default:
					t.Error("unexpected controller request")
					w.WriteHeader(http.StatusBadRequest)
				}
			}))
			defer server.Close()
			args := []string{"loopback-inform", "--controller-url", server.URL}
			if mode == unifiOSMode {
				args = uosArgs(t, "loopback-inform", server.URL)
			}
			output, err := runHelper(t, args...)
			if err != nil || stage.Load() != 3 || !strings.Contains(output, "guest-loopback") {
				t.Fatalf("workflow: %v, requests=%d, output=%s", err, stage.Load(), output)
			}
		})
	}
}

func TestUOSVersionUsesAuthenticatedSysinfo(t *testing.T) {
	for _, test := range []struct {
		name      string
		response  string
		wantError string
	}{
		{"ready", `{"meta":{"rc":"ok"},"data":[{"version":"11.0.97","uptime":4279}]}`, ""},
		{"missing-data", `{"meta":{"rc":"ok"}}`, "decode controller sysinfo"},
		{"empty-data", `{"meta":{"rc":"ok"},"data":[]}`, "exactly one record"},
		{"missing-version", `{"meta":{"rc":"ok"},"data":[{"uptime":4279}]}`, "did not include version"},
		{"blank-version", `{"meta":{"rc":"ok"},"data":[{"version":"  "}]}`, "did not include version"},
		{"multiple-records", `{"meta":{"rc":"ok"},"data":[{"version":"11.0.97"},{"version":"11.0.97"}]}`, "exactly one record"},
		{"mismatch", `{"meta":{"rc":"ok"},"data":[{"version":"11.0.94"}]}`, "!= expected"},
		{"untrimmed-mismatch", `{"meta":{"rc":"ok"},"data":[{"version":" 11.0.97 "}]}`, "!= expected"},
		{"rejected-meta", `{"meta":{"rc":"error"},"data":[{"version":"11.0.97"}]}`, "API rejected"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var requests atomic.Int32
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch requests.Add(1) {
				case 1:
					if r.Method != http.MethodPost || r.URL.Path != uosLoginPath {
						t.Error("version did not authenticate first")
					}
					loginReply(w)
				case 2:
					cookie, err := r.Cookie(sessionCookieName)
					if r.Method != http.MethodGet || r.URL.Path != "/proxy/network/api/s/default/stat/sysinfo" ||
						r.Header.Get("X-CSRF-Token") != firstCSRF || err != nil || cookie.Value != sessionCookieValue {
						t.Error("sysinfo did not use authenticated Network proxy")
					}
					replyJSON(w, test.response)
				default:
					t.Error("unexpected version request")
					w.WriteHeader(http.StatusBadRequest)
				}
			}))
			defer server.Close()
			args := append(uosArgs(t, "version", server.URL), "--expected-version", controllerVersion)
			output, err := runHelper(t, args...)
			if requests.Load() != 2 || strings.Contains(output, "up=") {
				t.Fatalf("incorrect version workflow: requests=%d, output=%s", requests.Load(), output)
			}
			if test.wantError != "" {
				if err == nil || !strings.Contains(output, test.wantError) || strings.Contains(output, "api=ready") {
					t.Fatalf("invalid sysinfo accepted: %v, output=%s", err, output)
				}
				return
			}
			if err != nil || !strings.Contains(output, "version="+controllerVersion+" api=ready") {
				t.Fatalf("version: %v, output=%s", err, output)
			}
		})
	}
}

func TestClassicVersionPreservesStatusReadiness(t *testing.T) {
	for _, up := range []bool{true, false} {
		t.Run(fmt.Sprintf("up=%t", up), func(t *testing.T) {
			var requests atomic.Int32
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				if r.Method != http.MethodGet || r.URL.Path != "/status" {
					t.Error("classic version route changed")
				}
				replyJSON(w, fmt.Sprintf(`{"meta":{"rc":"ok","server_version":%q,"up":%t}}`, controllerVersion, up))
			}))
			defer server.Close()
			output, err := runHelper(t, "version", "--controller-url", server.URL, "--expected-version", controllerVersion)
			if requests.Load() != 1 || (err == nil) != up {
				t.Fatalf("classic readiness changed: %v, requests=%d, output=%s", err, requests.Load(), output)
			}
			if up && !strings.Contains(output, "server_version="+controllerVersion+" up=true") {
				t.Fatalf("classic version output changed: %s", output)
			}
		})
	}
}

func TestUOSLoginFailsClosed(t *testing.T) {
	for _, failure := range []string{"unauthorized", "forbidden", "mfa", "setup", "no-cookie", "no-csrf"} {
		t.Run(failure, func(t *testing.T) {
			var requests atomic.Int32
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				if r.URL.Path != uosLoginPath {
					t.Error("Network request after incomplete login")
				}
				if failure != "no-cookie" {
					http.SetCookie(w, &http.Cookie{Name: sessionCookieName, Value: sessionCookieValue, Path: "/", Secure: true})
				}
				if failure != "no-csrf" {
					w.Header().Set("X-CSRF-Token", firstCSRF)
				}
				switch failure {
				case "unauthorized":
					w.WriteHeader(http.StatusUnauthorized)
				case "forbidden":
					w.WriteHeader(http.StatusForbidden)
				case "mfa":
					w.WriteHeader(http.StatusAccepted)
				case "setup":
					replyJSON(w, `{"uuid":"setup-only","password":"private-test-password"}`)
					return
				}
				replyJSON(w, `{"id":"lab-owner","password":"private-test-password"}`)
			}))
			defer server.Close()
			output, err := runHelper(t, uosArgs(t, "loopback-inform", server.URL)...)
			if err == nil || requests.Load() != 1 || strings.Contains(output, "private-test-password") || strings.Contains(output, "setup-only") {
				t.Fatalf("fail-closed login: %v, requests=%d, output=%s", err, requests.Load(), output)
			}
		})
	}
}

func TestControllerOriginAndPrivateAuthValidation(t *testing.T) {
	for _, origin := range []string{"https://user:secret@example.invalid", "https://example.invalid/api", "https://example.invalid?secret=value", "https://example.invalid#fragment", "https://example.invalid:0", "https://example.invalid:70000", "https://example.invalid:"} {
		if output, err := runHelper(t, uosArgs(t, "loopback-inform", origin)...); err == nil || strings.Contains(output, "secret") {
			t.Fatalf("invalid origin accepted or exposed: %v, %s", err, output)
		}
	}
	path := authFile(t)
	if err := os.Chmod(path, 0644); err != nil {
		t.Fatal(err)
	}
	if output, err := runHelper(t, "loopback-inform", "--controller-mode", unifiOSMode, "--controller-url", "https://example.invalid", "--auth-file", path); err == nil || !strings.Contains(output, "private regular file") {
		t.Fatalf("public auth file accepted: %v, %s", err, output)
	}
	if output, err := runHelper(t, uosArgs(t, "bootstrap", "https://example.invalid")...); err == nil || !strings.Contains(output, "companion") {
		t.Fatalf("UOS setup allowed: %v, %s", err, output)
	}
	if output, err := runHelper(t, "version", "--controller-mode", "unknown", "--expected-version", "11.0.97"); err == nil || !strings.Contains(output, "mode must be") {
		t.Fatalf("unknown mode accepted: %v, %s", err, output)
	}
}

func TestForeignRedirectAndResponseLimit(t *testing.T) {
	var foreignRequests atomic.Int32
	foreign := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		foreignRequests.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer foreign.Close()
	for _, test := range []string{redirectFailure, "limit", "network-rejection"} {
		t.Run(test, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if test == redirectFailure {
					http.Redirect(w, r, foreign.URL+"/?private-test-password", http.StatusTemporaryRedirect)
					return
				}
				if test == "limit" {
					replyJSON(w, strings.Repeat("x", 4*1024*1024+1))
					return
				}
				if r.URL.Path == uosLoginPath {
					loginReply(w)
					return
				}
				replyJSON(w, `{"meta":{"rc":"error","msg":"private-test-password"}}`)
			}))
			defer server.Close()
			output, err := runHelper(t, uosArgs(t, "loopback-inform", server.URL)...)
			if err == nil || foreignRequests.Load() != 0 || strings.Contains(output, "private-test-password") {
				t.Fatalf("unsafe response handling: %v, %s", err, output)
			}
			if test == redirectFailure {
				output, err = runHelper(t, "bootstrap", "--controller-url", server.URL)
				if err == nil || foreignRequests.Load() != 0 || strings.Contains(output, "private-test-password") {
					t.Fatalf("unsafe setup redirect handling: %v, %s", err, output)
				}
			}
		})
	}
}

func TestCancellationInterruptsControllerRequest(t *testing.T) {
	started, cancelled := make(chan struct{}), make(chan struct{})
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == uosLoginPath {
			loginReply(w)
			return
		}
		close(started)
		<-r.Context().Done()
		close(cancelled)
	}))
	defer server.Close()
	args := append(uosArgs(t, "wait-visible", server.URL), "--mac", "02:00:00:00:00:01")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, helperBinary, args...)
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-started:
	case <-ctx.Done():
		t.Fatal("request did not start")
	}
	if err := command.Process.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}
	if err := command.Wait(); err == nil {
		t.Error("cancelled helper succeeded")
	}
	select {
	case <-cancelled:
	case <-ctx.Done():
		t.Error("request context survived cancellation")
	}
}

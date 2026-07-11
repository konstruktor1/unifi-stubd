package cli_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

// TestDaemonStopsCleanlyOnSignals verifies both service-manager and terminal
// shutdown signals run normal daemon cleanup and release the instance lock.
func TestDaemonStopsCleanlyOnSignals(t *testing.T) {
	binary := buildStubdBinary(t)

	for _, tc := range []struct {
		name   string
		signal syscall.Signal
	}{
		{name: "SIGINT", signal: syscall.SIGINT},
		{name: "SIGTERM", signal: syscall.SIGTERM},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			lockPath := filepath.Join(dir, "instance.lock")
			var requests atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				requests.Add(1)
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			cmd, done, output := startBuiltStubdDaemon(t, binary,
				"-no-discovery",
				"-controller", server.URL+"/inform",
				"-instance-guard-path", lockPath,
				"-ssh-state", filepath.Join(dir, "adoption.env"),
				"-status-path", filepath.Join(dir, "status.json"),
				"-profile", "us8",
				"-ip", "192.0.2.50",
				"-hostname", "shutdown-test",
				"-uplink-speed", "profile",
				"-interval", "1h",
			)
			stopped := false
			t.Cleanup(func() {
				if !stopped {
					stopStubdDaemon(t, cmd, done, output)
				}
			})
			waitForRequestCount(t, &requests, 1, output)

			if err := syscall.Kill(-cmd.Process.Pid, tc.signal); err != nil {
				t.Fatalf("send %s: %v", tc.name, err)
			}
			select {
			case err := <-done:
				stopped = true
				if err != nil {
					t.Fatalf("daemon exit after %s: %v\n%s", tc.name, err, output.String())
				}
			case <-time.After(5 * time.Second):
				_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
				t.Fatalf("daemon did not exit after %s:\n%s", tc.name, output.String())
			}
			if !strings.Contains(output.String(), "stopping") {
				t.Fatalf("daemon did not log clean shutdown after %s:\n%s", tc.name, output.String())
			}

			check := exec.Command(binary,
				"-once",
				"-no-discovery",
				"-controller", server.URL+"/inform",
				"-instance-guard-path", lockPath,
				"-ssh-state", filepath.Join(dir, "adoption.env"),
				"-status-path", filepath.Join(dir, "status.json"),
				"-profile", "us8",
				"-ip", "192.0.2.50",
				"-hostname", "shutdown-lock-check",
				"-uplink-speed", "profile",
			)
			if checkOutput, err := check.CombinedOutput(); err != nil {
				t.Fatalf("instance lock remained held after %s: %v\n%s", tc.name, err, checkOutput)
			}
		})
	}
}

func buildStubdBinary(t *testing.T) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "unifi-stubd")
	cmd := exec.Command("go", "build", "-o", binary, "../../cmd/unifi-stubd")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build stubd binary: %v\n%s", err, output)
	}
	return binary
}

func startBuiltStubdDaemon(t *testing.T, binary string, args ...string) (*exec.Cmd, <-chan error, *bytes.Buffer) {
	t.Helper()
	cmd := exec.Command(binary, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	output := &bytes.Buffer{}
	cmd.Stdout = output
	cmd.Stderr = output
	if err := cmd.Start(); err != nil {
		t.Fatalf("start stubd binary: %v\n%s", err, output.String())
	}
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	return cmd, done, output
}

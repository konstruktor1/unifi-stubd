package cli_test

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDryRunPlanHonorsOperationModeOverride(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(configPath, []byte(`controller_url: http://192.0.2.10:8080/inform
operation_mode: stub
profile: us8
mac: auto
ip: 192.0.2.50
hostname: config-host
uplink_speed: profile
port_overrides:
  - port: 2
    speed: 1000
`), 0o600); err != nil {
		t.Fatal(err)
	}

	output := runStubd(t,
		"-config", configPath,
		"-dry-run-plan",
		"-operation-mode", "observe",
		"-profile", "usaggpro",
		"-hostname", "cli-host",
		"-uplink-port", "1",
		"-observe-interface", "eth0",
		"-observe-bridge", "br0",
	)
	if !strings.Contains(output, "operation_mode: observe") {
		t.Fatalf("output did not contain observe mode:\n%s", output)
	}
	if !strings.Contains(output, "profile: usaggpro (USAGGPRO)") {
		t.Fatalf("output did not contain profile override:\n%s", output)
	}
	if !strings.Contains(output, "hostname: cli-host") {
		t.Fatalf("output did not contain hostname override:\n%s", output)
	}
	if !strings.Contains(output, "uplink_port: 1") {
		t.Fatalf("output did not contain uplink override:\n%s", output)
	}
	if !strings.Contains(output, `port_override: port=2 speed=1000`) {
		t.Fatalf("output did not contain port override:\n%s", output)
	}
}

func TestInvalidUplinkPortIsRejected(t *testing.T) {
	cmd := stubdCommand("-dry-run-plan",
		"-profile", "usaggpro",
		"-uplink-port", "33",
	)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("command succeeded; output:\n%s", out)
	}
	if !strings.Contains(string(out), "invalid -uplink-port 33") {
		t.Fatalf("output did not contain uplink validation:\n%s", out)
	}
}

func TestInvalidPortOverrideIsRejected(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(configPath, []byte(`profile: usaggpro
port_overrides:
  - port: 33
    speed: 1000
`), 0o600); err != nil {
		t.Fatal(err)
	}

	cmd := stubdCommand("-config", configPath, "-dry-run-plan")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("command succeeded; output:\n%s", out)
	}
	if !strings.Contains(string(out), "invalid port override 33") {
		t.Fatalf("output did not contain port override validation:\n%s", out)
	}
}

func TestMacHostIsRejectedOutsideHostDirectMode(t *testing.T) {
	cmd := stubdCommand("-dry-run-plan",
		"-operation-mode", "stub",
		"-profile", "usaggpro",
		"-mac", "host",
		"-ip", "192.0.2.50",
		"-hostname", "cli-host",
		"-uplink-speed", "profile",
	)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("command succeeded; output:\n%s", out)
	}
	if !strings.Contains(string(out), "mac: host is only allowed") {
		t.Fatalf("output did not contain host MAC guard:\n%s", out)
	}
}

func TestMacvlanDryRunPlanDoesNotExecute(t *testing.T) {
	output := runStubd(t,
		"-dry-run-plan",
		"-operation-mode", "macvlan",
		"-profile", "usaggpro",
		"-mac", "auto",
		"-ip", "192.0.2.50",
		"-hostname", "cli-host",
		"-uplink-speed", "profile",
		"-observe-interface", "eth0",
	)
	if !strings.Contains(output, "operation_mode: macvlan") {
		t.Fatalf("output did not contain macvlan mode:\n%s", output)
	}
	if !strings.Contains(output, "actions: macvlan is not executed by this release") {
		t.Fatalf("output did not contain non-execution note:\n%s", output)
	}
	if !strings.Contains(output, "planned_command: ip link add link eth0") {
		t.Fatalf("output did not contain planned command:\n%s", output)
	}
}

func TestVersionFlagPrintsBinaryVersion(t *testing.T) {
	output := runStubdStdout(t, "-version")
	if strings.TrimSpace(output) != "dev" {
		t.Fatalf("version output = %q, want dev", output)
	}
}

func TestDoubleDashVersionFlagPrintsBinaryVersion(t *testing.T) {
	output := runStubdStdout(t, "--version")
	if strings.TrimSpace(output) != "dev" {
		t.Fatalf("version output = %q, want dev", output)
	}
}

func TestStatusJSONReportsAdoptionAndLastInformWithoutAuthKey(t *testing.T) {
	dir := t.TempDir()
	statePath := filepath.Join(dir, "adoption.env")
	statusPath := filepath.Join(dir, "status.json")
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(statePath, []byte(`STATE=connected
INFORM_URL=http://10.10.0.30:8080/inform
AUTHKEY=super-secret-test-key
CFGVERSION=abc123
USE_AES_GCM=true
`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(statusPath, []byte(`{
  "last_inform": {
    "time": "2026-05-16T21:00:00+02:00",
    "url": "http://10.10.0.30:8080/inform",
    "status_code": 200,
    "response_type": "noop",
    "controller_state": "connected",
    "cfgversion": "abc123",
    "used_aes_gcm": true,
    "raw_bytes": 128,
    "json_bytes": 64
  }
}
`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configPath, []byte(`controller_url: http://192.0.2.10:8080/inform
operation_mode: stub
profile: usaggpro
mac: auto
ip: 192.0.2.50
hostname: status-host
uplink_speed: profile
state_path: `+statePath+`
status_path: `+statusPath+`
`), 0o600); err != nil {
		t.Fatal(err)
	}

	output := runStubdStdout(t, "-config", configPath, "-status-json")
	if strings.Contains(output, "super-secret-test-key") {
		t.Fatalf("status leaked authkey:\n%s", output)
	}
	var doc struct {
		Config struct {
			OperationMode string `json:"operation_mode"`
			InformURL     string `json:"inform_url"`
		} `json:"config"`
		Adoption struct {
			State      string `json:"state"`
			Adopted    bool   `json:"adopted"`
			AuthKeySet bool   `json:"authkey_set"`
			CFGVersion string `json:"cfgversion"`
		} `json:"adoption"`
		Runtime struct {
			LastInform struct {
				StatusCode   int    `json:"status_code"`
				ResponseType string `json:"response_type"`
			} `json:"last_inform"`
		} `json:"runtime"`
	}
	if err := json.Unmarshal([]byte(output), &doc); err != nil {
		t.Fatalf("status JSON invalid: %v\n%s", err, output)
	}
	if doc.Config.OperationMode != "stub" {
		t.Fatalf("OperationMode = %q", doc.Config.OperationMode)
	}
	if doc.Config.InformURL != "http://10.10.0.30:8080/inform" {
		t.Fatalf("InformURL = %q", doc.Config.InformURL)
	}
	if !doc.Adoption.Adopted || !doc.Adoption.AuthKeySet {
		t.Fatalf("adoption flags = %+v", doc.Adoption)
	}
	if doc.Adoption.State != "connected" || doc.Adoption.CFGVersion != "abc123" {
		t.Fatalf("adoption state = %+v", doc.Adoption)
	}
	if doc.Runtime.LastInform.StatusCode != 200 || doc.Runtime.LastInform.ResponseType != "noop" {
		t.Fatalf("last inform = %+v", doc.Runtime.LastInform)
	}
}

func runStubd(t *testing.T, args ...string) string {
	t.Helper()
	cmd := stubdCommand(args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %v\n%s", err, out)
	}
	return string(out)
}

func runStubdStdout(t *testing.T, args ...string) string {
	t.Helper()
	cmd := stubdCommand(args...)
	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			t.Fatalf("command failed: %v\n%s", err, exitErr.Stderr)
		}
		t.Fatalf("command failed: %v", err)
	}
	return string(out)
}

func stubdCommand(args ...string) *exec.Cmd {
	cmdArgs := append([]string{"run", "../../cmd/unifi-stubd"}, args...)
	return exec.Command("go", cmdArgs...)
}

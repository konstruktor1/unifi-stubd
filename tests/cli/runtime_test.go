package cli_test

import (
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
`), 0o600); err != nil {
		t.Fatal(err)
	}

	output := runStubd(t,
		"-config", configPath,
		"-dry-run-plan",
		"-operation-mode", "observe",
		"-profile", "usaggpro",
		"-hostname", "cli-host",
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

func runStubd(t *testing.T, args ...string) string {
	t.Helper()
	cmd := stubdCommand(args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %v\n%s", err, out)
	}
	return string(out)
}

func stubdCommand(args ...string) *exec.Cmd {
	cmdArgs := append([]string{"run", "../../cmd/unifi-stubd"}, args...)
	return exec.Command("go", cmdArgs...)
}

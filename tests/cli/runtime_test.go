package cli_test

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/konstruktor1/unifi-stubd/internal/inform"
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
management_vlan: 42
discovery_interface: eth0
uplink_neighbor:
  mac: 02:aa:bb:cc:dd:01
  vlan: 1
  type: usw
port_neighbors:
  - port: 2
    mac: 02:00:5e:00:53:03
    vlan: 1
    type: usw
port_overrides:
  - port: 2
    role: lan
    network_group: LAN
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
	if !strings.Contains(output, "management_vlan: 42") {
		t.Fatalf("output did not contain management VLAN:\n%s", output)
	}
	if !strings.Contains(output, "discovery_interface: eth0") {
		t.Fatalf("output did not contain discovery interface:\n%s", output)
	}
	if !strings.Contains(output, `uplink_neighbor: mac=02:aa:bb:cc:dd:01`) {
		t.Fatalf("output did not contain uplink neighbor:\n%s", output)
	}
	if !strings.Contains(output, `port_neighbor: port=2 mac=02:00:5e:00:53:03`) {
		t.Fatalf("output did not contain port neighbor:\n%s", output)
	}
	if !strings.Contains(output, `port_override: port=2`) || !strings.Contains(output, `speed=1000`) {
		t.Fatalf("output did not contain port override:\n%s", output)
	}
	if !strings.Contains(output, `role="lan"`) || !strings.Contains(output, `network_group="LAN"`) {
		t.Fatalf("output did not contain port assignment override:\n%s", output)
	}
}

func TestInvalidManagementVLANIsRejected(t *testing.T) {
	cmd := stubdCommand("-dry-run-plan",
		"-profile", "usaggpro",
		"-management-vlan", "4095",
	)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("command succeeded; output:\n%s", out)
	}
	if !strings.Contains(string(out), "invalid -management-vlan 4095") {
		t.Fatalf("output did not contain management VLAN validation:\n%s", out)
	}
}

func TestLLDPSourceIsRejectedUntilImplemented(t *testing.T) {
	cmd := stubdCommand("-dry-run-plan",
		"-profile", "usaggpro",
		"-lldp-source", "lldpd",
	)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("command succeeded; output:\n%s", out)
	}
	if !strings.Contains(string(out), "lldpd is planned but not implemented yet") {
		t.Fatalf("output did not contain LLDP validation:\n%s", out)
	}
}

func TestValidateAcceptsPackagedConfigs(t *testing.T) {
	for _, path := range []string{
		"../../packaging/linux/etc/unifi-stubd/config.yaml",
		"../../packaging/freebsd/usr/local/etc/unifi-stubd/config.yaml",
	} {
		t.Run(path, func(t *testing.T) {
			output := runStubdStdout(t, "-validate", "-config", path)
			if !strings.Contains(output, "configuration valid: profile=us16p150") {
				t.Fatalf("validate output = %q", output)
			}
		})
	}
}

func TestProfileValidateTemplateAndExport(t *testing.T) {
	dir := t.TempDir()
	templatePath := filepath.Join(dir, "gateway.yaml")
	template := runStubdStdout(t, "-profile-template", "gateway")
	if err := os.WriteFile(templatePath, []byte(template), 0o600); err != nil {
		t.Fatal(err)
	}
	output := runStubdStdout(t, "-profile-validate", templatePath)
	if !strings.Contains(output, "profiles valid") {
		t.Fatalf("profile validate output = %q", output)
	}

	exportPath := filepath.Join(dir, "us8.yaml")
	exported := runStubdStdout(t, "-profile-export", "us8")
	if err := os.WriteFile(exportPath, []byte(exported), 0o600); err != nil {
		t.Fatal(err)
	}
	output = runStubdStdout(t, "-profile-validate", exportPath)
	if !strings.Contains(output, "profiles valid") {
		t.Fatalf("exported profile validate output = %q", output)
	}
}

func TestValidateUsesExternalProfile(t *testing.T) {
	dir := t.TempDir()
	profilePath := filepath.Join(dir, "lab-switch.yaml")
	if err := os.WriteFile(profilePath, []byte(`schema_version: 1
name: lab-switch
model: LABSW
model_display: Lab Switch
device_type: usw
version: 7.4.1.16850
ports: 4
port_speed: 1000
uplink_speed: 1000
port_media: GE
uplink_media: GE
stability: external
payload:
  kind: switch
description: external lab switch
`), 0o600); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(`operation_mode: stub
profile: lab-switch
profile_file: `+profilePath+`
mac: auto
ip: 192.0.2.50
hostname: auto
uplink_speed: profile
`), 0o600); err != nil {
		t.Fatal(err)
	}
	output := runStubdStdout(t, "-validate", "-config", configPath)
	if !strings.Contains(output, "configuration valid: profile=lab-switch") || !strings.Contains(output, "payload=switch") {
		t.Fatalf("validate output = %q", output)
	}
}

func TestProfileValidateRejectsSemanticError(t *testing.T) {
	dir := t.TempDir()
	profilePath := filepath.Join(dir, "bad.yaml")
	if err := os.WriteFile(profilePath, []byte(`schema_version: 1
name: bad-switch
model: BADSW
device_type: usw
ports: 4
port_groups:
  - count: 3
    speed: 1000
payload:
  kind: switch
`), 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := stubdCommand("-profile-validate", profilePath)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("command succeeded; output:\n%s", out)
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
		t.Fatalf("exit = %v, want code 1; output:\n%s", err, out)
	}
	if !strings.Contains(string(out), "port_groups total 3 != ports 4") {
		t.Fatalf("output did not contain profile error:\n%s", out)
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
INFORM_URL=http://192.0.2.10:8080/inform
AUTHKEY=super-secret-test-key
CFGVERSION=abc123
USE_AES_GCM=true
`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(statusPath, []byte(`{
  "last_inform": {
    "time": "2026-05-16T21:00:00+02:00",
    "url": "http://192.0.2.10:8080/inform",
    "status_code": 200,
    "response_type": "noop",
    "controller_state": "connected",
    "cfgversion": "abc123",
    "attempted_aes_gcm": true,
    "used_aes_gcm": true,
    "fallback_to_cbc": false,
    "interval_seconds": 10,
    "include_blocks": ["system-stats"],
    "has_system_cfg": true,
    "system_cfg_bytes": 42,
    "system_cfg_keys": ["ubntconf", "udapi"],
    "ignored": true,
    "ignored_reason": "system_cfg provisioning is recorded as metadata only",
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
management_vlan: 77
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
			OperationMode  string `json:"operation_mode"`
			InformURL      string `json:"inform_url"`
			ManagementVLAN int    `json:"management_vlan"`
		} `json:"config"`
		Adoption struct {
			State      string `json:"state"`
			Adopted    bool   `json:"adopted"`
			AuthKeySet bool   `json:"authkey_set"`
			CFGVersion string `json:"cfgversion"`
		} `json:"adoption"`
		Runtime struct {
			LastInform struct {
				StatusCode      int      `json:"status_code"`
				ResponseType    string   `json:"response_type"`
				AttemptedAESGCM bool     `json:"attempted_aes_gcm"`
				UsedAESGCM      bool     `json:"used_aes_gcm"`
				FallbackToCBC   bool     `json:"fallback_to_cbc"`
				IntervalSeconds int      `json:"interval_seconds"`
				IncludeBlocks   []string `json:"include_blocks"`
				HasSystemCFG    bool     `json:"has_system_cfg"`
				SystemCFGBytes  int      `json:"system_cfg_bytes"`
				SystemCFGKeys   []string `json:"system_cfg_keys"`
				Ignored         bool     `json:"ignored"`
				IgnoredReason   string   `json:"ignored_reason"`
			} `json:"last_inform"`
		} `json:"runtime"`
	}
	if err := json.Unmarshal([]byte(output), &doc); err != nil {
		t.Fatalf("status JSON invalid: %v\n%s", err, output)
	}
	if doc.Config.OperationMode != "stub" {
		t.Fatalf("OperationMode = %q", doc.Config.OperationMode)
	}
	if doc.Config.InformURL != "http://192.0.2.10:8080/inform" {
		t.Fatalf("InformURL = %q", doc.Config.InformURL)
	}
	if doc.Config.ManagementVLAN != 77 {
		t.Fatalf("ManagementVLAN = %d", doc.Config.ManagementVLAN)
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
	if !doc.Runtime.LastInform.AttemptedAESGCM || !doc.Runtime.LastInform.UsedAESGCM || doc.Runtime.LastInform.FallbackToCBC {
		t.Fatalf("last inform cipher status = %+v", doc.Runtime.LastInform)
	}
	if doc.Runtime.LastInform.IntervalSeconds != 10 ||
		len(doc.Runtime.LastInform.IncludeBlocks) != 1 ||
		doc.Runtime.LastInform.IncludeBlocks[0] != "system-stats" {
		t.Fatalf("last inform interval/include blocks = %+v", doc.Runtime.LastInform)
	}
	if !doc.Runtime.LastInform.HasSystemCFG ||
		doc.Runtime.LastInform.SystemCFGBytes != 42 ||
		len(doc.Runtime.LastInform.SystemCFGKeys) != 2 ||
		!doc.Runtime.LastInform.Ignored ||
		doc.Runtime.LastInform.IgnoredReason == "" {
		t.Fatalf("last inform safe provisioning metadata = %+v", doc.Runtime.LastInform)
	}
}

func TestInformCipherFallbackStatusIsRecorded(t *testing.T) {
	dir := t.TempDir()
	statePath := filepath.Join(dir, "adoption.env")
	statusPath := filepath.Join(dir, "status.json")
	if err := os.WriteFile(statePath, []byte(`AUTHKEY=0123456789abcdef
`), 0o600); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read request body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if len(body) >= 16 && binary.BigEndian.Uint16(body[14:16])&0x08 != 0 {
			http.Error(w, "gcm unavailable", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	output := runStubd(t,
		"-once",
		"-no-discovery",
		"-controller", server.URL+"/inform",
		"-ssh-state", statePath,
		"-status-path", statusPath,
		"-profile", "us8",
		"-ip", "192.0.2.50",
		"-hostname", "fallback-host",
		"-uplink-speed", "profile",
	)
	if strings.Contains(output, "inform send failed") {
		t.Fatalf("inform failed:\n%s", output)
	}

	data, err := os.ReadFile(statusPath)
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		LastInform struct {
			StatusCode      int  `json:"status_code"`
			AttemptedAESGCM bool `json:"attempted_aes_gcm"`
			UsedAESGCM      bool `json:"used_aes_gcm"`
			FallbackToCBC   bool `json:"fallback_to_cbc"`
		} `json:"last_inform"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("status JSON invalid: %v\n%s", err, data)
	}
	if doc.LastInform.StatusCode != http.StatusOK {
		t.Fatalf("status_code = %d, want 200", doc.LastInform.StatusCode)
	}
	if !doc.LastInform.AttemptedAESGCM || doc.LastInform.UsedAESGCM || !doc.LastInform.FallbackToCBC {
		t.Fatalf("cipher status = %+v", doc.LastInform)
	}
}

func TestControllerRestoreDefaultResetsAdoptionState(t *testing.T) {
	dir := t.TempDir()
	statePath := filepath.Join(dir, "adoption.env")
	statusPath := filepath.Join(dir, "status.json")
	if err := os.WriteFile(statePath, []byte(`STATE=connected
AUTHKEY=0123456789abcdef
CFGVERSION=abc123
USE_AES_GCM=true
VERSION=5.0.17.1
`), 0o600); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read request body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if len(body) < 14 {
			http.Error(w, "short inform packet", http.StatusBadRequest)
			return
		}
		packet, err := inform.EncodeJSON(body[8:14], []byte("0123456789abcdef"), []byte(`{"_type":"restore-default"}`), inform.Options{Zlib: true})
		if err != nil {
			t.Errorf("encode response: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(packet); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	defer server.Close()

	_ = runStubd(t,
		"-once",
		"-no-discovery",
		"-controller", server.URL+"/inform",
		"-ssh-state", statePath,
		"-status-path", statusPath,
		"-profile", "us8",
		"-ip", "192.0.2.50",
		"-hostname", "reset-host",
		"-uplink-speed", "profile",
	)

	state, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(state)) != "STATE=factory" {
		t.Fatalf("state after reset = %q", state)
	}
	data, err := os.ReadFile(statusPath)
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		LastInform struct {
			ResetRequested  bool   `json:"reset_requested"`
			ResetApplied    bool   `json:"reset_applied"`
			ResetReason     string `json:"reset_reason"`
			ControllerState string `json:"controller_state"`
		} `json:"last_inform"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("status JSON invalid: %v\n%s", err, data)
	}
	if !doc.LastInform.ResetRequested || !doc.LastInform.ResetApplied || doc.LastInform.ControllerState != "factory" {
		t.Fatalf("reset status = %+v", doc.LastInform)
	}
	if !strings.Contains(doc.LastInform.ResetReason, "restore-default") {
		t.Fatalf("reset reason = %q", doc.LastInform.ResetReason)
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

// Runtime CLI tests exercise unifi-stubd through its public command surface.
// They cover config/CLI precedence, dry-run output, status output, and inform
// behavior without placing tests inside internal packages.
package cli_test

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"net"
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
management_lan:
  enabled: true
  vlan: 42
  mode: metadata-only
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
	if !strings.Contains(output, "operation_mode: bridge-observe") {
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
	if !strings.Contains(output, "management_lan.vlan: 42") {
		t.Fatalf("output did not contain management LAN VLAN:\n%s", output)
	}
	if !strings.Contains(output, "bridge_observe.bridge: br0") {
		t.Fatalf("output did not contain bridge observe fallback:\n%s", output)
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

func TestValidatePortMapRejectsMissingMappedInterface(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(configPath, []byte(`operation_mode: port-map
profile: us8
mac: auto
ip: 192.0.2.50
hostname: auto
uplink_speed: profile
port_mappings:
  - port: 1
    interface: definitely_missing_unifi_stubd0
  - port: 2
    unmapped: true
  - port: 3
    unmapped: true
  - port: 4
    unmapped: true
  - port: 5
    unmapped: true
  - port: 6
    unmapped: true
  - port: 7
    unmapped: true
  - port: 8
    unmapped: true
`), 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := stubdCommand("-validate", "-config", configPath)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("command succeeded; output:\n%s", out)
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
		t.Fatalf("exit = %v, want code 1; output:\n%s", err, out)
	}
	if !strings.Contains(string(out), "port_mappings.interface") ||
		!strings.Contains(string(out), "not found") {
		t.Fatalf("output did not contain missing interface validation:\n%s", out)
	}
}

func TestValidatePortMapAcceptsDisabledAndUnmapped(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(configPath, []byte(`operation_mode: port-map
profile: us8
mac: auto
ip: 192.0.2.50
hostname: auto
uplink_speed: profile
port_mappings:
  - port: 1
    disabled: true
  - port: 2
    unmapped: true
  - port: 3
    unmapped: true
  - port: 4
    unmapped: true
  - port: 5
    unmapped: true
  - port: 6
    unmapped: true
  - port: 7
    unmapped: true
  - port: 8
    unmapped: true
`), 0o600); err != nil {
		t.Fatal(err)
	}
	output := runStubdStdout(t, "-validate", "-config", configPath)
	if !strings.Contains(output, "configuration valid: profile=us8") {
		t.Fatalf("validate output = %q", output)
	}
}

func TestPortMapRendersDisabledAndUnmappedDifferently(t *testing.T) {
	output := runStubd(t,
		"-dry-run",
		"-no-discovery",
		"-profile", "us8",
		"-operation-mode", "port-map",
		"-mac", "02:00:5e:00:53:41",
		"-ip", "192.0.2.50",
		"-hostname", "port-map-test",
		"-uplink-speed", "profile",
		"-port-map", "port=1,unmapped=true",
		"-port-map", "port=2,disabled=true",
		"-port-map", "port=3,unmapped=true",
		"-port-map", "port=4,unmapped=true",
		"-port-map", "port=5,unmapped=true",
		"-port-map", "port=6,unmapped=true",
		"-port-map", "port=7,unmapped=true",
		"-port-map", "port=8,unmapped=true",
	)
	payload := extractDryRunPayloadJSON(t, output)
	var doc struct {
		PortTable []map[string]any `json:"port_table"`
	}
	if err := json.Unmarshal([]byte(payload), &doc); err != nil {
		t.Fatalf("payload JSON invalid: %v\n%s", err, payload)
	}
	disabled := doc.PortTable[1]
	if up := disabled["up"].(bool); up {
		t.Fatal("disabled port is up")
	}
	if speed := int(disabled["speed"].(float64)); speed != 0 {
		t.Fatalf("disabled speed = %d, want 0", speed)
	}
	if enabled := disabled["enable"].(bool); enabled {
		t.Fatal("disabled port is still administratively enabled")
	}
	unmapped := doc.PortTable[2]
	if up := unmapped["up"].(bool); !up {
		t.Fatal("unmapped port is down")
	}
	if speed := int(unmapped["speed"].(float64)); speed == 0 {
		t.Fatal("unmapped port speed was cleared")
	}
	if enabled := unmapped["enable"].(bool); !enabled {
		t.Fatal("unmapped port was administratively disabled")
	}
}

func TestBridgeObserveDefaultsPhysicalUplinkToNormalPort(t *testing.T) {
	output := runStubd(t,
		"-dry-run-plan",
		"-no-discovery",
		"-profile", "us48p500",
		"-operation-mode", "bridge-observe",
		"-bridge-observe-bridge", "vmbr0",
		"-bridge-observe-uplink-interface", "enp100s0",
		"-mac", "02:00:5e:00:53:48",
		"-ip", "192.0.2.50",
		"-hostname", "bridge-observe-test",
		"-uplink-speed", "profile",
	)
	if !strings.Contains(output, "uplink_port: 48") {
		t.Fatalf("output did not contain inferred normal uplink port:\n%s", output)
	}
}

func TestBridgeObserveKeepsDefaultUplinkForSimpleSwitch(t *testing.T) {
	output := runStubd(t,
		"-dry-run-plan",
		"-no-discovery",
		"-profile", "us8",
		"-operation-mode", "bridge-observe",
		"-bridge-observe-bridge", "br0",
		"-bridge-observe-uplink-interface", "eth0",
		"-mac", "02:00:5e:00:53:08",
		"-ip", "192.0.2.50",
		"-hostname", "bridge-observe-us8",
		"-uplink-speed", "profile",
	)
	if !strings.Contains(output, "uplink_port: 0") {
		t.Fatalf("simple switch should keep profile default uplink:\n%s", output)
	}
}

func TestDryRunPayloadHonorsModelOverride(t *testing.T) {
	output := runStubd(t,
		"-dry-run",
		"-no-discovery",
		"-profile", "us8",
		"-mac", "02:00:5e:00:53:13",
		"-ip", "192.0.2.50",
		"-hostname", "gugus-13",
		"-model", "GUGUS13",
		"-model-display", "gugus 13",
		"-uplink-speed", "profile",
	)
	payload := extractDryRunPayloadJSON(t, output)
	var doc struct {
		Hostname     string `json:"hostname"`
		Model        string `json:"model"`
		ModelDisplay string `json:"model_display"`
	}
	if err := json.Unmarshal([]byte(payload), &doc); err != nil {
		t.Fatalf("payload JSON invalid: %v\n%s", err, payload)
	}
	if doc.Hostname != "gugus-13" || doc.Model != "GUGUS13" || doc.ModelDisplay != "gugus 13" {
		t.Fatalf("payload identity = %+v", doc)
	}
}

func TestInvalidManagementLANVLANIsRejected(t *testing.T) {
	cmd := stubdCommand("-dry-run-plan",
		"-profile", "usaggpro",
		"-management-lan-vlan", "4095",
	)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("command succeeded; output:\n%s", out)
	}
	if !strings.Contains(string(out), "invalid management_lan.vlan 4095") {
		t.Fatalf("output did not contain management LAN VLAN validation:\n%s", out)
	}
}

func TestManagementLANPlanReportsPreexistingInterface(t *testing.T) {
	iface, ip := loopbackInterface(t)
	output := runStubdStdout(t,
		"-dry-run-plan",
		"-profile", "us8",
		"-management-lan-enabled",
		"-management-lan-vlan", "42",
		"-management-lan-mode", "preexisting-interface",
		"-management-lan-interface", iface,
		"-management-lan-ip", ip,
		"-management-lan-network", "Management",
	)
	for _, want := range []string{
		"management_lan.vlan: 42",
		"management_lan.mode: preexisting-interface",
		"management_lan.interface: " + iface,
		"management_lan.ip: " + ip,
		"discovery_interface: " + iface,
		"management_lan.actions: use preexisting VLAN interface; no host VLAN changes",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output did not contain %q:\n%s", want, output)
		}
	}
}

func TestManagementLANRejectsGatewayProfiles(t *testing.T) {
	cmd := stubdCommand("-dry-run-plan",
		"-profile", "uxg-lite",
		"-management-lan-enabled",
		"-management-lan-vlan", "42",
	)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("command succeeded; output:\n%s", out)
	}
	if !strings.Contains(string(out), "management_lan is supported for switch profiles only") {
		t.Fatalf("output did not contain management LAN gateway validation:\n%s", out)
	}
}

func TestLegacyManagementVLANConfigIsRejected(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(configPath, []byte(`profile: us8
management_vlan: 42
`), 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := stubdCommand("-validate", "-config", configPath)
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("command succeeded; output:\n%s", out)
	}
	if !strings.Contains(string(out), "field management_vlan not found") {
		t.Fatalf("output did not reject legacy management_vlan:\n%s", out)
	}
}

func TestManagementLANPreexistingInterfaceFeedsPayloadIP(t *testing.T) {
	iface, ip := loopbackInterface(t)
	output := runStubdStdout(t,
		"-dry-run",
		"-profile", "us8",
		"-mac", "02:00:5e:00:53:42",
		"-management-lan-enabled",
		"-management-lan-vlan", "42",
		"-management-lan-mode", "preexisting-interface",
		"-management-lan-interface", iface,
		"-management-lan-ip", ip,
	)
	payload := extractDryRunPayloadJSON(t, output)
	var doc struct {
		IP             string           `json:"ip"`
		ManagementVLAN int              `json:"management_vlan"`
		IfTable        []map[string]any `json:"if_table"`
	}
	if err := json.Unmarshal([]byte(payload), &doc); err != nil {
		t.Fatalf("payload JSON invalid: %v\n%s", err, payload)
	}
	if doc.IP != ip {
		t.Fatalf("payload ip = %q, want %s", doc.IP, ip)
	}
	if doc.ManagementVLAN != 42 {
		t.Fatalf("management_vlan = %d, want 42", doc.ManagementVLAN)
	}
	if len(doc.IfTable) == 0 {
		t.Fatal("if_table is empty")
	}
	if got := int(doc.IfTable[0]["management_vlan"].(float64)); got != 42 {
		t.Fatalf("if_table management_vlan = %d, want 42", got)
	}
}

func TestLLDPSourceIsAcceptedForDryRunPlan(t *testing.T) {
	cmd := stubdCommand("-dry-run-plan",
		"-profile", "usaggpro",
		"-lldp-source", "lldpd",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed; output:\n%s", out)
	}
	if !strings.Contains(string(out), "lldp_source: lldpd") {
		t.Fatalf("output did not contain LLDP source:\n%s", out)
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

func TestProfileValidateRejectsUnknownPayloadField(t *testing.T) {
	dir := t.TempDir()
	profilePath := filepath.Join(dir, "bad-payload.yaml")
	if err := os.WriteFile(profilePath, []byte(`schema_version: 1
name: bad-payload
model: BADPAYLOAD
device_type: usw
ports: 4
payload:
  kind: switch
  unknown_flag: true
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
		t.Fatalf("exit = %v, want go run wrapper code 1; output:\n%s", err, out)
	}
	if !strings.Contains(string(out), "exit status 2") ||
		!strings.Contains(string(out), "unknown_flag") ||
		!strings.Contains(string(out), "PayloadProfile") {
		t.Fatalf("output did not contain strict YAML field error:\n%s", out)
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
management_lan:
  enabled: true
  vlan: 77
  mode: metadata-only
port_overrides:
  - port: 2
    name: Downlink
    mac: 02:00:5e:00:53:22
    ip: 192.0.2.51
    netmask: 255.255.255.0
    role: lan
    network_group: LAN
    speed: 1000
    media: GE
    up: false
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
			ManagementLAN struct {
				Enabled bool `json:"enabled"`
				VLAN    int  `json:"vlan"`
			} `json:"management_lan"`
			PortOverrides []struct {
				Port         int    `json:"port"`
				Name         string `json:"name"`
				MAC          string `json:"mac"`
				IP           string `json:"ip"`
				Netmask      string `json:"netmask"`
				Role         string `json:"role"`
				NetworkGroup string `json:"network_group"`
				Speed        int    `json:"speed"`
				Media        string `json:"media"`
				Up           *bool  `json:"up"`
			} `json:"port_overrides"`
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
	if !doc.Config.ManagementLAN.Enabled || doc.Config.ManagementLAN.VLAN != 77 {
		t.Fatalf("ManagementLAN = %+v", doc.Config.ManagementLAN)
	}
	if len(doc.Config.PortOverrides) != 1 {
		t.Fatalf("PortOverrides = %+v", doc.Config.PortOverrides)
	}
	override := doc.Config.PortOverrides[0]
	if override.Port != 2 ||
		override.Name != "Downlink" ||
		override.MAC != "02:00:5e:00:53:22" ||
		override.IP != "192.0.2.51" ||
		override.Netmask != "255.255.255.0" ||
		override.Role != "lan" ||
		override.NetworkGroup != "LAN" ||
		override.Speed != 1000 ||
		override.Media != "GE" ||
		override.Up == nil ||
		*override.Up {
		t.Fatalf("PortOverride = %+v", override)
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

func extractDryRunPayloadJSON(t *testing.T, output string) string {
	t.Helper()
	const marker = "minimal_inform_payload_json:\n"
	_, payload, ok := strings.Cut(output, marker)
	if !ok {
		t.Fatalf("dry-run payload marker missing:\n%s", output)
	}
	return strings.TrimSpace(payload)
}

func loopbackInterface(t *testing.T) (string, string) {
	t.Helper()
	interfaces, err := net.Interfaces()
	if err != nil {
		t.Fatalf("list interfaces: %v", err)
	}
	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback == 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			if ip := ipNet.IP.To4(); ip != nil {
				return iface.Name, ip.String()
			}
		}
	}
	t.Fatal("no IPv4 loopback interface found")
	return "", ""
}

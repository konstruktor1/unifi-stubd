// Lab-tool tests verify lifecycle evidence filters against sanitized fixtures.
package labtools_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestAssertEventsRecognizesControllerLifecycle verifies discovery, setparam,
// and AES-GCM evidence are attributed to the expected device MAC.
func TestAssertEventsRecognizesControllerLifecycle(t *testing.T) {
	const mac = "02:00:5e:00:53:01"
	eventsPath := filepath.Join(t.TempDir(), "events.jsonl")
	events := strings.Join([]string{
		`{"event":"request","tnbu":{"present":true,"mac":"` + mac + `","flags":3}}`,
		`{"event":"discovery","discovery":{"present":true,"mac":"` + mac + `"}}`,
		`{"event":"response","tnbu":{"present":true,"mac":"` + mac + `","flags":3},"controller_response":{"decoded_with_default_key":true,"recognized":true,"type":"setparam","has_mgmt_cfg":true,"authkey_set":true,"cfgversion_set":true,"use_aes_gcm":true}}`,
		`{"event":"request","tnbu":{"present":true,"mac":"` + mac + `","flags":11}}`,
	}, "\n") + "\n"
	if err := os.WriteFile(eventsPath, []byte(events), 0o600); err != nil {
		t.Fatal(err)
	}

	binary := filepath.Join(t.TempDir(), "assert-events")
	build := exec.Command("go", "build", "-o", binary, "../../lab/stub/tools/assert-events")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build assert-events: %v\n%s", err, output)
	}
	for _, kind := range []string{"inform", "discovery", "setparam", "gcm"} {
		cmd := exec.Command(binary, eventsPath, "0", "--kind", kind, mac)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("assert %s events: %v\n%s", kind, err, output)
		}
		if !strings.Contains(string(output), "found "+kind+" events") {
			t.Fatalf("assert %s output = %q", kind, output)
		}
	}
}

// TestAssertEventsRejectsIncompleteSetParam verifies adoption evidence requires
// all sanitized management fields rather than only the response type.
func TestAssertEventsRejectsIncompleteSetParam(t *testing.T) {
	const mac = "02:00:5e:00:53:02"
	eventsPath := filepath.Join(t.TempDir(), "events.jsonl")
	event := `{"event":"response","tnbu":{"present":true,"mac":"` + mac + `"},"controller_response":{"decoded_with_default_key":true,"recognized":true,"type":"setparam","has_mgmt_cfg":true}}` + "\n"
	if err := os.WriteFile(eventsPath, []byte(event), 0o600); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("go", "run", "../../lab/stub/tools/assert-events", eventsPath, "0", "--kind", "setparam", mac)
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("incomplete setparam was accepted:\n%s", output)
	}
	if !strings.Contains(string(output), "missing setparam events") {
		t.Fatalf("incomplete setparam output = %q", output)
	}
}

// TestControllerAdoptionReportsSanitizedAPIError verifies the fixture reset
// exposes a controller error code without logging the response data.
func TestControllerAdoptionReportsSanitizedAPIError(t *testing.T) {
	const (
		mac       = "02:15:6d:00:00:01"
		sensitive = "sensitive-response-data"
	)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		switch {
		case request.Method == http.MethodPost && request.URL.Path == "/api/login":
			_, _ = writer.Write([]byte(`{"meta":{"rc":"ok"},"data":[]}`))
		case request.Method == http.MethodGet && request.URL.Path == "/api/s/default/stat/device":
			_, _ = writer.Write([]byte(`{"meta":{"rc":"ok"},"data":[{"mac":"` + mac + `","adopted":true,"state":7}]}`))
		case request.Method == http.MethodPost && request.URL.Path == "/api/s/default/cmd/sitemgr":
			writer.WriteHeader(http.StatusBadRequest)
			_, _ = writer.Write([]byte(`{"meta":{"rc":"error","msg":"api.err.DeviceBusy"},"data":{"detail":"` + sensitive + `"}}`))
		default:
			http.NotFound(writer, request)
		}
	}))
	t.Cleanup(server.Close)

	cmd := exec.Command(
		"go", "run", "../../lab/stub/tools/controller-adoption",
		"reset-fixtures", "--base-url", server.URL, "--timeout", "1",
	)
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("busy fixture reset succeeded:\n%s", output)
	}
	if !strings.Contains(string(output), "api.err.DeviceBusy") {
		t.Fatalf("busy fixture error omitted controller code: %q", output)
	}
	if strings.Contains(string(output), sensitive) {
		t.Fatalf("busy fixture error exposed controller response data: %q", output)
	}
}

// TestControllerAdoptionRejectsGlobalFixturePrefix prevents the cleanup helper
// from targeting a real vendor MAC prefix.
func TestControllerAdoptionRejectsGlobalFixturePrefix(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"meta":{"rc":"ok"},"data":[]}`))
	}))
	t.Cleanup(server.Close)

	cmd := exec.Command(
		"go", "run", "../../lab/stub/tools/controller-adoption",
		"reset-fixtures", "--base-url", server.URL, "--mac-prefix", "00:11:22",
	)
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("global fixture prefix was accepted:\n%s", output)
	}
	if !strings.Contains(string(output), "locally administered unicast prefix") {
		t.Fatalf("unexpected global fixture prefix error: %q", output)
	}
}

// TestAssertStatusRequiresCompletedUniqueCommand verifies the Docker gate
// accepts one completed command and rejects a duplicate loop.
func TestAssertStatusRequiresCompletedUniqueCommand(t *testing.T) {
	valid := `{"config":{"controller_interval_min":"5s","controller_interval_max":"1m0s"},"runtime":{"last_inform":{"interval_seconds":12,"schedule":{"effective_interval":"12s","next_delay":"12s","controller_hint_applied":true},"controller_commands":{"last_command":"wg-genkeypair","handled_count":1,"result_sent_count":1}}}}`
	cmd := exec.Command("go", "run", "../../lab/stub/tools/assert-status", "wg-genkeypair-complete")
	cmd.Stdin = strings.NewReader(valid)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("valid command lifecycle rejected: %v\n%s", err, output)
	}

	duplicate := `{"config":{"controller_interval_min":"5s","controller_interval_max":"1m0s"},"runtime":{"last_inform":{"interval_seconds":12,"schedule":{"effective_interval":"12s","next_delay":"12s","controller_hint_applied":true},"controller_commands":{"last_command":"wg-genkeypair","handled_count":1,"result_sent_count":1,"duplicate_count":1}}}}`
	cmd = exec.Command("go", "run", "../../lab/stub/tools/assert-status", "wg-genkeypair-complete")
	cmd.Stdin = strings.NewReader(duplicate)
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("duplicate command lifecycle accepted:\n%s", output)
	}
	if !strings.Contains(string(output), "duplicate command count = 1") {
		t.Fatalf("unexpected duplicate command error: %q", output)
	}
}

// TestAssertStatusAcceptsStableNoCommandController verifies newer controllers
// can explicitly prove that no command was issued during the stability window.
func TestAssertStatusAcceptsStableNoCommandController(t *testing.T) {
	valid := `{"config":{"controller_interval_min":"5s","controller_interval_max":"1m0s"},"runtime":{"last_inform":{"interval_seconds":12,"schedule":{"effective_interval":"12s","next_delay":"12s","controller_hint_applied":true},"controller_commands":{}}}}`
	cmd := exec.Command("go", "run", "../../lab/stub/tools/assert-status", "no-command-stable")
	cmd.Stdin = strings.NewReader(valid)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("stable no-command state rejected: %v\n%s", err, output)
	}

	unexpected := `{"config":{"controller_interval_min":"5s","controller_interval_max":"1m0s"},"runtime":{"last_inform":{"interval_seconds":12,"schedule":{"effective_interval":"12s","next_delay":"12s","controller_hint_applied":true},"controller_commands":{"last_command":"reboot","handled_count":1}}}}`
	cmd = exec.Command("go", "run", "../../lab/stub/tools/assert-status", "no-command-stable")
	cmd.Stdin = strings.NewReader(unexpected)
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("unexpected command state accepted:\n%s", output)
	}
	if !strings.Contains(string(output), "unexpected controller command state") {
		t.Fatalf("unexpected no-command error: %q", output)
	}
}

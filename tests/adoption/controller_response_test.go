// Adoption response tests protect the boundary between controller payloads and
// local state. Provisioning blobs may be present, but only safe summaries and
// explicit adoption keys may reach the store.
package adoption_test

import (
	"bytes"
	"crypto/ecdh"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/konstruktor1/unifi-stubd/internal/adoption"
)

// TestEnvRoundTripPreservesKeyOrderAndValues verifies adoption state persists
// in deterministic env-file form.
func TestEnvRoundTripPreservesKeyOrderAndValues(t *testing.T) {
	path := filepath.Join(t.TempDir(), "adoption.env")
	store := adoption.Store{
		State:      adoption.StateProvisioning,
		InformURL:  "http://192.0.2.10:8080/inform",
		AuthKey:    "test-auth-key-placeholder",
		CFGVersion: "cfg-123",
		UseAESGCM:  true,
		Version:    "5.0.17.1",
	}
	if err := adoption.SaveEnv(path, store); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	const want = "STATE=provisioning\n" +
		"INFORM_URL=http://192.0.2.10:8080/inform\n" +
		"AUTHKEY=test-auth-key-placeholder\n" +
		"CFGVERSION=cfg-123\n" +
		"USE_AES_GCM=true\n" +
		"VERSION=5.0.17.1\n"
	if string(data) != want {
		t.Fatalf("adoption env = %q, want %q", data, want)
	}
	loaded, err := adoption.LoadEnv(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded != store {
		t.Fatalf("loaded store = %+v, want %+v", loaded, store)
	}
}

// TestParseControllerResponseInfoExtractsMgmtCFG verifies allowed management
// config fields are persisted from setparam responses.
func TestParseControllerResponseInfoExtractsMgmtCFG(t *testing.T) {
	info, err := adoption.ParseControllerResponseInfo([]byte(`{
  "_type": "setparam",
  "mgmt_cfg": "cfgversion=abc123\nauthkey=test-auth-key-placeholder\ninform_url=http://192.0.2.10:8080/inform\nuse_aes_gcm=true\nreport_crash=true\n"
}`))
	if err != nil {
		t.Fatal(err)
	}
	if info.Type != "setparam" || !info.HasMgmtCFG || !info.HasStateUpdate {
		t.Fatalf("response summary = %+v", info)
	}
	if info.Store.CFGVersion != "abc123" {
		t.Fatalf("CFGVERSION = %q", info.Store.CFGVersion)
	}
	if info.Store.AuthKey != "test-auth-key-placeholder" {
		t.Fatalf("AuthKey = %q", info.Store.AuthKey)
	}
	if info.Store.InformURL != "http://192.0.2.10:8080/inform" {
		t.Fatalf("InformURL = %q", info.Store.InformURL)
	}
	if !info.Store.UseAESGCM {
		t.Fatal("UseAESGCM was not parsed")
	}
}

// TestParseControllerResponseInfoSummarizesSystemCFGOnly verifies provisioning
// blocks are summarized but not applied.
func TestParseControllerResponseInfoSummarizesSystemCFGOnly(t *testing.T) {
	info, err := adoption.ParseControllerResponseInfo([]byte(`{
  "_type": "setparam",
  "system_cfg": "{\"udapi\":{\"users\":{\"agent\":\"redacted-user\"}},\"ubntconf\":{\"field\":\"do-not-copy\"}}"
}`))
	if err != nil {
		t.Fatal(err)
	}
	if info.Type != "setparam" || !info.HasSystemCFG || !info.Ignored {
		t.Fatalf("response summary = %+v", info)
	}
	if info.HasStateUpdate {
		t.Fatalf("system_cfg should not update adoption state: %+v", info.Store)
	}
	if got := strings.Join(info.SystemCFGKeys, ","); got != "ubntconf,udapi" {
		t.Fatalf("SystemCFGKeys = %q", got)
	}
	data, err := json.Marshal(info)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "redacted-user") || strings.Contains(string(data), "do-not-copy") {
		t.Fatalf("system_cfg content leaked into summary: %s", data)
	}
}

// TestParseControllerResponseInfoNoopCarriesInterval verifies noop responses
// update connected state and heartbeat hints.
func TestParseControllerResponseInfoNoopCarriesInterval(t *testing.T) {
	info, err := adoption.ParseControllerResponseInfo([]byte(`{
  "_type": "noop",
  "interval": 10,
  "include_blocks": ["system-stats", "network_table"]
}`))
	if err != nil {
		t.Fatal(err)
	}
	if info.Type != "noop" || !info.HasStateUpdate || info.Store.State != adoption.StateConnected {
		t.Fatalf("response summary = %+v", info)
	}
	if info.IntervalSeconds != 10 {
		t.Fatalf("IntervalSeconds = %d", info.IntervalSeconds)
	}
	if got := strings.Join(info.IncludeBlocks, ","); got != "system-stats,network_table" {
		t.Fatalf("IncludeBlocks = %q", got)
	}
}

// TestParseControllerResponseInfoUpgradeIsIgnoredButVersionCaptured verifies
// firmware requests are ignored while requested version metadata is recorded.
func TestParseControllerResponseInfoUpgradeIsIgnoredButVersionCaptured(t *testing.T) {
	info, err := adoption.ParseControllerResponseInfo([]byte(`{
  "_type": "upgrade",
  "version": "5.0.17.1"
}`))
	if err != nil {
		t.Fatal(err)
	}
	if info.Type != "upgrade" || !info.Ignored || !info.HasStateUpdate {
		t.Fatalf("response summary = %+v", info)
	}
	if info.Store.Version != "5.0.17.1" {
		t.Fatalf("Version = %q", info.Store.Version)
	}
	if !strings.Contains(info.IgnoredReason, "ignored") {
		t.Fatalf("IgnoredReason = %q", info.IgnoredReason)
	}
}

// TestParseControllerResponseInfoRestoreDefaultRequestsReset verifies reset-like
// response types request a local adoption reset.
func TestParseControllerResponseInfoRestoreDefaultRequestsReset(t *testing.T) {
	info, err := adoption.ParseControllerResponseInfo([]byte(`{"_type":"restore-default"}`))
	if err != nil {
		t.Fatal(err)
	}
	if !info.ResetRequested || info.Store.State != adoption.StateFactory || !info.HasStateUpdate {
		t.Fatalf("response summary = %+v", info)
	}
	if !strings.Contains(info.ResetReason, "restore-default") {
		t.Fatalf("ResetReason = %q", info.ResetReason)
	}
}

// TestParseControllerResponseInfoSetDefaultRequestsReset verifies set-default
// command text triggers only local adoption reset behavior.
func TestParseControllerResponseInfoSetDefaultRequestsReset(t *testing.T) {
	info, err := adoption.ParseControllerResponseInfo([]byte(`{"_type":"setdefault"}`))
	if err != nil {
		t.Fatal(err)
	}
	if !info.ResetRequested || info.Store.State != adoption.StateFactory || !info.HasStateUpdate {
		t.Fatalf("response summary = %+v", info)
	}
	if !strings.Contains(info.ResetReason, "setdefault") {
		t.Fatalf("ResetReason = %q", info.ResetReason)
	}
}

// TestParseControllerResponseInfoCommandRestoreDefaultRequestsReset verifies
// shell-wrapper reset commands are detected without executing them.
func TestParseControllerResponseInfoCommandRestoreDefaultRequestsReset(t *testing.T) {
	info, err := adoption.ParseControllerResponseInfo([]byte(`{
  "_type": "cmd",
  "cmd": "/usr/bin/syswrapper.sh restore-default"
}`))
	if err != nil {
		t.Fatal(err)
	}
	if !info.ResetRequested || info.Store.State != adoption.StateFactory {
		t.Fatalf("response summary = %+v", info)
	}
}

// TestWireGuardCommandLifecycleBuildsMemoryOnlyResult verifies the one allowed
// controller command produces a valid correlated X25519 result and reuses it
// for duplicate request IDs.
func TestWireGuardCommandLifecycleBuildsMemoryOnlyResult(t *testing.T) {
	const requestID = "controller-request-id-must-not-leak"
	info, err := adoption.ParseControllerResponseInfo([]byte(`{
  "_type": "cmd",
  "cmd": "wg-genkeypair",
  "request_id": "` + requestID + `"
}`))
	if err != nil {
		t.Fatal(err)
	}
	if info.CommandName != "wg-genkeypair" || !info.CommandRequestIDSet || info.Ignored {
		t.Fatalf("command summary before handling = %+v", info)
	}
	serializedInfo, err := json.Marshal(info)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(serializedInfo), requestID) {
		t.Fatalf("controller request ID leaked into response summary: %s", serializedInfo)
	}

	var lifecycle adoption.CommandLifecycle
	if err := lifecycle.Handle(&info); err != nil {
		t.Fatal(err)
	}
	if !info.CommandHandled || !info.CommandResultQueued || info.CommandDuplicate {
		t.Fatalf("command summary after handling = %+v", info)
	}
	firstPayload, attached, err := lifecycle.Attach([]byte(`{"mac":"02:00:5e:00:53:01","state":2}`))
	if err != nil {
		t.Fatal(err)
	}
	if !attached {
		t.Fatal("command result was not attached")
	}
	first := decodeWireGuardCommandResult(t, firstPayload)
	if first.EventString != "wg-genkeypair" || first.RequestID != requestID {
		t.Fatalf("command result correlation = %+v", first)
	}
	privateBytes, err := base64.StdEncoding.DecodeString(first.Params.PrivateKey)
	if err != nil {
		t.Fatalf("decode private key: %v", err)
	}
	publicBytes, err := base64.StdEncoding.DecodeString(first.Params.PublicKey)
	if err != nil {
		t.Fatalf("decode public key: %v", err)
	}
	if len(privateBytes) != 32 || privateBytes[0]&7 != 0 || privateBytes[31]&128 != 0 || privateBytes[31]&64 == 0 {
		t.Fatal("private key is not a WireGuard-clamped X25519 scalar")
	}
	privateKey, err := ecdh.X25519().NewPrivateKey(privateBytes)
	if err != nil {
		t.Fatalf("parse private key: %v", err)
	}
	if !bytes.Equal(privateKey.PublicKey().Bytes(), publicBytes) {
		t.Fatal("public key does not match private key")
	}

	lifecycle.MarkResultSent()
	if err := lifecycle.Handle(&info); err != nil {
		t.Fatal(err)
	}
	secondPayload, attached, err := lifecycle.Attach([]byte(`{"mac":"02:00:5e:00:53:01","state":2}`))
	if err != nil {
		t.Fatal(err)
	}
	if !attached {
		t.Fatal("duplicate command result was not attached")
	}
	second := decodeWireGuardCommandResult(t, secondPayload)
	if first.Params != second.Params {
		t.Fatal("duplicate request generated a different key pair")
	}
	summary := lifecycle.Summary()
	if summary.HandledCount != 1 || summary.ResultSentCount != 1 || summary.DuplicateCount != 1 || !summary.Pending {
		t.Fatalf("command lifecycle summary = %+v", summary)
	}
	serializedSummary, err := json.Marshal(summary)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(serializedSummary), first.Params.PrivateKey) || strings.Contains(string(serializedSummary), requestID) {
		t.Fatalf("command secret or request ID leaked into summary: %s", serializedSummary)
	}
}

// TestWireGuardCommandRequiresNarrowShape verifies arguments and missing
// correlation IDs are rejected without preparing a result.
func TestWireGuardCommandRequiresNarrowShape(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "missing request ID", body: `{"_type":"cmd","cmd":"wg-genkeypair"}`},
		{name: "unexpected params", body: `{"_type":"cmd","cmd":"wg-genkeypair","request_id":"id","params":{"path":"/tmp/key"}}`},
		{name: "unexpected args", body: `{"_type":"cmd","cmd":"wg-genkeypair","request_id":"id","args":["shell"]}`},
		{name: "unsupported command", body: `{"_type":"cmd","cmd":"shell","request_id":"id"}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			info, err := adoption.ParseControllerResponseInfo([]byte(test.body))
			if err != nil {
				t.Fatal(err)
			}
			if !info.Ignored || info.CommandHandled || info.CommandResultQueued {
				t.Fatalf("unsafe command summary = %+v", info)
			}
			var lifecycle adoption.CommandLifecycle
			if err := lifecycle.Handle(&info); err != nil {
				t.Fatal(err)
			}
			_, attached, err := lifecycle.Attach([]byte(`{"state":2}`))
			if err != nil {
				t.Fatal(err)
			}
			if attached {
				t.Fatal("unsafe command produced a result")
			}
		})
	}
}

type wireGuardCommandResult struct {
	EventString string `json:"event_string"`
	RequestID   string `json:"request_id"`
	Params      struct {
		PrivateKey string `json:"private_key"`
		PublicKey  string `json:"public_key"`
	} `json:"params"`
}

func decodeWireGuardCommandResult(t *testing.T, payload []byte) wireGuardCommandResult {
	t.Helper()
	var notification struct {
		InformAsNotification bool                   `json:"inform_as_notif"`
		NotificationReason   string                 `json:"notif_reason"`
		NotificationPayload  wireGuardCommandResult `json:"notif_payload"`
		EventString          string                 `json:"event_string"`
		RequestID            string                 `json:"request_id"`
		Params               map[string]any         `json:"params"`
	}
	if err := json.Unmarshal(payload, &notification); err != nil {
		t.Fatalf("decode command result: %v", err)
	}
	if !notification.InformAsNotification || notification.NotificationReason != "cmd" {
		t.Fatalf("command notification metadata = enabled %t reason %q", notification.InformAsNotification, notification.NotificationReason)
	}
	if notification.EventString != "" || notification.RequestID != "" || notification.Params != nil {
		t.Fatal("command result leaked into the ordinary inform root")
	}
	return notification.NotificationPayload
}

// FuzzParseControllerResponseInfo verifies arbitrary controller JSON bytes do
// not panic the response sanitizer.
func FuzzParseControllerResponseInfo(f *testing.F) {
	f.Add([]byte(`{"_type":"noop","interval":10}`))
	f.Add([]byte(`{"_type":"setparam","mgmt_cfg":"cfgversion=abc\nauthkey=0123456789abcdef\nuse_aes_gcm=true\n"}`))
	f.Add([]byte(`{"_type":"setparam","system_cfg":"{\"udapi\":{},\"ubntconf\":{}}"}`))
	f.Add([]byte(`{"_type":"setdefault"}`))
	f.Add([]byte{})

	f.Fuzz(func(_ *testing.T, data []byte) {
		_, _ = adoption.ParseControllerResponseInfo(data)
	})
}

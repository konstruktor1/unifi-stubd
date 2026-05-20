// Adoption response tests protect the boundary between controller payloads and
// local state. Provisioning blobs may be present, but only safe summaries and
// explicit adoption keys may reach the store.
package adoption_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/konstruktor1/unifi-stubd/internal/adoption"
)

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

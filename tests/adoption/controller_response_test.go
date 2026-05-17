package adoption_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/konstruktor1/unifi-stubd/internal/adoption"
)

func TestParseControllerResponseInfoExtractsMgmtCFG(t *testing.T) {
	info, err := adoption.ParseControllerResponseInfo([]byte(`{
  "_type": "setparam",
  "mgmt_cfg": "cfgversion=abc123\nauthkey=secret-test-key\ninform_url=http://192.0.2.10:8080/inform\nuse_aes_gcm=true\nreport_crash=true\n"
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
	if info.Store.AuthKey != "secret-test-key" {
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
  "system_cfg": "{\"udapi\":{\"users\":{\"agent\":\"secret\"}},\"ubntconf\":{\"token\":\"do-not-copy\"}}"
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
	if strings.Contains(string(data), "secret") || strings.Contains(string(data), "do-not-copy") {
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

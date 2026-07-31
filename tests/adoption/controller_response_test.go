package adoption_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/konstruktor1/unifi-stubd/internal/adoption"
)

const (
	individualKey = "00112233445566778899aabbccddeeff"
	controllerURL = "http://192.0.2.10:8080/inform"
)

// TestAdoptionStatePersistsOnlyInformIdentity fixes the complete local state
// needed for later informs and the private file representation that contains it.
func TestAdoptionStatePersistsOnlyInformIdentity(t *testing.T) {
	path := filepath.Join(t.TempDir(), "adoption.env")
	want := adoption.Store{
		State: adoption.StateConnected, InformURL: controllerURL, AuthKey: individualKey,
		CFGVersion: "42", UseAESGCM: true,
	}
	if err := adoption.SaveEnv(path, want); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	wantFile := "STATE=connected\nINFORM_URL=" + controllerURL + "\nAUTHKEY=" + individualKey +
		"\nCFGVERSION=42\nUSE_AES_GCM=true\n"
	if string(data) != wantFile {
		t.Fatalf("persisted adoption identity = %q; want %q", data, wantFile)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("adoption state mode = %v; want 0600", info.Mode().Perm())
	}
	got, err := adoption.LoadEnv(path)
	if err != nil || got != want {
		t.Fatalf("loaded adoption identity = %+v, error %v; want %+v", got, err, want)
	}
}

// TestControllerAdoptionPersistsOnlyInformIdentity accepts only fields needed
// by later informs; unrelated controller provisioning remains metadata.
func TestControllerAdoptionPersistsOnlyInformIdentity(t *testing.T) {
	response := `{"_type":"setparam","mgmt_cfg":"cfgversion=42\nauthkey=` + individualKey +
		`\ninform_url=` + controllerURL + `\nuse_aes_gcm=true\nreport_crash=true","system_cfg":{"hostname":"controller-owned"}}`
	info, err := adoption.ParseControllerResponseInfo([]byte(response))
	if err != nil {
		t.Fatal(err)
	}
	want := adoption.Store{
		State: adoption.StateConnected, InformURL: controllerURL, AuthKey: individualKey,
		CFGVersion: "42", UseAESGCM: true,
	}
	if info.Store != want || !info.HasMgmtCFG || !info.HasSystemCFG || !info.Ignored {
		t.Fatalf("sanitized setparam = %+v; want store %+v and provisioning metadata only", info, want)
	}
}

// TestInvalidControllerIdentityCannotBePersisted keeps the public adoption key
// and credential-bearing controller URLs outside the individual device store.
func TestInvalidControllerIdentityCannotBePersisted(t *testing.T) {
	for name, store := range map[string]adoption.Store{
		"public default key": {State: adoption.StateConnected, AuthKey: "ba86f2bbe107c7c57eb5f2690775c712"},
		"URL credentials":    {State: adoption.StateProvisioning, InformURL: "http://user:secret@192.0.2.10/inform"},
	} {
		t.Run(name, func(t *testing.T) {
			if err := adoption.SaveEnv(filepath.Join(t.TempDir(), "adoption.env"), store); err == nil {
				t.Fatal("unsafe controller identity was persisted")
			}
		})
	}
}

// TestControllerHostMutatorsAreMetadataOnly is the structural safety boundary:
// these controller messages never create an adoption-state update.
func TestControllerHostMutatorsAreMetadataOnly(t *testing.T) {
	for name, response := range map[string]string{
		"restart":          `{"_type":"reboot"}`,
		"upgrade":          `{"_type":"upgrade"}`,
		"network metadata": `{"_type":"geo_ip_update"}`,
		"shell":            `{"_type":"cmd","cmd":"run","args":"touch /controller-owned"}`,
	} {
		t.Run(name, func(t *testing.T) {
			info, err := adoption.ParseControllerResponseInfo([]byte(response))
			if err != nil {
				t.Fatal(err)
			}
			if !info.Ignored || info.HasStateUpdate || info.Store != (adoption.Store{}) {
				t.Fatalf("host mutator escaped metadata boundary: %+v", info)
			}
		})
	}
}

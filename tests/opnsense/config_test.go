package opnsense_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/konstruktor1/unifi-stubd/internal/opnsense"
)

func TestDecodeSourceConfigValidatesMappings(t *testing.T) {
	t.Parallel()

	cfg, err := opnsense.DecodeSourceConfig([]byte(`base_url: https://127.0.0.1
api_key_file: /root/key
api_secret_file: /root/secret
timeout_ms: 1500
uplink_port: 3
gateway_status: true
interfaces:
  - port: 3
    interface: ixl0
    name: WAN SFP+
    role: wan
    network_group: WAN
wan_health:
  source: static
  interval_seconds: 10
  timeout_ms: 1000
  targets: []
`))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.TimeoutMS != 1500 || cfg.UplinkPort != 3 || !cfg.GatewayStatus {
		t.Fatalf("source config = %+v", cfg)
	}
	if cfg.Interfaces[0].Role != testRoleWAN || cfg.Interfaces[0].Interface != testInterfaceIXL0 {
		t.Fatalf("mapping = %+v", cfg.Interfaces[0])
	}
}

func TestLoadCredentialsReadsFilesWithoutLeakingValues(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	keyPath := filepath.Join(dir, "key")
	secretPath := filepath.Join(dir, "secret")
	if err := os.WriteFile(keyPath, []byte("key-from-file\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(secretPath, []byte("secret-from-file\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	credentials, err := opnsense.LoadCredentials(opnsense.SourceConfig{
		APIKeyFile:    keyPath,
		APISecretFile: secretPath,
	})
	if err != nil {
		t.Fatal(err)
	}
	if credentials.Key != "key-from-file" || credentials.Secret != "secret-from-file" {
		t.Fatalf("credentials = %+v", credentials)
	}

	_, err = opnsense.LoadCredentials(opnsense.SourceConfig{
		APIKeyFile:    keyPath,
		APISecretFile: filepath.Join(dir, "missing-secret"),
	})
	if err == nil {
		t.Fatal("LoadCredentials missing secret error = nil")
	}
	if strings.Contains(err.Error(), "secret-from-file") {
		t.Fatalf("error leaked secret: %s", err)
	}
}

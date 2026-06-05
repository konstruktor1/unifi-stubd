package config_test

import (
	"strings"
	"testing"

	"github.com/konstruktor1/unifi-stubd/internal/config"
)

func TestMigrateDataNormalizesLegacyAliases(t *testing.T) {
	result, err := config.MigrateData([]byte(`controller: http://192.0.2.10:8080/inform
operation_mode: observe
observe_bridge: vmbr0
observe_interface: eno1
port_map:
  - port: 1
    interface: eno1
`))
	if err != nil {
		t.Fatal(err)
	}
	if !result.Changed {
		t.Fatal("migration did not report changes")
	}
	text := string(result.Data)
	for _, legacy := range []string{"\ncontroller:", "\nobserve_bridge:", "\nobserve_interface:", "\nport_map:"} {
		if strings.Contains(text, legacy) {
			t.Fatalf("migrated YAML still contains %s:\n%s", legacy, text)
		}
	}
	cfg, err := config.Decode(result.Data)
	if err != nil {
		t.Fatalf("migrated YAML did not decode: %v\n%s", err, result.Data)
	}
	if cfg.ControllerURL != "http://192.0.2.10:8080/inform" {
		t.Fatalf("ControllerURL = %q", cfg.ControllerURL)
	}
	if cfg.OperationMode != "bridge-observe" {
		t.Fatalf("OperationMode = %q", cfg.OperationMode)
	}
	if cfg.BridgeObserve.Bridge != "vmbr0" || cfg.BridgeObserve.UplinkInterface != testInterface {
		t.Fatalf("BridgeObserve = %+v", cfg.BridgeObserve)
	}
	if len(cfg.PortMappings) != 1 || cfg.PortMappings[0].Interface != testInterface {
		t.Fatalf("PortMappings = %+v", cfg.PortMappings)
	}
}

func TestMigrateDataRejectsConflictingAliases(t *testing.T) {
	_, err := config.MigrateData([]byte(`controller_url: http://192.0.2.10:8080/inform
controller: http://192.0.2.20:8080/inform
`))
	if err == nil {
		t.Fatal("conflicting aliases were accepted")
	}
	if !strings.Contains(err.Error(), "different values") {
		t.Fatalf("error = %v", err)
	}
}

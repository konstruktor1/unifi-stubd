//nolint:goconst // Repeated configured values document overlay preservation.
package config_test

import (
	"testing"

	"github.com/konstruktor1/unifi-stubd/internal/config"
)

func TestDecodeOverlaysSparseYAMLOnSafeDefaults(t *testing.T) {
	if profile := config.Default().Profile; profile != "" {
		t.Fatalf("implicit device profile = %q", profile)
	}
	cfg, err := config.Decode([]byte(`
operation_mode: observe
controller_url: http://192.0.2.10:8080/inform
profile: us8
observe_interface: eno1
traffic_rates_enabled: true
port_overrides:
  - port: 1
    interface: eno1
    speed: 1000
  - port: 5
    up: false
state_path: /var/lib/unifi-stubd/lab.env
`))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.OperationMode != "observe" || cfg.Profile != "us8" || cfg.ObserveInterface != "eno1" {
		t.Fatalf("YAML selection = mode %q profile %q interface %q", cfg.OperationMode, cfg.Profile, cfg.ObserveInterface)
	}
	if cfg.ControllerURL != "http://192.0.2.10:8080/inform" || !cfg.TrafficRatesEnabled {
		t.Fatalf("runtime overlay = controller %q rates %t", cfg.ControllerURL, cfg.TrafficRatesEnabled)
	}
	if len(cfg.PortOverrides) != 2 || cfg.PortOverrides[0].Interface != "eno1" || cfg.PortOverrides[0].Speed != 1000 {
		t.Fatalf("port overrides = %+v", cfg.PortOverrides)
	}
	if cfg.PortOverrides[1].Up == nil || *cfg.PortOverrides[1].Up {
		t.Fatalf("link-down override = %+v", cfg.PortOverrides[1])
	}
	if cfg.IntervalSeconds != 10 || cfg.SSHUser != "ubnt" || cfg.InstanceGuard != "fail" {
		t.Fatalf("defaults were not retained: interval %d user %q guard %q", cfg.IntervalSeconds, cfg.SSHUser, cfg.InstanceGuard)
	}
	if cfg.StatePath != "/var/lib/unifi-stubd/lab.env" || cfg.StatusPath != "/var/lib/unifi-stubd/status.json" {
		t.Fatalf("state paths = %q/%q", cfg.StatePath, cfg.StatusPath)
	}
}

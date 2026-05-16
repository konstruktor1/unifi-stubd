package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/konstruktor1/unifi-stubd/internal/config"
)

func TestDefaultSeparatesConfigAndStatePaths(t *testing.T) {
	cfg := config.Default()
	if cfg.OperationMode != "stub" {
		t.Fatalf("OperationMode = %q, want stub", cfg.OperationMode)
	}
	if cfg.SSHHostKeyPath != "/etc/unifi-stubd/ssh_host_rsa_key" {
		t.Fatalf("SSHHostKeyPath = %q", cfg.SSHHostKeyPath)
	}
	if cfg.StatePath != "/var/lib/unifi-stubd/adoption.env" {
		t.Fatalf("StatePath = %q", cfg.StatePath)
	}
	if cfg.StatusPath != "/var/lib/unifi-stubd/status.json" {
		t.Fatalf("StatusPath = %q", cfg.StatusPath)
	}
	if cfg.Profile == "" {
		t.Fatal("Profile default is empty")
	}
}

func TestLoadMergesWithDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(`controller_url: http://10.10.0.30:8080/inform
profile: us16p150
operation_mode: observe
observe_interface: eth0
observe_bridge: vmbr0
lldp_source: lldpd
ssh_listen: 0.0.0.0:22
state_path: /tmp/unifi-stubd/adoption.env
status_path: /tmp/unifi-stubd/status.json
`), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ControllerURL != "http://10.10.0.30:8080/inform" {
		t.Fatalf("ControllerURL = %q", cfg.ControllerURL)
	}
	if cfg.OperationMode != "observe" {
		t.Fatalf("OperationMode = %q", cfg.OperationMode)
	}
	if cfg.ObserveInterface != "eth0" {
		t.Fatalf("ObserveInterface = %q", cfg.ObserveInterface)
	}
	if cfg.ObserveBridge != "vmbr0" {
		t.Fatalf("ObserveBridge = %q", cfg.ObserveBridge)
	}
	if cfg.LLDPSource != "lldpd" {
		t.Fatalf("LLDPSource = %q", cfg.LLDPSource)
	}
	if cfg.TrafficSource != "off" {
		t.Fatalf("TrafficSource default was not preserved: %q", cfg.TrafficSource)
	}
	if cfg.SSHListen != "0.0.0.0:22" {
		t.Fatalf("SSHListen = %q", cfg.SSHListen)
	}
	if cfg.SSHUser != "ubnt" {
		t.Fatalf("SSHUser default was not preserved: %q", cfg.SSHUser)
	}
	if cfg.StatePath != "/tmp/unifi-stubd/adoption.env" {
		t.Fatalf("StatePath = %q", cfg.StatePath)
	}
	if cfg.StatusPath != "/tmp/unifi-stubd/status.json" {
		t.Fatalf("StatusPath = %q", cfg.StatusPath)
	}
}

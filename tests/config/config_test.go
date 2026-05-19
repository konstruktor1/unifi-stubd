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
	if cfg.SSHHostKeyPath != "/var/lib/unifi-stubd/ssh_host_rsa_key" {
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
	if err := os.WriteFile(path, []byte(`controller_url: http://192.0.2.10:8080/inform
profile: us16p150
profile_file: /etc/unifi-stubd/profiles/lab.yaml
profile_dir: /etc/unifi-stubd/profiles.d
operation_mode: observe
observe_interface: eth0
observe_bridge: vmbr0
discovery_interface: eth0
discovery_targets:
  - 192.0.2.255:10001
  - 233.89.188.1:10001
management_vlan: 42
uplink_port: 1
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
    interface: eth1
    mac: 02:00:5e:00:53:02
    ip: 192.0.2.2
    netmask: 255.255.255.0
    role: lan
    network_group: LAN
    speed: 1000
  - port: 5
    up: false
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
	if cfg.ControllerURL != "http://192.0.2.10:8080/inform" {
		t.Fatalf("ControllerURL = %q", cfg.ControllerURL)
	}
	if cfg.OperationMode != "observe" {
		t.Fatalf("OperationMode = %q", cfg.OperationMode)
	}
	if cfg.ProfileFile != "/etc/unifi-stubd/profiles/lab.yaml" {
		t.Fatalf("ProfileFile = %q", cfg.ProfileFile)
	}
	if cfg.ProfileDir != "/etc/unifi-stubd/profiles.d" {
		t.Fatalf("ProfileDir = %q", cfg.ProfileDir)
	}
	if cfg.ObserveInterface != "eth0" {
		t.Fatalf("ObserveInterface = %q", cfg.ObserveInterface)
	}
	if cfg.ObserveBridge != "vmbr0" {
		t.Fatalf("ObserveBridge = %q", cfg.ObserveBridge)
	}
	if len(cfg.DiscoveryTargets) != 2 || cfg.DiscoveryTargets[0] != "192.0.2.255:10001" {
		t.Fatalf("DiscoveryTargets = %#v", cfg.DiscoveryTargets)
	}
	if cfg.DiscoveryInterface != "eth0" {
		t.Fatalf("DiscoveryInterface = %q", cfg.DiscoveryInterface)
	}
	if cfg.ManagementVLAN != 42 {
		t.Fatalf("ManagementVLAN = %d", cfg.ManagementVLAN)
	}
	if cfg.UplinkPort != 1 {
		t.Fatalf("UplinkPort = %d", cfg.UplinkPort)
	}
	if cfg.UplinkNeighbor == nil {
		t.Fatal("UplinkNeighbor is nil")
	}
	if cfg.UplinkNeighbor.MAC != "02:aa:bb:cc:dd:01" || cfg.UplinkNeighbor.VLAN != 1 || cfg.UplinkNeighbor.Type != "usw" {
		t.Fatalf("UplinkNeighbor = %+v", cfg.UplinkNeighbor)
	}
	if len(cfg.PortNeighbors) != 1 {
		t.Fatalf("len(PortNeighbors) = %d, want 1", len(cfg.PortNeighbors))
	}
	if cfg.PortNeighbors[0].Port != 2 || cfg.PortNeighbors[0].MAC != "02:00:5e:00:53:03" {
		t.Fatalf("first PortNeighbor = %+v", cfg.PortNeighbors[0])
	}
	if len(cfg.PortOverrides) != 2 {
		t.Fatalf("len(PortOverrides) = %d, want 2", len(cfg.PortOverrides))
	}
	if cfg.PortOverrides[0].Port != 2 ||
		cfg.PortOverrides[0].Interface != "eth1" ||
		cfg.PortOverrides[0].MAC != "02:00:5e:00:53:02" ||
		cfg.PortOverrides[0].IP != "192.0.2.2" ||
		cfg.PortOverrides[0].Netmask != "255.255.255.0" ||
		cfg.PortOverrides[0].Role != "lan" ||
		cfg.PortOverrides[0].NetworkGroup != "LAN" ||
		cfg.PortOverrides[0].Speed != 1000 {
		t.Fatalf("first PortOverride = %+v", cfg.PortOverrides[0])
	}
	if cfg.PortOverrides[1].Up == nil || *cfg.PortOverrides[1].Up {
		t.Fatalf("second PortOverride.Up = %v, want false", cfg.PortOverrides[1].Up)
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

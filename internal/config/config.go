package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// DefaultPath is the system-wide YAML configuration path.
const DefaultPath = "/etc/unifi-stubd/config.yaml"

const automaticValue = "auto"

// Config describes the runtime settings loaded from YAML and CLI flags.
type Config struct {
	// OperationMode selects the runtime network behavior.
	OperationMode string `yaml:"operation_mode"`
	// ControllerURL is the UniFi inform endpoint.
	ControllerURL string `yaml:"controller_url"`
	// Profile selects the device profile to emulate.
	Profile string `yaml:"profile"`
	// MAC is the fake device MAC address or auto for a derived address.
	MAC string `yaml:"mac"`
	// IP is the IPv4 address reported to the controller.
	IP string `yaml:"ip"`
	// Hostname is the reported device hostname or auto for the OS hostname.
	Hostname string `yaml:"hostname"`
	// Model overrides the UniFi model identifier from the selected profile.
	Model string `yaml:"model"`
	// ModelDisplay overrides the display name from the selected profile.
	ModelDisplay string `yaml:"model_display"`
	// Ports overrides the switch port count from the selected profile.
	Ports int `yaml:"ports"`
	// LinkSpeed overrides regular switch port speed in Mbps.
	LinkSpeed int `yaml:"link_speed"`
	// UplinkSpeed is an explicit Mbps value, auto, or profile.
	UplinkSpeed string `yaml:"uplink_speed"`
	// UplinkPort overrides the profile-selected uplink port when positive.
	UplinkPort int `yaml:"uplink_port"`
	// UplinkNeighbor adds a configured fake neighbor to the selected uplink.
	UplinkNeighbor *UplinkNeighbor `yaml:"uplink_neighbor"`
	// PortOverrides applies per-port runtime overrides after profile generation.
	PortOverrides []PortOverride `yaml:"port_overrides"`
	// ObserveInterface is the host interface used for passive link data.
	ObserveInterface string `yaml:"observe_interface"`
	// ObserveBridge is the Linux bridge used for passive FDB data.
	ObserveBridge string `yaml:"observe_bridge"`
	// LLDPSource selects the passive LLDP source.
	LLDPSource string `yaml:"lldp_source"`
	// TrafficSource selects the passive traffic metadata source.
	TrafficSource string `yaml:"traffic_source"`
	// Version overrides the firmware version from the selected profile.
	Version string `yaml:"version"`
	// IntervalSeconds is the loop interval for discovery and inform traffic.
	IntervalSeconds int `yaml:"interval_seconds"`
	// NoDiscovery disables UDP discovery announcements.
	NoDiscovery bool `yaml:"no_discovery"`
	// SSHListen enables the built-in adoption SSH server when set.
	SSHListen string `yaml:"ssh_listen"`
	// SSHUser is the username accepted by the adoption SSH server.
	SSHUser string `yaml:"ssh_user"`
	// SSHPassword is the password accepted by the adoption SSH server.
	SSHPassword string `yaml:"ssh_password"`
	// SSHHostKeyPath stores the persistent adoption SSH host key.
	SSHHostKeyPath string `yaml:"ssh_host_key_path"`
	// StatePath stores adoption state learned from the controller.
	StatePath string `yaml:"state_path"`
	// StatusPath stores non-sensitive runtime status for health checks.
	StatusPath string `yaml:"status_path"`
}

// Default returns the built-in runtime defaults.
func Default() Config {
	return Config{
		OperationMode:    "stub",
		ControllerURL:    "",
		Profile:          "us16p150",
		MAC:              automaticValue,
		IP:               "192.168.1.50",
		Hostname:         automaticValue,
		Model:            "",
		ModelDisplay:     "",
		Ports:            0,
		LinkSpeed:        0,
		UplinkSpeed:      automaticValue,
		UplinkPort:       0,
		UplinkNeighbor:   nil,
		PortOverrides:    nil,
		ObserveInterface: "",
		ObserveBridge:    "",
		LLDPSource:       "off",
		TrafficSource:    "off",
		Version:          "",
		IntervalSeconds:  10,
		NoDiscovery:      false,
		SSHListen:        "",
		SSHUser:          "ubnt",
		SSHPassword:      "ubnt",
		SSHHostKeyPath:   "/etc/unifi-stubd/ssh_host_rsa_key",
		StatePath:        "/var/lib/unifi-stubd/adoption.env",
		StatusPath:       "/var/lib/unifi-stubd/status.json",
	}
}

// UplinkNeighbor describes a configured fake upstream neighbor.
type UplinkNeighbor struct {
	// MAC is the neighbor MAC address to expose on the uplink port.
	MAC string `yaml:"mac"`
	// VLAN is the optional VLAN associated with the neighbor.
	VLAN int `yaml:"vlan"`
	// Type is the controller-facing neighbor type.
	Type string `yaml:"type"`
	// Age is the controller-facing MAC-table age counter.
	Age int `yaml:"age"`
	// Uptime is the number of seconds the neighbor has been visible.
	Uptime int `yaml:"uptime"`
}

// PortOverride describes one per-port YAML override.
type PortOverride struct {
	// Port is the one-based switch port index.
	Port int `yaml:"port"`
	// Name overrides the controller-facing port label when set.
	Name string `yaml:"name"`
	// Speed overrides the negotiated speed in Mbps when positive.
	Speed int `yaml:"speed"`
	// Media overrides the controller-facing media label when set.
	Media string `yaml:"media"`
	// Up overrides link state when set.
	Up *bool `yaml:"up"`
}

// Load reads path and overlays its YAML values on top of Default.
func Load(path string) (Config, error) {
	cfg := Default()
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, fmt.Errorf("read config %s: %w", path, err)
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse config %s: %w", path, err)
	}
	return cfg, nil
}

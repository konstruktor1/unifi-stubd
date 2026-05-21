// Package config defines the user-facing YAML schema shared by packaged config
// files, CLI validation, and runtime startup. Defaults stay explicit so new
// fields are visible at the service boundary.
package config

import (
	"bytes"
	"fmt"
	"os"

	"github.com/konstruktor1/unifi-stubd/internal/device/payload"
	"gopkg.in/yaml.v3"
)

// DefaultPath is the system-wide YAML configuration path.
const DefaultPath = "/etc/unifi-stubd/config.yaml"

const (
	automaticValue = "auto"
	sourceOffValue = "off"
)

// Config describes the runtime settings loaded from YAML and CLI flags.
type Config struct {
	// OperationMode selects the runtime network behavior.
	OperationMode string `yaml:"operation_mode"`
	// ControllerURL is the UniFi inform endpoint.
	ControllerURL string `yaml:"controller_url"`
	// Profile selects the device profile to emulate.
	Profile string `yaml:"profile"`
	// ProfileFile loads one external device profile YAML file.
	ProfileFile string `yaml:"profile_file"`
	// ProfileDir loads external device profile YAML files from a directory.
	ProfileDir string `yaml:"profile_dir"`
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
	// PortNeighbors adds configured fake neighbors to specific ports.
	PortNeighbors []PortNeighbor `yaml:"port_neighbors"`
	// PortOverrides applies per-port runtime overrides after profile generation.
	PortOverrides []PortOverride `yaml:"port_overrides"`
	// BridgeObserve configures bridge-to-virtual-port observation.
	BridgeObserve BridgeObserve `yaml:"bridge_observe"`
	// PortMappings map controller ports to host interfaces or explicit states.
	PortMappings []PortMapping `yaml:"port_mappings"`
	// ObserveInterface is the host interface used for passive link data.
	ObserveInterface string `yaml:"observe_interface"`
	// ObserveBridge is the Linux bridge used for passive FDB data.
	ObserveBridge string `yaml:"observe_bridge"`
	// LLDPSource selects the passive LLDP source.
	LLDPSource string `yaml:"lldp_source"`
	// TrafficSource selects the passive traffic metadata source.
	TrafficSource string `yaml:"traffic_source"`
	// LogSource selects optional read-only runtime log metadata.
	LogSource string `yaml:"log_source"`
	// ProcSource selects optional Linux procfs metadata.
	ProcSource string `yaml:"proc_source"`
	// DBusEnabled enables optional D-Bus connectivity checks.
	DBusEnabled bool `yaml:"dbus_enabled"`
	// DBusBus selects the system or session bus for optional D-Bus checks.
	DBusBus string `yaml:"dbus_bus"`
	// SyslogPath is the syslog file read by log_source: syslog.
	SyslogPath string `yaml:"syslog_path"`
	// Version overrides the firmware version from the selected profile.
	Version string `yaml:"version"`
	// IntervalSeconds is the loop interval for discovery and inform traffic.
	IntervalSeconds int `yaml:"interval_seconds"`
	// NoDiscovery disables UDP discovery announcements.
	NoDiscovery bool `yaml:"no_discovery"`
	// DiscoveryInterface selects the local interface used for UDP discovery sends.
	DiscoveryInterface string `yaml:"discovery_interface"`
	// DiscoveryTargets adds explicit UDP discovery targets.
	DiscoveryTargets []string `yaml:"discovery_targets"`
	// ManagementLAN controls how switch management VLAN metadata maps to a local interface.
	ManagementLAN ManagementLAN `yaml:"management_lan"`
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
		OperationMode:      "stub",
		ControllerURL:      "",
		Profile:            "us16p150",
		ProfileFile:        "",
		ProfileDir:         "",
		MAC:                automaticValue,
		IP:                 "192.168.1.50",
		Hostname:           automaticValue,
		Model:              "",
		ModelDisplay:       "",
		Ports:              0,
		LinkSpeed:          0,
		UplinkSpeed:        automaticValue,
		UplinkPort:         0,
		UplinkNeighbor:     nil,
		PortNeighbors:      nil,
		PortOverrides:      nil,
		BridgeObserve:      BridgeObserve{},
		PortMappings:       nil,
		ObserveInterface:   "",
		ObserveBridge:      "",
		LLDPSource:         sourceOffValue,
		TrafficSource:      sourceOffValue,
		LogSource:          sourceOffValue,
		ProcSource:         sourceOffValue,
		DBusEnabled:        false,
		DBusBus:            "system",
		SyslogPath:         "/var/log/messages",
		Version:            "",
		IntervalSeconds:    10,
		NoDiscovery:        false,
		DiscoveryInterface: "",
		DiscoveryTargets:   nil,
		ManagementLAN:      ManagementLAN{},
		SSHListen:          "",
		SSHUser:            "ubnt",
		SSHPassword:        "ubnt",
		SSHHostKeyPath:     "/var/lib/unifi-stubd/ssh_host_rsa_key",
		StatePath:          "/var/lib/unifi-stubd/adoption.env",
		StatusPath:         "/var/lib/unifi-stubd/status.json",
	}
}

// ManagementLAN describes controller-facing switch management VLAN behavior.
type ManagementLAN struct {
	// Enabled opts into the structured management LAN model.
	Enabled bool `yaml:"enabled"`
	// VLAN is the management VLAN ID. 0 leaves it unset.
	VLAN int `yaml:"vlan"`
	// NetworkName is an optional controller-facing label for status and docs.
	NetworkName string `yaml:"network_name"`
	// Mode selects metadata-only, preexisting-interface, or planned-host-vlan.
	Mode string `yaml:"mode"`
	// Interface is the existing host VLAN interface used by preexisting-interface.
	Interface string `yaml:"interface"`
	// IP optionally pins the management IP instead of reading the interface IPv4.
	IP string `yaml:"ip"`
	// ControllerReachable controls optional live reachability validation.
	ControllerReachable string `yaml:"controller_reachable"`
	// AdoptionStrategy documents whether adoption starts untagged or tagged.
	AdoptionStrategy string `yaml:"adoption_strategy"`
}

// UplinkNeighbor describes a configured fake upstream neighbor.
type UplinkNeighbor struct {
	// MAC is the neighbor MAC address to expose on the uplink port.
	MAC string `yaml:"mac"`
	// Name is an optional lab label reported as the client hostname.
	Name string `yaml:"name"`
	// Hostname is the optional controller-facing client hostname.
	Hostname string `yaml:"hostname"`
	// IP is the optional client IPv4 address.
	IP string `yaml:"ip"`
	// VLAN is the optional VLAN associated with the neighbor.
	VLAN int `yaml:"vlan"`
	// Static reports the neighbor as a configured lab hint.
	Static bool `yaml:"static"`
	// Type is the controller-facing neighbor type.
	Type string `yaml:"type"`
	// Age is the controller-facing MAC-table age counter.
	Age int `yaml:"age"`
	// Uptime is the number of seconds the neighbor has been visible.
	Uptime int `yaml:"uptime"`
}

// PortNeighbor describes a configured fake neighbor on a specific port.
type PortNeighbor struct {
	// Port is the one-based switch port index.
	Port int `yaml:"port"`
	// MAC is the neighbor MAC address to expose on the port.
	MAC string `yaml:"mac"`
	// Name is an optional lab label reported as the client hostname.
	Name string `yaml:"name"`
	// Hostname is the optional controller-facing client hostname.
	Hostname string `yaml:"hostname"`
	// IP is the optional client IPv4 address.
	IP string `yaml:"ip"`
	// VLAN is the optional VLAN associated with the neighbor.
	VLAN int `yaml:"vlan"`
	// Static reports the neighbor as a configured lab hint.
	Static bool `yaml:"static"`
	// Type is the controller-facing neighbor type.
	Type string `yaml:"type"`
	// Age is the controller-facing MAC-table age counter.
	Age int `yaml:"age"`
	// Uptime is the number of seconds the neighbor has been visible.
	Uptime int `yaml:"uptime"`
}

// PortOverride describes one per-port YAML override.
type PortOverride = payload.PortOverride

// BridgeObserve describes a bridge represented as one virtual switch.
type BridgeObserve struct {
	// Bridge is the host bridge whose learned MAC table is observed.
	Bridge string `yaml:"bridge"`
	// UplinkInterface is the bridge member that points upstream.
	UplinkInterface string `yaml:"uplink_interface"`
	// MemberPortMap pins bridge members to one-based UniFi ports.
	MemberPortMap []BridgeMemberPortMap `yaml:"member_port_map"`
}

// BridgeMemberPortMap pins one bridge member interface to one UniFi port.
type BridgeMemberPortMap struct {
	// Member is the bridge member interface name, such as tap101i0.
	Member string `yaml:"member"`
	// Port is the one-based UniFi port index.
	Port int `yaml:"port"`
}

// PortMapping maps one UniFi port to a physical interface or explicit state.
type PortMapping struct {
	// Port is the one-based UniFi port index.
	Port int `yaml:"port"`
	// Interface is the physical OS interface used as the source for this port.
	Interface string `yaml:"interface"`
	// Disabled reports this port as link-down with speed 0.
	Disabled bool `yaml:"disabled"`
	// Unmapped leaves this port on profile defaults without an observation source.
	Unmapped bool `yaml:"unmapped"`
}

// Load reads path and overlays its YAML values on top of Default.
func Load(path string) (Config, error) {
	cfg := Default()
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, fmt.Errorf("read config %s: %w", path, err)
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return cfg, fmt.Errorf("parse config %s: %w", path, err)
	}
	return cfg, nil
}

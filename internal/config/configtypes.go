package config

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
	// WANHealth controls optional read-only gateway WAN health measurement.
	WANHealth WANHealthConfig `yaml:"wan_health"`
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
	// TrafficRatesEnabled reports interface byte rates to the controller.
	TrafficRatesEnabled bool `yaml:"traffic_rates_enabled"`
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

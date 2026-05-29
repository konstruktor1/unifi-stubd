package main

import (
	appconfig "github.com/konstruktor1/unifi-stubd/internal/config"
	"github.com/konstruktor1/unifi-stubd/internal/device"
	"github.com/konstruktor1/unifi-stubd/internal/observe"
	"github.com/konstruktor1/unifi-stubd/internal/platform"
)

// localStatus is the public status document printed by --status.
type localStatus struct {
	ConfigPath string             `json:"config_path"`
	Identity   statusIdentity     `json:"identity"`
	Config     statusConfig       `json:"config"`
	Adoption   statusAdoption     `json:"adoption"`
	Observe    statusObservation  `json:"observe,omitempty"`
	Platform   statusPlatform     `json:"platform"`
	Runtime    persistedRunStatus `json:"runtime"`
	Warnings   []string           `json:"warnings,omitempty"`
}

// statusIdentity contains the controller-facing fake device identity.
type statusIdentity struct {
	MAC        string `json:"mac"`
	IP         string `json:"ip"`
	Hostname   string `json:"hostname"`
	Serial     string `json:"serial"`
	Model      string `json:"model"`
	ModelName  string `json:"model_name"`
	DeviceType string `json:"device_type"`
	Profile    string `json:"profile"`
	Ports      int    `json:"ports"`
	UplinkPort int    `json:"uplink_port"`
}

// statusConfig contains non-secret runtime configuration selected for status output.
type statusConfig struct {
	OperationMode       string                   `json:"operation_mode"`
	ControllerURL       string                   `json:"controller_url,omitempty"`
	InformURL           string                   `json:"inform_url,omitempty"`
	Interval            string                   `json:"interval"`
	NoDiscovery         bool                     `json:"no_discovery"`
	DiscoveryInterface  string                   `json:"discovery_interface,omitempty"`
	DiscoveryTargets    []string                 `json:"discovery_targets,omitempty"`
	ManagementLAN       *appconfig.ManagementLAN `json:"management_lan,omitempty"`
	SSHListen           string                   `json:"ssh_listen,omitempty"`
	StatePath           string                   `json:"state_path"`
	StatusPath          string                   `json:"status_path"`
	UplinkNeighbor      *statusUplinkNeighbor    `json:"uplink_neighbor,omitempty"`
	PortNeighbors       []statusPortNeighbor     `json:"port_neighbors,omitempty"`
	PortOverrides       []device.PortOverride    `json:"port_overrides,omitempty"`
	BridgeObserve       appconfig.BridgeObserve  `json:"bridge_observe,omitempty"`
	PortMappings        []appconfig.PortMapping  `json:"port_mappings,omitempty"`
	LLDPSource          string                   `json:"lldp_source"`
	TrafficSource       string                   `json:"traffic_source"`
	TrafficRatesEnabled bool                     `json:"traffic_rates_enabled"`
	WANHealth           statusWANHealthConfig    `json:"wan_health"`
	LogSource           string                   `json:"log_source"`
	ProcSource          string                   `json:"proc_source"`
	DBusEnabled         bool                     `json:"dbus_enabled"`
	DBusBus             string                   `json:"dbus_bus"`
	SyslogPath          string                   `json:"syslog_path,omitempty"`
	InstanceGuard       string                   `json:"instance_guard"`
	InstanceGuardPath   string                   `json:"instance_guard_path,omitempty"`
}

// statusUplinkNeighbor summarizes the configured uplink MAC-table neighbor.
type statusUplinkNeighbor struct {
	MAC      string `json:"mac"`
	Hostname string `json:"hostname,omitempty"`
	IP       string `json:"ip,omitempty"`
	VLAN     int    `json:"vlan,omitempty"`
	Static   bool   `json:"static,omitempty"`
	Type     string `json:"type,omitempty"`
	Age      int    `json:"age,omitempty"`
	Uptime   int    `json:"uptime,omitempty"`
}

// statusPortNeighbor summarizes one configured per-port MAC-table neighbor.
type statusPortNeighbor struct {
	Port     int    `json:"port"`
	MAC      string `json:"mac"`
	Hostname string `json:"hostname,omitempty"`
	IP       string `json:"ip,omitempty"`
	VLAN     int    `json:"vlan,omitempty"`
	Static   bool   `json:"static,omitempty"`
	Type     string `json:"type,omitempty"`
	Age      int    `json:"age,omitempty"`
	Uptime   int    `json:"uptime,omitempty"`
}

// statusWANHealthConfig summarizes optional active gateway WAN probes.
type statusWANHealthConfig struct {
	Source          string                  `json:"source"`
	IntervalSeconds int                     `json:"interval_seconds,omitempty"`
	TimeoutMS       int                     `json:"timeout_ms,omitempty"`
	Targets         []statusWANHealthTarget `json:"targets,omitempty"`
	Results         []statusWANHealthResult `json:"results,omitempty"`
}

// statusWANHealthTarget mirrors one configured WAN probe target.
type statusWANHealthTarget struct {
	Port int    `json:"port"`
	Host string `json:"host"`
}

// statusWANHealthResult reports one sanitized WAN probe sample.
type statusWANHealthResult struct {
	Port            int     `json:"port"`
	Host            string  `json:"host"`
	Connected       bool    `json:"connected"`
	LatencyMS       int     `json:"latency_ms"`
	DowntimeSeconds int     `json:"downtime_seconds"`
	UptimePercent   float64 `json:"uptime_percent"`
	LastError       string  `json:"last_error,omitempty"`
}

// statusAdoption exposes adoption state without leaking the auth key.
type statusAdoption struct {
	State      string `json:"state"`
	Adopted    bool   `json:"adopted"`
	AuthKeySet bool   `json:"authkey_set"`
	CFGVersion string `json:"cfgversion,omitempty"`
	UseAESGCM  bool   `json:"use_aes_gcm"`
	Version    string `json:"version,omitempty"`
}

// statusObservation reports passive Linux observation inputs and counters.
type statusObservation struct {
	Interface string `json:"interface,omitempty"`
	Bridge    string `json:"bridge,omitempty"`
	observe.InterfaceStats
	BridgeDevices  int      `json:"bridge_devices,omitempty"`
	LearnedMACs    int      `json:"learned_macs,omitempty"`
	SourceWarnings []string `json:"source_warnings,omitempty"`
}

// statusPlatform reports optional OS integration capabilities.
type statusPlatform struct {
	Capabilities platform.CapabilityReport `json:"capabilities"`
}

// persistedRunStatus contains runtime data loaded from the status file.
type persistedRunStatus struct {
	LastInform lastInformStatus `json:"last_inform,omitempty"`
}

// lastInformStatus summarizes the latest controller inform exchange.
type lastInformStatus struct {
	Time            string                   `json:"time,omitempty"`
	URL             string                   `json:"url,omitempty"`
	StatusCode      int                      `json:"status_code,omitempty"`
	ResponseType    string                   `json:"response_type,omitempty"`
	ControllerState string                   `json:"controller_state,omitempty"`
	CFGVersion      string                   `json:"cfgversion,omitempty"`
	Version         string                   `json:"version,omitempty"`
	AttemptedAESGCM bool                     `json:"attempted_aes_gcm,omitempty"`
	UsedAESGCM      bool                     `json:"used_aes_gcm,omitempty"`
	FallbackToCBC   bool                     `json:"fallback_to_cbc,omitempty"`
	RawBytes        int                      `json:"raw_bytes,omitempty"`
	JSONBytes       int                      `json:"json_bytes,omitempty"`
	Traffic         *lastInformTrafficStatus `json:"traffic,omitempty"`
	IntervalSeconds int                      `json:"interval_seconds,omitempty"`
	IncludeBlocks   []string                 `json:"include_blocks,omitempty"`
	ResetRequested  bool                     `json:"reset_requested,omitempty"`
	ResetApplied    bool                     `json:"reset_applied,omitempty"`
	ResetReason     string                   `json:"reset_reason,omitempty"`
	HasMgmtCFG      bool                     `json:"has_mgmt_cfg,omitempty"`
	HasSystemCFG    bool                     `json:"has_system_cfg,omitempty"`
	SystemCFGBytes  int                      `json:"system_cfg_bytes,omitempty"`
	SystemCFGKeys   []string                 `json:"system_cfg_keys,omitempty"`
	Ignored         bool                     `json:"ignored,omitempty"`
	IgnoredReason   string                   `json:"ignored_reason,omitempty"`
	Error           string                   `json:"error,omitempty"`
}

// lastInformTrafficStatus reports the payload traffic fields from the most
// recent inform in explicit units so operators can compare them with UI graphs.
type lastInformTrafficStatus struct {
	Root lastInformTrafficRates `json:"root,omitempty"`
	Rows []lastInformTrafficRow `json:"rows,omitempty"`
}

// lastInformTrafficRow is one table row that carried traffic counters or rates.
type lastInformTrafficRow struct {
	Table           string                 `json:"table"`
	PortIdx         int                    `json:"port_idx,omitempty"`
	IfName          string                 `json:"ifname,omitempty"`
	SourceInterface string                 `json:"source_interface,omitempty"`
	Role            string                 `json:"role,omitempty"`
	NetworkGroup    string                 `json:"networkgroup,omitempty"`
	Up              *bool                  `json:"up,omitempty"`
	Rates           lastInformTrafficRates `json:"rates"`
}

// lastInformTrafficRates keeps payload counter values and their units explicit.
type lastInformTrafficRates struct {
	Bytes                     *int64 `json:"bytes,omitempty"`
	RXBytes                   *int64 `json:"rx_bytes,omitempty"`
	TXBytes                   *int64 `json:"tx_bytes,omitempty"`
	BytesRateBytesPerSecond   *int64 `json:"bytes_rate_bytes_per_second,omitempty"`
	RXBytesRateBytesPerSecond *int64 `json:"rx_bytes_rate_bytes_per_second,omitempty"`
	TXBytesRateBytesPerSecond *int64 `json:"tx_bytes_rate_bytes_per_second,omitempty"`
	RXRateBitsPerSecond       *int64 `json:"rx_rate_bits_per_second,omitempty"`
	TXRateBitsPerSecond       *int64 `json:"tx_rate_bits_per_second,omitempty"`
}

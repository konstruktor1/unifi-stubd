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
	OperationMode              string                          `json:"operation_mode"`
	ControllerURL              string                          `json:"controller_url,omitempty"`
	InformURL                  string                          `json:"inform_url,omitempty"`
	Interval                   string                          `json:"interval"`
	ControllerIntervalMin      string                          `json:"controller_interval_min"`
	ControllerIntervalMax      string                          `json:"controller_interval_max"`
	ControllerIntervalPolicy   statusControllerIntervalPolicy  `json:"controller_interval_policy"`
	ControllerReportingPolicy  statusControllerReportingPolicy `json:"controller_reporting_policy"`
	InformBackoffMax           string                          `json:"inform_backoff_max"`
	InformBackoffJitterPercent int                             `json:"inform_backoff_jitter_percent"`
	NoDiscovery                bool                            `json:"no_discovery"`
	DiscoveryInterface         string                          `json:"discovery_interface,omitempty"`
	DiscoveryTargets           []string                        `json:"discovery_targets,omitempty"`
	ManagementLAN              *appconfig.ManagementLAN        `json:"management_lan,omitempty"`
	SSHListen                  string                          `json:"ssh_listen,omitempty"`
	StatePath                  string                          `json:"state_path"`
	StatusPath                 string                          `json:"status_path"`
	UplinkNeighbor             *statusUplinkNeighbor           `json:"uplink_neighbor,omitempty"`
	PortNeighbors              []statusPortNeighbor            `json:"port_neighbors,omitempty"`
	PortOverrides              []device.PortOverride           `json:"port_overrides,omitempty"`
	BridgeObserve              appconfig.BridgeObserve         `json:"bridge_observe,omitempty"`
	PortMappings               []appconfig.PortMapping         `json:"port_mappings,omitempty"`
	LLDPSource                 string                          `json:"lldp_source"`
	TrafficSource              string                          `json:"traffic_source"`
	TrafficRatesEnabled        bool                            `json:"traffic_rates_enabled"`
	WANHealth                  statusWANHealthConfig           `json:"wan_health"`
	LogSource                  string                          `json:"log_source"`
	ProcSource                 string                          `json:"proc_source"`
	DBusEnabled                bool                            `json:"dbus_enabled"`
	DBusBus                    string                          `json:"dbus_bus"`
	SyslogPath                 string                          `json:"syslog_path,omitempty"`
	InstanceGuard              string                          `json:"instance_guard"`
	InstanceGuardPath          string                          `json:"instance_guard_path,omitempty"`
}

// statusControllerIntervalPolicy reports whether an exact firmware binding
// permits positive controller interval hints for the selected product profile.
type statusControllerIntervalPolicy struct {
	Supported        bool   `json:"supported"`
	EvidenceID       string `json:"evidence_id,omitempty"`
	FirmwareMinimum  string `json:"firmware_minimum,omitempty"`
	EffectiveMinimum string `json:"effective_minimum,omitempty"`
	Reason           string `json:"reason,omitempty"`
}

// statusControllerReportingPolicy exposes the exact evidence decision for
// controller-selected payload producers and controller time.
type statusControllerReportingPolicy struct {
	EvidenceID      string `json:"evidence_id,omitempty"`
	BehaviorID      string `json:"behavior_id,omitempty"`
	BlockSelection  bool   `json:"block_selection"`
	ControllerClock bool   `json:"controller_clock"`
	Reason          string `json:"reason"`
}

// statusUplinkNeighbor summarizes the configured uplink MAC-table neighbor.
type statusUplinkNeighbor struct {
	MAC        string `json:"mac"`
	Hostname   string `json:"hostname,omitempty"`
	IP         string `json:"ip,omitempty"`
	VLAN       int    `json:"vlan,omitempty"`
	Static     bool   `json:"static,omitempty"`
	Type       string `json:"type,omitempty"`
	Age        int    `json:"age,omitempty"`
	Uptime     int    `json:"uptime,omitempty"`
	RemotePort int    `json:"remote_port,omitempty"`
}

// statusPortNeighbor summarizes one configured per-port MAC-table neighbor.
type statusPortNeighbor struct {
	Port       int    `json:"port"`
	MAC        string `json:"mac"`
	Hostname   string `json:"hostname,omitempty"`
	IP         string `json:"ip,omitempty"`
	VLAN       int    `json:"vlan,omitempty"`
	Static     bool   `json:"static,omitempty"`
	Type       string `json:"type,omitempty"`
	Age        int    `json:"age,omitempty"`
	Uptime     int    `json:"uptime,omitempty"`
	RemotePort int    `json:"remote_port,omitempty"`
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
	State         string `json:"state"`
	WireState     int    `json:"wire_state"`
	WireStateName string `json:"wire_state_name"`
	Adopted       bool   `json:"adopted"`
	AuthKeySet    bool   `json:"authkey_set"`
	CFGVersion    string `json:"cfgversion,omitempty"`
	UseAESGCM     bool   `json:"use_aes_gcm"`
	Version       string `json:"version,omitempty"`
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
	LastInform    lastInformStatus    `json:"last_inform,omitempty"`
	InformSession informSessionStatus `json:"inform_session"`
}

// informSessionStatus is a bounded, non-secret transcript of the current
// daemon run. It preserves enough ordering to distinguish a real decoded
// CBC setparam exchange from subsequent accepted GCM exchanges.
type informSessionStatus struct {
	StartedAt         string                 `json:"started_at"`
	Attempts          int                    `json:"attempts"`
	HTTP200           int                    `json:"http_200"`
	DecodedResponses  int                    `json:"decoded_responses"`
	ParsedResponses   int                    `json:"parsed_responses"`
	CBCRequests       int                    `json:"cbc_requests"`
	GCMRequests       int                    `json:"gcm_requests"`
	CBCFallbacks      int                    `json:"cbc_fallbacks"`
	SetParamResponses int                    `json:"setparam_responses"`
	NoopResponses     int                    `json:"noop_responses"`
	Events            []informExchangeStatus `json:"events,omitempty"`
}

// informExchangeStatus records protocol metadata only. It intentionally omits
// URLs, MAC addresses, keys, payloads, and controller-provided values.
type informExchangeStatus struct {
	Sequence              int    `json:"sequence"`
	RequestFlags          uint16 `json:"request_flags"`
	KeySource             string `json:"key_source"`
	StatusCode            int    `json:"status_code,omitempty"`
	ResponseBodyPresent   bool   `json:"response_body_present"`
	ResponsePacketDecoded bool   `json:"response_packet_decoded"`
	ResponseJSONParsed    bool   `json:"response_json_parsed"`
	ResponseFlags         uint16 `json:"response_flags,omitempty"`
	ResponseType          string `json:"response_type,omitempty"`
	HasMgmtCFG            bool   `json:"has_mgmt_cfg,omitempty"`
	UsedAESGCM            bool   `json:"used_aes_gcm"`
	FallbackToCBC         bool   `json:"fallback_to_cbc"`
	Failed                bool   `json:"failed"`
}

// lastInformStatus summarizes the latest controller inform exchange.
type lastInformStatus struct {
	Time                  string                   `json:"time,omitempty"`
	URL                   string                   `json:"url,omitempty"`
	StatusCode            int                      `json:"status_code,omitempty"`
	ResponseType          string                   `json:"response_type,omitempty"`
	ControllerState       string                   `json:"controller_state,omitempty"`
	WireState             *int                     `json:"wire_state,omitempty"`
	WireStateName         string                   `json:"wire_state_name,omitempty"`
	CFGVersion            string                   `json:"cfgversion,omitempty"`
	Version               string                   `json:"version,omitempty"`
	AttemptedAESGCM       bool                     `json:"attempted_aes_gcm,omitempty"`
	UsedAESGCM            bool                     `json:"used_aes_gcm,omitempty"`
	FallbackToCBC         bool                     `json:"fallback_to_cbc,omitempty"`
	RequestFlags          uint16                   `json:"request_flags,omitempty"`
	KeySource             string                   `json:"key_source,omitempty"`
	ResponseBodyPresent   bool                     `json:"response_body_present,omitempty"`
	ResponsePacketDecoded bool                     `json:"response_packet_decoded,omitempty"`
	ResponseJSONParsed    bool                     `json:"response_json_parsed,omitempty"`
	ResponseFlags         uint16                   `json:"response_flags,omitempty"`
	RawBytes              int                      `json:"raw_bytes,omitempty"`
	JSONBytes             int                      `json:"json_bytes,omitempty"`
	IntervalSeconds       int                      `json:"interval_seconds,omitempty"`
	Schedule              lastInformScheduleStatus `json:"schedule"`
	IncludeBlocks         []string                 `json:"include_blocks,omitempty"`
	ExcludeBlocks         []string                 `json:"exclude_blocks,omitempty"`
	UDAPIBlocks           []string                 `json:"udapi_blocks,omitempty"`
	BlockSelectionLimited bool                     `json:"block_selection_limited,omitempty"`
	Immediate             bool                     `json:"immediate,omitempty"`
	ServerTimeUTC         int64                    `json:"server_time_in_utc,omitempty"`
	ServerTimeSet         bool                     `json:"server_time_set,omitempty"`
	ControllerTimeOffset  int64                    `json:"controller_time_offset_seconds,omitempty"`
	ControllerTimeSet     bool                     `json:"controller_time_offset_set,omitempty"`
	ResetRequested        bool                     `json:"reset_requested,omitempty"`
	ResetApplied          bool                     `json:"reset_applied,omitempty"`
	ResetReason           string                   `json:"reset_reason,omitempty"`
	ResetDisposition      string                   `json:"reset_disposition,omitempty"`
	HasMgmtCFG            bool                     `json:"has_mgmt_cfg,omitempty"`
	HasSystemCFG          bool                     `json:"has_system_cfg,omitempty"`
	SystemCFGBytes        int                      `json:"system_cfg_bytes,omitempty"`
	SystemCFGKeys         []string                 `json:"system_cfg_keys,omitempty"`
	Ignored               bool                     `json:"ignored,omitempty"`
	IgnoredReason         string                   `json:"ignored_reason,omitempty"`
	Error                 string                   `json:"error,omitempty"`
}

// lastInformScheduleStatus exposes only effective timing and retry counters.
type lastInformScheduleStatus struct {
	EffectiveInterval     string `json:"effective_interval"`
	NextDelay             string `json:"next_delay"`
	ConsecutiveFailures   int    `json:"consecutive_failures"`
	Backoff               bool   `json:"backoff"`
	ControllerHintApplied bool   `json:"controller_hint_applied"`
	ControllerHintIgnored bool   `json:"controller_hint_ignored"`
	ControllerHints       bool   `json:"controller_hints_supported"`
	ImmediateApplied      bool   `json:"immediate_applied"`
	ImmediateSuppressed   bool   `json:"immediate_suppressed"`
}

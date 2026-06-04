package payload

type emptyObject struct{}

type featureStatusPayload struct {
	FeatureStatus string `json:"feature_status"`
}

type hashPayload struct {
	Hash string `json:"hash"`
}

type gatewayRulePayload struct {
	RuleCount     int    `json:"rule_count"`
	SHA256        string `json:"sha256"`
	SignatureType string `json:"signature_type"`
	UpdateTime    string `json:"update_time"`
}

type ledStatePayload struct {
	Pattern string `json:"pattern"`
	Tempo   int    `json:"tempo"`
}

type speedtestServerPayload struct {
	CountryCode string  `json:"cc"`
	City        string  `json:"city"`
	Country     string  `json:"country"`
	Latitude    float64 `json:"lat"`
	Longitude   float64 `json:"lon"`
	Provider    string  `json:"provider"`
	ProviderURL string  `json:"provider_url"`
}

type speedtestStatusPayload struct {
	Latency        int                    `json:"latency"`
	RunDate        int                    `json:"rundate"`
	Runtime        int                    `json:"runtime"`
	Server         speedtestServerPayload `json:"server"`
	SourceIf       string                 `json:"source_interface"`
	StatusDownload int                    `json:"status_download"`
	StatusPing     int                    `json:"status_ping"`
	StatusSummary  int                    `json:"status_summary"`
	StatusUpload   int                    `json:"status_upload"`
	XputDownload   float64                `json:"xput_download"`
	XputUpload     float64                `json:"xput_upload"`
}

type switchCapsPayload struct {
	FeatureCaps          int `json:"feature_caps"`
	MaxAggregateSessions int `json:"max_aggregate_sessions"`
	MaxMirrorSessions    int `json:"max_mirror_sessions"`
}

type udapiVersionPayload struct {
	Path          string         `json:"path"`
	Version       int            `json:"version"`
	VersionDetail map[string]int `json:"versionDetail"`
}

type gatewayTelemetry struct {
	AnonID                 string                 `json:"anon_id"`
	Architecture           string                 `json:"architecture"`
	BLECaps                int                    `json:"ble_caps"`
	BoardRevision          int                    `json:"board_rev"`
	BOMRevision            string                 `json:"bomrev"`
	BOMRevisionID          string                 `json:"bomrev_id"`
	Boot                   emptyObject            `json:"boot"`
	BootID                 int                    `json:"bootid"`
	BootROMVersion         string                 `json:"bootrom_version"`
	CFGVersionEffective    string                 `json:"cfgversion_effective"`
	Connections            []emptyObject          `json:"connections"`
	ContentFilteringStatus featureStatusPayload   `json:"content_filtering_status"`
	DNSShield              hashPayload            `json:"dns_shield"`
	DPIStats               []emptyObject          `json:"dpi_stats"`
	Dualboot               bool                   `json:"dualboot"`
	EverCrash              bool                   `json:"ever_crash"`
	Fingerprint            string                 `json:"fingerprint"`
	Fingerprints           []emptyObject          `json:"fingerprints"`
	FW2Caps                int                    `json:"fw2_caps"`
	FWCaps                 int                    `json:"fw_caps"`
	GuestKicks             int                    `json:"guest_kicks"`
	GuestToken             string                 `json:"guest_token"`
	GWCapabilities         gatewayCapabilities    `json:"gw_caps"`
	HardwareUUID           string                 `json:"hardware_uuid"`
	HasDefaultRouteDist    bool                   `json:"has_default_route_distance"`
	HasSpeaker             bool                   `json:"has_speaker"`
	HasSSHDisable          bool                   `json:"has_ssh_disable"`
	HasVTI                 bool                   `json:"has_vti"`
	HWCapabilities         int                    `json:"hw_caps"`
	IDsIPSRule             gatewayRulePayload     `json:"ids_ips_rule"`
	InformMinInterval      int                    `json:"inform_min_interval"`
	IPv4ActiveLeases       []emptyObject          `json:"ipv4_active_leases"`
	Isolated               bool                   `json:"isolated"`
	KernelVersion          string                 `json:"kernel_version"`
	LastErrorConns         []emptyObject          `json:"last_error_conns"`
	LEDState               ledStatePayload        `json:"led_state"`
	LLDPTable              []emptyObject          `json:"lldp_table"`
	Locating               bool                   `json:"locating"`
	ManufacturerID         int                    `json:"manufacturer_id"`
	Netmask                string                 `json:"netmask"`
	OutletEnabled          bool                   `json:"outlet_enabled"`
	OutletOverrides        []emptyObject          `json:"outlet_overrides"`
	OutletTable            []emptyObject          `json:"outlet_table"`
	PingtestStatus         []emptyObject          `json:"pingtest-status"`
	QRID                   string                 `json:"qrid"`
	RebootDuration         int                    `json:"reboot_duration"`
	SelfrunBeacon          bool                   `json:"selfrun_beacon"`
	SpeedtestStatus        speedtestStatusPayload `json:"speedtest-status"`
	SpeedtestStatusSaved   bool                   `json:"speedtest-status-saved"`
	SpeedtestStatusUDAPI   []emptyObject          `json:"speedtest-status-udapi"`
	SSHSessionTable        []emptyObject          `json:"ssh_session_table"`
	StatsInformInterval    int                    `json:"stats_inform_interval"`
	SwitchCaps             switchCapsPayload      `json:"switch_caps"`
	SysErrorCaps           int                    `json:"sys_error_caps"`
	SysID                  int                    `json:"sysid"`
	TeleportVersion        int                    `json:"teleport_version"`
	TimeMS                 int64                  `json:"time_ms"`
	Timestamp              string                 `json:"timestamp"`
	TMReady                bool                   `json:"tm_ready"`
	Triggers               []emptyObject          `json:"triggers"`
	TriggersDNSFilter      []emptyObject          `json:"triggers_dns_filter"`
	TriggersGeo            []emptyObject          `json:"triggers_geo"`
	UDAPICaps              int                    `json:"udapi_caps"`
	UDAPIVersion           udapiVersionPayload    `json:"udapi_version"`
	UpgradeDuration        int                    `json:"upgrade_duration"`
	UptimeText             string                 `json:"uptime_str"`
	USG2Caps               int                    `json:"usg2_caps"`
	USGCaps                int                    `json:"usg_caps"`
	WiFiCaps               int                    `json:"wifi_caps"`
}

type gatewayCapabilities struct {
	GTI                bool `json:"gti,omitempty"`
	HB46PPIPIP         bool `json:"hb46pp_ipip,omitempty"`
	HB46PPMapEHubSpoke bool `json:"hb46pp_map_e_hubspoke,omitempty"`
	JPIXMapE           bool `json:"jpix_map_e,omitempty"`
	MDNSTable          bool `json:"mdns_table,omitempty"`
	NTTMapE            bool `json:"ntt_map_e,omitempty"`
	WANMagic           bool `json:"wan_magic,omitempty"`
}

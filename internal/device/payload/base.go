package payload

type basePayload struct {
	MAC               string             `json:"mac"`
	IP                string             `json:"ip"`
	Hostname          string             `json:"hostname"`
	Model             string             `json:"model"`
	ModelDisplay      string             `json:"model_display"`
	Type              string             `json:"type"`
	Version           string             `json:"version"`
	Serial            string             `json:"serial"`
	NumPort           int                `json:"num_port"`
	State             int                `json:"state"`
	Adopted           bool               `json:"adopted"`
	Default           bool               `json:"default"`
	DiscoveryResponse bool               `json:"discovery_response"`
	RequiredVersion   string             `json:"required_version"`
	CFGVersion        string             `json:"cfgversion"`
	Uptime            int                `json:"uptime"`
	Time              int64              `json:"time"`
	InformURL         string             `json:"inform_url"`
	SysStats          sysStatsPayload    `json:"sys_stats"`
	SystemStats       systemStatsPayload `json:"system-stats"`
	ManagementVLAN    int                `json:"management_vlan,omitempty"`
	InformIP          string             `json:"inform_ip,omitempty"`
}

type sysStatsPayload struct {
	LoadAverage1  float64 `json:"loadavg_1"`
	LoadAverage5  float64 `json:"loadavg_5"`
	LoadAverage15 float64 `json:"loadavg_15"`
	MemoryTotal   int     `json:"mem_total"`
	MemoryUsed    int     `json:"mem_used"`
	MemoryBuffer  int     `json:"mem_buffer"`
	Uptime        int     `json:"uptime"`
}

type systemStatsPayload struct {
	CPU    float64 `json:"cpu"`
	Memory float64 `json:"mem"`
	Uptime int     `json:"uptime"`
}

// identityUptime clamps reported uptime to a positive value because controller
// freshness checks treat zero-like uptime as suspicious.
func identityUptime(uptime int) int {
	if uptime < 1 {
		return 1
	}
	return uptime
}

// informState maps adoption state to the controller-facing numeric state.
func informState(adopted bool) int {
	if adopted {
		return 2
	}
	return 1
}

// sysStats returns deterministic low-load system counters for lab payloads.
func sysStats(uptime int) sysStatsPayload {
	return sysStatsPayload{
		LoadAverage1:  0.01,
		LoadAverage5:  0.01,
		LoadAverage15: 0.01,
		MemoryTotal:   536870912,
		MemoryUsed:    67108864,
		MemoryBuffer:  0,
		Uptime:        uptime,
	}
}

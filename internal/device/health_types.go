package device

// WANHealthResult is a sanitized active WAN measurement ready for payload merge.
type WANHealthResult struct {
	// Port is the one-based UniFi port index.
	Port int
	// Connected reports whether the target probe succeeded.
	Connected bool
	// LatencyMS is the measured latency in milliseconds.
	LatencyMS int
	// DowntimeSeconds is the approximate downtime for a failed probe.
	DowntimeSeconds int
	// UptimePercent is the reported availability percentage.
	UptimePercent float64
}

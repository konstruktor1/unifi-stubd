package adoption

// knownControllerResponseType reports whether the response discriminator has
// an evidence-bound top-level branch. Every other value is a dispatcher error;
// it must not become a successful no-op or trigger a local reset.
func knownControllerResponseType(value string) bool {
	switch value {
	case "noop", "setparam", "setdefault", "reboot", "upgrade", "cmd", "geo_ip_update":
		return true
	default:
		return false
	}
}

// isUnsafeControllerCommand identifies known response types that imply host or
// firmware changes and must remain metadata-only in the stub.
func isUnsafeControllerCommand(responseType string) bool {
	switch responseType {
	case "reboot", "upgrade", "geo_ip_update":
		return true
	default:
		return false
	}
}

// isRejectOnlyControllerCommand binds the 73 Linux-agent command names whose
// recovered vendor effects are not authorized in the Stub. The fixed list is
// metadata for safe audit; it is never an execution allowlist.
func isRejectOnlyControllerCommand(value string) bool {
	switch value {
	case "authorize-guest", "block-sta", "build-ssh-session",
		"clear-all-counters", "clear-all-dpi-counters", "clear-counters",
		"clear-crash-log", "clear-crashlog", "close-ssh-session",
		"dhclient-renew-dhcprange", "dhclient-renew-subnet", "dhcpv6detect-start",
		"dhcpv6detect-stop", "dnsresolve-start", "dnsresolve-stop",
		"dynamic-accept-mac", "dynamic-psk", "execute-idsips-command",
		"execute-wanmagic-command", "force-update-neighbors", "fwupdate",
		"get-ssl-inspector-domain-category-file", "get-ssl-inspector-sandbox-report", "get-support-file",
		"hide-lcm-tracker", "lcm-sync", "locate",
		"pcap-start", "pcap-stop", "ping-test-start",
		"ping-test-stop", "port-block", "port-recovery",
		"port-unblock", "power-cycle", "probe-dhcpv6",
		"probe-slaac", "rate-update", "relayctl",
		"rps-port-recovery", "run", "scan_band",
		"send-crash-log", "send-crashlog", "set-etherlight-device-type",
		"set-etherlight-mode", "set-locate", "set-port-locate",
		"set-vbms", "shadow-mode-adopt", "shadow-mode-failover",
		"show-lcm-tracker", "slaacdetect-start", "slaacdetect-stop",
		"soft-restart", "speed-test", "speed-test-udapi-start",
		"speed-test-udapi-stop", "ssh-sdp-answer", "teleport-connect",
		"teleport-disconnect", "traceroute-start", "traceroute-stop",
		"unauthorize-guest", "unblock-sta", "unset-locate",
		"unset-port-locate", "vpn-remote-disconnect", "wan_diag_dhcp_log",
		"wan_diag_ipv6_log", "wan_diag_port_log", "wan_diag_ppp_log",
		"wol-test":
		return true
	default:
		return false
	}
}

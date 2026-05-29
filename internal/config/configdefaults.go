package config

// Default returns the built-in runtime defaults.
func Default() Config {
	return Config{
		OperationMode: "stub",
		Profile:       "us16p150",
		MAC:           automaticValue,
		IP:            "192.168.1.50",
		Hostname:      automaticValue,
		UplinkSpeed:   automaticValue,
		LLDPSource:    sourceOffValue,
		TrafficSource: sourceOffValue,
		WANHealth: WANHealthConfig{
			Source:          sourceOffValue,
			IntervalSeconds: 10,
			TimeoutMS:       1000,
		},
		LogSource:         sourceOffValue,
		ProcSource:        sourceOffValue,
		DBusBus:           "system",
		SyslogPath:        "/var/log/messages",
		InstanceGuard:     "fail",
		InstanceGuardPath: "/var/lib/unifi-stubd/instance.lock",
		IntervalSeconds:   10,
		SSHUser:           "ubnt",
		SSHPassword:       "ubnt",
		SSHHostKeyPath:    "/var/lib/unifi-stubd/ssh_host_rsa_key",
		StatePath:         "/var/lib/unifi-stubd/adoption.env",
		StatusPath:        "/var/lib/unifi-stubd/status.json",
	}
}

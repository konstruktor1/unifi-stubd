package main

import appconfig "github.com/konstruktor1/unifi-stubd/internal/config"

func serviceRuntimeSettings() []runtimeSetting {
	return []runtimeSetting{
		stringSetting("ssh-listen", "optional built-in adoption SSH listen address, for example 0.0.0.0:22",
			func(flags *runtimeFlags) *string { return &flags.sshListen },
			func(cfg *appconfig.Config) *string { return &cfg.SSHListen },
		),
		stringSetting("ssh-user", "built-in adoption SSH username",
			func(flags *runtimeFlags) *string { return &flags.sshUser },
			func(cfg *appconfig.Config) *string { return &cfg.SSHUser },
		),
		stringSetting("ssh-password", "built-in adoption SSH password",
			func(flags *runtimeFlags) *string { return &flags.sshPassword },
			func(cfg *appconfig.Config) *string { return &cfg.SSHPassword },
		),
		stringSetting("ssh-host-key", "built-in adoption SSH host key path",
			func(flags *runtimeFlags) *string { return &flags.sshHostKey },
			func(cfg *appconfig.Config) *string { return &cfg.SSHHostKeyPath },
		),
		stringSetting("ssh-state", "built-in adoption SSH state file path",
			func(flags *runtimeFlags) *string { return &flags.sshState },
			func(cfg *appconfig.Config) *string { return &cfg.StatePath },
		),
		stringSetting("status-path", "non-sensitive runtime status file path",
			func(flags *runtimeFlags) *string { return &flags.statusPath },
			func(cfg *appconfig.Config) *string { return &cfg.StatusPath },
		),
	}
}

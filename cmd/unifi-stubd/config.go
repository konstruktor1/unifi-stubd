package main

import (
	"errors"
	"log"
	"os"
	"strings"
	"time"

	appconfig "github.com/konstruktor1/unifi-stubd/internal/config"
)

func loadConfig(path string, explicit bool) (appconfig.Config, error) {
	if strings.TrimSpace(path) == "" {
		return appconfig.Default(), nil
	}
	cfg, err := appconfig.Load(path)
	if err == nil {
		log.Printf("loaded config from %s", path)
		return cfg, nil
	}
	if errors.Is(err, os.ErrNotExist) && !explicit {
		return appconfig.Default(), nil
	}
	return appconfig.Config{}, err
}

func applyConfig(cfg appconfig.Config, changed map[string]bool, flags runtimeFlags) {
	if !changed["profile"] {
		*flags.profileName = cfg.Profile
	}
	if !changed["operation-mode"] {
		*flags.operationMode = cfg.OperationMode
	}
	if !changed["mac"] {
		*flags.macText = cfg.MAC
	}
	if !changed["ip"] {
		*flags.ipText = cfg.IP
	}
	if !changed["hostname"] {
		*flags.hostname = cfg.Hostname
	}
	if !changed["model"] {
		*flags.model = cfg.Model
	}
	if !changed["model-display"] {
		*flags.modelDisplay = cfg.ModelDisplay
	}
	if !changed["version"] {
		*flags.version = cfg.Version
	}
	if !changed["ports"] {
		*flags.portCount = cfg.Ports
	}
	if !changed["link-speed"] {
		*flags.linkSpeed = cfg.LinkSpeed
	}
	if !changed["uplink-speed"] {
		*flags.uplinkSpeed = cfg.UplinkSpeed
	}
	if !changed["observe-interface"] {
		*flags.observeInterface = cfg.ObserveInterface
	}
	if !changed["observe-bridge"] {
		*flags.observeBridge = cfg.ObserveBridge
	}
	if !changed["lldp-source"] {
		*flags.lldpSource = cfg.LLDPSource
	}
	if !changed["traffic-source"] {
		*flags.trafficSource = cfg.TrafficSource
	}
	if !changed["controller"] {
		*flags.controller = cfg.ControllerURL
	}
	if !changed["interval"] && cfg.IntervalSeconds > 0 {
		*flags.interval = time.Duration(cfg.IntervalSeconds) * time.Second
	}
	if !changed["no-discovery"] {
		*flags.noDiscovery = cfg.NoDiscovery
	}
	if !changed["ssh-listen"] {
		*flags.sshListen = cfg.SSHListen
	}
	if !changed["ssh-user"] {
		*flags.sshUser = cfg.SSHUser
	}
	if !changed["ssh-password"] {
		*flags.sshPassword = cfg.SSHPassword
	}
	if !changed["ssh-host-key"] {
		*flags.sshHostKey = cfg.SSHHostKeyPath
	}
	if !changed["ssh-state"] {
		*flags.sshState = cfg.StatePath
	}
	if !changed["status-path"] {
		*flags.statusPath = cfg.StatusPath
	}
}

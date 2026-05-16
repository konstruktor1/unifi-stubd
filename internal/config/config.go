package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	ControllerURL   string `json:"controller_url"`
	Profile         string `json:"profile"`
	MAC             string `json:"mac"`
	IP              string `json:"ip"`
	Hostname        string `json:"hostname"`
	Model           string `json:"model"`
	ModelDisplay    string `json:"model_display"`
	Ports           int    `json:"ports"`
	LinkSpeed       int    `json:"link_speed"`
	UplinkSpeed     string `json:"uplink_speed"`
	Version         string `json:"version"`
	IntervalSeconds int    `json:"interval_seconds"`
}

func Default() Config {
	return Config{
		ControllerURL:   "http://unifi:8080/inform",
		Profile:         "us16p150",
		MAC:             "auto",
		IP:              "192.168.1.50",
		Hostname:        "proxmox-vmbr0",
		Model:           "US16P150",
		ModelDisplay:    "UniFi Switch 16 POE-150W",
		Ports:           16,
		LinkSpeed:       0,
		UplinkSpeed:     "auto",
		Version:         "6.6.0",
		IntervalSeconds: 10,
	}
}

func Load(path string) (Config, error) {
	cfg := Default()
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// Package opnsense contains the read-only OPNsense API companion used by the
// unifi-stubd-opnsense generator command.
package opnsense

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	appconfig "github.com/konstruktor1/unifi-stubd/internal/config"
	"gopkg.in/yaml.v3"
)

const defaultTimeoutMS = 2000

// SourceConfig describes the separate OPNsense generator source file.
type SourceConfig struct {
	// BaseURL is the OPNsense WebGUI/API base URL, for example https://127.0.0.1.
	BaseURL string `yaml:"base_url"`
	// APIKeyFile contains the API key when APIKeyEnv is not set.
	APIKeyFile string `yaml:"api_key_file"`
	// APISecretFile contains the API secret when APISecretEnv is not set.
	APISecretFile string `yaml:"api_secret_file"`
	// APIKeyEnv names the environment variable containing the API key.
	APIKeyEnv string `yaml:"api_key_env"`
	// APISecretEnv names the environment variable containing the API secret.
	APISecretEnv string `yaml:"api_secret_env"`
	// CAFile optionally pins the CA bundle used to verify OPNsense TLS.
	CAFile string `yaml:"ca_file"`
	// InsecureSkipVerify permits self-signed lab endpoints when explicitly set.
	InsecureSkipVerify bool `yaml:"insecure_skip_verify"`
	// TimeoutMS is the per-request HTTP timeout.
	TimeoutMS int `yaml:"timeout_ms"`
	// UplinkPort optionally sets the generated unifi-stubd uplink_port.
	UplinkPort int `yaml:"uplink_port"`
	// GatewayStatus reads /api/routes/gateway/status and applies WAN health hints.
	GatewayStatus bool `yaml:"gateway_status"`
	// Interfaces maps represented UniFi ports to OPNsense interfaces.
	Interfaces []InterfaceMapping `yaml:"interfaces"`
	// WANHealth optionally sets the generated unifi-stubd wan_health block.
	WANHealth appconfig.WANHealthConfig `yaml:"wan_health"`
}

// InterfaceMapping maps one represented UniFi profile port to one OPNsense interface.
type InterfaceMapping struct {
	// Port is the one-based represented UniFi port index.
	Port int `yaml:"port"`
	// Interface is the OPNsense/FreeBSD source interface, such as ixl0.
	Interface string `yaml:"interface"`
	// Name optionally overrides the rendered UniFi port label.
	Name string `yaml:"name"`
	// Role is the effective gateway role, such as wan, lan, wan2, lan2, or unassigned.
	Role string `yaml:"role"`
	// NetworkGroup is the UniFi network group label, such as WAN, WAN2, or LAN.
	NetworkGroup string `yaml:"network_group"`
	// PortConfID mirrors a known controller port-profile assignment ID.
	PortConfID string `yaml:"portconf_id"`
	// NetworkConfID mirrors a known controller network assignment ID.
	NetworkConfID string `yaml:"networkconf_id"`
	// NativeNetworkConfID mirrors a known controller native-network assignment ID.
	NativeNetworkConfID string `yaml:"native_networkconf_id"`
	// NetworkName mirrors a controller network display name.
	NetworkName string `yaml:"network_name"`
	// VLAN mirrors the controller display VLAN ID.
	VLAN int `yaml:"vlan"`
	// Speed is an optional represented link speed override in Mbps.
	Speed int `yaml:"speed"`
	// Media is an optional represented media label such as GE or SFP+.
	Media string `yaml:"media"`
}

// LoadSourceConfig reads and validates one OPNsense generator source file.
func LoadSourceConfig(path string) (SourceConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return SourceConfig{}, fmt.Errorf("read OPNsense source config %s: %w", path, err)
	}
	cfg, err := DecodeSourceConfig(data)
	if err != nil {
		return SourceConfig{}, fmt.Errorf("parse OPNsense source config %s: %w", path, err)
	}
	return cfg, nil
}

// DecodeSourceConfig parses one OPNsense generator source document.
func DecodeSourceConfig(data []byte) (SourceConfig, error) {
	cfg := SourceConfig{TimeoutMS: defaultTimeoutMS}
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return cfg, fmt.Errorf("decode OPNsense source YAML: %w", err)
	}
	normalizeSourceConfig(&cfg)
	if err := ValidateSourceConfig(cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// ValidateSourceConfig checks fields that are independent from live credentials.
func ValidateSourceConfig(cfg SourceConfig) error {
	if strings.TrimSpace(cfg.BaseURL) == "" {
		return fmt.Errorf("opnsense base_url is required")
	}
	if cfg.TimeoutMS <= 0 {
		return fmt.Errorf("opnsense timeout_ms must be positive")
	}
	if len(cfg.Interfaces) == 0 {
		return fmt.Errorf("opnsense interfaces must contain at least one mapping")
	}
	seenPorts := map[int]bool{}
	for _, mapping := range cfg.Interfaces {
		if mapping.Port < 1 {
			return fmt.Errorf("opnsense interface mapping has invalid port %d", mapping.Port)
		}
		if seenPorts[mapping.Port] {
			return fmt.Errorf("opnsense interface mapping has duplicate port %d", mapping.Port)
		}
		seenPorts[mapping.Port] = true
		if strings.TrimSpace(mapping.Interface) == "" {
			return fmt.Errorf("opnsense interface mapping on port %d requires interface", mapping.Port)
		}
		if strings.Contains(mapping.Interface, "/") {
			return fmt.Errorf("opnsense interface mapping on port %d has invalid interface %q", mapping.Port, mapping.Interface)
		}
	}
	return nil
}

// Timeout returns the configured request timeout.
func (cfg SourceConfig) Timeout() time.Duration {
	timeout := cfg.TimeoutMS
	if timeout <= 0 {
		timeout = defaultTimeoutMS
	}
	return time.Duration(timeout) * time.Millisecond
}

func normalizeSourceConfig(cfg *SourceConfig) {
	cfg.BaseURL = strings.TrimSpace(cfg.BaseURL)
	cfg.APIKeyFile = strings.TrimSpace(cfg.APIKeyFile)
	cfg.APISecretFile = strings.TrimSpace(cfg.APISecretFile)
	cfg.APIKeyEnv = strings.TrimSpace(cfg.APIKeyEnv)
	cfg.APISecretEnv = strings.TrimSpace(cfg.APISecretEnv)
	cfg.CAFile = strings.TrimSpace(cfg.CAFile)
	for index := range cfg.Interfaces {
		mapping := &cfg.Interfaces[index]
		mapping.Interface = strings.TrimSpace(mapping.Interface)
		mapping.Name = strings.TrimSpace(mapping.Name)
		mapping.Role = strings.ToLower(strings.TrimSpace(mapping.Role))
		mapping.NetworkGroup = strings.TrimSpace(mapping.NetworkGroup)
		mapping.PortConfID = strings.TrimSpace(mapping.PortConfID)
		mapping.NetworkConfID = strings.TrimSpace(mapping.NetworkConfID)
		mapping.NativeNetworkConfID = strings.TrimSpace(mapping.NativeNetworkConfID)
		mapping.NetworkName = strings.TrimSpace(mapping.NetworkName)
		mapping.Media = strings.TrimSpace(mapping.Media)
	}
}

package config

import (
	"bytes"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Load reads one operator-owned YAML document and overlays it on top of
// Default. Controller setparam/system_cfg data is persisted elsewhere and is
// not treated as runtime authority for host networking or gateway mapping.
func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config %s: %w", path, err)
	}
	cfg, err := Decode(data)
	if err != nil {
		return Config{}, fmt.Errorf("parse config %s: %w", path, err)
	}
	return cfg, nil
}

// Decode parses one YAML config document and overlays it on Default.
func Decode(data []byte) (Config, error) {
	cfg := Default()
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return cfg, fmt.Errorf("decode config YAML: %w", err)
	}
	return cfg, nil
}

package opnsense

import (
	"fmt"
	"os"
	"strings"
)

// Credentials contains OPNsense API key material. It must not be logged.
type Credentials struct {
	Key    string
	Secret string
}

// LoadCredentials resolves API credentials from environment variables or files.
func LoadCredentials(cfg SourceConfig) (Credentials, error) {
	key, err := secretValue("api_key", cfg.APIKeyEnv, cfg.APIKeyFile)
	if err != nil {
		return Credentials{}, err
	}
	secret, err := secretValue("api_secret", cfg.APISecretEnv, cfg.APISecretFile)
	if err != nil {
		return Credentials{}, err
	}
	return Credentials{Key: key, Secret: secret}, nil
}

func secretValue(label, envName, fileName string) (string, error) {
	if envName != "" {
		value := strings.TrimSpace(os.Getenv(envName))
		if value != "" {
			return value, nil
		}
	}
	if fileName != "" {
		data, err := os.ReadFile(fileName)
		if err != nil {
			return "", fmt.Errorf("read %s file %s: %w", label, fileName, err)
		}
		value := strings.TrimSpace(string(data))
		if value != "" {
			return value, nil
		}
	}
	if envName == "" && fileName == "" {
		return "", fmt.Errorf("%s requires an env var or file", label)
	}
	return "", fmt.Errorf("%s is empty", label)
}

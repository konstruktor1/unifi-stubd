package adoptionssh

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"
)

// loadOrCreateHostKey loads a persistent shim host key when configured, or
// creates one so SSH adoption clients see a stable server identity.
func loadOrCreateHostKey(hostKeyPath string) (ssh.Signer, error) {
	if hostKeyPath != "" {
		if data, err := os.ReadFile(hostKeyPath); err == nil {
			signer, err := ssh.ParsePrivateKey(data)
			if err != nil {
				return nil, fmt.Errorf("parse SSH host key %s: %w", hostKeyPath, err)
			}
			return signer, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("read SSH host key %s: %w", hostKeyPath, err)
		}
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("generate SSH host key: %w", err)
	}
	signer, err := ssh.NewSignerFromKey(key)
	if err != nil {
		return nil, fmt.Errorf("create SSH signer: %w", err)
	}
	if hostKeyPath == "" {
		return signer, nil
	}

	if err := os.MkdirAll(filepath.Dir(hostKeyPath), 0o700); err != nil {
		return nil, fmt.Errorf("create SSH host key directory: %w", err)
	}
	privateKey := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	if err := os.WriteFile(hostKeyPath, privateKey, 0o600); err != nil {
		return nil, fmt.Errorf("write SSH host key %s: %w", hostKeyPath, err)
	}
	return signer, nil
}

package adoptionssh

import (
	"crypto/subtle"
	"fmt"
	"log"
	"net"

	"golang.org/x/crypto/ssh"
)

// Start starts the built-in adoption SSH server when cfg.Listen is set.
func Start(cfg Config) (*Server, error) {
	if cfg.Listen == "" {
		return nil, nil
	}
	if cfg.User == "" {
		cfg.User = defaultSSHUser
	}
	if cfg.Password == "" {
		cfg.Password = defaultSSHPassword
	}
	if cfg.StatePath == "" {
		cfg.StatePath = defaultStatePath
	}

	signer, err := loadOrCreateHostKey(cfg.HostKeyPath)
	if err != nil {
		return nil, fmt.Errorf("load adoption SSH host key: %w", err)
	}

	sshConfig := &ssh.ServerConfig{
		ServerVersion: "SSH-2.0-dropbear_2019.78",
		PasswordCallback: func(meta ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			userOK := subtle.ConstantTimeCompare([]byte(meta.User()), []byte(cfg.User)) == 1
			passOK := subtle.ConstantTimeCompare(pass, []byte(cfg.Password)) == 1
			if userOK && passOK {
				return nil, nil
			}
			return nil, fmt.Errorf("invalid credentials for %s", meta.User())
		},
	}
	sshConfig.AddHostKey(signer)

	listener, err := net.Listen("tcp", cfg.Listen)
	if err != nil {
		return nil, fmt.Errorf("listen for adoption SSH: %w", err)
	}

	server := &Server{
		listener: listener,
		handler:  &Handler{config: cfg},
	}
	go server.serve(sshConfig)
	log.Printf("adoption ssh listening on %s as %s", listener.Addr(), cfg.User)
	return server, nil
}

// Close shuts down the adoption SSH listener.
func (s *Server) Close() error {
	if s == nil || s.listener == nil {
		return nil
	}
	if err := s.listener.Close(); err != nil {
		return fmt.Errorf("close adoption SSH listener: %w", err)
	}
	return nil
}

// Addr returns the listener address, or nil when the server is not running.
func (s *Server) Addr() net.Addr {
	if s == nil || s.listener == nil {
		return nil
	}
	return s.listener.Addr()
}

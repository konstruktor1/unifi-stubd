package adoptionssh

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/subtle"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/konstruktor1/unifi-stubd/internal/adoption"
	"golang.org/x/crypto/ssh"
)

const (
	defaultSSHUser     = "ubnt"
	defaultSSHPassword = "ubnt"
	defaultStatePath   = "/var/lib/unifi-stubd/adoption.env"
	commandInfo        = "info"
	commandSetInform   = "set-inform"
	okOutput           = "OK\n"
)

// Identity describes the device data exposed through adoption SSH commands.
type Identity struct {
	// MAC is the fake device MAC address.
	MAC string
	// IP is the fake device management IP address.
	IP string
	// Hostname is the fake device hostname.
	Hostname string
	// Model is the UniFi model identifier.
	Model string
	// Version is the firmware version reported by the SSH shim.
	Version string
	// InformURL is the controller inform URL currently known by the device.
	InformURL string
}

// Config controls the built-in adoption SSH server.
type Config struct {
	// Listen is the TCP listen address, such as 0.0.0.0:22.
	Listen string
	// User is the accepted SSH username.
	User string
	// Password is the accepted SSH password.
	Password string
	// HostKeyPath stores the persistent SSH host key.
	HostKeyPath string
	// StatePath stores adoption state learned through SSH commands.
	StatePath string
	// Identity is the fake device identity exposed to the controller.
	Identity Identity
}

// Server is a running adoption SSH listener.
type Server struct {
	listener net.Listener
	handler  *Handler
}

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

func (s *Server) serve(config *ssh.ServerConfig) {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			log.Printf("adoption ssh accept failed: %v", err)
			continue
		}
		go s.handleConn(conn, config)
	}
}

func (s *Server) handleConn(conn net.Conn, config *ssh.ServerConfig) {
	sshConn, chans, reqs, err := ssh.NewServerConn(conn, config)
	if err != nil {
		_ = conn.Close()
		return
	}
	log.Printf("adoption ssh login user=%s remote=%s", sshConn.User(), sshConn.RemoteAddr())
	go ssh.DiscardRequests(reqs)

	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			_ = newChannel.Reject(ssh.UnknownChannelType, "only session channels are supported")
			continue
		}
		channel, requests, err := newChannel.Accept()
		if err != nil {
			continue
		}
		go s.handleSession(channel, requests)
	}
}

func (s *Server) handleSession(channel ssh.Channel, requests <-chan *ssh.Request) {
	defer func() {
		_ = channel.Close()
	}()
	for req := range requests {
		switch req.Type {
		case "pty-req", "env":
			_ = req.Reply(true, nil)
		case "exec":
			var payload struct {
				Command string
			}
			if err := ssh.Unmarshal(req.Payload, &payload); err != nil {
				_ = req.Reply(false, nil)
				return
			}
			_ = req.Reply(true, nil)
			output, status := s.handler.Execute(payload.Command)
			_, _ = io.WriteString(channel, output)
			sendExitStatus(channel, status)
			return
		case "shell":
			_ = req.Reply(true, nil)
			s.handler.Shell(channel)
			sendExitStatus(channel, 0)
			return
		default:
			_ = req.Reply(false, nil)
		}
	}
}

func sendExitStatus(channel ssh.Channel, status int) {
	_, _ = channel.SendRequest("exit-status", false, ssh.Marshal(struct {
		Status uint32
	}{Status: uint32(status)}))
}

// Handler executes the small command subset used during UniFi adoption.
type Handler struct {
	config Config
	mu     sync.Mutex
}

// Execute runs one or more shell-like adoption commands.
func (h *Handler) Execute(command string) (string, int) {
	var out strings.Builder
	status := 0
	for _, part := range splitCommands(command) {
		text, code := h.executeOne(part)
		out.WriteString(text)
		if code != 0 {
			status = code
		}
	}
	return out.String(), status
}

// Shell serves a minimal interactive CLI over rw.
func (h *Handler) Shell(rw io.ReadWriter) {
	_, _ = io.WriteString(rw, "UniFi CLI shim\n")
	scanner := bufio.NewScanner(rw)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if line == "exit" || line == "quit" {
			return
		}
		output, _ := h.Execute(line)
		_, _ = io.WriteString(rw, output)
	}
}

func (h *Handler) executeOne(command string) (string, int) {
	args := CommandFields(strings.TrimSpace(command))
	if len(args) == 0 {
		return "", 0
	}
	if args[0] == "sudo" {
		args = args[1:]
		if len(args) == 0 {
			return okOutput, 0
		}
	}
	if (args[0] == "sh" || args[0] == "/bin/sh") && len(args) >= 3 && args[1] == "-c" {
		return h.Execute(strings.Join(args[2:], " "))
	}

	name := path.Base(args[0])
	log.Printf("adoption ssh command: %s", strings.Join(args, " "))

	switch name {
	case "syswrapper.sh":
		return h.handleSyswrapper(args)
	case "mca-cli-op":
		if len(args) > 1 && args[1] == commandInfo {
			return h.info(), 0
		}
		return h.handleSetInform(args)
	case commandInfo:
		return h.info(), 0
	case commandSetInform:
		return h.handleSetInform(args)
	case "mca-cli":
		return h.handleMCA(args)
	case "ubntbox":
		h.saveState("", "", strings.Join(args, " "))
		return okOutput, 0
	case "reset2defaults", "restore-default":
		h.resetState(strings.Join(args, " "))
		return "Factory reset accepted\n", 0
	case "hostname":
		return h.config.Identity.Hostname + "\n", 0
	case "uname":
		return "Linux unifi-stubd-lab 6.18.22-0-virt #1-Alpine SMP aarch64 GNU/Linux\n", 0
	case "cat":
		return h.handleCat(args)
	case "echo":
		return strings.Join(args[1:], " ") + "\n", 0
	case "true", "exit", "quit":
		return "", 0
	default:
		if url := findInformURL(args); url != "" {
			h.saveState(url, "", strings.Join(args, " "))
			return fmt.Sprintf("Adoption request accepted: %s\n", url), 0
		}
		h.saveState("", "", strings.Join(args, " "))
		return okOutput, 0
	}
}

func (h *Handler) handleSyswrapper(args []string) (string, int) {
	if len(args) < 2 {
		return h.info(), 0
	}
	switch args[1] {
	case "set-adopt":
		url := ""
		if len(args) > 2 {
			url = args[2]
		}
		authKey := ""
		if len(args) > 3 {
			authKey = args[3]
		}
		h.saveState(url, authKey, strings.Join(args, " "))
		if url == "" {
			return "Adoption command accepted\n", 0
		}
		return fmt.Sprintf("Adoption request accepted: %s\n", url), 0
	case commandSetInform:
		return h.handleSetInform(args)
	case commandInfo, "status", "get-info":
		return h.info(), 0
	case "restore-default", "reset2defaults":
		h.resetState(strings.Join(args, " "))
		return "Factory reset accepted\n", 0
	default:
		h.saveState("", "", strings.Join(args, " "))
		return okOutput, 0
	}
}

func (h *Handler) handleMCA(args []string) (string, int) {
	if len(args) < 2 {
		return "UniFi CLI shim\n", 0
	}
	if args[1] == "op" && len(args) > 2 {
		return h.handleSetInform(args[1:])
	}
	switch args[1] {
	case commandSetInform:
		return h.handleSetInform(args)
	case commandInfo, "status":
		return h.info(), 0
	default:
		h.saveState("", "", strings.Join(args, " "))
		return okOutput, 0
	}
}

func (h *Handler) handleSetInform(args []string) (string, int) {
	url := findInformURL(args)
	h.saveState(url, "", strings.Join(args, " "))
	if url == "" {
		return "Inform command accepted\n", 0
	}
	return fmt.Sprintf("Inform URL set: %s\n", url), 0
}

func (h *Handler) handleCat(args []string) (string, int) {
	if len(args) < 2 {
		return "", 0
	}
	switch args[1] {
	case "/etc/version", "/usr/lib/version":
		return h.config.Identity.Version + "\n", 0
	case "/proc/ubnthal/system.info":
		return h.info(), 0
	default:
		return "", 0
	}
}

func (h *Handler) info() string {
	id := h.config.Identity
	if id.Model == "" {
		id.Model = "US8"
	}
	if id.Version == "" {
		id.Version = "6.6.0"
	}
	if id.Hostname == "" {
		id.Hostname = "unifi-stubd"
	}
	informURL := id.InformURL
	if informURL == "" {
		informURL = "http://unifi:8080/inform"
	}
	return fmt.Sprintf(
		"Model:       %s\nVersion:     %s\nMAC Address: %s\nIP Address:  %s\nHostname:    %s\nUptime:      1 seconds\nStatus:      Not Adopted (%s)\n",
		id.Model,
		id.Version,
		id.MAC,
		id.IP,
		id.Hostname,
		informURL,
	)
}

func (h *Handler) saveState(informURL, authKey, command string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	statePath := h.config.StatePath
	if statePath == "" {
		statePath = defaultStatePath
	}
	store, err := adoption.LoadEnv(statePath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Printf("adoption state read failed: %v", err)
	}
	update := adoption.Store{
		State:     adoption.StateAdopting,
		InformURL: informURL,
		AuthKey:   authKey,
	}
	store, _ = adoption.Merge(store, update)
	if err := adoption.SaveEnv(statePath, store); err != nil {
		log.Printf("adoption state write failed: %v", err)
	}
	if command != "" {
		log.Printf("adoption state command accepted at %s: %s", time.Now().UTC().Format(time.RFC3339), command)
	}
}

func (h *Handler) resetState(command string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	statePath := h.config.StatePath
	if statePath == "" {
		statePath = defaultStatePath
	}
	if _, err := adoption.ResetEnv(statePath); err != nil {
		log.Printf("adoption state reset failed: %v", err)
		return
	}
	if command != "" {
		log.Printf("adoption state reset accepted at %s: %s", time.Now().UTC().Format(time.RFC3339), command)
	}
}

func splitCommands(command string) []string {
	raw := strings.NewReplacer("&&", ";", "\n", ";").Replace(command)
	parts := strings.Split(raw, ";")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

// CommandFields splits a shell-like command line for the adoption command shim.
func CommandFields(input string) []string {
	var fields []string
	var current strings.Builder
	var quote rune
	escaped := false

	flush := func() {
		if current.Len() == 0 {
			return
		}
		fields = append(fields, current.String())
		current.Reset()
	}

	for _, r := range input {
		switch {
		case escaped:
			current.WriteRune(r)
			escaped = false
		case r == '\\':
			escaped = true
		case quote != 0:
			if r == quote {
				quote = 0
			} else {
				current.WriteRune(r)
			}
		case r == '\'' || r == '"':
			quote = r
		case r == ' ' || r == '\t' || r == '\r' || r == '\n':
			flush()
		default:
			current.WriteRune(r)
		}
	}
	if escaped {
		current.WriteRune('\\')
	}
	flush()
	return fields
}

func findInformURL(args []string) string {
	for _, arg := range args {
		if (strings.HasPrefix(arg, "http://") || strings.HasPrefix(arg, "https://")) && strings.Contains(arg, "/inform") {
			return arg
		}
	}
	return ""
}

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

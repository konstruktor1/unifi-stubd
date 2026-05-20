//nolint:goconst // Repeated CLI fixture literals keep adoption tests readable.
package adoptionssh_test

import (
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/konstruktor1/unifi-stubd/internal/adoptionssh"
	"golang.org/x/crypto/ssh"
)

func TestInfoOutputMatchesUniFiCLIShape(t *testing.T) {
	server, err := adoptionssh.Start(adoptionssh.Config{
		Listen:      "127.0.0.1:0",
		User:        "ubnt",
		Password:    "ubnt",
		HostKeyPath: t.TempDir() + "/ssh_host_rsa_key",
		StatePath:   t.TempDir() + "/adoption.env",
		Identity: adoptionssh.Identity{
			MAC:       "a2:4b:45:16:50:51",
			IP:        "192.0.2.151",
			Hostname:  "unifi-stubd-lab",
			Model:     "USWProXG48",
			Version:   "7.4.1.16850",
			InformURL: "http://192.0.2.10:8080/inform",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := server.Close(); err != nil {
			t.Fatalf("server close: %v", err)
		}
	}()

	output := runSSHCommand(t, server.Addr().String(), "syswrapper.sh info")
	for _, want := range []string{
		"Model:       USWProXG48",
		"Version:     7.4.1.16850",
		"MAC Address: a2:4b:45:16:50:51",
		"IP Address:  192.0.2.151",
		"Hostname:    unifi-stubd-lab",
		"Status:      Not Adopted (http://192.0.2.10:8080/inform)",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("info output missing %q:\n%s", want, output)
		}
	}
}

func TestRestoreDefaultCommandResetsState(t *testing.T) {
	dir := t.TempDir()
	statePath := filepath.Join(dir, "adoption.env")
	if err := os.WriteFile(statePath, []byte(`STATE=connected
INFORM_URL=http://192.0.2.10:8080/inform
AUTHKEY=0123456789abcdef
CFGVERSION=abc123
USE_AES_GCM=true
VERSION=5.0.17.1
`), 0o600); err != nil {
		t.Fatal(err)
	}
	server, err := adoptionssh.Start(adoptionssh.Config{
		Listen:      "127.0.0.1:0",
		User:        "ubnt",
		Password:    "ubnt",
		HostKeyPath: filepath.Join(dir, "ssh_host_rsa_key"),
		StatePath:   statePath,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := server.Close(); err != nil {
			t.Fatalf("server close: %v", err)
		}
	}()

	output := runSSHCommand(t, server.Addr().String(), "syswrapper.sh restore-default")
	if !strings.Contains(output, "Factory reset accepted") {
		t.Fatalf("reset output = %q", output)
	}
	data, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(data)) != "STATE=factory" {
		t.Fatalf("state after reset = %q", data)
	}
}

func FuzzCommandFields(f *testing.F) {
	f.Add("syswrapper.sh set-adopt http://192.0.2.10:8080/inform 0123456789abcdef")
	f.Add(`sh -c "mca-cli-op set-inform http://192.0.2.10:8080/inform"`)
	f.Add("sudo /usr/bin/syswrapper.sh info && echo ok")
	f.Add("")

	f.Fuzz(func(t *testing.T, command string) {
		if len(command) > 4096 {
			t.Skip()
		}
		_ = adoptionssh.CommandFields(command)
	})
}

func runSSHCommand(t *testing.T, addr, command string) string {
	t.Helper()
	client, err := ssh.Dial("tcp", addr, &ssh.ClientConfig{
		User:            "ubnt",
		Auth:            []ssh.AuthMethod{ssh.Password("ubnt")},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := client.Close(); err != nil && !strings.Contains(err.Error(), net.ErrClosed.Error()) {
			t.Fatalf("client close: %v", err)
		}
	}()

	session, err := client.NewSession()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = session.Close()
	}()

	output, err := session.CombinedOutput(command)
	if err != nil {
		t.Fatal(err)
	}
	return string(output)
}

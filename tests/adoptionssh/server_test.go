package adoptionssh_test

import (
	"net"
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
			IP:        "10.0.0.151",
			Hostname:  "unifi-stubd-lab",
			Model:     "USWProXG48",
			Version:   "7.4.1.16850",
			InformURL: "http://10.10.0.30:8080/inform",
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
		"IP Address:  10.0.0.151",
		"Hostname:    unifi-stubd-lab",
		"Status:      Not Adopted (http://10.10.0.30:8080/inform)",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("info output missing %q:\n%s", want, output)
		}
	}
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

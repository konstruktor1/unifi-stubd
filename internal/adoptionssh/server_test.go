package adoptionssh

import (
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestHandlerSetAdopt(t *testing.T) {
	statePath := filepath.Join(t.TempDir(), "adoption.env")
	handler := &Handler{config: Config{
		StatePath: statePath,
		Identity: Identity{
			MAC:       "02:11:22:33:44:55",
			IP:        "10.0.0.151",
			Hostname:  "unifi-stubd-lab",
			Model:     "US8",
			Version:   "6.6.0",
			InformURL: "http://10.0.0.194:8080/inform",
		},
	}}

	output, status := handler.Execute("/usr/bin/syswrapper.sh set-adopt http://10.0.0.194:8080/inform test-authkey")
	if status != 0 {
		t.Fatalf("status = %d", status)
	}
	if !strings.Contains(output, "Adoption request accepted") {
		t.Fatalf("unexpected output: %q", output)
	}
	state, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"STATE=adopting",
		"INFORM_URL=http://10.0.0.194:8080/inform",
		"AUTHKEY=test-authkey",
	} {
		if !strings.Contains(string(state), want) {
			t.Fatalf("state missing %q in:\n%s", want, state)
		}
	}
}

func TestServerPasswordExec(t *testing.T) {
	server, err := Start(Config{
		Listen:      "127.0.0.1:0",
		User:        "ubnt",
		Password:    "ubnt",
		HostKeyPath: filepath.Join(t.TempDir(), "ssh_host_rsa_key"),
		StatePath:   filepath.Join(t.TempDir(), "adoption.env"),
		Identity: Identity{
			MAC:       "02:11:22:33:44:55",
			IP:        "10.0.0.151",
			Hostname:  "unifi-stubd-lab",
			Model:     "US8",
			Version:   "6.6.0",
			InformURL: "http://10.0.0.194:8080/inform",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()

	addr := server.Addr()
	if addr == nil {
		t.Fatal("server has no address")
	}
	client, err := ssh.Dial("tcp", addr.String(), &ssh.ClientConfig{
		User:            "ubnt",
		Auth:            []ssh.AuthMethod{ssh.Password("ubnt")},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	output, err := session.Output("syswrapper.sh info")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(output), "Model: US8") {
		t.Fatalf("unexpected output: %s", output)
	}

	if _, ok := addr.(*net.TCPAddr); !ok {
		t.Fatalf("addr = %T, want TCP", addr)
	}
}

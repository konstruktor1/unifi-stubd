package adoptionssh_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/konstruktor1/unifi-stubd/internal/adoption"
	"github.com/konstruktor1/unifi-stubd/internal/adoptionssh"
	"golang.org/x/crypto/ssh"
)

// TestAdvancedAdoptionChangesOnlyStubState exercises the controller-facing SSH
// product flow and proves an apparent host command remains a harmless no-op.
func TestAdvancedAdoptionChangesOnlyStubState(t *testing.T) {
	const credentials = "ubnt"

	dir := t.TempDir()
	cfg := adoptionssh.Config{Listen: "127.0.0.1:0", User: credentials, Password: credentials,
		HostKeyPath: filepath.Join(dir, "hostkey"), StatePath: filepath.Join(dir, "adoption.env")}
	server, err := adoptionssh.Start(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = server.Close() })
	client, err := ssh.Dial("tcp", server.Addr().String(), &ssh.ClientConfig{
		User: credentials, Auth: []ssh.AuthMethod{ssh.Password(credentials)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec // Disposable loopback server.
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	run := func(command string) string {
		session, err := client.NewSession()
		if err != nil {
			t.Fatal(err)
		}
		output, err := session.CombinedOutput(command)
		_ = session.Close()
		if err != nil {
			t.Fatal(err)
		}
		return string(output)
	}
	canary := filepath.Join(dir, "controller-must-not-create")
	if output := run("touch " + canary); output != "OK\n" {
		t.Fatalf("unsupported host command response = %q", output)
	}
	if _, err := os.Stat(canary); !os.IsNotExist(err) {
		t.Fatalf("controller command touched host path: %v", err)
	}
	const informURL = "http://192.0.2.10:8080/inform"
	if output := run("mca-cli-op set-inform " + informURL); !strings.Contains(output, informURL) {
		t.Fatalf("set-inform response = %q", output)
	}
	state, err := adoption.LoadEnv(cfg.StatePath)
	if err != nil || state.State != adoption.StateProvisioning || state.InformURL != informURL {
		t.Fatalf("stub adoption state = %+v, error %v", state, err)
	}
}

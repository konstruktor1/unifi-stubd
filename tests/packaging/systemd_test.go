// Packaging tests protect service hardening choices. Packaged defaults keep the
// adoption SSH shim closed while still allowing isolated labs to opt into a
// low-port listener without running the daemon as root.
package packaging_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLinuxSystemdServiceRunsNonRootWithBindCapability verifies the packaged
// unit grants only the required bind capability.
func TestLinuxSystemdServiceRunsNonRootWithBindCapability(t *testing.T) {
	data, err := os.ReadFile("../../packaging/linux/usr/lib/systemd/system/unifi-stubd.service")
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, want := range []string{
		"User=unifi-stubd",
		"Group=unifi-stubd",
		"AmbientCapabilities=CAP_NET_BIND_SERVICE",
		"CapabilityBoundingSet=CAP_NET_BIND_SERVICE",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("systemd unit missing %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "User=root") || strings.Contains(text, "Group=root") {
		t.Fatalf("systemd unit still runs as root:\n%s", text)
	}
}

// TestPackagedConfigsKeepAdoptionSSHClosed verifies inform-based adoption is
// the default package path. Advanced-adoption SSH is opt-in.
func TestPackagedConfigsKeepAdoptionSSHClosed(t *testing.T) {
	for _, path := range []string{
		"../../packaging/linux/etc/unifi-stubd/config.yaml",
		"../../packaging/freebsd/usr/local/etc/unifi-stubd/config.yaml",
	} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		text := string(data)
		if !strings.Contains(text, `ssh_listen: ""`) {
			t.Fatalf("%s does not keep ssh_listen closed by default", filepath.Clean(path))
		}
		if strings.Contains(text, "ssh_listen: 0.0.0.0:22") {
			t.Fatalf("%s enables adoption SSH by default", filepath.Clean(path))
		}
	}
}

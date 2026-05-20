// Packaging tests protect service hardening choices. The systemd unit must keep
// low-port SSH compatibility without running the daemon as root.
package packaging_test

import (
	"os"
	"strings"
	"testing"
)

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

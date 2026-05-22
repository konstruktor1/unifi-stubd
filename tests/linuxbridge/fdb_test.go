// Linux bridge FDB tests keep bridge-observe command-output parsing
// deterministic and fixture-based.
package linuxbridge_test

import (
	"strings"
	"testing"

	"github.com/konstruktor1/unifi-stubd/internal/adapters/linuxbridge"
)

// TestParseFDBCapturesDevicesVLANsAndFlags verifies Linux bridge FDB parsing
// captures topology fields used by observation.
func TestParseFDBCapturesDevicesVLANsAndFlags(t *testing.T) {
	input := `00:11:22:33:44:55 dev tap101i0 vlan 20 master vmbr0 dynamic
02:aa:bb:cc:dd:ee dev eth0 self permanent
02:cc:dd:ee:ff:00 dev veth0 static
not a forwarding row
`
	entries := linuxbridge.ParseFDB(strings.NewReader(input))
	if len(entries) != 3 {
		t.Fatalf("len(entries) = %d, want 3", len(entries))
	}
	if entries[0].MAC != "00:11:22:33:44:55" {
		t.Fatalf("MAC = %q", entries[0].MAC)
	}
	if entries[0].Device != "tap101i0" {
		t.Fatalf("Device = %q", entries[0].Device)
	}
	if entries[0].VLAN != 20 {
		t.Fatalf("VLAN = %d", entries[0].VLAN)
	}
	if !entries[0].Dynamic {
		t.Fatal("Dynamic flag not parsed")
	}
	if !entries[1].Self || !entries[1].Permanent {
		t.Fatalf("self/permanent flags not parsed: %+v", entries[1])
	}
	if !entries[2].Static {
		t.Fatal("Static flag not parsed")
	}
}

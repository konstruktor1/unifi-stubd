// FreeBSD bridge tests lock down ifconfig forwarding-output parsing with small
// fixtures, avoiding host-specific test dependencies.
package freebsdifconfig_test

import (
	"strings"
	"testing"

	"github.com/konstruktor1/unifi-stubd/internal/adapters/freebsdifconfig"
)

func TestParseBridgeAddr(t *testing.T) {
	entries := freebsdifconfig.ParseBridgeAddr(strings.NewReader(`
00:11:22:33:44:55 Vlan20 tap101 1199 flags=0<>
02:aa:bb:cc:dd:ee Vlan1 em0 0 flags=0<LOCAL>
33:33:00:00:00:01 Vlan1 em0 0 flags=0<>
00:11:22:33:44:66 vtnet0 42 flags=0<STATIC>
`))
	if len(entries) != 4 {
		t.Fatalf("len(entries) = %d, want 4: %+v", len(entries), entries)
	}
	first := entries[0]
	if first.MAC != "00:11:22:33:44:55" || first.VLAN != 20 || first.Interface != "tap101" || first.Age != 1199 {
		t.Fatalf("first entry = %+v", first)
	}
	if !entries[1].Local {
		t.Fatalf("local flag not parsed: %+v", entries[1])
	}
	if entries[3].VLAN != 0 || entries[3].Interface != "vtnet0" || !entries[3].Static {
		t.Fatalf("static untagged entry = %+v", entries[3])
	}
}

package discovery_test

import (
	"net"
	"testing"

	"github.com/konstruktor1/unifi-stubd/internal/discovery"
)

func TestAnnouncementMarshalBinary(t *testing.T) {
	mac, err := net.ParseMAC("02:11:22:33:44:55")
	if err != nil {
		t.Fatal(err)
	}
	packet, err := discovery.Announcement{
		MAC:      mac,
		IP:       net.ParseIP("192.168.1.50"),
		Model:    "US8",
		Version:  "6.6.0",
		Hostname: "proxmox-vmbr0",
		Default:  true,
		Uptime:   1,
		Sequence: 1,
	}.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	if len(packet) < 4 {
		t.Fatalf("packet too short: %d", len(packet))
	}
	if packet[0] != 0x02 || packet[1] != 0x06 {
		t.Fatalf("unexpected header: %x", packet[:2])
	}
}

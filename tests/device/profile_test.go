//nolint:goconst // Repeated profile fixture literals keep expected models explicit.
package device_test

import (
	"testing"

	"github.com/konstruktor1/unifi-stubd/internal/device"
)

func TestLookupProfile(t *testing.T) {
	profile, ok := device.LookupProfile("us16p150")
	if !ok {
		t.Fatal("profile not found")
	}
	if profile.Model != "US16P150" {
		t.Fatalf("Model = %q, want US16P150", profile.Model)
	}
	if profile.Ports != 18 {
		t.Fatalf("Ports = %d, want 18", profile.Ports)
	}
}

func TestLookupProfileByModel(t *testing.T) {
	profile, ok := device.LookupProfile("US8P60")
	if !ok {
		t.Fatal("profile not found")
	}
	if profile.Name != "us8p60" {
		t.Fatalf("Name = %q, want us8p60", profile.Name)
	}
}

func TestTenGigProfile(t *testing.T) {
	profile, ok := device.LookupProfile("us16xg")
	if !ok {
		t.Fatal("profile not found")
	}
	if profile.Model != "US16XG" {
		t.Fatalf("Model = %q, want US16XG", profile.Model)
	}
	if profile.Ports != 16 {
		t.Fatalf("Ports = %d, want 16", profile.Ports)
	}
	if profile.PortSpeed != 10000 {
		t.Fatalf("PortSpeed = %d, want 10000", profile.PortSpeed)
	}
	ports := device.SwitchPortsWithOptions(profile.Ports, profile.PortOptions())
	if ports[0].Media != "SFP+" || ports[0].Speed != 10000 {
		t.Fatalf("port 1 = media %q speed %d, want 10G SFP+", ports[0].Media, ports[0].Speed)
	}
	if ports[11].Media != "SFP+" || ports[11].Speed != 10000 {
		t.Fatalf("port 12 = media %q speed %d, want 10G SFP+", ports[11].Media, ports[11].Speed)
	}
	if ports[12].Media != "GE" || ports[12].Speed != 10000 {
		t.Fatalf("port 13 = media %q speed %d, want 10G GE", ports[12].Media, ports[12].Speed)
	}
}

func TestGen1PoEProfilesIncludeSFPUplinkPorts(t *testing.T) {
	tests := []struct {
		profile string
		ports   int
		uplink  int
	}{
		{profile: "us16p150", ports: 18, uplink: 17},
		{profile: "us24p250", ports: 26, uplink: 25},
		{profile: "us48p500", ports: 52, uplink: 49},
	}
	for _, test := range tests {
		t.Run(test.profile, func(t *testing.T) {
			profile, ok := device.LookupProfile(test.profile)
			if !ok {
				t.Fatal("profile not found")
			}
			if profile.Ports != test.ports {
				t.Fatalf("Ports = %d, want %d", profile.Ports, test.ports)
			}
			ports := device.SwitchPortsWithOptions(profile.Ports, profile.PortOptions())
			if len(ports) != test.ports {
				t.Fatalf("len(ports) = %d, want %d", len(ports), test.ports)
			}
			if !ports[test.uplink-1].Uplink {
				t.Fatalf("port %d is not marked as uplink", test.uplink)
			}
		})
	}
}

func TestLargestTenGigProfile(t *testing.T) {
	profile, ok := device.LookupProfile("usw-pro-xg-48")
	if !ok {
		t.Fatal("profile not found")
	}
	if profile.Model != "USWProXG48" {
		t.Fatalf("Model = %q, want USWProXG48", profile.Model)
	}
	if profile.Ports != 52 {
		t.Fatalf("Ports = %d, want 52", profile.Ports)
	}
	if profile.PortSpeed != 10000 {
		t.Fatalf("PortSpeed = %d, want 10000", profile.PortSpeed)
	}
	if profile.UplinkSpeed != 25000 {
		t.Fatalf("UplinkSpeed = %d, want 25000", profile.UplinkSpeed)
	}
	if profile.UplinkMedia != "SFP28" {
		t.Fatalf("UplinkMedia = %q, want SFP28", profile.UplinkMedia)
	}
}

func TestLargestControllerKnownTenGigProfile(t *testing.T) {
	profile, ok := device.LookupProfile("usaggpro")
	if !ok {
		t.Fatal("profile not found")
	}
	if profile.Model != "USAGGPRO" {
		t.Fatalf("Model = %q, want USAGGPRO", profile.Model)
	}
	if profile.Ports != 32 {
		t.Fatalf("Ports = %d, want 32", profile.Ports)
	}
	if profile.PortSpeed != 10000 {
		t.Fatalf("PortSpeed = %d, want 10000", profile.PortSpeed)
	}
	if profile.UplinkSpeed != 25000 {
		t.Fatalf("UplinkSpeed = %d, want 25000", profile.UplinkSpeed)
	}
}

func TestGatewayProfile(t *testing.T) {
	profile, ok := device.LookupProfile("ugw3")
	if !ok {
		t.Fatal("profile not found")
	}
	if profile.Model != "UGW3" {
		t.Fatalf("Model = %q, want UGW3", profile.Model)
	}
	if profile.DeviceType != "ugw" {
		t.Fatalf("DeviceType = %q, want ugw", profile.DeviceType)
	}
	if profile.Ports != 3 {
		t.Fatalf("Ports = %d, want 3", profile.Ports)
	}
	if profile.PortSpeed != 1000 {
		t.Fatalf("PortSpeed = %d, want 1000", profile.PortSpeed)
	}
}

func TestTenGigGatewayProfile(t *testing.T) {
	profile, ok := device.LookupProfile("uxgpro")
	if !ok {
		t.Fatal("profile not found")
	}
	if profile.Model != "UXGPRO" {
		t.Fatalf("Model = %q, want UXGPRO", profile.Model)
	}
	if profile.DeviceType != "uxg" {
		t.Fatalf("DeviceType = %q, want uxg", profile.DeviceType)
	}
	if profile.Ports != 4 {
		t.Fatalf("Ports = %d, want 4", profile.Ports)
	}
	if profile.Version != "5.0.16.30689" {
		t.Fatalf("Version = %q, want 5.0.16.30689", profile.Version)
	}
	if profile.PortSpeed != 1000 {
		t.Fatalf("PortSpeed = %d, want 1000", profile.PortSpeed)
	}
	if profile.UplinkSpeed != 1000 {
		t.Fatalf("UplinkSpeed = %d, want 1000", profile.UplinkSpeed)
	}
	if profile.UplinkMedia != "GE" {
		t.Fatalf("UplinkMedia = %q, want GE", profile.UplinkMedia)
	}
	ports := device.SwitchPortsWithOptions(profile.Ports, profile.PortOptions())
	if ports[0].Role != "wan" || ports[0].NetworkGroup != "WAN" {
		t.Fatalf("port 1 assignment = role %q group %q, want WAN", ports[0].Role, ports[0].NetworkGroup)
	}
	if ports[0].Media != "GE" || ports[0].Speed != 1000 || !ports[0].Uplink {
		t.Fatalf("port 1 = media %q speed %d uplink %v, want 1G GE WAN uplink", ports[0].Media, ports[0].Speed, ports[0].Uplink)
	}
	if ports[1].Role != "lan" || ports[1].NetworkGroup != "LAN" {
		t.Fatalf("port 2 assignment = role %q group %q, want LAN", ports[1].Role, ports[1].NetworkGroup)
	}
	if ports[2].Role != "wan2" || ports[2].NetworkGroup != "WAN2" {
		t.Fatalf("port 3 assignment = role %q group %q, want WAN2", ports[2].Role, ports[2].NetworkGroup)
	}
	if ports[2].Media != "SFP+" || ports[2].Speed != 10000 || ports[2].Uplink {
		t.Fatalf("port 3 = media %q speed %d uplink %v, want 10G SFP+ WAN2", ports[2].Media, ports[2].Speed, ports[2].Uplink)
	}
	options := profile.PortOptions()
	options.UplinkPort = 3
	remappedPorts := device.SwitchPortsWithOptions(profile.Ports, options)
	if !remappedPorts[2].Uplink || remappedPorts[2].Speed != 10000 || remappedPorts[2].Media != "SFP+" {
		t.Fatalf("remapped port 3 = media %q speed %d uplink %v, want 10G SFP+ uplink", remappedPorts[2].Media, remappedPorts[2].Speed, remappedPorts[2].Uplink)
	}
}

func TestGatewayLiteProfile(t *testing.T) {
	profile, ok := device.LookupProfile("uxg-lite")
	if !ok {
		t.Fatal("profile not found")
	}
	if profile.Model != "UXG" {
		t.Fatalf("Model = %q, want UXG", profile.Model)
	}
	if profile.DeviceType != "uxg" {
		t.Fatalf("DeviceType = %q, want uxg", profile.DeviceType)
	}
	if profile.Ports != 2 {
		t.Fatalf("Ports = %d, want 2", profile.Ports)
	}
	ports := device.SwitchPortsWithOptions(profile.Ports, profile.PortOptions())
	if len(ports) != 2 {
		t.Fatalf("len(ports) = %d, want 2", len(ports))
	}
	if ports[0].Name != "LAN" || ports[0].Uplink {
		t.Fatalf("port 1 = %+v, want LAN access", ports[0])
	}
	if ports[1].Name != "WAN" || !ports[1].Uplink {
		t.Fatalf("port 2 = %+v, want WAN uplink", ports[1])
	}
}

func TestAutoMACIsStableAndProfileSensitive(t *testing.T) {
	first := device.AutoMAC("host|us16p150")
	second := device.AutoMAC("host|us16p150")
	third := device.AutoMAC("host|us16xg")

	if first.String() != second.String() {
		t.Fatalf("AutoMAC is not stable: %s != %s", first, second)
	}
	if first.String() == third.String() {
		t.Fatalf("AutoMAC should change when seed changes: %s", first)
	}
	if first[0]&0x01 != 0 {
		t.Fatalf("AutoMAC generated multicast address: %s", first)
	}
	if first[0]&0x02 == 0 {
		t.Fatalf("AutoMAC is not locally administered: %s", first)
	}
}

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
	if profile.Ports != 16 {
		t.Fatalf("Ports = %d, want 16", profile.Ports)
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

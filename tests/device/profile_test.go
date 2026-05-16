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

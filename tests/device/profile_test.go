//nolint:goconst // Repeated profile fixture literals keep expected models explicit.
package device_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/konstruktor1/unifi-stubd/internal/device"
)

// TestLookupProfile verifies built-in profiles can be found by profile name.
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

// TestLookupProfileByModel verifies built-in profiles can be found by UniFi
// model identifier.
func TestLookupProfileByModel(t *testing.T) {
	profile, ok := device.LookupProfile("US8P60")
	if !ok {
		t.Fatal("profile not found")
	}
	if profile.Name != "us8p60" {
		t.Fatalf("Name = %q, want us8p60", profile.Name)
	}
}

// TestExternalProfileRegistryLoadsDerivedProfile verifies external YAML
// inheritance and profile defaults.
func TestExternalProfileRegistryLoadsDerivedProfile(t *testing.T) {
	dir := t.TempDir()
	profilePath := filepath.Join(dir, "derived.yaml")
	if err := os.WriteFile(profilePath, []byte(`schema_version: 1
extends: us8
name: lab-us8
model: LABUS8
model_display: Lab US8
stability: external
payload:
  kind: switch
description: derived lab profile
`), 0o600); err != nil {
		t.Fatal(err)
	}
	registry := device.NewProfileRegistry()
	if err := registry.LoadProfilePath(profilePath); err != nil {
		t.Fatal(err)
	}
	profile, ok := registry.LookupProfile("lab-us8")
	if !ok {
		t.Fatal("derived profile not found")
	}
	if profile.Ports != 8 || profile.PortSpeed != 1000 {
		t.Fatalf("derived profile did not inherit base port data: %+v", profile)
	}
	if profile.SourceType != "external" || profile.Payload.Kind != "switch" {
		t.Fatalf("derived profile metadata = %+v", profile)
	}
}

// TestExternalProfileRegistryDerivedProfileOverridesZeroValues verifies YAML
// inheritance preserves explicit zero and false overrides.
func TestExternalProfileRegistryDerivedProfileOverridesZeroValues(t *testing.T) {
	dir := t.TempDir()
	basePath := filepath.Join(dir, "00-base.yaml")
	if err := os.WriteFile(basePath, []byte(`schema_version: 1
name: lab-base-gateway
model: LABBASEGW
model_display: Lab Base Gateway
device_type: uxg
version: 5.0.16.30689
ports: 2
port_names:
  - WAN
  - LAN
port_roles:
  - wan
  - lan
port_network_groups:
  - WAN
  - LAN
port_speed: 1000
uplink_speed: 1000
recommended: true
payload:
  kind: gateway
  has_dpi: true
description: external base gateway
`), 0o600); err != nil {
		t.Fatal(err)
	}
	derivedPath := filepath.Join(dir, "01-derived.yaml")
	if err := os.WriteFile(derivedPath, []byte(`schema_version: 1
extends: lab-base-gateway
name: lab-derived-gateway
model: LABDERIVEDGW
recommended: false
port_names: []
payload:
  has_dpi: false
`), 0o600); err != nil {
		t.Fatal(err)
	}
	registry := device.NewProfileRegistry()
	if err := registry.LoadProfilePath(dir); err != nil {
		t.Fatal(err)
	}
	profile, ok := registry.LookupProfile("lab-derived-gateway")
	if !ok {
		t.Fatal("derived profile not found")
	}
	if profile.Recommended {
		t.Fatalf("Recommended = true, want false")
	}
	if profile.Payload.HasDPI {
		t.Fatalf("Payload.HasDPI = true, want false")
	}
	if len(profile.PortNames) != 0 {
		t.Fatalf("PortNames = %#v, want cleared", profile.PortNames)
	}
	if profile.Payload.Kind != "gateway" || profile.Ports != 2 {
		t.Fatalf("derived profile did not inherit base fields: %+v", profile)
	}
}

// TestExternalProfileCanOverrideBuiltinModel verifies explicit override markers
// can replace a built-in profile.
func TestExternalProfileCanOverrideBuiltinModel(t *testing.T) {
	dir := t.TempDir()
	profilePath := filepath.Join(dir, "us8-override.yaml")
	if err := os.WriteFile(profilePath, []byte(`schema_version: 1
extends: us8
allow_builtin_override: true
name: us8
model: GUGUS13
model_display: gugus 13
description: overridden lab identity
`), 0o600); err != nil {
		t.Fatal(err)
	}
	registry := device.NewProfileRegistry()
	if err := registry.LoadProfilePath(profilePath); err != nil {
		t.Fatal(err)
	}
	profile, ok := registry.LookupProfile("us8")
	if !ok {
		t.Fatal("overridden profile not found by name")
	}
	if profile.Model != "GUGUS13" || profile.ModelDisplay != "gugus 13" {
		t.Fatalf("overridden profile identity = %+v", profile)
	}
	if _, ok := registry.LookupProfile("GUGUS13"); !ok {
		t.Fatal("overridden profile not found by new model")
	}
}

// TestTenGigProfile verifies the US-16-XG profile's high-speed port layout.
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
	ports := device.BuildPorts(profile, device.PortBuildOptions{})
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

// TestGen1PoEProfilesIncludeSFPUplinkPorts verifies older PoE profiles include
// expected SFP uplink groups.
func TestGen1PoEProfilesIncludeSFPUplinkPorts(t *testing.T) {
	tests := []struct {
		profile           string
		ports             int
		uplink            int
		lastProfileUplink int
	}{
		{profile: "us16p150", ports: 18, uplink: 17, lastProfileUplink: 18},
		{profile: "us24p250", ports: 26, uplink: 25, lastProfileUplink: 26},
		{profile: "us48p500", ports: 52, uplink: 49, lastProfileUplink: 50},
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
			ports := device.BuildPorts(profile, device.PortBuildOptions{})
			if len(ports) != test.ports {
				t.Fatalf("len(ports) = %d, want %d", len(ports), test.ports)
			}
			if !ports[test.uplink-1].Uplink {
				t.Fatalf("port %d is not marked as uplink", test.uplink)
			}
			if !ports[test.uplink-1].ProfileUplink {
				t.Fatalf("port %d is not marked as profile uplink", test.uplink)
			}
			if !ports[test.lastProfileUplink-1].ProfileUplink {
				t.Fatalf("last profile uplink port %d is not marked as profile uplink", test.lastProfileUplink)
			}
		})
	}
}

// TestLargestTenGigProfile verifies the aggregation profile with the largest
// ten-gig switch layout.
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

// TestLargestControllerKnownTenGigProfile verifies the controller-known
// high-port-count ten-gig profile metadata.
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

// TestGatewayProfile verifies the UXG-Pro profile selects gateway payload
// metadata and WAN/LAN roles.
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

// TestTenGigGatewayProfile verifies the ten-gig gateway profile's role and
// speed layout.
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
	ports := device.BuildPorts(profile, device.PortBuildOptions{})
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
	remappedPorts := device.BuildPorts(profile, device.PortBuildOptions{UplinkPort: 3})
	if !remappedPorts[2].Uplink || remappedPorts[2].Speed != 10000 || remappedPorts[2].Media != "SFP+" {
		t.Fatalf("remapped port 3 = media %q speed %d uplink %v, want 10G SFP+ uplink", remappedPorts[2].Media, remappedPorts[2].Speed, remappedPorts[2].Uplink)
	}
}

// TestGatewayLiteProfile verifies the UXG-Lite gateway profile metadata.
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
	ports := device.BuildPorts(profile, device.PortBuildOptions{})
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

// TestCloudGatewayFiberProfile verifies the UCG-Fiber profile metadata and
// port layout.
func TestCloudGatewayFiberProfile(t *testing.T) {
	profile, ok := device.LookupProfile("ucg-fiber")
	if !ok {
		t.Fatal("profile not found")
	}
	if profile.Model != "UCGF" {
		t.Fatalf("Model = %q, want UCGF", profile.Model)
	}
	if profile.DeviceType != "udm" {
		t.Fatalf("DeviceType = %q, want udm", profile.DeviceType)
	}
	if profile.Ports != 7 {
		t.Fatalf("Ports = %d, want 7", profile.Ports)
	}
	if profile.Version != "5.0.16" {
		t.Fatalf("Version = %q, want 5.0.16", profile.Version)
	}
	byModel, ok := device.LookupProfile("UCGF")
	if !ok || byModel.Name != "ucg-fiber" {
		t.Fatalf("LookupProfile(UCGF) = %+v, %v; want ucg-fiber", byModel, ok)
	}
	ports := device.BuildPorts(profile, device.PortBuildOptions{})
	assertPort := func(index int, name string, speed int, media string, role string, group string, uplink bool) {
		t.Helper()
		port := ports[index-1]
		if port.Name != name {
			t.Fatalf("port %d name = %q, want %q", index, port.Name, name)
		}
		if port.Speed != speed {
			t.Fatalf("port %d speed = %d, want %d", index, port.Speed, speed)
		}
		if port.Media != media {
			t.Fatalf("port %d media = %q, want %q", index, port.Media, media)
		}
		if port.Role != role || port.NetworkGroup != group {
			t.Fatalf("port %d assignment = role %q group %q, want %s/%s", index, port.Role, port.NetworkGroup, role, group)
		}
		if port.Uplink != uplink {
			t.Fatalf("port %d uplink = %v, want %v", index, port.Uplink, uplink)
		}
	}
	assertPort(1, "LAN 1", 2500, "GE", "lan", "LAN", false)
	assertPort(4, "LAN 4", 2500, "GE", "lan", "LAN", false)
	assertPort(5, "WAN 2", 10000, "GE", "wan2", "WAN2", false)
	assertPort(6, "WAN", 10000, "SFP+", "wan", "WAN", true)
	assertPort(7, "LAN 5", 10000, "SFP+", "lan", "LAN", false)
}

// TestAutoMACIsStableAndProfileSensitive verifies deterministic fake MACs vary
// by profile seed.
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

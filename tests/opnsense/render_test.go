package opnsense_test

import (
	"strings"
	"testing"

	appconfig "github.com/konstruktor1/unifi-stubd/internal/config"
	"github.com/konstruktor1/unifi-stubd/internal/device"
	"github.com/konstruktor1/unifi-stubd/internal/opnsense"
)

func TestGenerateConfigRendersReviewableStubYAML(t *testing.T) {
	t.Parallel()

	base := appconfig.Default()
	base.Profile = "uxgpro"
	base.UplinkPort = 0
	base.PortOverrides = []device.PortOverride{
		{Port: 3, IP: testManualWANIP},
	}
	source := opnsense.SourceConfig{
		UplinkPort: 3,
		Interfaces: []opnsense.InterfaceMapping{
			{Port: 3, Interface: testInterfaceIXL0, Role: testRoleWAN, NetworkGroup: testNetworkWAN},
		},
	}
	generated := opnsense.GenerateConfig(base, source, map[string]opnsense.InterfaceStatus{
		testInterfaceIXL0: {
			Interface: testInterfaceIXL0,
			MAC:       testWANMAC,
			IP:        testWANIP,
			Netmask:   testWANNetmask,
			SpeedMbps: 10000,
			Media:     "SFP+",
		},
	}, nil)
	if generated.UplinkPort != 3 {
		t.Fatalf("UplinkPort = %d, want 3", generated.UplinkPort)
	}
	if len(generated.PortOverrides) != 1 {
		t.Fatalf("PortOverrides = %+v", generated.PortOverrides)
	}
	override := generated.PortOverrides[0]
	if override.IP != testManualWANIP || override.Interface != testInterfaceIXL0 || override.MAC != testWANMAC {
		t.Fatalf("generated override = %+v", override)
	}
	data, err := opnsense.MarshalConfig(generated)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, forbidden := range []string{"api_key", "api_secret", "secret-value"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("generated YAML contains secret field %q:\n%s", forbidden, text)
		}
	}
	if !strings.Contains(text, "port_overrides:") || !strings.Contains(text, "interface: "+testInterfaceIXL0) {
		t.Fatalf("generated YAML missing override:\n%s", text)
	}
}

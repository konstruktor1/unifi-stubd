package observe_test

import (
	"testing"

	appconfig "github.com/konstruktor1/unifi-stubd/internal/config"
	"github.com/konstruktor1/unifi-stubd/internal/observe"
	"github.com/konstruktor1/unifi-stubd/internal/observe/portmap"
)

const (
	observedInterface = "eth0"
	observedMAC       = "00:11:22:33:44:55"
)

func TestPortMapOverridesFromObservation(t *testing.T) {
	t.Parallel()

	up := true
	overrides := portmap.OverridesFromObservation(
		[]appconfig.PortMapping{
			{Port: 1, Interface: observedInterface},
			{Port: 2, Disabled: true},
			{Port: 3, Unmapped: true},
		},
		observe.PortMapObservation{
			Ports: map[int]observe.PortObservation{
				1: {
					Port:      1,
					Interface: observedInterface,
					MAC:       observedMAC,
					IP:        "192.0.2.10",
					Netmask:   "255.255.255.0",
					Up:        &up,
					SpeedMbps: 2500,
					Media:     "GE",
					Stats: observe.InterfaceStats{
						RXBytes: 42,
						TXBytes: 84,
					},
				},
			},
		},
	)

	if len(overrides) != 2 {
		t.Fatalf("got %d overrides, want 2: %#v", len(overrides), overrides)
	}
	if overrides[0].Interface != observedInterface || overrides[0].Speed != 2500 {
		t.Fatalf("interface override = %#v", overrides[0])
	}
	if overrides[0].RXBytes != 42 || overrides[0].TXBytes != 84 {
		t.Fatalf("counters = rx %d tx %d", overrides[0].RXBytes, overrides[0].TXBytes)
	}
	if overrides[1].Up == nil || *overrides[1].Up {
		t.Fatalf("disabled override did not force link down: %#v", overrides[1])
	}
	if !overrides[1].Disabled {
		t.Fatalf("disabled override did not mark port disabled: %#v", overrides[1])
	}
}

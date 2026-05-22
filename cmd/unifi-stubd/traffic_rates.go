package main

import (
	"strings"
	"time"

	"github.com/konstruktor1/unifi-stubd/internal/device"
	"github.com/konstruktor1/unifi-stubd/internal/observe"
)

// applyTrafficRates converts observed monotonic counters into per-second rates
// for ports that have a real source interface and a persistent tracker.
func applyTrafficRates(ports []device.Port, tracker *observe.TrafficRateTracker, now time.Time) []device.Port {
	if len(ports) == 0 || tracker == nil {
		return ports
	}
	out := make([]device.Port, len(ports))
	copy(out, ports)
	for index := range out {
		port := &out[index]
		port.TrafficRatesEnabled = true
		key := trafficRateKey(*port)
		if key == "" {
			continue
		}
		rates := tracker.Rates(now, key, observe.InterfaceStats{
			RXBytes:   port.RXBytes,
			TXBytes:   port.TXBytes,
			RXPackets: port.RXPackets,
			TXPackets: port.TXPackets,
			RXErrors:  port.RXErrors,
			TXErrors:  port.TXErrors,
		}, port.Up)
		port.RXBytesRate = rates.RXBytesRate
		port.TXBytesRate = rates.TXBytesRate
		port.TrafficRatesSet = true
	}
	return out
}

// trafficRateKey ties rate samples to real source interfaces, avoiding
// synthetic rates from accumulating across unrelated generated ports.
func trafficRateKey(port device.Port) string {
	if iface := strings.TrimSpace(port.Interface); iface != "" {
		return iface
	}
	return ""
}

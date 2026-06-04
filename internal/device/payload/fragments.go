// Package payload normalizes link, counter, and source-interface fragments for
// both switch and gateway renderers. Centralizing them keeps renderer logic
// data-driven rather than model-name driven.
package payload

import (
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/device"
)

// effectivePortSpeed keeps connected synthetic ports from rendering as
// zero-speed links.
func effectivePortSpeed(port device.Port) int {
	speed := port.Speed
	if port.Up && speed <= 0 {
		return 1000
	}
	return speed
}

// effectivePortMedia derives a controller media label from speed unless the
// profile or override supplied one explicitly.
func effectivePortMedia(port device.Port, speed int) string {
	media := strings.TrimSpace(port.Media)
	if media == "" && speed > 0 {
		return mediaForSpeed(speed)
	}
	return media
}

type linkFields struct {
	Speed     int    `json:"speed"`
	MaxSpeed  int    `json:"max_speed"`
	SpeedCaps []int  `json:"speed_caps"`
	Media     string `json:"media"`
}

type counterFields struct {
	RXBytes   int64 `json:"rx_bytes"`
	TXBytes   int64 `json:"tx_bytes"`
	RXPackets int64 `json:"rx_packets"`
	TXPackets int64 `json:"tx_packets"`
	RXErrors  int64 `json:"rx_errors"`
	TXErrors  int64 `json:"tx_errors"`
}

type optionalRateFields struct {
	BytesRate   *int64 `json:"bytes-r,omitempty"`
	RXBytesRate *int64 `json:"rx_bytes-r,omitempty"`
	TXBytesRate *int64 `json:"tx_bytes-r,omitempty"`
}

type gatewayRateFields struct {
	RXRate *int64 `json:"rx_rate,omitempty"`
	TXRate *int64 `json:"tx_rate,omitempty"`
}

// portLinkFields renders the common link speed, capability, and media fields
// shared by switch and gateway tables.
func portLinkFields(speed int, media string) linkFields {
	return linkFields{
		Speed:     speed,
		MaxSpeed:  speed,
		SpeedCaps: speedCaps(speed, media),
		Media:     media,
	}
}

func gatewayPortLinkFieldsFor(speed int, media string) gatewayPortLinkFields {
	return gatewayPortLinkFields{
		Speed:     speed,
		MaxSpeed:  speed,
		SpeedCaps: gatewaySpeedCapsCode(speed, media),
		Media:     media,
	}
}

func gatewaySpeedCapsCode(speed int, media string) int {
	media = strings.ToUpper(strings.TrimSpace(media))
	switch {
	case speed >= 25000 || strings.Contains(media, "SFP28"):
		return 1048864
	case speed >= 10000 || strings.Contains(media, "SFP+"):
		return 1048864
	default:
		return 1048623
	}
}

// portCounterFields renders raw counters plus packet fallbacks shared by switch
// and gateway payload rows.
func portCounterFields(port device.Port) counterFields {
	// Packet counters stay non-zero for synthetic connected ports because some
	// controller views treat all-zero rows as stale.
	return counterFields{
		RXBytes:   port.RXBytes,
		TXBytes:   port.TXBytes,
		RXPackets: firstNonZeroInt64(port.RXPackets, 1),
		TXPackets: firstNonZeroInt64(port.TXPackets, 1),
		RXErrors:  port.RXErrors,
		TXErrors:  port.TXErrors,
	}
}

// portRateFields returns explicit observed rates when available, otherwise a
// synthetic low heartbeat for connected synthetic ports.
func portRateFields(port device.Port) (int64, int64) {
	if port.TrafficRatesSet || port.TrafficRatesEnabled {
		return port.RXBytesRate, port.TXBytesRate
	}
	rxRate := int64(0)
	txRate := int64(0)
	if port.Up && effectivePortSpeed(port) > 0 {
		// Synthetic rates provide a small heartbeat when no real traffic source
		// is enabled. Observed or explicitly configured rates bypass this path.
		rxRate = 64 + int64(port.Index)
		txRate = 48 + int64(port.Index)
	}
	return rxRate, txRate
}

// explicitPortRateFields returns rate fields only when observation or operator
// input made rates explicit.
func explicitPortRateFields(port device.Port) optionalRateFields {
	if !port.TrafficRatesSet && !port.TrafficRatesEnabled {
		return optionalRateFields{}
	}
	return optionalRateFields{
		BytesRate:   int64Ref(port.RXBytesRate + port.TXBytesRate),
		RXBytesRate: int64Ref(port.RXBytesRate),
		TXBytesRate: int64Ref(port.TXBytesRate),
	}
}

func gatewayPortRateFields(port device.Port) gatewayRateFields {
	if !port.TrafficRatesSet && !port.TrafficRatesEnabled {
		return gatewayRateFields{}
	}
	return gatewayRateFields{
		RXRate: int64Ref(bitsPerSecond(port.RXBytesRate)),
		TXRate: int64Ref(bitsPerSecond(port.TXBytesRate)),
	}
}

func bitsPerSecond(value int64) int64 {
	return value * 8
}

func intRef(value int) *int {
	out := value
	return &out
}

func int64Ref(value int64) *int64 {
	out := value
	return &out
}

func float64Ref(value float64) *float64 {
	out := value
	return &out
}

func boolRef(value bool) *bool {
	out := value
	return &out
}

func stringRef(value string) *string {
	out := value
	return &out
}

// mediaForSpeed returns the UniFi media label implied by a link speed.
func mediaForSpeed(speed int) string {
	if speed >= 10000 {
		return mediaSFPPlus
	}
	return "GE"
}

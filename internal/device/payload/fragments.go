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
	RXBytes   *int64 `json:"rx_bytes,omitempty"`
	TXBytes   *int64 `json:"tx_bytes,omitempty"`
	RXPackets *int64 `json:"rx_packets,omitempty"`
	TXPackets *int64 `json:"tx_packets,omitempty"`
	RXErrors  *int64 `json:"rx_errors,omitempty"`
	TXErrors  *int64 `json:"tx_errors,omitempty"`
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

// portCounterFields renders a complete local observation, including real zero
// values. A port without a source-backed sample contributes no counters.
func portCounterFields(port device.Port) counterFields {
	if !port.CountersSet {
		return counterFields{}
	}
	return counterFields{
		RXBytes:   int64Ref(port.RXBytes),
		TXBytes:   int64Ref(port.TXBytes),
		RXPackets: int64Ref(port.RXPackets),
		TXPackets: int64Ref(port.TXPackets),
		RXErrors:  int64Ref(port.RXErrors),
		TXErrors:  int64Ref(port.TXErrors),
	}
}

// portRateFields returns only a complete observed rate sample.
func portRateFields(port device.Port) (int64, int64) {
	if !port.TrafficRatesSet || !port.Up {
		return 0, 0
	}
	return port.RXBytesRate, port.TXBytesRate
}

// explicitPortRateFields returns rate fields only when observation or operator
// input made rates explicit.
func explicitPortRateFields(port device.Port) optionalRateFields {
	if !port.TrafficRatesSet {
		return optionalRateFields{}
	}
	rxRate, txRate := portRateFields(port)
	return optionalRateFields{
		BytesRate:   int64Ref(rxRate + txRate),
		RXBytesRate: int64Ref(rxRate),
		TXBytesRate: int64Ref(txRate),
	}
}

func gatewayPortRateFields(port device.Port) gatewayRateFields {
	if !port.TrafficRatesSet {
		return gatewayRateFields{}
	}
	rxRate, txRate := portRateFields(port)
	return gatewayRateFields{
		RXRate: int64Ref(bitsPerSecond(rxRate)),
		TXRate: int64Ref(bitsPerSecond(txRate)),
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

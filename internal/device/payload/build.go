// Package payload builds UniFi inform payloads from typed device data.
package payload

// Build assembles common inform fields before switch or gateway renderers add
// their controller-specific tables.

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/konstruktor1/unifi-stubd/internal/device"
)

// defaultRequiredVersion is the conservative controller version floor reported
// by sparse payload profiles.
const defaultRequiredVersion = "5.0.0"

// Build returns a JSON inform payload using profile-driven renderer metadata.
func Build(profile device.Profile, id device.Identity, ports []device.Port) ([]byte, error) {
	return BuildAt(profile, id, ports, time.Now())
}

// BuildAt builds an inform payload using an injected local wall clock. The
// controller offset affects reported payload timestamps only; uptime remains
// the separate monotonic value supplied on Identity.
func BuildAt(profile device.Profile, id device.Identity, ports []device.Port, localNow time.Time) ([]byte, error) {
	profile = normalizePayloadProfile(profile, id)
	now := reportedWallClock(localNow, id.ControllerTimeOffsetSeconds)
	uptime := identityUptime(id.UptimeSeconds)
	numPorts := len(ports)
	informURL := id.InformURL
	if informURL == "" {
		informURL = "http://unifi:8080/inform"
	}
	cfgVersion := id.CFGVersion
	if cfgVersion == "" {
		cfgVersion = "?"
	}
	deviceType := deviceTypeOrDefault(id.DeviceType)
	wireState, err := informState(id)
	if err != nil {
		return nil, err
	}

	base := basePayload{
		MAC:               id.MAC,
		IP:                id.IP,
		Hostname:          id.Hostname,
		Model:             id.Model,
		ModelDisplay:      id.ModelDisplay,
		Type:              deviceType,
		Version:           id.Version,
		Serial:            id.Serial,
		NumPort:           numPorts,
		State:             int(wireState),
		Adopted:           id.Adopted,
		Default:           wireState == device.InformStateDefault,
		DiscoveryResponse: true,
		RequiredVersion:   profile.Payload.RequiredVersion,
		CFGVersion:        cfgVersion,
		Uptime:            uptime,
		Time:              now.Unix(),
		InformURL:         informURL,
		ManagementVLAN:    id.ManagementVLAN,
		InformIP:          id.InformIP,
	}
	topology, topologyErr := device.ResolveTopology(ports)
	if topologyErr != nil {
		return nil, fmt.Errorf("resolve topology: %w", topologyErr)
	}
	portViews := BuildPortViews(profile, id, ports, topology)
	var data []byte
	if profile.Payload.Kind == payloadKindGateway {
		data, err = json.MarshalIndent(buildGatewayPayload(base, profile, id, portViews, now, uptime), "", "  ")
	} else {
		data, err = json.MarshalIndent(buildSwitchPayload(base, profile, id, portViews, numPorts, managementSpeed(ports)), "", "  ")
	}
	if err != nil {
		return nil, fmt.Errorf("marshal switch payload: %w", err)
	}
	return data, nil
}

func reportedWallClock(local time.Time, offsetSeconds int64) time.Time {
	seconds := local.Unix()
	switch {
	case offsetSeconds > 0 && seconds > math.MaxInt64-offsetSeconds:
		seconds = math.MaxInt64
	case offsetSeconds < 0 && seconds < math.MinInt64-offsetSeconds:
		seconds = math.MinInt64
	default:
		seconds += offsetSeconds
	}
	return time.Unix(seconds, int64(local.Nanosecond())).In(local.Location())
}

// Package payload builds UniFi inform payloads from typed device data.
package payload

// Build assembles common inform fields before switch or gateway renderers add
// their controller-specific tables.

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/konstruktor1/unifi-stubd/internal/device"
)

// defaultRequiredVersion is the conservative controller version floor reported
// by sparse payload profiles.
const defaultRequiredVersion = "5.0.0"

// Build returns a JSON inform payload using profile-driven renderer metadata.
func Build(profile device.Profile, id device.Identity, ports []device.Port) ([]byte, error) {
	profile = normalizePayloadProfile(profile, id)
	now := time.Now()
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
		State:             informState(id.Adopted),
		Adopted:           id.Adopted,
		Default:           !id.Adopted,
		DiscoveryResponse: true,
		RequiredVersion:   profile.Payload.RequiredVersion,
		CFGVersion:        cfgVersion,
		Uptime:            uptime,
		Time:              now.Unix(),
		InformURL:         informURL,
		SysStats:          sysStats(uptime),
		SystemStats: systemStatsPayload{
			CPU:    1.0,
			Memory: 10.0,
			Uptime: uptime,
		},
		ManagementVLAN: id.ManagementVLAN,
		InformIP:       id.InformIP,
	}
	portViews := BuildPortViews(profile, id, ports)
	var data []byte
	var err error
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

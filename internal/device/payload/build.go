// Package payload builds UniFi inform payloads from typed device data.
package payload

// BuildPayload assembles common inform fields before switch or gateway renderers
// add their controller-specific tables.

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// defaultRequiredVersion is the conservative controller version floor reported
// by sparse payload profiles.
const defaultRequiredVersion = "5.0.0"

// MinimalSwitchPayload returns a JSON inform payload with a switch-shaped port table.
func MinimalSwitchPayload(id Identity, ports []Port) ([]byte, error) {
	return BuildPayload(defaultPayloadProfile(id), id, ports)
}

// BuildPayload returns a JSON inform payload using profile-driven renderer metadata.
func BuildPayload(profile Profile, id Identity, ports []Port) ([]byte, error) {
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
	ifSpeed := 1000
	if speed := managementInterfaceSpeed(ports); speed > 0 {
		ifSpeed = speed
	}
	deviceType := deviceTypeOrDefault(id.DeviceType)

	payload := map[string]any{
		jsonKeyMAC:           id.MAC,
		"ip":                 id.IP,
		"hostname":           id.Hostname,
		"model":              id.Model,
		"model_display":      id.ModelDisplay,
		jsonKeyType:          deviceType,
		"version":            id.Version,
		"serial":             id.Serial,
		jsonKeyNumPort:       numPorts,
		"state":              informState(id.Adopted),
		"adopted":            id.Adopted,
		"default":            !id.Adopted,
		"discovery_response": true,
		"required_version":   profile.RequiredVersion,
		"cfgversion":         cfgVersion,
		jsonKeyUptime:        uptime,
		"time":               now.Unix(),
		"inform_url":         informURL,
		"sys_stats":          sysStats(uptime),
		"system-stats":       map[string]any{"cpu": 1.0, "mem": 10.0, jsonKeyUptime: uptime},
	}
	if id.ManagementVLAN > 0 {
		payload["management_vlan"] = id.ManagementVLAN
	}
	if id.InformIP != "" {
		payload["inform_ip"] = id.InformIP
	}
	portViews := BuildPortViews(profile, id, ports)
	if profile.Kind == payloadKindGateway {
		applyGatewayPayload(payload, profile, id, portViews, now, uptime)
	} else {
		applySwitchPayload(payload, profile, id, portViews, numPorts, ifSpeed)
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal switch payload: %w", err)
	}
	return data, nil
}

// identityUptime clamps reported uptime to a positive value because controller
// freshness checks treat zero-like uptime as suspicious.
func identityUptime(uptime int) int {
	if uptime < 1 {
		return 1
	}
	return uptime
}

// applySwitchPayload fills the tables expected by UniFi switch devices.
func applySwitchPayload(payload map[string]any, profile Profile, id Identity, ports []PortView, numPorts int, ifSpeed int) {
	ifaceName := profile.ManagementInterface
	iface := map[string]any{
		jsonKeyName:       ifaceName,
		jsonKeyMAC:        id.MAC,
		"ip":              id.IP,
		jsonKeyNumPort:    numPorts,
		"up":              true,
		jsonKeySpeed:      ifSpeed,
		jsonKeyFullDuplex: true,
	}
	addManagementVLAN(iface, id.ManagementVLAN)
	payload["if_table"] = []map[string]any{iface}
	payload["ethernet_table"] = []map[string]any{
		{
			jsonKeyName:    ifaceName,
			jsonKeyMAC:     id.MAC,
			jsonKeyNumPort: numPorts,
		},
		{
			// srv0 is a synthetic secondary interface seen by controllers on
			// switch-like payloads. It is derived from the fake MAC and does not
			// represent a host interface.
			jsonKeyName: "srv0",
			jsonKeyMAC:  incrementMAC(id.MAC),
		},
	}
	payload["port_table"] = portTable(ports)
}

// addManagementVLAN writes both legacy and newer management VLAN field names so
// controller versions can recognize the same intent.
func addManagementVLAN(row map[string]any, vlan int) {
	if vlan > 0 {
		row["vlan"] = vlan
		row["management_vlan"] = vlan
	}
}

// informState maps adoption state to the controller-facing numeric state.
func informState(adopted bool) int {
	if adopted {
		return 2
	}
	return 1
}

// isGatewayDeviceType reports whether a device type needs gateway-shaped tables.
func isGatewayDeviceType(deviceType string) bool {
	switch strings.TrimSpace(deviceType) {
	case deviceTypeUGW, deviceTypeUXG, deviceTypeUDM:
		return true
	default:
		return false
	}
}

// defaultPayloadProfile infers switch or gateway payload shape from the UniFi
// device type when no profile renderer metadata is available.
func defaultPayloadProfile(id Identity) Profile {
	profile := Profile{Kind: payloadKindSwitch}
	if isGatewayDeviceType(deviceTypeOrDefault(id.DeviceType)) {
		profile.Kind = payloadKindGateway
	}
	return normalizePayloadProfile(profile, id)
}

// normalizePayloadProfile turns sparse profile metadata into the renderer
// defaults used by both legacy switch payloads and gateway-shaped payloads.
func normalizePayloadProfile(profile Profile, id Identity) Profile {
	profile.Kind = strings.ToLower(strings.TrimSpace(profile.Kind))
	if profile.Kind == "" {
		if isGatewayDeviceType(deviceTypeOrDefault(id.DeviceType)) {
			profile.Kind = payloadKindGateway
		} else {
			profile.Kind = payloadKindSwitch
		}
	}
	if profile.Kind != payloadKindGateway {
		profile.Kind = payloadKindSwitch
	}
	if strings.TrimSpace(profile.RequiredVersion) == "" {
		profile.RequiredVersion = defaultRequiredVersion
	}
	if strings.TrimSpace(profile.ManagementInterface) == "" {
		profile.ManagementInterface = "eth0"
	}
	if strings.TrimSpace(profile.GatewayInterfacePrefix) == "" {
		profile.GatewayInterfacePrefix = "eth"
	}
	return profile
}

// deviceTypeOrDefault keeps older switch payloads usable when no type is configured.
func deviceTypeOrDefault(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return deviceTypeUSW
	}
	return value
}

// managementInterfaceSpeed chooses a stable management speed from generated ports.
func managementInterfaceSpeed(ports []Port) int {
	for _, port := range ports {
		if port.Uplink && port.Speed > 0 {
			return port.Speed
		}
	}
	if len(ports) > 0 && ports[0].Speed > 0 {
		return ports[0].Speed
	}
	return 0
}

// sysStats returns deterministic low-load system counters for lab payloads.
func sysStats(uptime int) map[string]any {
	return map[string]any{
		"loadavg_1":   0.01,
		"loadavg_5":   0.01,
		"loadavg_15":  0.01,
		"mem_total":   536870912,
		"mem_used":    67108864,
		"mem_buffer":  0,
		jsonKeyUptime: uptime,
	}
}

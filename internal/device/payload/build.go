// Package payload builds UniFi inform payloads from typed device data.
package payload

// Build assembles common inform fields before switch or gateway renderers add
// their controller-specific tables.

import (
	"encoding/json"
	"fmt"
	"strings"
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
		data, err = json.MarshalIndent(buildSwitchPayload(base, profile, id, portViews, numPorts, managementInterfaceSpeedOrDefault(ports)), "", "  ")
	}
	if err != nil {
		return nil, fmt.Errorf("marshal switch payload: %w", err)
	}
	return data, nil
}

type basePayload struct {
	MAC               string             `json:"mac"`
	IP                string             `json:"ip"`
	Hostname          string             `json:"hostname"`
	Model             string             `json:"model"`
	ModelDisplay      string             `json:"model_display"`
	Type              string             `json:"type"`
	Version           string             `json:"version"`
	Serial            string             `json:"serial"`
	NumPort           int                `json:"num_port"`
	State             int                `json:"state"`
	Adopted           bool               `json:"adopted"`
	Default           bool               `json:"default"`
	DiscoveryResponse bool               `json:"discovery_response"`
	RequiredVersion   string             `json:"required_version"`
	CFGVersion        string             `json:"cfgversion"`
	Uptime            int                `json:"uptime"`
	Time              int64              `json:"time"`
	InformURL         string             `json:"inform_url"`
	SysStats          sysStatsPayload    `json:"sys_stats"`
	SystemStats       systemStatsPayload `json:"system-stats"`
	ManagementVLAN    int                `json:"management_vlan,omitempty"`
	InformIP          string             `json:"inform_ip,omitempty"`
}

type sysStatsPayload struct {
	LoadAverage1  float64 `json:"loadavg_1"`
	LoadAverage5  float64 `json:"loadavg_5"`
	LoadAverage15 float64 `json:"loadavg_15"`
	MemoryTotal   int     `json:"mem_total"`
	MemoryUsed    int     `json:"mem_used"`
	MemoryBuffer  int     `json:"mem_buffer"`
	Uptime        int     `json:"uptime"`
}

type systemStatsPayload struct {
	CPU    float64 `json:"cpu"`
	Memory float64 `json:"mem"`
	Uptime int     `json:"uptime"`
}

type switchPayload struct {
	basePayload
	IfTable       []switchInterfaceRow `json:"if_table"`
	EthernetTable []switchEthernetRow  `json:"ethernet_table"`
	PortTable     []switchPortRow      `json:"port_table"`
}

// identityUptime clamps reported uptime to a positive value because controller
// freshness checks treat zero-like uptime as suspicious.
func identityUptime(uptime int) int {
	if uptime < 1 {
		return 1
	}
	return uptime
}

// buildSwitchPayload fills the tables expected by UniFi switch devices.
func buildSwitchPayload(base basePayload, profile device.Profile, id device.Identity, ports []PortView, numPorts int, ifSpeed int) switchPayload {
	ifaceName := profile.Payload.ManagementInterface
	iface := switchInterfaceRow{
		Name:           ifaceName,
		MAC:            id.MAC,
		IP:             id.IP,
		NumPort:        numPorts,
		Up:             true,
		Speed:          ifSpeed,
		FullDuplex:     true,
		VLAN:           id.ManagementVLAN,
		ManagementVLAN: id.ManagementVLAN,
	}
	return switchPayload{
		basePayload: base,
		IfTable:     []switchInterfaceRow{iface},
		EthernetTable: []switchEthernetRow{
			{
				Name:    ifaceName,
				MAC:     id.MAC,
				NumPort: intRef(numPorts),
			},
			{
				// srv0 is a synthetic secondary interface seen by controllers on
				// switch-like payloads. It is derived from the fake MAC and does not
				// represent a host interface.
				Name: "srv0",
				MAC:  incrementMAC(id.MAC),
			},
		},
		PortTable: portTable(ports),
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

// normalizePayloadProfile turns sparse profile metadata into the renderer
// defaults used by both legacy switch payloads and gateway-shaped payloads.
func normalizePayloadProfile(profile device.Profile, id device.Identity) device.Profile {
	profile.Payload.Kind = strings.ToLower(strings.TrimSpace(profile.Payload.Kind))
	if profile.Payload.Kind == "" {
		if isGatewayDeviceType(deviceTypeOrDefault(id.DeviceType)) {
			profile.Payload.Kind = payloadKindGateway
		} else {
			profile.Payload.Kind = payloadKindSwitch
		}
	}
	if profile.Payload.Kind != payloadKindGateway {
		profile.Payload.Kind = payloadKindSwitch
	}
	if strings.TrimSpace(profile.Payload.RequiredVersion) == "" {
		profile.Payload.RequiredVersion = defaultRequiredVersion
	}
	if strings.TrimSpace(profile.Payload.ManagementInterface) == "" {
		profile.Payload.ManagementInterface = "eth0"
	}
	if strings.TrimSpace(profile.Payload.GatewayInterfacePrefix) == "" {
		profile.Payload.GatewayInterfacePrefix = "eth"
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
func managementInterfaceSpeed(ports []device.Port) int {
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

func managementInterfaceSpeedOrDefault(ports []device.Port) int {
	if speed := managementInterfaceSpeed(ports); speed > 0 {
		return speed
	}
	return 1000
}

// sysStats returns deterministic low-load system counters for lab payloads.
func sysStats(uptime int) sysStatsPayload {
	return sysStatsPayload{
		LoadAverage1:  0.01,
		LoadAverage5:  0.01,
		LoadAverage15: 0.01,
		MemoryTotal:   536870912,
		MemoryUsed:    67108864,
		MemoryBuffer:  0,
		Uptime:        uptime,
	}
}

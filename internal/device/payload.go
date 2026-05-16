package device

import (
	"encoding/json"
	"net"
	"strconv"
	"time"
)

// Identity contains the device attributes reported in inform payloads.
type Identity struct {
	// MAC is the fake device MAC address in controller-facing text form.
	MAC string
	// IP is the device management IP address reported to UniFi.
	IP string
	// Hostname is the device name reported to UniFi.
	Hostname string
	// Model is the UniFi model identifier.
	Model string
	// ModelDisplay is the human-readable UniFi model name.
	ModelDisplay string
	// Version is the firmware version reported by the stub.
	Version string
	// Serial is the serial number reported by the stub.
	Serial string
	// InformURL is the controller inform URL currently known by the device.
	InformURL string
	// CFGVersion is the controller configuration version applied to the device.
	CFGVersion string
	// Adopted reports whether the stub should present itself as adopted.
	Adopted bool
}

// MacTableEntry represents a learned MAC entry for a switch port.
type MacTableEntry struct {
	// MAC is the learned client or neighbor MAC address.
	MAC string `json:"mac"`
	// Age is the controller-facing age counter for this entry.
	Age int `json:"age"`
	// Uptime is the number of seconds the entry has been visible.
	Uptime int `json:"uptime"`
	// VLAN is the optional VLAN associated with the entry.
	VLAN int `json:"vlan,omitempty"`
	// Type describes the learned device type when known.
	Type string `json:"type,omitempty"`
}

// Port describes one fake switch port in the UniFi payload.
type Port struct {
	// Index is the one-based UniFi port index.
	Index int
	// Name is the display name reported for the port.
	Name string
	// Media is the UniFi media label, such as GE or SFP+.
	Media string
	// Uplink marks the port as the upstream connection.
	Uplink bool
	// Up reports whether link is up.
	Up bool
	// Speed is the negotiated speed in Mbps.
	Speed int
	// RXBytes is the receive byte counter.
	RXBytes int64
	// TXBytes is the transmit byte counter.
	TXBytes int64
	// MACs contains learned MAC entries for this port.
	MACs []MacTableEntry
}

// PortOptions configures generated switch port defaults.
type PortOptions struct {
	// Speed is the default access port speed in Mbps.
	Speed int
	// UplinkSpeed is the uplink port speed in Mbps.
	UplinkSpeed int
	// Media is the default access port media label.
	Media string
	// UplinkMedia is the uplink port media label.
	UplinkMedia string
}

// MinimalSwitchPayload returns a JSON inform payload for a fake UniFi switch.
func MinimalSwitchPayload(id Identity, ports []Port) ([]byte, error) {
	now := time.Now().Unix()
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
	if len(ports) > 0 && ports[0].Speed > 0 {
		ifSpeed = ports[0].Speed
	}

	return json.MarshalIndent(map[string]any{
		"mac":                id.MAC,
		"ip":                 id.IP,
		"hostname":           id.Hostname,
		"model":              id.Model,
		"model_display":      id.ModelDisplay,
		"type":               "usw",
		"version":            id.Version,
		"serial":             id.Serial,
		"num_port":           numPorts,
		"state":              1,
		"default":            !id.Adopted,
		"discovery_response": true,
		"required_version":   "5.0.0",
		"cfgversion":         cfgVersion,
		"uptime":             1,
		"time":               now,
		"inform_url":         informURL,
		"if_table": []map[string]any{
			{
				"name":        "eth0",
				"mac":         id.MAC,
				"ip":          id.IP,
				"num_port":    numPorts,
				"up":          true,
				"speed":       ifSpeed,
				"full_duplex": true,
			},
		},
		"ethernet_table": []map[string]any{
			{
				"name":     "eth0",
				"mac":      id.MAC,
				"num_port": numPorts,
			},
			{
				"name": "srv0",
				"mac":  incrementMAC(id.MAC),
			},
		},
		"port_table":   portTable(ports),
		"sys_stats":    sysStats(),
		"system-stats": map[string]any{"cpu": 1.0, "mem": 10.0, "uptime": 1},
	}, "", "  ")
}

// SwitchPorts returns count generated switch ports with profile-neutral defaults.
func SwitchPorts(count int) []Port {
	return SwitchPortsWithOptions(count, PortOptions{})
}

// SwitchPortsWithOptions returns count generated switch ports using options.
func SwitchPortsWithOptions(count int, options PortOptions) []Port {
	if count < 1 {
		count = 1
	}
	options = normalizePortOptions(options)
	ports := make([]Port, 0, count)
	for i := 1; i <= count; i++ {
		speed := options.Speed
		media := options.Media
		if i == 1 {
			speed = options.UplinkSpeed
			media = options.UplinkMedia
		}
		port := Port{
			Index:   i,
			Name:    portName(i),
			Media:   media,
			Uplink:  i == 1,
			Up:      true,
			Speed:   speed,
			RXBytes: int64(1000 * i),
			TXBytes: int64(900 * i),
		}
		if i == 1 {
			port.MACs = []MacTableEntry{
				{MAC: "02:aa:bb:cc:dd:01", Age: 4, Uptime: 1200, VLAN: 1, Type: "usw"},
			}
		}
		ports = append(ports, port)
	}
	return ports
}

func portName(index int) string {
	if index < 1 {
		index = 1
	}
	return "Port " + strconv.Itoa(index)
}

func incrementMAC(macText string) string {
	mac, err := net.ParseMAC(macText)
	if err != nil || len(mac) == 0 {
		return macText
	}
	out := append(net.HardwareAddr{}, mac...)
	out[len(out)-1]++
	return out.String()
}

func portTable(ports []Port) []map[string]any {
	out := make([]map[string]any, 0, len(ports))
	for _, p := range ports {
		speed := p.Speed
		if speed <= 0 {
			speed = 1000
		}
		media := p.Media
		if media == "" {
			media = mediaForSpeed(speed)
		}
		out = append(out, map[string]any{
			"port_idx":     p.Index,
			"name":         p.Name,
			"media":        media,
			"enable":       true,
			"up":           p.Up,
			"is_uplink":    p.Uplink,
			"speed":        speed,
			"full_duplex":  true,
			"autoneg":      true,
			"port_poe":     false,
			"poe_caps":     0,
			"rx_bytes":     p.RXBytes,
			"tx_bytes":     p.TXBytes,
			"rx_packets":   1,
			"tx_packets":   1,
			"rx_errors":    0,
			"tx_errors":    0,
			"stp_state":    "forwarding",
			"stp_pathcost": 20000,
			"mac_table":    p.MACs,
		})
	}
	return out
}

func normalizePortOptions(options PortOptions) PortOptions {
	if options.Speed <= 0 {
		options.Speed = 1000
	}
	if options.UplinkSpeed <= 0 {
		options.UplinkSpeed = options.Speed
	}
	if options.Media == "" {
		options.Media = mediaForSpeed(options.Speed)
	}
	if options.UplinkMedia == "" {
		options.UplinkMedia = mediaForSpeed(options.UplinkSpeed)
	}
	return options
}

func mediaForSpeed(speed int) string {
	if speed >= 10000 {
		return "SFP+"
	}
	return "GE"
}

func sysStats() map[string]any {
	return map[string]any{
		"loadavg_1":  0.01,
		"loadavg_5":  0.01,
		"loadavg_15": 0.01,
		"mem_total":  536870912,
		"mem_used":   67108864,
		"mem_buffer": 0,
	}
}

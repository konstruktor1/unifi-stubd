package device

import (
	"encoding/json"
	"net"
	"strconv"
	"time"
)

type Identity struct {
	MAC          string
	IP           string
	Hostname     string
	Model        string
	ModelDisplay string
	Version      string
	Serial       string
	InformURL    string
	CFGVersion   string
	Adopted      bool
}

type MacTableEntry struct {
	MAC    string `json:"mac"`
	Age    int    `json:"age"`
	Uptime int    `json:"uptime"`
	VLAN   int    `json:"vlan,omitempty"`
	Type   string `json:"type,omitempty"`
}

type Port struct {
	Index   int
	Name    string
	Media   string
	Uplink  bool
	Up      bool
	Speed   int
	RXBytes int64
	TXBytes int64
	MACs    []MacTableEntry
}

type PortOptions struct {
	Speed       int
	UplinkSpeed int
	Media       string
	UplinkMedia string
}

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

func SwitchPorts(count int) []Port {
	return SwitchPortsWithOptions(count, PortOptions{})
}

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

func ExampleUplinkPort() Port {
	return Port{
		Index:   1,
		Name:    "vmbr0-uplink",
		Media:   "GE",
		Uplink:  true,
		Up:      true,
		Speed:   1000,
		RXBytes: 1000,
		TXBytes: 1000,
		MACs: []MacTableEntry{
			{MAC: "02:aa:bb:cc:dd:01", Age: 4, Uptime: 1200, VLAN: 1, Type: "usw"},
		},
	}
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

package main

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/device"
)

type payloadTopologyDocument struct {
	Uplink    json.RawMessage          `json:"uplink"`
	PortTable []payloadTopologyPortRow `json:"port_table"`
	LLDPTable []payloadTopologyLLDPRow `json:"lldp_table"`
}

type payloadTopologyUplinkRow struct {
	PortIdx          int    `json:"port_idx"`
	UplinkMAC        string `json:"uplink_mac"`
	UplinkDeviceName string `json:"uplink_device_name"`
	UplinkRemotePort int    `json:"uplink_remote_port"`
	UplinkSource     string `json:"uplink_source"`
}

type payloadTopologyPortRow struct {
	PortIdx        int                            `json:"port_idx"`
	LastConnection *payloadTopologyLastConnection `json:"last_connection"`
}

type payloadTopologyLastConnection struct {
	MAC      string `json:"mac"`
	Hostname string `json:"hostname"`
	Type     string `json:"type"`
}

type payloadTopologyLLDPRow struct {
	ChassisID    string `json:"chassis_id"`
	LocalPortIdx int    `json:"local_port_idx"`
	PortID       string `json:"port_id"`
}

func payloadTopology(payload []byte) *lastInformTopologyStatus {
	var document payloadTopologyDocument
	if err := json.Unmarshal(payload, &document); err != nil {
		return nil
	}
	lldpByPortMAC := make(map[string]int, len(document.LLDPTable))
	for _, row := range document.LLDPTable {
		lldpByPortMAC[topologyStatusKey(row.LocalPortIdx, row.ChassisID)] = device.ParseRemotePortID(row.PortID)
	}
	edges := make([]lastInformTopologyEdge, 0)
	if len(document.Uplink) > 0 && string(document.Uplink) != "null" && string(document.Uplink) != `""` {
		var uplink payloadTopologyUplinkRow
		if err := json.Unmarshal(document.Uplink, &uplink); err == nil && strings.TrimSpace(uplink.UplinkMAC) != "" {
			source := "config"
			if strings.EqualFold(strings.TrimSpace(uplink.UplinkSource), "lldp_uplink") {
				source = topologySourceLLDP
			}
			edges = append(edges, lastInformTopologyEdge{
				LocalPort:        uplink.PortIdx,
				Direction:        "upstream",
				Source:           source,
				NeighborMAC:      strings.ToLower(strings.TrimSpace(uplink.UplinkMAC)),
				NeighborHostname: strings.TrimSpace(uplink.UplinkDeviceName),
				RemotePort:       uplink.UplinkRemotePort,
			})
		}
	}
	for _, port := range document.PortTable {
		connection := port.LastConnection
		if connection == nil || !infrastructureConnectionType(connection.Type) {
			continue
		}
		source := "config"
		remotePort, lldp := lldpByPortMAC[topologyStatusKey(port.PortIdx, connection.MAC)]
		if lldp {
			source = topologySourceLLDP
		}
		edges = append(edges, lastInformTopologyEdge{
			LocalPort:        port.PortIdx,
			Direction:        "downstream",
			Source:           source,
			NeighborMAC:      strings.ToLower(strings.TrimSpace(connection.MAC)),
			NeighborHostname: strings.TrimSpace(connection.Hostname),
			NeighborType:     strings.TrimSpace(connection.Type),
			RemotePort:       remotePort,
		})
	}
	if len(edges) == 0 {
		return nil
	}
	return &lastInformTopologyStatus{Edges: edges}
}

func infrastructureConnectionType(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return value != "" && value != "client"
}

func topologyStatusKey(port int, mac string) string {
	return strconv.Itoa(port) + "|" + strings.ToLower(strings.TrimSpace(mac))
}

//nolint:goconst // Repeated documentation identities make topology fixtures readable.
package device_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/konstruktor1/unifi-stubd/internal/device"
)

func TestResolveTopologyUsesUplinkDirectionAndConfigPrecedence(t *testing.T) {
	ports := device.SwitchPorts(4)
	ports = device.ApplyUplinkNeighbor(ports, &device.TopologyNeighbor{
		MAC:        "02:00:5e:00:53:01",
		Hostname:   "configured-upstream",
		DeviceType: "uxg",
		RemotePort: 4,
	})
	ports[0].Neighbors = append(ports[0].Neighbors, device.TopologyNeighbor{
		MAC:          "02:00:5e:00:53:02",
		Hostname:     "lldp-upstream",
		DeviceType:   "lldp",
		RemotePort:   12,
		RemotePortID: "Port 12",
		Source:       device.NeighborSourceLLDP,
	})
	ports = device.ApplyPortNeighbors(ports, []device.PortNeighbor{{
		Port:       3,
		RemotePort: 7,
		Entry: device.MacTableEntry{
			MAC:      "02:00:5e:00:53:03",
			Hostname: "configured-downstream",
			Type:     "usw",
		},
	}})

	resolution, err := device.ResolveTopology(ports)
	if err != nil {
		t.Fatal(err)
	}
	if len(resolution.Edges) != 2 {
		t.Fatalf("edges = %+v, want two", resolution.Edges)
	}
	if resolution.Edges[0].Direction != device.DirectionUpstream || resolution.Edges[0].Neighbor.MAC != "02:00:5e:00:53:01" {
		t.Fatalf("upstream edge = %+v", resolution.Edges[0])
	}
	if resolution.Edges[1].Direction != device.DirectionDownstream || resolution.Edges[1].LocalPort != 3 {
		t.Fatalf("downstream edge = %+v", resolution.Edges[1])
	}
	if len(resolution.Warnings) != 1 || !strings.Contains(resolution.Warnings[0], "overrides LLDP") {
		t.Fatalf("warnings = %#v", resolution.Warnings)
	}
}

func TestParseRemotePortID(t *testing.T) {
	for input, want := range map[string]int{
		"Port 50":   50,
		"port 4":    4,
		"12":        12,
		"Ethernet1": 0,
		"Port 0":    0,
	} {
		if got := device.ParseRemotePortID(input); got != want {
			t.Fatalf("ParseRemotePortID(%q) = %d, want %d", input, got, want)
		}
	}
}

func TestResolveTopologyRejectsConflictingConfiguredNeighbors(t *testing.T) {
	ports := device.SwitchPorts(4)
	ports = device.ApplyPortNeighbors(ports, []device.PortNeighbor{
		{Port: 2, Entry: device.MacTableEntry{MAC: "02:00:5e:00:53:11", Type: "usw"}},
		{Port: 2, Entry: device.MacTableEntry{MAC: "02:00:5e:00:53:12", Type: "usw"}},
	})
	if _, err := device.ResolveTopology(ports); err == nil {
		t.Fatal("ResolveTopology accepted conflicting configured neighbors")
	}
}

func TestResolveTopologyRejectsConfiguredNeighborInBothDirections(t *testing.T) {
	ports := device.SwitchPorts(4)
	ports = device.ApplyUplinkNeighbor(ports, &device.TopologyNeighbor{
		MAC:        "02:00:5e:00:53:21",
		DeviceType: "uxg",
	})
	ports = device.ApplyPortNeighbors(ports, []device.PortNeighbor{{
		Port: 2,
		Entry: device.MacTableEntry{
			MAC:  "02:00:5e:00:53:21",
			Type: "uxg",
		},
	}})
	if _, err := device.ResolveTopology(ports); err == nil {
		t.Fatal("ResolveTopology accepted one configured neighbor in both directions")
	}
}

func TestSwitchPayloadProjectsLLDPUplinkWithoutClientRow(t *testing.T) {
	ports := device.SwitchPorts(8)
	ports[0].Neighbors = []device.TopologyNeighbor{{
		MAC:          "02:00:5e:00:53:31",
		Hostname:     "upstream-switch",
		DeviceType:   "lldp",
		RemotePort:   12,
		RemotePortID: "Port 12",
		Source:       device.NeighborSourceLLDP,
	}}
	payload, err := switchPayload(device.Identity{
		MAC:      "02:00:5e:00:53:30",
		IP:       "192.0.2.30",
		Hostname: "represented-switch",
	}, ports)
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		MAC    string `json:"mac"`
		Uplink struct {
			PortIdx          int    `json:"port_idx"`
			MAC              string `json:"mac"`
			UplinkMAC        string `json:"uplink_mac"`
			UplinkSource     string `json:"uplink_source"`
			UplinkRemotePort int    `json:"uplink_remote_port"`
		} `json:"uplink"`
		LastUplink struct {
			PortIdx   int    `json:"port_idx"`
			UplinkMAC string `json:"uplink_mac"`
		} `json:"last_uplink"`
		LLDPTable []map[string]any `json:"lldp_table"`
		PortTable []map[string]any `json:"port_table"`
	}
	if err := json.Unmarshal(payload, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.MAC != "02:00:5e:00:53:30" || doc.Uplink.MAC != doc.MAC {
		t.Fatalf("real identity MAC changed: device=%q uplink=%q", doc.MAC, doc.Uplink.MAC)
	}
	if doc.Uplink.PortIdx != 1 || doc.Uplink.UplinkMAC != "02:00:5e:00:53:31" || doc.Uplink.UplinkSource != "lldp_uplink" || doc.Uplink.UplinkRemotePort != 12 {
		t.Fatalf("uplink = %+v", doc.Uplink)
	}
	if doc.LastUplink.PortIdx != 1 || doc.LastUplink.UplinkMAC != doc.Uplink.UplinkMAC {
		t.Fatalf("last_uplink = %+v", doc.LastUplink)
	}
	if len(doc.LLDPTable) != 1 || doc.LLDPTable[0]["port_id"] != "Port 12" {
		t.Fatalf("lldp_table = %#v", doc.LLDPTable)
	}
	if macs, ok := doc.PortTable[0]["mac_table"].([]any); ok && len(macs) > 0 {
		t.Fatalf("uplink mac_table = %#v, want no infrastructure row", macs)
	}
	if doc.PortTable[0]["last_connection"] != nil {
		t.Fatalf("uplink last_connection = %#v", doc.PortTable[0]["last_connection"])
	}
}

func TestSwitchPayloadSuppressesConflictingLLDPWhenConfigWins(t *testing.T) {
	ports := device.SwitchPorts(4)
	ports = device.ApplyUplinkNeighbor(ports, &device.TopologyNeighbor{
		MAC:        "02:00:5e:00:53:01",
		DeviceType: "uxg",
	})
	ports[0].Neighbors = append(ports[0].Neighbors, device.TopologyNeighbor{
		MAC:          "02:00:5e:00:53:02",
		DeviceType:   "lldp",
		RemotePortID: "Port 12",
		Source:       device.NeighborSourceLLDP,
	})
	payload, err := switchPayload(device.Identity{
		MAC:      "02:00:5e:00:53:03",
		IP:       "192.0.2.3",
		Hostname: "config-wins",
	}, ports)
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Uplink struct {
			UplinkMAC    string `json:"uplink_mac"`
			UplinkSource string `json:"uplink_source"`
		} `json:"uplink"`
		LLDPTable []map[string]any `json:"lldp_table"`
	}
	if err := json.Unmarshal(payload, &doc); err != nil {
		t.Fatal(err)
	}
	if doc.Uplink.UplinkMAC != "02:00:5e:00:53:01" || doc.Uplink.UplinkSource != "" {
		t.Fatalf("uplink = %+v", doc.Uplink)
	}
	if len(doc.LLDPTable) != 0 {
		t.Fatalf("conflicting lldp_table = %#v, want empty", doc.LLDPTable)
	}
}

func TestFDBClientDoesNotCreateTopologyUplink(t *testing.T) {
	ports := device.SwitchPorts(4)
	ports[0].MACs = []device.MacTableEntry{{
		MAC:  "02:00:5e:00:53:41",
		Type: "client",
	}}
	payload, err := switchPayload(device.Identity{
		MAC:      "02:00:5e:00:53:40",
		IP:       "192.0.2.40",
		Hostname: "fdb-only-switch",
	}, ports)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]json.RawMessage
	if err := json.Unmarshal(payload, &doc); err != nil {
		t.Fatal(err)
	}
	if _, ok := doc["uplink"]; ok {
		t.Fatalf("FDB client created uplink: %s", doc["uplink"])
	}
	if _, ok := doc["last_uplink"]; ok {
		t.Fatalf("FDB client created last_uplink: %s", doc["last_uplink"])
	}
}

func TestResolveTopologyBuildsGenericGatewaySwitchChain(t *testing.T) {
	gatewayProfile, ok := device.LookupProfile("uxgpro")
	if !ok {
		t.Fatal("gateway profile not found")
	}
	gatewayPorts := device.BuildPorts(gatewayProfile, device.PortBuildOptions{UplinkPort: 3})
	gatewayPorts = device.ApplyPortNeighbors(gatewayPorts, []device.PortNeighbor{{
		Port: 4,
		Entry: device.MacTableEntry{
			MAC:  "02:00:5e:00:53:51",
			Type: "usw",
		},
	}})

	middlePorts := device.BuildPorts(device.Profile{Ports: 6}, device.PortBuildOptions{UplinkPort: 5})
	middlePorts = device.ApplyUplinkNeighbor(middlePorts, &device.TopologyNeighbor{
		MAC:        "02:00:5e:00:53:50",
		DeviceType: "uxg",
	})
	middlePorts = device.ApplyPortNeighbors(middlePorts, []device.PortNeighbor{{
		Port: 2,
		Entry: device.MacTableEntry{
			MAC:  "02:00:5e:00:53:52",
			Type: "usw",
		},
	}})

	leafPorts := device.BuildPorts(device.Profile{Ports: 8}, device.PortBuildOptions{UplinkPort: 7})
	leafPorts = device.ApplyUplinkNeighbor(leafPorts, &device.TopologyNeighbor{
		MAC:        "02:00:5e:00:53:51",
		DeviceType: "usw",
	})

	for name, test := range map[string]struct {
		ports          []device.Port
		wantDirections []device.TopologyDirection
	}{
		"gateway": {gatewayPorts, []device.TopologyDirection{device.DirectionDownstream}},
		"middle":  {middlePorts, []device.TopologyDirection{device.DirectionDownstream, device.DirectionUpstream}},
		"leaf":    {leafPorts, []device.TopologyDirection{device.DirectionUpstream}},
	} {
		t.Run(name, func(t *testing.T) {
			resolution, err := device.ResolveTopology(test.ports)
			if err != nil {
				t.Fatal(err)
			}
			if len(resolution.Edges) != len(test.wantDirections) {
				t.Fatalf("edges = %+v", resolution.Edges)
			}
			for _, want := range test.wantDirections {
				found := false
				for _, edge := range resolution.Edges {
					found = found || edge.Direction == want
				}
				if !found {
					t.Fatalf("edges = %+v, missing %s", resolution.Edges, want)
				}
			}
		})
	}
}

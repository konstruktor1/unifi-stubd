package observe

import (
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/device"
)

// Apply merges a passive snapshot into generated switch ports.
func Apply(ports []device.Port, snapshot Snapshot) []device.Port {
	if len(ports) == 0 {
		return ports
	}
	out := make([]device.Port, len(ports))
	copy(out, ports)
	index := snapshot.UplinkPortIndex
	if index < 1 || index > len(out) {
		index = uplinkPortIndex(out)
	}
	port := &out[index-1]
	if snapshot.Stats.SpeedMbps > 0 {
		port.Speed = snapshot.Stats.SpeedMbps
	}
	if hasCounters(snapshot.Stats) {
		applyInterfaceStatsToPort(port, snapshot.Stats)
	}
	if len(snapshot.DeviceMACs) > 0 {
		applyDeviceMACs(out, snapshot, index)
	} else if len(snapshot.MACs) > 0 {
		port.MACs = snapshot.MACs
	}
	return out
}

// uplinkPortIndex finds the represented uplink and falls back to port 1 for
// minimal synthetic profiles.
func uplinkPortIndex(ports []device.Port) int {
	for _, port := range ports {
		if port.Uplink {
			return port.Index
		}
	}
	return 1
}

// applyDeviceMACs assigns bridge-member MAC groups to represented UniFi ports,
// honoring pinned mappings, uplink filtering, and deterministic fallback order.
func applyDeviceMACs(ports []device.Port, snapshot Snapshot, uplinkIndex int) {
	for index := range ports {
		ports[index].MACs = nil
	}

	// Remote MACs are learned behind the physical uplink neighbor. Filtering
	// them keeps the represented virtual switch from claiming clients that
	// actually live behind the real upstream switch.
	remoteMACs := normalizeRemoteMACSet(snapshot.RemoteMACs)
	if len(remoteMACs) == 0 {
		remoteMACs = RemoteMACsByBridgeMember(snapshot.DeviceMACs, snapshot.MemberRoles, snapshot.Interface, snapshot.Bridge)
	}
	usedPorts := map[int]bool{uplinkIndex: true}
	accessIndexes := make([]int, 0, len(ports)-1)
	pinned := validPinnedPortSet(snapshot.MemberPortMap, len(ports))
	for _, port := range ports {
		if port.Index != uplinkIndex && !pinned[port.Index] {
			accessIndexes = append(accessIndexes, port.Index)
		}
	}
	nextAccess := 0

	for _, deviceName := range sortedDeviceNames(snapshot.DeviceMACs, snapshot.Interface, snapshot.Bridge) {
		macs := snapshot.DeviceMACs[deviceName]
		if len(macs) == 0 {
			continue
		}
		role := bridgeMemberRole(snapshot.MemberRoles, deviceName)
		if role == BridgeMemberRoleBridge || role == BridgeMemberRoleIgnored {
			continue
		}

		portIndex := uplinkIndex
		isUplink := role == BridgeMemberRoleUplink || isUplinkDevice(deviceName, snapshot.Interface, snapshot.Bridge)
		if isUplink {
			port := &ports[portIndex-1]
			applyMemberPortObservation(port, snapshot.MemberPorts, deviceName)
			usedPorts[portIndex] = true
			if deviceName != "" {
				port.Name = deviceName
			}
			continue
		}
		macs = filterRemoteMACEntries(macs, remoteMACs)
		if len(macs) == 0 {
			continue
		}
		if pinnedPort := snapshot.MemberPortMap[strings.TrimSpace(deviceName)]; pinnedPort >= 1 && pinnedPort <= len(ports) {
			portIndex = pinnedPort
		} else if nextAccess < len(accessIndexes) {
			// Unpinned bridge members are assigned deterministically by sorted
			// interface name, leaving the selected uplink and pinned ports alone.
			portIndex = accessIndexes[nextAccess]
			nextAccess++
		}
		if portIndex < 1 || portIndex > len(ports) {
			portIndex = uplinkIndex
		}

		port := &ports[portIndex-1]
		applyMemberPortObservation(port, snapshot.MemberPorts, deviceName)
		port.MACs = append(port.MACs, macs...)
		usedPorts[portIndex] = true
		if deviceName != "" && (isUplink || portIndex != uplinkIndex) {
			port.Name = deviceName
		}
	}
	for index := range ports {
		if !usedPorts[ports[index].Index] {
			// A profile port without a mapped bridge member is rendered as
			// disconnected instead of inventing a synthetic link.
			markBridgePortDisconnected(&ports[index])
		}
	}
}

// markBridgePortDisconnected clears link, counters, and MACs for generated
// ports with no observed bridge member.
func markBridgePortDisconnected(port *device.Port) {
	port.Up = false
	port.Speed = 0
	port.MACs = nil
	for _, field := range interfaceStatsFields {
		field.setPort(port, 0)
	}
}

// applyMemberPortObservation overlays per-member interface state onto the port
// selected for that bridge member.
func applyMemberPortObservation(port *device.Port, observations map[string]PortObservation, member string) {
	observation, ok := memberPortObservation(observations, member)
	if !ok {
		return
	}
	if iface := strings.TrimSpace(observation.Interface); iface != "" {
		port.Interface = iface
	}
	if observation.Up != nil {
		port.Up = *observation.Up
	}
	if observation.SpeedMbps > 0 {
		port.Speed = observation.SpeedMbps
	}
	if media := strings.TrimSpace(observation.Media); media != "" {
		port.Media = media
	}
	if hasCounters(observation.Stats) {
		applyInterfaceStatsToPort(port, observation.Stats)
	}
	if !port.Up && observation.SpeedMbps <= 0 {
		port.Speed = 0
	}
}

// memberPortObservation resolves member observations case-insensitively before
// overlaying interface state onto a port.
func memberPortObservation(observations map[string]PortObservation, member string) (PortObservation, bool) {
	if len(observations) == 0 {
		return PortObservation{}, false
	}
	if observation, ok := observations[strings.TrimSpace(member)]; ok {
		return observation, true
	}
	lower := strings.ToLower(strings.TrimSpace(member))
	for key, observation := range observations {
		if strings.ToLower(strings.TrimSpace(key)) == lower {
			return observation, true
		}
	}
	return PortObservation{}, false
}

// validPinnedPortSet reserves operator-pinned ports before automatic member
// assignment chooses fallback ports.
func validPinnedPortSet(values map[string]int, portCount int) map[int]bool {
	out := map[int]bool{}
	for _, port := range values {
		if port >= 1 && port <= portCount {
			out[port] = true
		}
	}
	return out
}

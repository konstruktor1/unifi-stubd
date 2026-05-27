package main

import (
	"fmt"
	"net"
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/adoption"
	"github.com/konstruktor1/unifi-stubd/internal/device"
	"github.com/konstruktor1/unifi-stubd/internal/device/payload"
)

// payloadForIdentity converts resolved runtime flags and adoption state into
// the stable device identity model consumed by the payload renderer.
func payloadForIdentity(
	mac net.HardwareAddr,
	ip net.IP,
	hostname string,
	informURL string,
	store adoption.Store,
	flags runtimeFlags,
	profile device.Profile,
	ports []device.Port,
	uptimeSeconds int,
) ([]byte, error) {
	return buildPayload(device.Identity{
		MAC:            mac.String(),
		IP:             ip.String(),
		Hostname:       hostname,
		Model:          flags.model,
		ModelDisplay:   flags.modelDisplay,
		DeviceType:     profile.DeviceType,
		Version:        flags.version,
		Serial:         serialFromMAC(mac),
		InformURL:      informURL,
		InformIP:       resolveInformIP(informURL),
		ManagementVLAN: effectiveManagementVLAN(flags),
		UptimeSeconds:  uptimeSeconds,
	}, profile, store, refreshPortFreshness(ports, uptimeSeconds))
}

// buildPayload overlays adoption-derived controller state onto the identity
// before delegating switch or gateway table construction to the device package.
func buildPayload(id device.Identity, profile device.Profile, store adoption.Store, ports []device.Port) ([]byte, error) {
	id.CFGVersion = store.CFGVersion
	id.Adopted = store.AuthKey != ""
	if store.Version != "" {
		id.Version = store.Version
	}
	payload, err := payload.Build(profile, id, ports)
	if err != nil {
		return nil, fmt.Errorf("build device payload: %w", err)
	}
	return payload, nil
}

// refreshPortFreshness increments synthetic connected-port counters and learned
// MAC uptimes so repeated informs do not look like a frozen first-boot payload.
func refreshPortFreshness(ports []device.Port, uptimeSeconds int) []device.Port {
	if uptimeSeconds < 1 || len(ports) == 0 {
		return ports
	}
	out := make([]device.Port, len(ports))
	copy(out, ports)
	for index := range out {
		port := &out[index]
		if port.Up && strings.TrimSpace(port.Interface) == "" {
			port.RXBytes += int64((uptimeSeconds * 64) + port.Index)
			port.TXBytes += int64((uptimeSeconds * 48) + port.Index)
			port.RXPackets += int64(uptimeSeconds)
			port.TXPackets += int64(uptimeSeconds)
		}
		for macIndex := range port.MACs {
			if port.MACs[macIndex].Uptime < uptimeSeconds {
				port.MACs[macIndex].Uptime = uptimeSeconds
			}
		}
	}
	return out
}

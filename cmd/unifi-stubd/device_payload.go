package main

import (
	"fmt"
	"net"

	"github.com/konstruktor1/unifi-stubd/internal/adoption"
	"github.com/konstruktor1/unifi-stubd/internal/device"
	"github.com/konstruktor1/unifi-stubd/internal/device/payload"
)

// controllerPayloadState contains only process-local reporting decisions
// learned from accepted controller responses. It cannot mutate host state.
type controllerPayloadState struct {
	TimeOffsetSeconds int64
	Blocks            adoption.BlockSelection
}

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
	controllerState controllerPayloadState,
) ([]byte, error) {
	return buildPayload(device.Identity{
		MAC:                         mac.String(),
		IP:                          ip.String(),
		Hostname:                    hostname,
		Model:                       flags.model,
		ModelDisplay:                flags.modelDisplay,
		DeviceType:                  profile.DeviceType,
		Version:                     flags.version,
		Serial:                      serialFromMAC(mac),
		InformURL:                   informURL,
		InformIP:                    resolveInformIP(informURL),
		ManagementVLAN:              effectiveManagementVLAN(flags),
		UptimeSeconds:               uptimeSeconds,
		ControllerTimeOffsetSeconds: controllerState.TimeOffsetSeconds,
		IPv4ActiveLeasesRequested:   controllerState.Blocks.Includes("ipv4_active_leases"),
		GatewayMACIPv6Requested:     controllerState.Blocks.Includes("port_table.mac_table_ipv6"),
		GWCapabilitiesRequested:     controllerState.Blocks.Includes("gw_caps"),
		UDAPIVersionExcluded:        controllerState.Blocks.Excludes("udapi_version"),
		DPIStatsExcluded:            controllerState.Blocks.Excludes("dpi_stats"),
		FingerprintsExcluded:        controllerState.Blocks.Excludes("fingerprints"),
	}, profile, store, refreshPortUptimes(ports, uptimeSeconds))
}

// buildPayload overlays adoption-derived controller state onto the identity
// before delegating switch or gateway table construction to the device package.
func buildPayload(id device.Identity, profile device.Profile, store adoption.Store, ports []device.Port) ([]byte, error) {
	id.CFGVersion = store.CFGVersion
	id.Adopted = store.AuthKey != ""
	id.InformState = controllerInformState(store)
	id.InformStateSet = true
	if store.Version != "" {
		id.Version = store.Version
	}
	payload, err := payload.Build(profile, id, ports)
	if err != nil {
		return nil, fmt.Errorf("build device payload: %w", err)
	}
	return payload, nil
}

// controllerInformState maps the persisted adoption phase to the observed
// wire lifecycle. A keyed phase remains at 0 until a controller noop has
// confirmed it as connected; only then may the Stub report managed state 2.
func controllerInformState(store adoption.Store) device.InformState {
	if store.AuthKey == "" {
		return device.InformStateDefault
	}
	if store.State == adoption.StateConnected {
		return device.InformStateManaged
	}
	return device.InformStateUnknown
}

// refreshPortUptimes advances only the age of locally known MAC and neighbor
// observations. Runtime counters and rates remain absent until a named local
// source supplies them.
func refreshPortUptimes(ports []device.Port, uptimeSeconds int) []device.Port {
	if uptimeSeconds < 1 || len(ports) == 0 {
		return ports
	}
	out := make([]device.Port, len(ports))
	copy(out, ports)
	for index := range out {
		port := &out[index]
		for macIndex := range port.MACs {
			if port.MACs[macIndex].Uptime < uptimeSeconds {
				port.MACs[macIndex].Uptime = uptimeSeconds
			}
		}
		for neighborIndex := range port.Neighbors {
			if port.Neighbors[neighborIndex].Uptime < uptimeSeconds {
				port.Neighbors[neighborIndex].Uptime = uptimeSeconds
			}
		}
	}
	return out
}

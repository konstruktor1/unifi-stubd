package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"

	"github.com/konstruktor1/unifi-stubd/internal/adoption"
	"github.com/konstruktor1/unifi-stubd/internal/device"
	"github.com/konstruktor1/unifi-stubd/internal/platform"
)

// buildLocalStatus assembles status from config, adoption state, optional
// platform capabilities, and passive observations without exposing secrets.
func buildLocalStatus(flags runtimeFlags, profile device.Profile, mac net.HardwareAddr, ip net.IP, hostname string, portBuildOptions device.PortBuildOptions, plt platform.Platform) localStatus {
	store, adoptionWarnings := loadAdoptionStateForStatus(flags.sshState)
	informURL := effectiveInformURL(flags.controller, store)
	ports := device.BuildPorts(profile, portBuildOptions)
	status := localStatus{
		ConfigPath: flags.configPath,
		Identity: statusIdentity{
			MAC:        mac.String(),
			IP:         ip.String(),
			Hostname:   hostname,
			Serial:     serialFromMAC(mac),
			Model:      flags.model,
			ModelName:  flags.modelDisplay,
			DeviceType: profile.DeviceType,
			Profile:    profile.Name,
			Ports:      len(ports),
			UplinkPort: uplinkPortIndex(ports),
		},
		Config: statusConfig{
			OperationMode:       flags.operationMode,
			ControllerURL:       flags.controller,
			InformURL:           informURL,
			Interval:            flags.interval.String(),
			NoDiscovery:         flags.noDiscovery,
			DiscoveryInterface:  effectiveDiscoveryInterface(flags),
			DiscoveryTargets:    cloneStrings(flags.discoveryTargets),
			ManagementLAN:       statusManagementLAN(flags),
			SSHListen:           flags.sshListen,
			StatePath:           flags.sshState,
			StatusPath:          flags.statusPath,
			UplinkNeighbor:      statusUplinkNeighborEntry(flags.uplinkNeighbor),
			PortNeighbors:       statusPortNeighbors(flags.portNeighbors),
			PortOverrides:       statusPortOverrides(flags.portOverrides),
			BridgeObserve:       cloneBridgeObserve(flags.bridgeObserve),
			PortMappings:        clonePortMappings(flags.portMappings),
			LLDPSource:          flags.lldpSource,
			TrafficSource:       flags.trafficSource,
			TrafficRatesEnabled: flags.trafficRatesEnabled,
			WANHealth:           buildStatusWANHealth(flags, profile),
			LogSource:           flags.logSource,
			ProcSource:          flags.procSource,
			DBusEnabled:         flags.dbusEnabled,
			DBusBus:             flags.dbusBus,
			SyslogPath:          flags.syslogPath,
			InstanceGuard:       flags.instanceGuard,
			InstanceGuardPath:   flags.instanceGuardPath,
		},
		Adoption: statusAdoption{
			State:      adoptionStateText(store),
			Adopted:    store.AuthKey != "",
			AuthKeySet: store.AuthKey != "",
			CFGVersion: store.CFGVersion,
			UseAESGCM:  store.UseAESGCM,
			Version:    store.Version,
		},
	}
	status.Warnings = append(status.Warnings, adoptionWarnings...)
	if plt == nil {
		plt = runtimePlatform(flags)
	}
	status.Platform = statusPlatform{Capabilities: plt.Capabilities(context.Background(), runtimePlatformConfig(flags))}

	runStatus, err := loadPersistedRunStatus(flags.statusPath)
	if err != nil {
		status.Warnings = append(status.Warnings, fmt.Sprintf("runtime status: %v", err))
	}
	status.Runtime = runStatus

	if shouldObserveStatus(flags.operationMode) {
		status.Observe = buildObservationStatus(flags, ports, plt)
	}
	return status
}

// loadAdoptionStateForStatus reports adoption-state read failures as warnings
// so status remains usable on a fresh or partially configured host.
func loadAdoptionStateForStatus(path string) (adoption.Store, []string) {
	store, err := adoption.LoadEnv(path)
	if err == nil {
		return store, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return adoption.Store{}, []string{"adoption state: not found"}
	}
	return adoption.Store{}, []string{fmt.Sprintf("adoption state: %v", err)}
}

// adoptionStateText maps an empty store to factory state for human and JSON
// status output.
func adoptionStateText(store adoption.Store) string {
	if store.State != "" {
		return string(store.State)
	}
	if store.AuthKey != "" {
		return string(adoption.StateProvisioning)
	}
	return string(adoption.StateFactory)
}

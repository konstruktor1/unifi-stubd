package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/konstruktor1/unifi-stubd/internal/adoption"
	"github.com/konstruktor1/unifi-stubd/internal/device"
	"github.com/konstruktor1/unifi-stubd/internal/observe"
)

type localStatus struct {
	ConfigPath string             `json:"config_path"`
	Identity   statusIdentity     `json:"identity"`
	Config     statusConfig       `json:"config"`
	Adoption   statusAdoption     `json:"adoption"`
	Observe    statusObservation  `json:"observe,omitempty"`
	Runtime    persistedRunStatus `json:"runtime"`
	Warnings   []string           `json:"warnings,omitempty"`
}

type statusIdentity struct {
	MAC       string `json:"mac"`
	IP        string `json:"ip"`
	Hostname  string `json:"hostname"`
	Serial    string `json:"serial"`
	Model     string `json:"model"`
	ModelName string `json:"model_name"`
	Profile   string `json:"profile"`
	Ports     int    `json:"ports"`
}

type statusConfig struct {
	OperationMode string `json:"operation_mode"`
	ControllerURL string `json:"controller_url,omitempty"`
	InformURL     string `json:"inform_url,omitempty"`
	Interval      string `json:"interval"`
	NoDiscovery   bool   `json:"no_discovery"`
	SSHListen     string `json:"ssh_listen,omitempty"`
	StatePath     string `json:"state_path"`
	StatusPath    string `json:"status_path"`
}

type statusAdoption struct {
	State      string `json:"state"`
	Adopted    bool   `json:"adopted"`
	AuthKeySet bool   `json:"authkey_set"`
	CFGVersion string `json:"cfgversion,omitempty"`
	UseAESGCM  bool   `json:"use_aes_gcm"`
	Version    string `json:"version,omitempty"`
}

type statusObservation struct {
	Interface      string   `json:"interface,omitempty"`
	Bridge         string   `json:"bridge,omitempty"`
	SpeedMbps      int      `json:"speed_mbps,omitempty"`
	RXBytes        int64    `json:"rx_bytes,omitempty"`
	TXBytes        int64    `json:"tx_bytes,omitempty"`
	RXPackets      int64    `json:"rx_packets,omitempty"`
	TXPackets      int64    `json:"tx_packets,omitempty"`
	BridgeDevices  int      `json:"bridge_devices,omitempty"`
	LearnedMACs    int      `json:"learned_macs,omitempty"`
	SourceWarnings []string `json:"source_warnings,omitempty"`
}

type persistedRunStatus struct {
	LastInform lastInformStatus `json:"last_inform,omitempty"`
}

type lastInformStatus struct {
	Time            string `json:"time,omitempty"`
	URL             string `json:"url,omitempty"`
	StatusCode      int    `json:"status_code,omitempty"`
	ResponseType    string `json:"response_type,omitempty"`
	ControllerState string `json:"controller_state,omitempty"`
	CFGVersion      string `json:"cfgversion,omitempty"`
	Version         string `json:"version,omitempty"`
	UsedAESGCM      bool   `json:"used_aes_gcm,omitempty"`
	RawBytes        int    `json:"raw_bytes,omitempty"`
	JSONBytes       int    `json:"json_bytes,omitempty"`
	Error           string `json:"error,omitempty"`
}

func printLocalStatus(flags runtimeFlags, profile device.Profile, mac net.HardwareAddr, ip net.IP, hostname string, portOptions device.PortOptions) error {
	status := buildLocalStatus(flags, profile, mac, ip, hostname, portOptions)
	if *flags.statusJSON {
		data, err := json.MarshalIndent(status, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}
	printHumanStatus(status)
	return nil
}

func buildLocalStatus(flags runtimeFlags, profile device.Profile, mac net.HardwareAddr, ip net.IP, hostname string, portOptions device.PortOptions) localStatus {
	store, adoptionWarnings := loadAdoptionStateForStatus(*flags.sshState)
	informURL := effectiveInformURL(*flags.controller, store)
	ports := device.SwitchPortsWithOptions(*flags.portCount, portOptions)
	status := localStatus{
		ConfigPath: *flags.configPath,
		Identity: statusIdentity{
			MAC:       mac.String(),
			IP:        ip.String(),
			Hostname:  hostname,
			Serial:    serialFromMAC(mac),
			Model:     *flags.model,
			ModelName: *flags.modelDisplay,
			Profile:   profile.Name,
			Ports:     len(ports),
		},
		Config: statusConfig{
			OperationMode: *flags.operationMode,
			ControllerURL: *flags.controller,
			InformURL:     informURL,
			Interval:      flags.interval.String(),
			NoDiscovery:   *flags.noDiscovery,
			SSHListen:     *flags.sshListen,
			StatePath:     *flags.sshState,
			StatusPath:    *flags.statusPath,
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

	runStatus, err := loadPersistedRunStatus(*flags.statusPath)
	if err != nil {
		status.Warnings = append(status.Warnings, fmt.Sprintf("runtime status: %v", err))
	}
	status.Runtime = runStatus

	if shouldObserveStatus(*flags.operationMode) {
		status.Observe = buildObservationStatus(flags, ports)
	}
	return status
}

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

func adoptionStateText(store adoption.Store) string {
	if store.State != "" {
		return string(store.State)
	}
	if store.AuthKey != "" {
		return string(adoption.StateProvisioning)
	}
	return string(adoption.StateFactory)
}

func shouldObserveStatus(mode string) bool {
	mode = normalizeMode(mode)
	return mode == operationModeObserve || mode == operationModeHostDirect
}

func buildObservationStatus(flags runtimeFlags, ports []device.Port) statusObservation {
	ctx, cancel := context.WithTimeout(context.Background(), observeTimeout)
	defer cancel()

	snapshot, errs := observe.LinuxSnapshot(ctx, observe.Config{
		Interface: strings.TrimSpace(*flags.observeInterface),
		Bridge:    strings.TrimSpace(*flags.observeBridge),
	}, uplinkPortIndex(ports))

	out := statusObservation{
		Interface:     strings.TrimSpace(*flags.observeInterface),
		Bridge:        strings.TrimSpace(*flags.observeBridge),
		SpeedMbps:     snapshot.Stats.SpeedMbps,
		RXBytes:       snapshot.Stats.RXBytes,
		TXBytes:       snapshot.Stats.TXBytes,
		RXPackets:     snapshot.Stats.RXPackets,
		TXPackets:     snapshot.Stats.TXPackets,
		BridgeDevices: len(snapshot.DeviceMACs),
		LearnedMACs:   len(snapshot.MACs),
	}
	for _, err := range errs {
		out.SourceWarnings = append(out.SourceWarnings, err.Error())
	}
	return out
}

func printHumanStatus(status localStatus) {
	fmt.Println("unifi-stubd status")
	fmt.Printf("config_path: %s\n", status.ConfigPath)
	fmt.Printf("operation_mode: %s\n", status.Config.OperationMode)
	fmt.Printf("profile: %s (%s)\n", status.Identity.Profile, status.Identity.Model)
	fmt.Printf("model_name: %s\n", status.Identity.ModelName)
	fmt.Printf("mac: %s\n", status.Identity.MAC)
	fmt.Printf("ip: %s\n", status.Identity.IP)
	fmt.Printf("hostname: %s\n", status.Identity.Hostname)
	fmt.Printf("serial: %s\n", status.Identity.Serial)
	fmt.Printf("ports: %d\n", status.Identity.Ports)
	fmt.Printf("controller_url: %s\n", valueOrDash(status.Config.ControllerURL))
	fmt.Printf("inform_url: %s\n", valueOrDash(status.Config.InformURL))
	fmt.Printf("interval: %s\n", status.Config.Interval)
	fmt.Printf("no_discovery: %t\n", status.Config.NoDiscovery)
	fmt.Printf("ssh_listen: %s\n", valueOrDash(status.Config.SSHListen))
	fmt.Printf("state_path: %s\n", status.Config.StatePath)
	fmt.Printf("status_path: %s\n", status.Config.StatusPath)
	fmt.Printf("adoption_state: %s\n", status.Adoption.State)
	fmt.Printf("adopted: %t\n", status.Adoption.Adopted)
	fmt.Printf("authkey_set: %t\n", status.Adoption.AuthKeySet)
	fmt.Printf("cfgversion: %s\n", valueOrDash(status.Adoption.CFGVersion))
	fmt.Printf("use_aes_gcm: %t\n", status.Adoption.UseAESGCM)
	fmt.Printf("version: %s\n", valueOrDash(status.Adoption.Version))
	printObservationStatus(status.Observe)
	printLastInform(status.Runtime.LastInform)
	for _, warning := range status.Warnings {
		fmt.Printf("warning: %s\n", warning)
	}
}

func printObservationStatus(status statusObservation) {
	if status.Interface == "" && status.Bridge == "" {
		return
	}
	fmt.Printf("observe_interface: %s\n", valueOrDash(status.Interface))
	fmt.Printf("observe_bridge: %s\n", valueOrDash(status.Bridge))
	fmt.Printf("observe_speed_mbps: %d\n", status.SpeedMbps)
	fmt.Printf("observe_rx_bytes: %d\n", status.RXBytes)
	fmt.Printf("observe_tx_bytes: %d\n", status.TXBytes)
	fmt.Printf("observe_bridge_devices: %d\n", status.BridgeDevices)
	fmt.Printf("observe_learned_macs: %d\n", status.LearnedMACs)
	for _, warning := range status.SourceWarnings {
		fmt.Printf("observe_warning: %s\n", warning)
	}
}

func printLastInform(last lastInformStatus) {
	if last.Time == "" {
		fmt.Println("last_inform: none")
		return
	}
	fmt.Printf("last_inform_time: %s\n", last.Time)
	fmt.Printf("last_inform_url: %s\n", valueOrDash(last.URL))
	fmt.Printf("last_inform_status: %d\n", last.StatusCode)
	fmt.Printf("last_inform_type: %s\n", valueOrDash(last.ResponseType))
	fmt.Printf("last_inform_state: %s\n", valueOrDash(last.ControllerState))
	fmt.Printf("last_inform_cfgversion: %s\n", valueOrDash(last.CFGVersion))
	fmt.Printf("last_inform_version: %s\n", valueOrDash(last.Version))
	fmt.Printf("last_inform_used_aes_gcm: %t\n", last.UsedAESGCM)
	fmt.Printf("last_inform_raw_bytes: %d\n", last.RawBytes)
	fmt.Printf("last_inform_json_bytes: %d\n", last.JSONBytes)
	if last.Error != "" {
		fmt.Printf("last_inform_error: %s\n", last.Error)
	}
}

func valueOrDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}

func loadPersistedRunStatus(path string) (persistedRunStatus, error) {
	var status persistedRunStatus
	data, err := os.ReadFile(path)
	if err != nil {
		return status, err
	}
	if err := json.Unmarshal(data, &status); err != nil {
		return status, err
	}
	return status, nil
}

func saveLastInformStatus(path string, last lastInformStatus) error {
	if path == "" {
		return errors.New("status path is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(persistedRunStatus{LastInform: last}, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o600)
}

func newLastInformStatus(url string, store adoption.Store) lastInformStatus {
	return lastInformStatus{
		Time:            time.Now().Format(time.RFC3339),
		URL:             url,
		ControllerState: adoptionStateText(store),
		CFGVersion:      store.CFGVersion,
		Version:         store.Version,
	}
}

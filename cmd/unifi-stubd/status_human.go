package main

import (
	"fmt"
	"strings"
)

// printHumanStatus emits stable key-value lines rather than prose so operators
// and tests can grep individual safety, adoption, and observation fields.
func printHumanStatus(status localStatus) {
	fmt.Println("unifi-stubd status")
	fmt.Printf("config_path: %s\n", status.ConfigPath)
	fmt.Printf("operation_mode: %s\n", status.Config.OperationMode)
	fmt.Printf("profile: %s (%s)\n", status.Identity.Profile, status.Identity.Model)
	fmt.Printf("model_name: %s\n", status.Identity.ModelName)
	fmt.Printf("device_type: %s\n", valueOrDash(status.Identity.DeviceType))
	fmt.Printf("mac: %s\n", status.Identity.MAC)
	fmt.Printf("ip: %s\n", status.Identity.IP)
	fmt.Printf("hostname: %s\n", status.Identity.Hostname)
	fmt.Printf("serial: %s\n", status.Identity.Serial)
	fmt.Printf("ports: %d\n", status.Identity.Ports)
	fmt.Printf("uplink_port: %d\n", status.Identity.UplinkPort)
	fmt.Printf("controller_url: %s\n", valueOrDash(status.Config.ControllerURL))
	fmt.Printf("inform_url: %s\n", valueOrDash(status.Config.InformURL))
	fmt.Printf("interval: %s\n", status.Config.Interval)
	fmt.Printf("no_discovery: %t\n", status.Config.NoDiscovery)
	printHumanManagementLANStatus(status)
	printHumanSourceStatus(status)
	printHumanDiscoveryStatus(status)
	printHumanNeighborStatus(status)
	printHumanPortOverrideStatus(status)
	fmt.Printf("ssh_listen: %s\n", valueOrDash(status.Config.SSHListen))
	fmt.Printf("state_path: %s\n", status.Config.StatePath)
	fmt.Printf("status_path: %s\n", status.Config.StatusPath)
	fmt.Printf("instance_guard: %s\n", status.Config.InstanceGuard)
	fmt.Printf("instance_guard_path: %s\n", valueOrDash(status.Config.InstanceGuardPath))
	fmt.Printf("adoption_state: %s\n", status.Adoption.State)
	fmt.Printf("adopted: %t\n", status.Adoption.Adopted)
	fmt.Printf("authkey_set: %t\n", status.Adoption.AuthKeySet)
	fmt.Printf("cfgversion: %s\n", valueOrDash(status.Adoption.CFGVersion))
	fmt.Printf("use_aes_gcm: %t\n", status.Adoption.UseAESGCM)
	fmt.Printf("version: %s\n", valueOrDash(status.Adoption.Version))
	printObservationStatus(status.Observe)
	printPlatformStatus(status.Platform)
	printLastInform(status.Runtime.LastInform)
	for _, warning := range status.Warnings {
		fmt.Printf("warning: %s\n", warning)
	}
}

func printHumanManagementLANStatus(status localStatus) {
	if status.Config.ManagementLAN == nil || !status.Config.ManagementLAN.Enabled {
		return
	}
	fmt.Printf("management_lan.mode: %s\n", status.Config.ManagementLAN.Mode)
	fmt.Printf("management_lan.vlan: %d\n", status.Config.ManagementLAN.VLAN)
	fmt.Printf("management_lan.interface: %s\n", valueOrDash(status.Config.ManagementLAN.Interface))
	fmt.Printf("management_lan.ip: %s\n", valueOrDash(status.Config.ManagementLAN.IP))
	fmt.Printf("management_lan.network_name: %s\n", valueOrDash(status.Config.ManagementLAN.NetworkName))
	fmt.Printf("management_lan.controller_reachable: %s\n", status.Config.ManagementLAN.ControllerReachable)
	fmt.Printf("management_lan.adoption_strategy: %s\n", status.Config.ManagementLAN.AdoptionStrategy)
}

func printHumanSourceStatus(status localStatus) {
	fmt.Printf("lldp_source: %s\n", status.Config.LLDPSource)
	fmt.Printf("traffic_source: %s\n", status.Config.TrafficSource)
	fmt.Printf("traffic_rates_enabled: %t\n", status.Config.TrafficRatesEnabled)
	printWANHealthStatus(status.Config.WANHealth)
	fmt.Printf("log_source: %s\n", status.Config.LogSource)
	fmt.Printf("proc_source: %s\n", status.Config.ProcSource)
	fmt.Printf("dbus_enabled: %t\n", status.Config.DBusEnabled)
	fmt.Printf("dbus_bus: %s\n", valueOrDash(status.Config.DBusBus))
	if status.Config.SyslogPath != "" {
		fmt.Printf("syslog_path: %s\n", status.Config.SyslogPath)
	}
}

func printHumanDiscoveryStatus(status localStatus) {
	if status.Config.DiscoveryInterface != "" {
		fmt.Printf("discovery_interface: %s\n", status.Config.DiscoveryInterface)
	}
	for _, target := range status.Config.DiscoveryTargets {
		fmt.Printf("discovery_target: %s\n", target)
	}
}

func printHumanNeighborStatus(status localStatus) {
	if status.Config.UplinkNeighbor != nil {
		fmt.Printf("uplink_neighbor: mac=%s hostname=%s ip=%s vlan=%d type=%s\n",
			status.Config.UplinkNeighbor.MAC,
			valueOrDash(status.Config.UplinkNeighbor.Hostname),
			valueOrDash(status.Config.UplinkNeighbor.IP),
			status.Config.UplinkNeighbor.VLAN,
			valueOrDash(status.Config.UplinkNeighbor.Type),
		)
	}
	for _, neighbor := range status.Config.PortNeighbors {
		fmt.Printf("port_neighbor: port=%d mac=%s hostname=%s ip=%s vlan=%d type=%s\n",
			neighbor.Port,
			neighbor.MAC,
			valueOrDash(neighbor.Hostname),
			valueOrDash(neighbor.IP),
			neighbor.VLAN,
			valueOrDash(neighbor.Type),
		)
	}
}

func printHumanPortOverrideStatus(status localStatus) {
	for _, override := range status.Config.PortOverrides {
		fmt.Printf("port_override: port=%d interface=%s mac=%s ip=%s netmask=%s role=%s network_group=%s speed=%d media=%s up=%s name=%s\n",
			override.Port,
			valueOrDash(override.Interface),
			valueOrDash(override.MAC),
			valueOrDash(override.IP),
			valueOrDash(override.Netmask),
			valueOrDash(override.Role),
			valueOrDash(override.NetworkGroup),
			override.Speed,
			valueOrDash(override.Media),
			boolPointerText(override.Up),
			valueOrDash(override.Name),
		)
	}
}

// printPlatformStatus renders optional host integration availability without
// implying that missing tools are required.
func printPlatformStatus(status statusPlatform) {
	if status.Capabilities.GOOS == "" {
		return
	}
	fmt.Printf("platform_goos: %s\n", status.Capabilities.GOOS)
	for _, capability := range status.Capabilities.Capabilities {
		fmt.Printf("platform_capability: name=%s source=%s state=%s detail=%s\n",
			capability.Name,
			valueOrDash(capability.Source),
			capability.State,
			valueOrDash(capability.Detail),
		)
	}
}

// printObservationStatus reports only the passive observation summary and
// warnings, never raw host command output.
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

// printLastInform explains the most recent controller exchange without writing
// raw inform payloads, auth keys, or controller provisioning bodies.
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
	fmt.Printf("last_inform_attempted_aes_gcm: %t\n", last.AttemptedAESGCM)
	fmt.Printf("last_inform_used_aes_gcm: %t\n", last.UsedAESGCM)
	fmt.Printf("last_inform_fallback_to_cbc: %t\n", last.FallbackToCBC)
	fmt.Printf("last_inform_raw_bytes: %d\n", last.RawBytes)
	fmt.Printf("last_inform_json_bytes: %d\n", last.JSONBytes)
	printLastInformTraffic(last.Traffic)
	if last.IntervalSeconds > 0 {
		fmt.Printf("last_inform_interval_seconds: %d\n", last.IntervalSeconds)
	}
	for _, block := range last.IncludeBlocks {
		fmt.Printf("last_inform_include_block: %s\n", block)
	}
	if last.ResetRequested {
		fmt.Println("last_inform_reset_requested: true")
		fmt.Printf("last_inform_reset_applied: %t\n", last.ResetApplied)
		fmt.Printf("last_inform_reset_reason: %s\n", valueOrDash(last.ResetReason))
	}
	if last.HasMgmtCFG {
		fmt.Println("last_inform_has_mgmt_cfg: true")
	}
	if last.HasSystemCFG {
		fmt.Printf("last_inform_has_system_cfg: true\n")
		fmt.Printf("last_inform_system_cfg_bytes: %d\n", last.SystemCFGBytes)
		for _, key := range last.SystemCFGKeys {
			fmt.Printf("last_inform_system_cfg_key: %s\n", key)
		}
	}
	if last.Ignored {
		fmt.Println("last_inform_ignored: true")
		fmt.Printf("last_inform_ignored_reason: %s\n", valueOrDash(last.IgnoredReason))
	}
	if last.Error != "" {
		fmt.Printf("last_inform_error: %s\n", last.Error)
	}
}

func printLastInformTraffic(traffic *lastInformTrafficStatus) {
	if traffic == nil {
		return
	}
	printTrafficRates("last_inform_traffic_root", traffic.Root)
	for _, row := range traffic.Rows {
		prefix := "last_inform_traffic_row"
		fmt.Printf("%s: table=%s port_idx=%d ifname=%s source_interface=%s\n",
			prefix,
			row.Table,
			row.PortIdx,
			valueOrDash(row.IfName),
			valueOrDash(row.SourceInterface),
		)
		printTrafficRates(prefix, row.Rates)
	}
}

func printTrafficRates(prefix string, rates lastInformTrafficRates) {
	if rates.Bytes != nil {
		fmt.Printf("%s_bytes: %d\n", prefix, *rates.Bytes)
	}
	if rates.RXBytes != nil {
		fmt.Printf("%s_rx_bytes: %d\n", prefix, *rates.RXBytes)
	}
	if rates.TXBytes != nil {
		fmt.Printf("%s_tx_bytes: %d\n", prefix, *rates.TXBytes)
	}
	if rates.BytesRateBytesPerSecond != nil {
		fmt.Printf("%s_bytes_rate_bytes_per_second: %d\n", prefix, *rates.BytesRateBytesPerSecond)
	}
	if rates.RXBytesRateBytesPerSecond != nil {
		fmt.Printf("%s_rx_bytes_rate_bytes_per_second: %d\n", prefix, *rates.RXBytesRateBytesPerSecond)
	}
	if rates.TXBytesRateBytesPerSecond != nil {
		fmt.Printf("%s_tx_bytes_rate_bytes_per_second: %d\n", prefix, *rates.TXBytesRateBytesPerSecond)
	}
	if rates.RXRateBitsPerSecond != nil {
		fmt.Printf("%s_rx_rate_bits_per_second: %d\n", prefix, *rates.RXRateBitsPerSecond)
	}
	if rates.TXRateBitsPerSecond != nil {
		fmt.Printf("%s_tx_rate_bits_per_second: %d\n", prefix, *rates.TXRateBitsPerSecond)
	}
}

// valueOrDash keeps human status output readable when optional fields are
// absent.
func valueOrDash(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}

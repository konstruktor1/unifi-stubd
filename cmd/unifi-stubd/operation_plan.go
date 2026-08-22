package main

import (
	"fmt"
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/device"
)

// printRuntimePlan describes the non-mutating runtime plan, including actions
// that are intentionally dry-run-only for host-networking safety.
func printRuntimePlan(flags runtimeFlags, profile device.Profile, informSchedulePolicy device.InformSchedulePolicy, reportingPolicy device.ControllerReportingPolicy, macText, ipText, hostname string) {
	mode := normalizeMode(flags.operationMode)
	fmt.Printf("operation_mode: %s\n", mode)
	fmt.Printf("profile: %s (%s)\n", profile.Name, profile.Model)
	fmt.Printf("device_type: %s\n", profile.DeviceType)
	fmt.Printf("mac: %s\n", macText)
	fmt.Printf("ip: %s\n", ipText)
	fmt.Printf("hostname: %s\n", hostname)
	fmt.Printf("interval: %s\n", flags.interval)
	fmt.Printf("controller_interval_min: %s\n", flags.controllerIntervalMin)
	fmt.Printf("controller_interval_max: %s\n", flags.controllerIntervalMax)
	policyStatus := controllerIntervalPolicyStatus(flags, informSchedulePolicy)
	fmt.Printf("controller_interval_hints_supported: %t\n", policyStatus.Supported)
	if policyStatus.FirmwareMinimum != "" {
		fmt.Printf("controller_interval_firmware_minimum: %s\n", policyStatus.FirmwareMinimum)
	}
	if policyStatus.EffectiveMinimum != "" {
		fmt.Printf("controller_interval_effective_minimum: %s\n", policyStatus.EffectiveMinimum)
	}
	if policyStatus.Reason != "" {
		fmt.Printf("controller_interval_policy_reason: %s\n", policyStatus.Reason)
	}
	reportingStatus := controllerReportingPolicyStatus(reportingPolicy)
	fmt.Printf("controller_block_selection_supported: %t\n", reportingStatus.BlockSelection)
	fmt.Printf("controller_clock_supported: %t\n", reportingStatus.ControllerClock)
	fmt.Printf("controller_idle_speedtest: %t\n", reportingStatus.IdleSpeedtest)
	if reportingStatus.Reason != "" {
		fmt.Printf("controller_reporting_policy_reason: %s\n", reportingStatus.Reason)
	}
	fmt.Printf("inform_backoff_max: %s\n", flags.informBackoffMax)
	fmt.Printf("inform_backoff_jitter_percent: %d\n", flags.informBackoffJitterPercent)
	fmt.Printf("uplink_port: %d\n", effectiveUplinkPort(profile, flags))
	if flags.uplinkNeighbor != nil {
		fmt.Printf("uplink_neighbor: mac=%s hostname=%q ip=%q vlan=%d type=%q remote_port=%d\n",
			flags.uplinkNeighbor.MAC,
			strings.TrimSpace(flags.uplinkNeighbor.Hostname),
			strings.TrimSpace(flags.uplinkNeighbor.IP),
			flags.uplinkNeighbor.VLAN,
			strings.TrimSpace(flags.uplinkNeighbor.DeviceType),
			flags.uplinkNeighbor.RemotePort,
		)
	}
	for _, neighbor := range flags.portNeighbors {
		fmt.Printf("port_neighbor: port=%d mac=%s hostname=%q ip=%q vlan=%d type=%q remote_port=%d\n",
			neighbor.Port,
			neighbor.Entry.MAC,
			strings.TrimSpace(neighbor.Entry.Hostname),
			strings.TrimSpace(neighbor.Entry.IP),
			neighbor.Entry.VLAN,
			strings.TrimSpace(neighbor.Entry.Type),
			neighbor.RemotePort,
		)
	}
	printTopologyPlan(flags, profile)
	for _, override := range flags.portOverrides {
		override = device.NormalizePortOverride(override)
		fmt.Printf("port_override: port=%d interface=%q mac=%q ip=%q netmask=%q role=%q network_group=%q speed=%d media=%q up=%s name=%q\n",
			override.Port,
			override.Interface,
			override.MAC,
			override.IP,
			override.Netmask,
			override.Role,
			override.NetworkGroup,
			override.Speed,
			override.Media,
			boolPointerText(override.Up),
			override.Name,
		)
	}
	fmt.Printf("observe_interface: %s\n", strings.TrimSpace(flags.observeInterface))
	fmt.Printf("observe_bridge: %s\n", strings.TrimSpace(flags.observeBridge))
	bridgeObserve := effectiveBridgeObserve(flags)
	fmt.Printf("bridge_observe.bridge: %s\n", strings.TrimSpace(bridgeObserve.Bridge))
	fmt.Printf("bridge_observe.uplink_interface: %s\n", strings.TrimSpace(bridgeObserve.UplinkInterface))
	for _, member := range bridgeObserve.IgnoredMembers {
		fmt.Printf("bridge_ignore_member: member=%s\n", strings.TrimSpace(member))
	}
	for _, mapping := range bridgeObserve.MemberPortMap {
		fmt.Printf("bridge_member_port: member=%s port=%d\n", strings.TrimSpace(mapping.Member), mapping.Port)
	}
	for _, mapping := range flags.portMappings {
		fmt.Printf("port_mapping: port=%d interface=%q disabled=%t unmapped=%t\n",
			mapping.Port,
			strings.TrimSpace(mapping.Interface),
			mapping.Disabled,
			mapping.Unmapped,
		)
	}
	fmt.Printf("lldp_source: %s\n", strings.TrimSpace(flags.lldpSource))
	fmt.Printf("traffic_rates_enabled: %t\n", flags.trafficRatesEnabled)
	fmt.Printf("wan_health_source: %s\n", strings.TrimSpace(flags.wanHealth.Source))
	if flags.wanHealth.IntervalSeconds > 0 {
		fmt.Printf("wan_health_interval_seconds: %d\n", flags.wanHealth.IntervalSeconds)
	}
	if flags.wanHealth.TimeoutMS > 0 {
		fmt.Printf("wan_health_timeout_ms: %d\n", flags.wanHealth.TimeoutMS)
	}
	for _, target := range flags.wanHealth.Targets {
		fmt.Printf("wan_health_target: port=%d host=%s\n", target.Port, strings.TrimSpace(target.Host))
	}
	fmt.Printf("proc_source: %s\n", strings.TrimSpace(flags.procSource))
	fmt.Printf("instance_guard: %s\n", strings.TrimSpace(flags.instanceGuard))
	fmt.Printf("instance_guard_path: %s\n", strings.TrimSpace(flags.instanceGuardPath))
	printManagementLANPlan(flags)
	if iface := effectiveDiscoveryInterface(flags); iface != "" {
		fmt.Printf("discovery_interface: %s\n", iface)
	}
	for _, target := range flags.discoveryTargets {
		fmt.Printf("discovery_target: %s\n", strings.TrimSpace(target))
	}
	switch mode {
	case operationModeStub:
		fmt.Println("actions: synthetic stub only; no host network changes")
	case operationModeBridgeObserve:
		fmt.Println("actions: read-only bridge observation; no host network changes")
	case operationModePortMap:
		fmt.Println("actions: read-only explicit port mapping; no host network changes")
	case operationModeHostDirect:
		fmt.Println("actions: direct host identity mode; no host network changes")
	case operationModeMacvlan:
		parent := strings.TrimSpace(flags.observeInterface)
		if parent == "" {
			parent = "<required-parent-interface>"
		}
		fmt.Println("actions: macvlan is not executed by this release")
		fmt.Printf("planned_command: ip link add link %s name unifi-stubd0 type macvlan mode bridge\n", parent)
		fmt.Printf("planned_command: ip link set unifi-stubd0 address %s up\n", macText)
		fmt.Printf("planned_note: assign %s to unifi-stubd0 after subnet/prefix config exists\n", ipText)
	}
}

func printTopologyPlan(flags runtimeFlags, profile device.Profile) {
	ports := device.BuildPorts(profile, device.PortBuildOptions{
		Count:      flags.portCount,
		LinkSpeed:  flags.linkSpeed,
		UplinkPort: effectiveUplinkPort(profile, flags),
	})
	ports = device.ApplyPortNeighbors(ports, flags.portNeighbors)
	ports = device.ApplyUplinkNeighbor(ports, flags.uplinkNeighbor)
	resolution, err := device.ResolveTopology(ports)
	if err != nil {
		fmt.Printf("topology_error: %v\n", err)
		return
	}
	for _, warning := range resolution.Warnings {
		fmt.Printf("topology_warning: %s\n", warning)
	}
	for _, edge := range resolution.Edges {
		fmt.Printf("topology_edge: local_port=%d direction=%s source=%s neighbor_type=%s remote_port=%d\n",
			edge.LocalPort,
			edge.Direction,
			edge.Neighbor.Source,
			edge.Neighbor.DeviceType,
			edge.Neighbor.RemotePort,
		)
	}
}

// printManagementLANPlan renders management-LAN intent separately so operators
// can review whether it is metadata-only, preexisting-interface, or dry-run
// host VLAN planning.
func printManagementLANPlan(flags runtimeFlags) {
	cfg := effectiveManagementLAN(flags)
	fmt.Printf("management_lan.enabled: %t\n", cfg.Enabled)
	if !cfg.Enabled {
		return
	}
	fmt.Printf("management_lan.mode: %s\n", cfg.Mode)
	fmt.Printf("management_lan.vlan: %d\n", cfg.VLAN)
	if cfg.NetworkName != "" {
		fmt.Printf("management_lan.network_name: %s\n", cfg.NetworkName)
	}
	if cfg.Interface != "" {
		fmt.Printf("management_lan.interface: %s\n", cfg.Interface)
	}
	if cfg.IP != "" {
		fmt.Printf("management_lan.ip: %s\n", cfg.IP)
	}
	fmt.Printf("management_lan.controller_reachable: %s\n", cfg.ControllerReachable)
	fmt.Printf("management_lan.adoption_strategy: %s\n", cfg.AdoptionStrategy)
	switch cfg.Mode {
	case managementLANModeMetadataOnly:
		fmt.Println("management_lan.actions: controller metadata only; no host VLAN changes")
	case managementLANModePreexistingInterface:
		fmt.Println("management_lan.actions: use preexisting VLAN interface; no host VLAN changes")
	case managementLANModePlannedHostVLAN:
		fmt.Println("management_lan.actions: planned host VLAN only; no host VLAN changes")
		fmt.Printf("management_lan.planned_note: create and address VLAN interface %s for VLAN %d outside unifi-stubd\n", valueOrDash(cfg.Interface), cfg.VLAN)
	}
}

// boolPointerText distinguishes unset optional booleans from explicit true or
// false in plans and status output.
func boolPointerText(value *bool) string {
	if value == nil {
		return "unset"
	}
	return fmt.Sprintf("%t", *value)
}

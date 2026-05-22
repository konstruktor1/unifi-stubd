// Operation-mode validation is the safety gate between synthetic stubbing,
// read-only host observation, and planned host-network modes. Anything capable
// of mutating the host remains rejected or dry-run-only here.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	appconfig "github.com/konstruktor1/unifi-stubd/internal/config"
	"github.com/konstruktor1/unifi-stubd/internal/device"
	"github.com/konstruktor1/unifi-stubd/internal/observe"
	"github.com/konstruktor1/unifi-stubd/internal/observe/portmap"
	"github.com/konstruktor1/unifi-stubd/internal/platform"
)

// Operation modes define the host-observation boundary for the daemon.
const (
	operationModeStub          = "stub"
	operationModeObserve       = "observe"
	operationModeBridgeObserve = "bridge-observe"
	operationModePortMap       = "port-map"
	operationModeHostDirect    = "host-direct"
	operationModeMacvlan       = "macvlan"

	trafficSourceOff = "off"
	observeTimeout   = 2 * time.Second
)

// validateOperationFlags normalizes operator-selected modes and optional host
// sources before later validation can decide whether live checks are required.
func validateOperationFlags(flags *runtimeFlags) error {
	mode := normalizeMode(flags.operationMode)
	flags.operationMode = mode
	switch mode {
	case operationModeStub, operationModeBridgeObserve, operationModePortMap, operationModeHostDirect, operationModeMacvlan:
	default:
		return fmt.Errorf("invalid -operation-mode %q; use stub, bridge-observe, observe, port-map, host-direct, or macvlan", mode)
	}

	lldpSource := strings.ToLower(strings.TrimSpace(flags.lldpSource))
	if lldpSource == "" {
		lldpSource = platform.SourceOff
	}
	flags.lldpSource = lldpSource
	switch lldpSource {
	case platform.SourceOff, platform.LLDPSourceLLDPD:
	default:
		return fmt.Errorf("invalid -lldp-source %q; use off or lldpd", lldpSource)
	}

	trafficSource := strings.ToLower(strings.TrimSpace(flags.trafficSource))
	if trafficSource == "" {
		trafficSource = trafficSourceOff
	}
	flags.trafficSource = trafficSource
	if trafficSource != trafficSourceOff {
		return fmt.Errorf("invalid -traffic-source %q; only off is implemented", trafficSource)
	}

	flags.logSource = strings.ToLower(strings.TrimSpace(flags.logSource))
	if flags.logSource == "" {
		flags.logSource = platform.SourceOff
	}
	switch flags.logSource {
	case platform.SourceOff, platform.LogSourceJournalctl, platform.LogSourceSyslog:
	default:
		return fmt.Errorf("invalid -log-source %q; use off, journalctl, or syslog", flags.logSource)
	}

	flags.procSource = strings.ToLower(strings.TrimSpace(flags.procSource))
	if flags.procSource == "" {
		flags.procSource = platform.SourceOff
	}
	switch flags.procSource {
	case platform.SourceOff, platform.ProcSourceProcFS:
	default:
		return fmt.Errorf("invalid -proc-source %q; use off or procfs", flags.procSource)
	}
	flags.dbusBus = strings.ToLower(strings.TrimSpace(flags.dbusBus))
	if flags.dbusBus == "" {
		flags.dbusBus = platform.DBusBusSystem
	}
	switch flags.dbusBus {
	case platform.DBusBusSystem, platform.DBusBusSession:
	default:
		return fmt.Errorf("invalid -dbus-bus %q; use system or session", flags.dbusBus)
	}

	if strings.EqualFold(strings.TrimSpace(flags.macText), "host") && mode != operationModeHostDirect {
		return fmt.Errorf("mac: host is only allowed with -operation-mode host-direct")
	}
	// Planned host-networking modes remain review-only. The daemon may print
	// the intended macvlan commands, but it must not create interfaces itself.
	if mode == operationModeMacvlan && !flags.dryRunPlan {
		return fmt.Errorf("operation-mode macvlan is planned only; use -dry-run-plan to inspect the non-mutating plan")
	}
	if iface := strings.TrimSpace(flags.discoveryInterface); strings.Contains(iface, "/") {
		return fmt.Errorf("invalid -discovery-interface %q", iface)
	}
	return nil
}

// validatePortOverrides checks configured payload metadata before it can reach
// generated ports, keeping invalid MAC/IP/speed data out of inform payloads.
func validatePortOverrides(flags runtimeFlags) error {
	if flags.uplinkPort < 0 || flags.uplinkPort > flags.portCount {
		return fmt.Errorf("invalid -uplink-port %d; use 0 or 1..%d", flags.uplinkPort, flags.portCount)
	}
	if flags.uplinkNeighbor != nil {
		if _, err := net.ParseMAC(flags.uplinkNeighbor.MAC); err != nil {
			return fmt.Errorf("invalid uplink_neighbor mac %q: %w", flags.uplinkNeighbor.MAC, err)
		}
		if flags.uplinkNeighbor.VLAN < 0 {
			return fmt.Errorf("invalid uplink_neighbor vlan %d; use 0 or a positive VLAN ID", flags.uplinkNeighbor.VLAN)
		}
		if ip := strings.TrimSpace(flags.uplinkNeighbor.IP); ip != "" && net.ParseIP(ip).To4() == nil {
			return fmt.Errorf("invalid uplink_neighbor ip %q; use an IPv4 address", flags.uplinkNeighbor.IP)
		}
	}
	for _, neighbor := range flags.portNeighbors {
		if neighbor.Port < 1 || neighbor.Port > flags.portCount {
			return fmt.Errorf("invalid port neighbor %d; use 1..%d", neighbor.Port, flags.portCount)
		}
		if _, err := net.ParseMAC(neighbor.Entry.MAC); err != nil {
			return fmt.Errorf("invalid port neighbor mac %q on port %d: %w", neighbor.Entry.MAC, neighbor.Port, err)
		}
		if neighbor.Entry.VLAN < 0 {
			return fmt.Errorf("invalid port neighbor vlan %d on port %d; use 0 or a positive VLAN ID", neighbor.Entry.VLAN, neighbor.Port)
		}
		if ip := strings.TrimSpace(neighbor.Entry.IP); ip != "" && net.ParseIP(ip).To4() == nil {
			return fmt.Errorf("invalid port neighbor ip %q on port %d; use an IPv4 address", neighbor.Entry.IP, neighbor.Port)
		}
	}
	for _, override := range flags.portOverrides {
		if err := device.ValidatePortOverride(override, flags.portCount); err != nil {
			return fmt.Errorf("validate port overrides: %w", err)
		}
	}
	return nil
}

// validateSourceMappings validates bridge-observe and port-map inputs in two
// phases: structural checks first, optional live interface existence checks when
// requested by -validate or runtime startup.
func validateSourceMappings(flags runtimeFlags, live bool) error {
	var errs []error
	mode := normalizeMode(flags.operationMode)
	switch mode {
	case operationModeBridgeObserve:
		cfg := effectiveBridgeObserve(flags)
		if err := validateOptionalInterfaceName("bridge_observe.bridge", cfg.Bridge, live); err != nil {
			errs = append(errs, err)
		}
		if err := validateOptionalInterfaceName("bridge_observe.uplink_interface", cfg.UplinkInterface, live); err != nil {
			errs = append(errs, err)
		}
		for _, member := range cfg.IgnoredMembers {
			if err := validateOptionalInterfaceName("bridge_observe.ignored_members", member, live); err != nil {
				errs = append(errs, err)
			}
		}
		for _, mapping := range cfg.MemberPortMap {
			member := strings.TrimSpace(mapping.Member)
			if member == "" || strings.Contains(member, "/") {
				errs = append(errs, fmt.Errorf("invalid bridge_observe.member_port_map member %q", mapping.Member))
			}
			if mapping.Port < 1 || mapping.Port > flags.portCount {
				errs = append(errs, fmt.Errorf("invalid bridge_observe.member_port_map port %d for %q; use 1..%d", mapping.Port, member, flags.portCount))
			}
		}
		errs = append(errs, validateBridgeMemberPortMap(cfg.MemberPortMap)...)
		errs = append(errs, validateBridgeIgnoredMembers(cfg)...)
	case operationModePortMap:
		seenPorts := map[int]bool{}
		for _, mapping := range flags.portMappings {
			if mapping.Port < 1 || mapping.Port > flags.portCount {
				errs = append(errs, fmt.Errorf("invalid port_mappings port %d; use 1..%d", mapping.Port, flags.portCount))
				continue
			}
			if seenPorts[mapping.Port] {
				errs = append(errs, fmt.Errorf("duplicate port_mappings entry for port %d", mapping.Port))
			}
			seenPorts[mapping.Port] = true
			sources := 0
			if strings.TrimSpace(mapping.Interface) != "" {
				sources++
				if err := validateOptionalInterfaceName("port_mappings.interface", mapping.Interface, live); err != nil {
					errs = append(errs, fmt.Errorf("port %d: %w", mapping.Port, err))
				}
			}
			if mapping.Disabled {
				sources++
			}
			if mapping.Unmapped {
				sources++
			}
			if sources != 1 {
				errs = append(errs, fmt.Errorf("invalid port_mappings entry on port %d; set exactly one of interface, disabled, or unmapped", mapping.Port))
			}
		}
		for port := 1; port <= flags.portCount; port++ {
			if !seenPorts[port] {
				errs = append(errs, fmt.Errorf("missing port_mappings entry for port %d", port))
			}
		}
	}
	if err := errors.Join(errs...); err != nil {
		return fmt.Errorf("validate source mappings: %w", err)
	}
	return nil
}

// validateOptionalInterfaceName rejects path-like names and optionally checks
// local existence for modes that will read host interface data.
func validateOptionalInterfaceName(field, value string, live bool) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	if strings.Contains(value, "/") {
		return fmt.Errorf("invalid %s %q", field, value)
	}
	if live {
		if _, err := net.InterfaceByName(value); err != nil {
			return fmt.Errorf("%s %q not found: %w", field, value, err)
		}
	}
	return nil
}

// validateIdentityFlags checks only locally supplied identity values; controller
// adoption responses are not allowed to redefine the host-facing identity.
func validateIdentityFlags(flags runtimeFlags) error {
	if ip := net.ParseIP(strings.TrimSpace(flags.ipText)).To4(); ip == nil {
		return fmt.Errorf("invalid IPv4 address: %q", flags.ipText)
	}
	macText := strings.TrimSpace(flags.macText)
	if macText == "" || strings.EqualFold(macText, automaticText) || strings.EqualFold(macText, "host") {
		return nil
	}
	if _, err := net.ParseMAC(macText); err != nil {
		return fmt.Errorf("invalid MAC address: %w", err)
	}
	return nil
}

// normalizeMode keeps the legacy observe alias while making stub mode the
// default safety posture.
func normalizeMode(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return operationModeStub
	}
	if value == operationModeObserve {
		return operationModeBridgeObserve
	}
	return value
}

// portsForRuntime merges profile defaults, passive observations, LLDP hints,
// operator overrides, and configured neighbors into one ordered port list.
func portsForRuntime(flags runtimeFlags, portOptions device.PortOptions, plt platform.Platform) []device.Port {
	ports := device.SwitchPortsWithOptions(flags.portCount, portOptions)
	mode := normalizeMode(flags.operationMode)
	if mode == operationModePortMap {
		// Explicit port-map sources become ordinary overrides first, then user
		// overrides win. This preserves the operator's final say over observed
		// host data while keeping renderer code on one merge path.
		ctx, cancel := context.WithTimeout(context.Background(), observeTimeout)
		defer cancel()
		overrides, errs := portmap.OverridesFromSource(ctx, plt, flags.portMappings)
		for _, err := range errs {
			log.Printf("port-map observation warning: %v", err)
		}
		ports = device.ApplyPortOverrides(ports, overrides)
		ports = device.ApplyPortOverrides(ports, flags.portOverrides)
		ports = applyLLDPNeighbors(ports, flags, plt)
		ports = device.ApplyPortNeighbors(ports, flags.portNeighbors)
		return device.ApplyUplinkNeighbor(ports, flags.uplinkNeighbor)
	}
	if mode != operationModeBridgeObserve && mode != operationModeHostDirect {
		// Stub mode stays synthetic unless the operator explicitly supplies
		// payload metadata. No host bridge or interface data is guessed here.
		ports = device.ApplyPortOverrides(ports, flags.portOverrides)
		ports = device.ApplyPortNeighbors(ports, flags.portNeighbors)
		return device.ApplyUplinkNeighbor(ports, flags.uplinkNeighbor)
	}
	ctx, cancel := context.WithTimeout(context.Background(), observeTimeout)
	defer cancel()

	bridgeObserve := effectiveBridgeObserve(flags)
	snapshot, errs := observe.HostSnapshotFromSource(ctx, plt, observe.Config{
		Interface:      strings.TrimSpace(bridgeObserve.UplinkInterface),
		Bridge:         strings.TrimSpace(bridgeObserve.Bridge),
		IgnoredMembers: cloneStrings(bridgeObserve.IgnoredMembers),
		MemberPortMap:  bridgeMemberPortMap(bridgeObserve.MemberPortMap),
	}, uplinkPortIndex(ports))
	for _, err := range errs {
		log.Printf("passive observation warning: %v", err)
	}
	observedPorts := observe.Apply(ports, snapshot)
	// Bridge observation is read-only input. Operator overrides are applied
	// after the passive snapshot so a config file can correct or mask host facts
	// without the controller mutating the host.
	if flags.trafficRatesEnabled {
		observedPorts = markTrafficRateUplinkInterface(observedPorts, bridgeObserve.UplinkInterface)
	}
	ports = device.ApplyPortOverrides(observedPorts, flags.portOverrides)
	ports = applyLLDPNeighbors(ports, flags, plt)
	ports = device.ApplyPortNeighbors(ports, flags.portNeighbors)
	return device.ApplyUplinkNeighbor(ports, flags.uplinkNeighbor)
}

// applyLLDPNeighbors adds passive LLDP neighbors as MAC-table hints on the
// matching represented UniFi ports.
func applyLLDPNeighbors(ports []device.Port, flags runtimeFlags, plt platform.Platform) []device.Port {
	if strings.TrimSpace(flags.lldpSource) == "" || strings.EqualFold(strings.TrimSpace(flags.lldpSource), platform.SourceOff) {
		return ports
	}
	ctx, cancel := context.WithTimeout(context.Background(), observeTimeout)
	defer cancel()
	neighbors, errs := plt.LLDP(ctx, platform.LLDPConfig{Source: flags.lldpSource, Timeout: observeTimeout})
	for _, err := range errs {
		log.Printf("lldp observation warning: %v", err)
	}
	if len(neighbors) == 0 {
		return ports
	}
	out := make([]device.Port, len(ports))
	copy(out, ports)
	portByInterface := lldpInterfacePortMap(flags, out)
	for _, neighbor := range neighbors {
		portIndex := portByInterface[strings.ToLower(strings.TrimSpace(neighbor.Interface))]
		if portIndex < 1 || portIndex > len(out) {
			continue
		}
		entry := lldpNeighborMACEntry(neighbor)
		if strings.TrimSpace(entry.MAC) == "" {
			continue
		}
		// LLDP neighbors are represented only as controller-facing MAC-table
		// hints. They are never used to configure host networking.
		out[portIndex-1].MACs = append(out[portIndex-1].MACs, entry)
	}
	return out
}

// lldpInterfacePortMap maps observed interface names back to represented port
// indexes using explicit bridge, port-map, and override bindings.
func lldpInterfacePortMap(flags runtimeFlags, ports []device.Port) map[string]int {
	out := map[string]int{}
	bridgeObserve := effectiveBridgeObserve(flags)
	if iface := strings.ToLower(strings.TrimSpace(bridgeObserve.UplinkInterface)); iface != "" {
		out[iface] = uplinkPortIndex(ports)
	}
	for _, mapping := range bridgeObserve.MemberPortMap {
		if iface := strings.ToLower(strings.TrimSpace(mapping.Member)); iface != "" {
			out[iface] = mapping.Port
		}
	}
	for _, mapping := range flags.portMappings {
		if iface := strings.ToLower(strings.TrimSpace(mapping.Interface)); iface != "" {
			out[iface] = mapping.Port
		}
	}
	for _, override := range flags.portOverrides {
		if iface := strings.ToLower(strings.TrimSpace(override.Interface)); iface != "" {
			out[iface] = override.Port
		}
	}
	return out
}

// lldpNeighborMACEntry turns one LLDP neighbor into the same MAC-table metadata
// shape used for configured neighbors.
func lldpNeighborMACEntry(neighbor platform.LLDPNeighbor) device.MacTableEntry {
	mac := strings.TrimSpace(neighbor.ChassisMAC)
	if mac == "" {
		mac = strings.TrimSpace(neighbor.ChassisID)
	}
	if parsed, err := net.ParseMAC(mac); err == nil {
		mac = parsed.String()
	} else {
		return device.MacTableEntry{}
	}
	return device.MacTableEntry{
		MAC:      mac,
		Hostname: strings.TrimSpace(neighbor.SystemName),
		IP:       ipv4Text(neighbor.ManagementIP),
		Age:      4,
		Uptime:   1200,
		Type:     "lldp",
	}
}

// ipv4Text keeps LLDP management addresses limited to IPv4 strings accepted by
// the UniFi MAC-table payload.
func ipv4Text(value string) string {
	ip := net.ParseIP(strings.TrimSpace(value))
	if ip == nil || ip.To4() == nil {
		return ""
	}
	return ip.String()
}

// printRuntimePlan describes the non-mutating runtime plan, including actions
// that are intentionally dry-run-only for host-networking safety.
func printRuntimePlan(flags runtimeFlags, profile device.Profile, macText, ipText, hostname string) {
	mode := normalizeMode(flags.operationMode)
	fmt.Printf("operation_mode: %s\n", mode)
	fmt.Printf("profile: %s (%s)\n", profile.Name, profile.Model)
	fmt.Printf("device_type: %s\n", profile.DeviceType)
	fmt.Printf("mac: %s\n", macText)
	fmt.Printf("ip: %s\n", ipText)
	fmt.Printf("hostname: %s\n", hostname)
	fmt.Printf("uplink_port: %d\n", effectiveUplinkPort(profile, flags))
	if flags.uplinkNeighbor != nil {
		fmt.Printf("uplink_neighbor: mac=%s hostname=%q ip=%q vlan=%d type=%q\n",
			flags.uplinkNeighbor.MAC,
			strings.TrimSpace(flags.uplinkNeighbor.Hostname),
			strings.TrimSpace(flags.uplinkNeighbor.IP),
			flags.uplinkNeighbor.VLAN,
			strings.TrimSpace(flags.uplinkNeighbor.Type),
		)
	}
	for _, neighbor := range flags.portNeighbors {
		fmt.Printf("port_neighbor: port=%d mac=%s hostname=%q ip=%q vlan=%d type=%q\n",
			neighbor.Port,
			neighbor.Entry.MAC,
			strings.TrimSpace(neighbor.Entry.Hostname),
			strings.TrimSpace(neighbor.Entry.IP),
			neighbor.Entry.VLAN,
			strings.TrimSpace(neighbor.Entry.Type),
		)
	}
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
	fmt.Printf("traffic_source: %s\n", strings.TrimSpace(flags.trafficSource))
	fmt.Printf("traffic_rates_enabled: %t\n", flags.trafficRatesEnabled)
	fmt.Printf("log_source: %s\n", strings.TrimSpace(flags.logSource))
	fmt.Printf("proc_source: %s\n", strings.TrimSpace(flags.procSource))
	fmt.Printf("dbus_enabled: %t\n", flags.dbusEnabled)
	fmt.Printf("dbus_bus: %s\n", strings.TrimSpace(flags.dbusBus))
	fmt.Printf("syslog_path: %s\n", strings.TrimSpace(flags.syslogPath))
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

// uplinkPortIndex finds the represented uplink and falls back to port 1 for
// sparse or synthetic profiles.
func uplinkPortIndex(ports []device.Port) int {
	for _, port := range ports {
		if port.Uplink {
			return port.Index
		}
	}
	return 1
}

// markTrafficRateUplinkInterface preserves the bridge-observe uplink interface
// name so rate tracking has a stable key on later heartbeats.
func markTrafficRateUplinkInterface(ports []device.Port, iface string) []device.Port {
	iface = strings.TrimSpace(iface)
	if iface == "" || len(ports) == 0 {
		return ports
	}
	index := uplinkPortIndex(ports)
	if index < 1 || index > len(ports) || strings.TrimSpace(ports[index-1].Interface) != "" {
		return ports
	}
	out := make([]device.Port, len(ports))
	copy(out, ports)
	out[index-1].Interface = iface
	return out
}

// effectiveBridgeObserve merges structured bridge_observe config with the older
// observe_bridge/observe_interface flags.
func effectiveBridgeObserve(flags runtimeFlags) appconfig.BridgeObserve {
	cfg := cloneBridgeObserve(flags.bridgeObserve)
	if strings.TrimSpace(cfg.Bridge) == "" {
		cfg.Bridge = strings.TrimSpace(flags.observeBridge)
	}
	if strings.TrimSpace(cfg.UplinkInterface) == "" {
		cfg.UplinkInterface = strings.TrimSpace(flags.observeInterface)
	}
	return cfg
}

// bridgeMemberPortMap converts operator pinning into the observe package shape
// used for deterministic bridge-member assignment.
func bridgeMemberPortMap(mappings []appconfig.BridgeMemberPortMap) map[string]int {
	if len(mappings) == 0 {
		return nil
	}
	out := make(map[string]int, len(mappings))
	for _, mapping := range mappings {
		member := strings.TrimSpace(mapping.Member)
		if member != "" {
			out[member] = mapping.Port
		}
	}
	return out
}

// validateBridgeMemberPortMap rejects ambiguous bridge-member pinning before
// passive observation can assign represented UniFi ports.
func validateBridgeMemberPortMap(mappings []appconfig.BridgeMemberPortMap) []error {
	var errs []error
	seenMembers := map[string]bool{}
	seenPorts := map[int]string{}
	for _, mapping := range mappings {
		member := strings.TrimSpace(mapping.Member)
		if member == "" {
			continue
		}
		if seenMembers[member] {
			errs = append(errs, fmt.Errorf("duplicate bridge_observe.member_port_map member %q", member))
		}
		seenMembers[member] = true
		if previous := seenPorts[mapping.Port]; previous != "" {
			errs = append(errs, fmt.Errorf("duplicate bridge_observe.member_port_map port %d for %q and %q", mapping.Port, previous, member))
		}
		seenPorts[mapping.Port] = member
	}
	return errs
}

// validateBridgeIgnoredMembers catches configuration that both pins and ignores
// the same bridge member, because ignored members must never consume ports.
func validateBridgeIgnoredMembers(cfg appconfig.BridgeObserve) []error {
	var errs []error
	seen := map[string]bool{}
	pinned := map[string]bool{}
	for _, mapping := range cfg.MemberPortMap {
		if key := bridgeMemberKey(mapping.Member); key != "" {
			pinned[key] = true
		}
	}
	bridgeKey := bridgeMemberKey(cfg.Bridge)
	uplinkKey := bridgeMemberKey(cfg.UplinkInterface)
	for _, member := range cfg.IgnoredMembers {
		key := bridgeMemberKey(member)
		if key == "" {
			continue
		}
		switch {
		case seen[key]:
			errs = append(errs, fmt.Errorf("duplicate bridge_observe.ignored_members member %q", strings.TrimSpace(member)))
		case key == bridgeKey:
			errs = append(errs, fmt.Errorf("bridge_observe.ignored_members cannot ignore bridge %q", strings.TrimSpace(member)))
		case key == uplinkKey:
			errs = append(errs, fmt.Errorf("bridge_observe.ignored_members cannot ignore uplink_interface %q", strings.TrimSpace(member)))
		case pinned[key]:
			errs = append(errs, fmt.Errorf("bridge_observe.ignored_members member %q also appears in member_port_map", strings.TrimSpace(member)))
		}
		seen[key] = true
	}
	return errs
}

// bridgeMemberKey normalizes bridge member names for duplicate and ignore-list
// policy checks.
func bridgeMemberKey(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

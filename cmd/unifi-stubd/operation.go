package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/konstruktor1/unifi-stubd/internal/device"
	"github.com/konstruktor1/unifi-stubd/internal/observe"
)

const (
	operationModeStub       = "stub"
	operationModeObserve    = "observe"
	operationModeHostDirect = "host-direct"
	operationModeMacvlan    = "macvlan"

	lldpSourceOff    = "off"
	lldpSourceLLDPD  = "lldpd"
	trafficSourceOff = "off"
	observeTimeout   = 2 * time.Second

	portRoleLAN  = "lan"
	portRoleLAN2 = "lan2"
	portRoleWAN  = "wan"
	portRoleWAN2 = "wan2"
)

func validateOperationFlags(flags runtimeFlags) error {
	mode := normalizeMode(*flags.operationMode)
	*flags.operationMode = mode
	switch mode {
	case operationModeStub, operationModeObserve, operationModeHostDirect, operationModeMacvlan:
	default:
		return fmt.Errorf("invalid -operation-mode %q; use stub, observe, host-direct, or macvlan", mode)
	}

	lldpSource := strings.ToLower(strings.TrimSpace(*flags.lldpSource))
	if lldpSource == "" {
		lldpSource = lldpSourceOff
	}
	*flags.lldpSource = lldpSource
	switch lldpSource {
	case lldpSourceOff:
	case lldpSourceLLDPD:
		return fmt.Errorf("invalid -lldp-source %q; lldpd is planned but not implemented yet", lldpSource)
	default:
		return fmt.Errorf("invalid -lldp-source %q; only off is implemented", lldpSource)
	}

	trafficSource := strings.ToLower(strings.TrimSpace(*flags.trafficSource))
	if trafficSource == "" {
		trafficSource = trafficSourceOff
	}
	*flags.trafficSource = trafficSource
	if trafficSource != trafficSourceOff {
		return fmt.Errorf("invalid -traffic-source %q; only off is implemented", trafficSource)
	}

	if strings.EqualFold(strings.TrimSpace(*flags.macText), "host") && mode != operationModeHostDirect {
		return fmt.Errorf("mac: host is only allowed with -operation-mode host-direct")
	}
	if mode == operationModeMacvlan && !*flags.dryRunPlan {
		return fmt.Errorf("operation-mode macvlan is planned only; use -dry-run-plan to inspect the non-mutating plan")
	}
	if iface := strings.TrimSpace(*flags.discoveryInterface); strings.Contains(iface, "/") {
		return fmt.Errorf("invalid -discovery-interface %q", iface)
	}
	if *flags.managementVLAN < 0 || *flags.managementVLAN > 4094 {
		return fmt.Errorf("invalid -management-vlan %d; use 0..4094", *flags.managementVLAN)
	}
	return nil
}

func validatePortOverrides(flags runtimeFlags) error {
	if *flags.uplinkPort < 0 || *flags.uplinkPort > *flags.portCount {
		return fmt.Errorf("invalid -uplink-port %d; use 0 or 1..%d", *flags.uplinkPort, *flags.portCount)
	}
	if flags.uplinkNeighbor != nil {
		if _, err := net.ParseMAC(flags.uplinkNeighbor.MAC); err != nil {
			return fmt.Errorf("invalid uplink_neighbor mac %q: %w", flags.uplinkNeighbor.MAC, err)
		}
		if flags.uplinkNeighbor.VLAN < 0 {
			return fmt.Errorf("invalid uplink_neighbor vlan %d; use 0 or a positive VLAN ID", flags.uplinkNeighbor.VLAN)
		}
	}
	for _, neighbor := range flags.portNeighbors {
		if neighbor.Port < 1 || neighbor.Port > *flags.portCount {
			return fmt.Errorf("invalid port neighbor %d; use 1..%d", neighbor.Port, *flags.portCount)
		}
		if _, err := net.ParseMAC(neighbor.Entry.MAC); err != nil {
			return fmt.Errorf("invalid port neighbor mac %q on port %d: %w", neighbor.Entry.MAC, neighbor.Port, err)
		}
		if neighbor.Entry.VLAN < 0 {
			return fmt.Errorf("invalid port neighbor vlan %d on port %d; use 0 or a positive VLAN ID", neighbor.Entry.VLAN, neighbor.Port)
		}
	}
	for _, override := range flags.portOverrides {
		if override.Port < 1 || override.Port > *flags.portCount {
			return fmt.Errorf("invalid port override %d; use 1..%d", override.Port, *flags.portCount)
		}
		if strings.TrimSpace(override.MAC) != "" {
			if _, err := net.ParseMAC(override.MAC); err != nil {
				return fmt.Errorf("invalid port override mac %q on port %d: %w", override.MAC, override.Port, err)
			}
		}
		if override.Speed < 0 {
			return fmt.Errorf("invalid speed override %d on port %d; use 0 or a positive Mbps value", override.Speed, override.Port)
		}
		if iface := strings.TrimSpace(override.Interface); strings.Contains(iface, "/") {
			return fmt.Errorf("invalid interface override %q on port %d", iface, override.Port)
		}
		if ip := strings.TrimSpace(override.IP); ip != "" && net.ParseIP(ip).To4() == nil {
			return fmt.Errorf("invalid port override ip %q on port %d", ip, override.Port)
		}
		if netmask := strings.TrimSpace(override.Netmask); netmask != "" && net.ParseIP(netmask).To4() == nil {
			return fmt.Errorf("invalid port override netmask %q on port %d", netmask, override.Port)
		}
		if role := strings.ToLower(strings.TrimSpace(override.Role)); role != "" && !validPortRole(role) {
			return fmt.Errorf("invalid port override role %q on port %d; use wan, lan, wan2, or lan2", override.Role, override.Port)
		}
		if networkGroup := strings.TrimSpace(override.NetworkGroup); strings.ContainsAny(networkGroup, "\r\n\t") {
			return fmt.Errorf("invalid port override network_group %q on port %d", networkGroup, override.Port)
		}
		if override.Speed == 0 && override.Up == nil &&
			strings.TrimSpace(override.Name) == "" &&
			strings.TrimSpace(override.Interface) == "" &&
			strings.TrimSpace(override.MAC) == "" &&
			strings.TrimSpace(override.IP) == "" &&
			strings.TrimSpace(override.Netmask) == "" &&
			strings.TrimSpace(override.Role) == "" &&
			strings.TrimSpace(override.NetworkGroup) == "" &&
			strings.TrimSpace(override.Media) == "" {
			return fmt.Errorf("empty port override on port %d", override.Port)
		}
	}
	return nil
}

func validateIdentityFlags(flags runtimeFlags) error {
	if ip := net.ParseIP(strings.TrimSpace(*flags.ipText)).To4(); ip == nil {
		return fmt.Errorf("invalid IPv4 address: %q", *flags.ipText)
	}
	macText := strings.TrimSpace(*flags.macText)
	if macText == "" || strings.EqualFold(macText, automaticText) || strings.EqualFold(macText, "host") {
		return nil
	}
	if _, err := net.ParseMAC(macText); err != nil {
		return fmt.Errorf("invalid MAC address: %w", err)
	}
	return nil
}

func validPortRole(role string) bool {
	switch role {
	case portRoleWAN, portRoleLAN, portRoleWAN2, portRoleLAN2:
		return true
	default:
		return false
	}
}

func normalizeMode(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return operationModeStub
	}
	return value
}

func portsForRuntime(flags runtimeFlags, portOptions device.PortOptions) []device.Port {
	ports := device.SwitchPortsWithOptions(*flags.portCount, portOptions)
	mode := normalizeMode(*flags.operationMode)
	if mode != operationModeObserve && mode != operationModeHostDirect {
		ports = device.ApplyPortOverrides(ports, flags.portOverrides)
		ports = device.ApplyPortNeighbors(ports, flags.portNeighbors)
		return device.ApplyUplinkNeighbor(ports, flags.uplinkNeighbor)
	}
	ctx, cancel := context.WithTimeout(context.Background(), observeTimeout)
	defer cancel()

	snapshot, errs := observe.LinuxSnapshot(ctx, observe.Config{
		Interface: strings.TrimSpace(*flags.observeInterface),
		Bridge:    strings.TrimSpace(*flags.observeBridge),
	}, uplinkPortIndex(ports))
	for _, err := range errs {
		log.Printf("passive observation warning: %v", err)
	}
	ports = device.ApplyPortOverrides(observe.Apply(ports, snapshot), flags.portOverrides)
	ports = device.ApplyPortNeighbors(ports, flags.portNeighbors)
	return device.ApplyUplinkNeighbor(ports, flags.uplinkNeighbor)
}

func printRuntimePlan(flags runtimeFlags, profile device.Profile, macText, ipText, hostname string) {
	mode := normalizeMode(*flags.operationMode)
	fmt.Printf("operation_mode: %s\n", mode)
	fmt.Printf("profile: %s (%s)\n", profile.Name, profile.Model)
	fmt.Printf("device_type: %s\n", profile.DeviceType)
	fmt.Printf("mac: %s\n", macText)
	fmt.Printf("ip: %s\n", ipText)
	fmt.Printf("hostname: %s\n", hostname)
	fmt.Printf("uplink_port: %d\n", *flags.uplinkPort)
	if flags.uplinkNeighbor != nil {
		fmt.Printf("uplink_neighbor: mac=%s vlan=%d type=%q\n",
			flags.uplinkNeighbor.MAC,
			flags.uplinkNeighbor.VLAN,
			strings.TrimSpace(flags.uplinkNeighbor.Type),
		)
	}
	for _, neighbor := range flags.portNeighbors {
		fmt.Printf("port_neighbor: port=%d mac=%s vlan=%d type=%q\n",
			neighbor.Port,
			neighbor.Entry.MAC,
			neighbor.Entry.VLAN,
			strings.TrimSpace(neighbor.Entry.Type),
		)
	}
	for _, override := range flags.portOverrides {
		fmt.Printf("port_override: port=%d interface=%q mac=%q ip=%q netmask=%q role=%q network_group=%q speed=%d media=%q up=%s name=%q\n",
			override.Port,
			strings.TrimSpace(override.Interface),
			strings.TrimSpace(override.MAC),
			strings.TrimSpace(override.IP),
			strings.TrimSpace(override.Netmask),
			strings.TrimSpace(override.Role),
			strings.TrimSpace(override.NetworkGroup),
			override.Speed,
			strings.TrimSpace(override.Media),
			boolPointerText(override.Up),
			strings.TrimSpace(override.Name),
		)
	}
	fmt.Printf("observe_interface: %s\n", strings.TrimSpace(*flags.observeInterface))
	fmt.Printf("observe_bridge: %s\n", strings.TrimSpace(*flags.observeBridge))
	fmt.Printf("lldp_source: %s\n", strings.TrimSpace(*flags.lldpSource))
	fmt.Printf("traffic_source: %s\n", strings.TrimSpace(*flags.trafficSource))
	fmt.Printf("management_vlan: %d\n", *flags.managementVLAN)
	if iface := strings.TrimSpace(*flags.discoveryInterface); iface != "" {
		fmt.Printf("discovery_interface: %s\n", iface)
	}
	for _, target := range flags.discoveryTargets {
		fmt.Printf("discovery_target: %s\n", strings.TrimSpace(target))
	}
	switch mode {
	case operationModeStub:
		fmt.Println("actions: synthetic stub only; no host network changes")
	case operationModeObserve:
		fmt.Println("actions: read-only Linux sysfs/FDB observation; no host network changes")
	case operationModeHostDirect:
		fmt.Println("actions: direct host identity mode; no host network changes")
	case operationModeMacvlan:
		parent := strings.TrimSpace(*flags.observeInterface)
		if parent == "" {
			parent = "<required-parent-interface>"
		}
		fmt.Println("actions: macvlan is not executed by this release")
		fmt.Printf("planned_command: ip link add link %s name unifi-stubd0 type macvlan mode bridge\n", parent)
		fmt.Printf("planned_command: ip link set unifi-stubd0 address %s up\n", macText)
		fmt.Printf("planned_note: assign %s to unifi-stubd0 after subnet/prefix config exists\n", ipText)
	}
}

func boolPointerText(value *bool) string {
	if value == nil {
		return "unset"
	}
	return fmt.Sprintf("%t", *value)
}

func uplinkPortIndex(ports []device.Port) int {
	for _, port := range ports {
		if port.Uplink {
			return port.Index
		}
	}
	return 1
}

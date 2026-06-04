package payload

import (
	"net"
	"strconv"
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/device"
)

func gatewayPortRole(port device.Port) string {
	if role := normalizeGatewayRole(port.Role); role != "" {
		return role
	}
	switch port.Index {
	case 1:
		return gatewayPortRoleWAN
	case 2:
		return gatewayPortRoleLAN
	case 3:
		return gatewayPortRoleWAN2
	default:
		return gatewayPortRoleLAN
	}
}

func gatewayNetworkGroup(port device.Port) string {
	if networkGroup := normalizeNetworkGroup(port.NetworkGroup); networkGroup != "" {
		return networkGroup
	}
	switch gatewayPortRole(port) {
	case gatewayPortRoleWAN:
		return gatewayNetworkGroupWAN
	case gatewayPortRoleWAN2:
		return gatewayNetworkGroupWAN2
	default:
		return gatewayNetworkGroupLAN
	}
}

func normalizeGatewayRole(role string) string {
	return strings.ToLower(strings.TrimSpace(role))
}

func normalizeNetworkGroup(networkGroup string) string {
	return strings.TrimSpace(networkGroup)
}

// gatewayInterfaceName maps a one-based physical profile port index to the
// controller-facing gateway interface name. It deliberately ignores host
// interface names such as ixl0 or vtnet0; those belong in source_interface.
func gatewayInterfaceName(profile device.Profile, portIndex int) string {
	if portIndex < 1 {
		portIndex = 1
	}
	prefix := strings.TrimSpace(profile.Payload.GatewayInterfacePrefix)
	if prefix == "" {
		prefix = "eth"
	}
	return prefix + strconv.Itoa(portIndex-1)
}

// gatewayInterfaceNameForPort keeps the controller-facing interface identity
// tied to the physical profile port. Gateway remaps change Role/NetworkGroup;
// they do not rename the profile ethN identity.
func gatewayInterfaceNameForPort(profile device.Profile, port device.Port) string {
	return gatewayInterfaceName(profile, port.Index)
}

func gatewayUplinkInterfaceName(profile device.Profile, ports []PortView) string {
	for _, view := range ports {
		if view.Uplink && strings.TrimSpace(view.GatewayInterface.IfName) != "" {
			return view.GatewayInterface.IfName
		}
	}
	return gatewayInterfaceName(profile, uplinkPortIndex(ports))
}

func portMAC(baseMAC string, port device.Port) string {
	if mac := strings.TrimSpace(port.MAC); mac != "" {
		return strings.ToLower(mac)
	}
	return interfaceMAC(baseMAC, port.Index)
}

func interfaceMAC(baseMAC string, portIndex int) string {
	mac, err := net.ParseMAC(baseMAC)
	if err != nil || len(mac) == 0 {
		return baseMAC
	}
	out := append(net.HardwareAddr{}, mac...)
	out[len(out)-1] += byte(portIndex - 1)
	return out.String()
}

func interfaceIP(id device.Identity, port device.Port) string {
	if ip := strings.TrimSpace(port.IP); ip != "" {
		return ip
	}
	// Gateway WAN fallbacks use documentation addresses so payload examples do
	// not leak or invent real lab network data.
	switch gatewayPortRole(port) {
	case gatewayPortRoleLAN, gatewayPortRoleLAN2:
		return id.IP
	case gatewayPortRoleWAN, gatewayPortRoleWAN2:
		return "192.0.2.2"
	}
	return gatewayNoIP
}

func interfaceNetmask(port device.Port) string {
	if netmask := strings.TrimSpace(port.Netmask); netmask != "" {
		return netmask
	}
	return "255.255.255.0"
}

func interfaceAddressCIDR(ip, netmask string) string {
	prefix := netmaskPrefixLength(netmask)
	if prefix < 0 {
		prefix = 24
	}
	return strings.TrimSpace(ip) + "/" + strconv.Itoa(prefix)
}

func netmaskPrefixLength(netmask string) int {
	parsed := net.ParseIP(strings.TrimSpace(netmask)).To4()
	if parsed == nil {
		return -1
	}
	ones, bits := net.IPMask(parsed).Size()
	if bits != 32 {
		return -1
	}
	return ones
}

func boolText(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

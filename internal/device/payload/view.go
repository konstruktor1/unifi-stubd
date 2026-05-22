// Package payload resolves profile ports into a renderer-neutral view before
// switch or gateway tables are encoded. This keeps physical port data, host
// observation data, and controller payload shape separated.
package payload

import "strings"

// InterfaceView is the resolved controller-facing state of one interface.
type InterfaceView struct {
	Name            string
	IfName          string
	Comment         string
	MAC             string
	IP              string
	Netmask         string
	Address         string
	NetworkGroup    string
	SourceInterface string
	PortIndex       int
	Up              bool
	Speed           int
	Media           string
}

// PortView is the canonical resolved port state shared by all payload
// renderers. It is derived from profile data plus safe runtime observations.
type PortView struct {
	Port                Port
	Index               int
	Name                string
	SwitchInterfaceName string
	Role                string
	NetworkGroup        string
	SourceInterface     string
	Uplink              bool
	ProfileUplink       bool
	Enabled             bool
	Up                  bool
	Speed               int
	Media               string
	MACs                []MacTableEntry
	GatewayInterface    InterfaceView
}

// BuildPortViews resolves profile ports once so switch and gateway renderers
// cannot drift when roles, speed, source interfaces, or management metadata
// change.
func BuildPortViews(profile Profile, id Identity, ports []Port) []PortView {
	views := make([]PortView, 0, len(ports))
	for _, port := range ports {
		// Resolve all profile, override, and observation data once. Switch and
		// gateway renderers then consume the same view, which keeps MACs, speed,
		// role, source-interface, and management metadata aligned.
		speed := effectivePortSpeed(port)
		media := effectivePortMedia(port, speed)
		role := gatewayPortRole(port)
		networkGroup := gatewayNetworkGroup(port)
		ip := gatewayInterfaceIP(id, port)
		netmask := gatewayInterfaceNetmask(port)
		gatewayName := gatewayInterfaceName(profile, port.Index)
		sourceInterface := strings.TrimSpace(port.Interface)
		enabled := !port.Disabled
		macs := append([]MacTableEntry(nil), port.MACs...)
		view := PortView{
			Port:                port,
			Index:               port.Index,
			Name:                port.Name,
			SwitchInterfaceName: switchInterfaceName(port.Index),
			Role:                role,
			NetworkGroup:        networkGroup,
			SourceInterface:     sourceInterface,
			Uplink:              port.Uplink,
			ProfileUplink:       port.ProfileUplink,
			Enabled:             enabled,
			Up:                  port.Up,
			Speed:               speed,
			Media:               media,
			MACs:                macs,
			GatewayInterface: InterfaceView{
				Name:            gatewayName,
				IfName:          gatewayName,
				Comment:         port.Name,
				MAC:             gatewayPortMAC(id.MAC, port),
				IP:              ip,
				Netmask:         netmask,
				Address:         interfaceAddressCIDR(ip, netmask),
				NetworkGroup:    networkGroup,
				SourceInterface: sourceInterface,
				PortIndex:       port.Index,
				Up:              port.Up,
				Speed:           speed,
				Media:           media,
			},
		}
		views = append(views, view)
	}
	return views
}

// sourceFields keeps the host interface provenance visible in generated payload
// rows without implying controller ownership of that interface.
func sourceFields(sourceInterface string) map[string]any {
	return map[string]any{
		jsonKeySourceIf: strings.TrimSpace(sourceInterface),
	}
}

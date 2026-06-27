package opnsense

import (
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/device"
)

// InterfaceStatus is the normalized read-only state of one OPNsense interface.
type InterfaceStatus struct {
	Interface string
	Name      string
	MAC       string
	IP        string
	Netmask   string
	IPv6      []string
	Up        *bool
	SpeedMbps int
	Media     string
}

// GatewayStatus is the normalized state of one OPNsense gateway monitor row.
type GatewayStatus struct {
	Interface string
	Name      string
	Online    *bool
	LatencyMS int
}

// DecodeInterfaces accepts common OPNsense overview response shapes.
func DecodeInterfaces(raw any) map[string]InterfaceStatus {
	out := map[string]InterfaceStatus{}
	for key, value := range interfaceObjects(raw) {
		status := decodeInterfaceObject(key, value)
		addInterfaceStatus(out, status)
	}
	return out
}

// DecodeInterface accepts common OPNsense single-interface response shapes.
func DecodeInterface(raw any, requested string) InterfaceStatus {
	objects := interfaceObjects(raw)
	if len(objects) == 0 {
		return InterfaceStatus{Interface: strings.TrimSpace(requested)}
	}
	for key, value := range objects {
		status := decodeInterfaceObject(key, value)
		if status.Interface == "" {
			status.Interface = strings.TrimSpace(requested)
		}
		return status
	}
	return InterfaceStatus{Interface: strings.TrimSpace(requested)}
}

// DecodeGatewayStatuses accepts common OPNsense gateway status response shapes.
func DecodeGatewayStatuses(raw any) map[string]GatewayStatus {
	out := map[string]GatewayStatus{}
	for key, value := range interfaceObjects(raw) {
		m, ok := value.(map[string]any)
		if !ok {
			continue
		}
		status := GatewayStatus{
			Interface: firstString(m, "interface", "if", "ifname", "device", "name"),
			Name:      firstString(m, "name", "descr", "description", "gateway"),
			Online:    parseOnline(firstString(m, "status", "monitor", "online", "up", "gateway_status")),
			LatencyMS: firstInt(m, "latency", "latency_ms", "delay", "rtt", "avgdelay"),
		}
		if status.Interface == "" {
			status.Interface = strings.TrimSpace(key)
		}
		addGatewayStatus(out, status)
	}
	return out
}

// OverridesFromState maps source config and observed OPNsense facts to port overrides.
func OverridesFromState(mappings []InterfaceMapping, interfaces map[string]InterfaceStatus, gateways map[string]GatewayStatus) []device.PortOverride {
	out := make([]device.PortOverride, 0, len(mappings))
	for _, mapping := range mappings {
		status := lookupInterfaceStatus(interfaces, mapping.Interface)
		override := overrideFromMapping(mapping, status)
		if gateway, ok := lookupGatewayStatus(gateways, mapping.Interface); ok {
			applyGatewayStatus(&override, gateway)
		}
		out = append(out, override)
	}
	return out
}

// MergeOverrides overlays base overrides onto generated overrides, with base values winning.
func MergeOverrides(generated, base []device.PortOverride) []device.PortOverride {
	merged := map[int]device.PortOverride{}
	for _, override := range generated {
		if override.Port < 1 {
			continue
		}
		merged[override.Port] = override
	}
	for _, override := range base {
		if override.Port < 1 {
			continue
		}
		current := merged[override.Port]
		if current.Port == 0 {
			current.Port = override.Port
		}
		overlayOverride(&current, override)
		merged[override.Port] = current
	}
	ports := make([]int, 0, len(merged))
	for port := range merged {
		ports = append(ports, port)
	}
	sort.Ints(ports)
	out := make([]device.PortOverride, 0, len(ports))
	for _, port := range ports {
		out = append(out, merged[port])
	}
	return out
}

func overrideFromMapping(mapping InterfaceMapping, status InterfaceStatus) device.PortOverride {
	override := device.PortOverride{
		Port:                mapping.Port,
		Name:                mapping.Name,
		Interface:           mapping.Interface,
		Role:                mapping.Role,
		NetworkGroup:        mapping.NetworkGroup,
		PortConfID:          mapping.PortConfID,
		NetworkConfID:       mapping.NetworkConfID,
		NativeNetworkConfID: mapping.NativeNetworkConfID,
		NetworkName:         mapping.NetworkName,
		VLAN:                mapping.VLAN,
		Speed:               mapping.Speed,
		Media:               mapping.Media,
	}
	if override.Interface == "" {
		override.Interface = status.Interface
	}
	if override.MAC == "" {
		override.MAC = status.MAC
	}
	if override.IP == "" {
		override.IP = status.IP
	}
	if override.Netmask == "" {
		override.Netmask = status.Netmask
	}
	if len(override.IPv6) == 0 {
		override.IPv6 = append([]string(nil), status.IPv6...)
	}
	if override.Up == nil {
		override.Up = cloneBool(status.Up)
	}
	if override.Speed <= 0 {
		override.Speed = status.SpeedMbps
	}
	if override.Media == "" {
		override.Media = status.Media
	}
	return override
}

func applyGatewayStatus(override *device.PortOverride, status GatewayStatus) {
	if !isWANRole(override.Role) {
		return
	}
	if override.WANConnected == nil {
		override.WANConnected = cloneBool(status.Online)
	}
	if override.WANLatencyMS == 0 {
		override.WANLatencyMS = status.LatencyMS
	}
	if override.WANUptimePercent == nil && status.Online != nil {
		uptime := 0.0
		if *status.Online {
			uptime = 100.0
		}
		override.WANUptimePercent = &uptime
	}
}

func overlayOverride(target *device.PortOverride, source device.PortOverride) {
	if source.Name != "" {
		target.Name = source.Name
	}
	if source.Interface != "" {
		target.Interface = source.Interface
	}
	if source.MAC != "" {
		target.MAC = source.MAC
	}
	if source.IP != "" {
		target.IP = source.IP
	}
	if source.Netmask != "" {
		target.Netmask = source.Netmask
	}
	if len(source.IPv6) > 0 {
		target.IPv6 = append([]string(nil), source.IPv6...)
	}
	if source.Role != "" {
		target.Role = source.Role
	}
	if source.NetworkGroup != "" {
		target.NetworkGroup = source.NetworkGroup
	}
	if source.PortConfID != "" {
		target.PortConfID = source.PortConfID
	}
	if source.NetworkConfID != "" {
		target.NetworkConfID = source.NetworkConfID
	}
	if source.NativeNetworkConfID != "" {
		target.NativeNetworkConfID = source.NativeNetworkConfID
	}
	if source.NetworkName != "" {
		target.NetworkName = source.NetworkName
	}
	if source.VLAN > 0 {
		target.VLAN = source.VLAN
	}
	if source.WANUptimePercent != nil {
		target.WANUptimePercent = source.WANUptimePercent
	}
	if source.WANLatencyMS > 0 {
		target.WANLatencyMS = source.WANLatencyMS
	}
	if source.WANDowntimeSeconds > 0 {
		target.WANDowntimeSeconds = source.WANDowntimeSeconds
	}
	if source.WANConnected != nil {
		target.WANConnected = source.WANConnected
	}
	if source.Speed > 0 {
		target.Speed = source.Speed
	}
	if source.Media != "" {
		target.Media = source.Media
	}
	if source.Up != nil {
		target.Up = source.Up
	}
}

func interfaceObjects(raw any) map[string]any {
	switch value := raw.(type) {
	case map[string]any:
		for _, key := range []string{"interfaces", "rows", "items", "data", "response", "message", "result"} {
			if nested, ok := value[key]; ok {
				objects := interfaceObjects(nested)
				if len(objects) > 0 {
					return objects
				}
			}
		}
		if looksLikeInterface(value) {
			return map[string]any{"": value}
		}
		out := map[string]any{}
		for key, nested := range value {
			if nestedMap, ok := nested.(map[string]any); ok {
				out[key] = nestedMap
			}
		}
		return out
	case []any:
		out := map[string]any{}
		for index, nested := range value {
			key := fmt.Sprintf("%d", index)
			if nestedMap, ok := nested.(map[string]any); ok {
				if name := firstString(nestedMap, "interface", "if", "ifname", "device", "name"); name != "" {
					key = name
				}
				out[key] = nestedMap
			}
		}
		return out
	default:
		return nil
	}
}

func decodeInterfaceObject(key string, raw any) InterfaceStatus {
	m, ok := raw.(map[string]any)
	if !ok {
		return InterfaceStatus{}
	}
	status := InterfaceStatus{
		Interface: firstString(m, "interface", "if", "ifname", "device", "identifier", "name"),
		Name:      firstString(m, "descr", "description", "label", "name"),
		MAC:       normalizeMAC(firstString(m, "mac", "macaddr", "mac_address", "hwaddr")),
		IP:        firstAddress(m, "ip", "ipaddr", "ipv4", "addr4", "address"),
		Netmask:   firstNetmask(m, "netmask", "subnet", "subnet_bits", "prefix", "cidr"),
		IPv6:      firstIPv6(m, "ipv6", "ipaddrv6", "addr6", "addresses6"),
		Up:        parseUp(firstString(m, "status", "link_state", "up", "enabled", "enable", "carrier")),
		SpeedMbps: firstInt(m, "speed", "speed_mbps", "media_speed"),
		Media:     firstString(m, "media", "media_type"),
	}
	if status.Interface == "" {
		status.Interface = strings.TrimSpace(key)
	}
	if status.SpeedMbps <= 0 {
		status.SpeedMbps = speedFromText(firstString(m, "media", "media_type", "status"))
	}
	status.Media = mediaFor(status.Media, status.SpeedMbps)
	if status.Netmask == "" {
		status.Netmask = netmaskFromAddress(status.IP)
		status.IP = trimCIDR(status.IP)
	}
	return status
}

func looksLikeInterface(m map[string]any) bool {
	for _, key := range []string{"interface", "if", "ifname", "device", "mac", "macaddr", "ipaddr", "ipv4", "addr4"} {
		if _, ok := m[key]; ok {
			return true
		}
	}
	return false
}

func addInterfaceStatus(out map[string]InterfaceStatus, status InterfaceStatus) {
	if status.Interface == "" {
		return
	}
	out[strings.ToLower(status.Interface)] = status
	if status.Name != "" {
		out[strings.ToLower(status.Name)] = status
	}
}

func addGatewayStatus(out map[string]GatewayStatus, status GatewayStatus) {
	if status.Interface == "" && status.Name == "" {
		return
	}
	if status.Interface != "" {
		out[strings.ToLower(status.Interface)] = status
	}
	if status.Name != "" {
		out[strings.ToLower(status.Name)] = status
	}
}

func lookupInterfaceStatus(values map[string]InterfaceStatus, name string) InterfaceStatus {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return InterfaceStatus{}
	}
	if value, ok := values[name]; ok {
		return value
	}
	for _, value := range values {
		if strings.EqualFold(value.Interface, name) || strings.EqualFold(value.Name, name) {
			return value
		}
	}
	return InterfaceStatus{Interface: name}
}

func lookupGatewayStatus(values map[string]GatewayStatus, name string) (GatewayStatus, bool) {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return GatewayStatus{}, false
	}
	if value, ok := values[name]; ok {
		return value, true
	}
	for _, value := range values {
		if strings.EqualFold(value.Interface, name) || strings.EqualFold(value.Name, name) {
			return value, true
		}
	}
	return GatewayStatus{}, false
}

func firstString(m map[string]any, keys ...string) string {
	for _, key := range keys {
		value, ok := m[key]
		if !ok {
			continue
		}
		switch typed := value.(type) {
		case string:
			if trimmed := strings.TrimSpace(typed); trimmed != "" {
				return trimmed
			}
		case float64:
			return strconv.Itoa(int(typed))
		case bool:
			return strconv.FormatBool(typed)
		}
	}
	return ""
}

func firstInt(m map[string]any, keys ...string) int {
	for _, key := range keys {
		value, ok := m[key]
		if !ok {
			continue
		}
		switch typed := value.(type) {
		case float64:
			return int(typed)
		case int:
			return typed
		case string:
			if number := firstNumber(typed); number > 0 {
				return number
			}
		}
	}
	return 0
}

func firstAddress(m map[string]any, keys ...string) string {
	for _, key := range keys {
		value := firstString(m, key)
		if value == "" {
			continue
		}
		if strings.Contains(value, ",") {
			value = strings.TrimSpace(strings.Split(value, ",")[0])
		}
		return value
	}
	return ""
}

func firstNetmask(m map[string]any, keys ...string) string {
	for _, key := range keys {
		value := firstString(m, key)
		if value == "" {
			continue
		}
		if strings.Contains(value, ".") {
			return value
		}
		prefix, err := strconv.Atoi(strings.Trim(value, "/ "))
		if err == nil {
			return prefixToNetmask(prefix)
		}
	}
	return ""
}

func firstIPv6(m map[string]any, keys ...string) []string {
	var out []string
	for _, key := range keys {
		value, ok := m[key]
		if !ok {
			continue
		}
		switch typed := value.(type) {
		case []any:
			for _, item := range typed {
				if text, ok := item.(string); ok {
					out = appendIPv6(out, text)
				}
			}
		case string:
			for _, part := range strings.Split(typed, ",") {
				out = appendIPv6(out, part)
			}
		}
	}
	return out
}

func appendIPv6(out []string, value string) []string {
	value = strings.TrimSpace(value)
	if value == "" || strings.HasPrefix(strings.ToLower(value), "fe80:") {
		return out
	}
	if strings.Contains(value, ":") {
		return append(out, value)
	}
	return out
}

func parseUp(value string) *bool {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return nil
	}
	switch value {
	case "1", "true", "up", "active", "online", "ok", "connected":
		up := true
		return &up
	case "0", "false", "down", "inactive", "offline", "no carrier", "disconnected":
		up := false
		return &up
	default:
		if strings.Contains(value, "active") || strings.Contains(value, "up") {
			up := true
			return &up
		}
		if strings.Contains(value, "down") || strings.Contains(value, "offline") {
			up := false
			return &up
		}
		return nil
	}
}

func parseOnline(value string) *bool {
	return parseUp(value)
}

func normalizeMAC(value string) string {
	if mac, err := net.ParseMAC(strings.TrimSpace(value)); err == nil {
		return strings.ToLower(mac.String())
	}
	return strings.ToLower(strings.TrimSpace(value))
}

func mediaFor(value string, speed int) string {
	value = strings.TrimSpace(value)
	lower := strings.ToLower(value)
	switch {
	case strings.Contains(lower, "sfp28"), strings.Contains(lower, "25gbase"):
		return "SFP28"
	case strings.Contains(lower, "sfp+"), strings.Contains(lower, "10gbase-sr"),
		strings.Contains(lower, "10gbase-lr"), strings.Contains(lower, "10gbase-cr"):
		return "SFP+"
	case speed >= 25000:
		return "SFP28"
	case speed >= 10000 && strings.Contains(lower, "sfp"):
		return "SFP+"
	case speed > 0:
		return "GE"
	default:
		return value
	}
}

func speedFromText(value string) int {
	lower := strings.ToLower(value)
	switch {
	case strings.Contains(lower, "25g"):
		return 25000
	case strings.Contains(lower, "10g"):
		return 10000
	case strings.Contains(lower, "5g"):
		return 5000
	case strings.Contains(lower, "2.5g"), strings.Contains(lower, "2500base"):
		return 2500
	case strings.Contains(lower, "1000base"), strings.Contains(lower, "1g"):
		return 1000
	case strings.Contains(lower, "100base"):
		return 100
	case strings.Contains(lower, "10base"):
		return 10
	default:
		return firstNumber(value)
	}
}

func firstNumber(value string) int {
	fields := strings.FieldsFunc(value, func(r rune) bool {
		return r < '0' || r > '9'
	})
	for _, field := range fields {
		number, err := strconv.Atoi(field)
		if err == nil && number > 0 {
			return number
		}
	}
	return 0
}

func netmaskFromAddress(value string) string {
	if !strings.Contains(value, "/") {
		return ""
	}
	_, network, err := net.ParseCIDR(value)
	if err != nil {
		return ""
	}
	ones, bits := network.Mask.Size()
	if bits != 32 {
		return ""
	}
	return prefixToNetmask(ones)
}

func trimCIDR(value string) string {
	if strings.Contains(value, "/") {
		return strings.Split(value, "/")[0]
	}
	return value
}

func prefixToNetmask(prefix int) string {
	if prefix < 0 || prefix > 32 {
		return ""
	}
	mask := net.CIDRMask(prefix, 32)
	return net.IP(mask).String()
}

func cloneBool(value *bool) *bool {
	if value == nil {
		return nil
	}
	clone := *value
	return &clone
}

func isWANRole(role string) bool {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "wan", "wan2":
		return true
	default:
		return false
	}
}

package platform

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/adapters/freebsdifconfig"
	"github.com/konstruktor1/unifi-stubd/internal/adapters/linuxbridge"
	"github.com/konstruktor1/unifi-stubd/internal/device"
	"github.com/konstruktor1/unifi-stubd/internal/observe"
)

// Bridge dispatches bridge observation to the OS-specific read-only adapter and
// returns a portable observation plus warnings instead of mutating the host.
func (p hostPlatform) Bridge(ctx context.Context, cfg observe.BridgeConfig) (observe.BridgeObservation, []error) {
	switch p.goos {
	case goosLinux:
		return p.linuxBridge(ctx, cfg)
	case goosFreeBSD:
		return p.freebsdBridge(ctx, cfg)
	default:
		return observe.BridgeObservation{
			Bridge:          strings.TrimSpace(cfg.Bridge),
			UplinkInterface: strings.TrimSpace(cfg.UplinkInterface),
			MemberPortMap:   cloneMemberPortMap(cfg.MemberPortMap),
		}, []error{fmt.Errorf("bridge observation is not implemented on %s", p.goos)}
	}
}

// linuxBridge reads Linux bridge FDB rows, ARP metadata, member roles, and
// optional per-member interface observations without applying any settings.
func (p hostPlatform) linuxBridge(ctx context.Context, cfg observe.BridgeConfig) (observe.BridgeObservation, []error) {
	observation := observe.BridgeObservation{
		Bridge:          strings.TrimSpace(cfg.Bridge),
		UplinkInterface: strings.TrimSpace(cfg.UplinkInterface),
		MemberPortMap:   cloneMemberPortMap(cfg.MemberPortMap),
	}
	var errs []error
	if observation.UplinkInterface != "" {
		uplink, warnings := p.Interface(ctx, observation.UplinkInterface)
		observation.Uplink = uplink
		errs = append(errs, warnings...)
	}
	if observation.Bridge != "" {
		entries, err := p.linuxBridgeFDB(ctx, observation.Bridge)
		if err != nil {
			errs = append(errs, err)
		} else {
			observation.MemberMACs = observe.MACEntriesByDevice(entries)
			if err := enrichMACsFromLocalARP(observation.MemberMACs); err != nil {
				errs = append(errs, err)
			}
			observation.MemberRoles = observe.ClassifyMembersWithIgnores(observation.MemberMACs, observation.Bridge, observation.UplinkInterface, cfg.IgnoredMembers)
			observation.RemoteMACs = observe.RemoteMACsByBridgeMember(observation.MemberMACs, observation.MemberRoles, observation.UplinkInterface, observation.Bridge)
			observation.MemberPorts, errs = p.bridgeMemberObservations(ctx, observation.MemberMACs, observation.MemberRoles, errs)
		}
	}
	return observation, errs
}

// freebsdBridge reads FreeBSD bridge forwarding rows through ifconfig and then
// uses the same portable role and member-observation model as Linux.
func (p hostPlatform) freebsdBridge(ctx context.Context, cfg observe.BridgeConfig) (observe.BridgeObservation, []error) {
	observation := observe.BridgeObservation{
		Bridge:          strings.TrimSpace(cfg.Bridge),
		UplinkInterface: strings.TrimSpace(cfg.UplinkInterface),
		MemberPortMap:   cloneMemberPortMap(cfg.MemberPortMap),
	}
	var errs []error
	if observation.Bridge != "" {
		entries, err := p.freeBSDBridgeAddr(ctx, observation.Bridge)
		if err != nil {
			errs = append(errs, err)
		} else {
			observation.MemberMACs = observe.FreeBSDMACsByInterface(entries)
			observation.MemberRoles = observe.ClassifyMembersWithIgnores(observation.MemberMACs, observation.Bridge, observation.UplinkInterface, cfg.IgnoredMembers)
			observation.RemoteMACs = observe.RemoteMACsByBridgeMember(observation.MemberMACs, observation.MemberRoles, observation.UplinkInterface, observation.Bridge)
			observation.MemberPorts, errs = p.bridgeMemberObservations(ctx, observation.MemberMACs, observation.MemberRoles, errs)
		}
	}
	if observation.UplinkInterface != "" {
		uplink, warnings := p.Interface(ctx, observation.UplinkInterface)
		observation.Uplink = uplink
		errs = append(errs, warnings...)
	}
	return observation, errs
}

// bridgeMemberObservations reads interface state for eligible bridge members
// while excluding bridge metadata and explicitly ignored members.
func (p hostPlatform) bridgeMemberObservations(ctx context.Context, memberMACs map[string][]device.MacTableEntry, roles map[string]observe.BridgeMemberRole, errs []error) (map[string]observe.PortObservation, []error) {
	if len(memberMACs) == 0 {
		return nil, errs
	}
	out := map[string]observe.PortObservation{}
	for member := range memberMACs {
		role := roleForMember(roles, member)
		if role == observe.BridgeMemberRoleBridge || role == observe.BridgeMemberRoleIgnored {
			continue
		}
		observation, warnings := p.Interface(ctx, member)
		if strings.TrimSpace(observation.Interface) != "" || observation.SpeedMbps > 0 || observation.Up != nil {
			out[member] = observation
		}
		for _, warning := range warnings {
			errs = append(errs, fmt.Errorf("bridge member %s: %w", member, warning))
		}
	}
	if len(out) == 0 {
		return nil, errs
	}
	return out, errs
}

// roleForMember resolves bridge roles case-insensitively before platform
// member observations are read.
func roleForMember(roles map[string]observe.BridgeMemberRole, member string) observe.BridgeMemberRole {
	if len(roles) == 0 {
		return observe.BridgeMemberRoleUnknown
	}
	if role, ok := roles[strings.TrimSpace(member)]; ok {
		return role
	}
	lower := strings.ToLower(strings.TrimSpace(member))
	for name, role := range roles {
		if strings.ToLower(strings.TrimSpace(name)) == lower {
			return role
		}
	}
	return observe.BridgeMemberRoleUnknown
}

// cloneMemberPortMap detaches bridge-member pinning maps from caller-owned
// config.
func cloneMemberPortMap(values map[string]int) map[string]int {
	if len(values) == 0 {
		return nil
	}
	out := make(map[string]int, len(values))
	for member, port := range values {
		member = strings.TrimSpace(member)
		if member == "" {
			continue
		}
		out[member] = port
	}
	return out
}

func (p hostPlatform) linuxBridgeFDB(ctx context.Context, bridge string) ([]linuxbridge.FDBEntry, error) {
	bridge = strings.TrimSpace(bridge)
	if bridge == "" || strings.Contains(bridge, "/") {
		return nil, fmt.Errorf("invalid bridge name %q", bridge)
	}
	out, err := commandContext(ctx, p.cfg.CommandTimeout, "bridge", "fdb", "show", "br", bridge)
	if err != nil {
		return nil, fmt.Errorf("read bridge FDB for %s: %w", bridge, err)
	}
	return linuxbridge.ParseFDB(strings.NewReader(string(out))), nil
}

func (p hostPlatform) freeBSDBridgeAddr(ctx context.Context, bridge string) ([]freebsdifconfig.BridgeAddress, error) {
	bridge = strings.TrimSpace(bridge)
	if bridge == "" || strings.Contains(bridge, "/") {
		return nil, fmt.Errorf("invalid bridge name %q", bridge)
	}
	out, err := commandContext(ctx, p.cfg.CommandTimeout, "ifconfig", bridge, "addr")
	if err != nil {
		return nil, fmt.Errorf("read FreeBSD bridge addresses for %s: %w", bridge, err)
	}
	return freebsdifconfig.ParseBridgeAddr(strings.NewReader(string(out))), nil
}

func enrichMACsFromLocalARP(memberMACs map[string][]device.MacTableEntry) error {
	if len(memberMACs) == 0 {
		return nil
	}
	file, err := os.Open("/proc/net/arp")
	if err != nil {
		return fmt.Errorf("read ARP table /proc/net/arp: %w", err)
	}
	defer func() { _ = file.Close() }()
	observe.EnrichMACsFromARP(memberMACs, observe.ParseARPTable(file))
	return nil
}

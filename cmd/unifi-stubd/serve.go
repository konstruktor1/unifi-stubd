// serveSwitchEmulation orchestrates the daemon lifecycle after configuration is
// resolved. Protocol encoding, profile rendering, observation, and SSH handling
// stay in lower-level packages; this layer wires them together.
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/konstruktor1/unifi-stubd/internal/adoption"
	"github.com/konstruktor1/unifi-stubd/internal/discovery"
	"github.com/konstruktor1/unifi-stubd/internal/observe"
)

// serveSwitchEmulation applies configuration, validates the selected runtime
// mode, builds the initial controller-facing identity, and then either exits
// for one-shot commands or enters the presence loop.
func serveSwitchEmulation() error {
	flags, changed := parseServeFlags()
	if handled, err := handleServeEarlyExit(flags); handled || err != nil {
		return err
	}
	prepared, done, err := prepareServeRuntime(&flags, changed)
	if done || err != nil {
		return err
	}
	identity, err := resolveServeIdentity(flags, prepared.profile, prepared.platform)
	if err != nil {
		return err
	}

	if flags.dryRunPlan {
		printRuntimePlan(flags, prepared.profile, identity.mac.String(), identity.ip.String(), identity.hostname)
		return nil
	}
	if flags.status || flags.statusJSON {
		return printLocalStatus(flags, prepared.profile, identity.mac, identity.ip, identity.hostname, identity.portBuildOptions, prepared.platform)
	}

	ann := discovery.Announcement{
		MAC:      identity.mac,
		IP:       identity.ip,
		Model:    flags.model,
		Version:  flags.version,
		Hostname: identity.hostname,
		Default:  true,
		Uptime:   1,
		Sequence: 1,
	}

	packet, err := ann.MarshalBinary()
	if err != nil {
		return fmt.Errorf("marshal discovery announcement: %w", err)
	}

	ports := portsForRuntime(flags, prepared.profile, identity.portBuildOptions, prepared.platform)
	if flags.trafficRatesEnabled {
		ports = applyTrafficRates(ports, observe.NewTrafficRateTracker(), time.Now())
	}
	payload, err := payloadForIdentity(identity.mac, identity.ip, identity.hostname, flags.controller, adoption.Store{}, flags, prepared.profile, ports, 1)
	if err != nil {
		return err
	}

	if flags.dryRun {
		printDryRun(packet, payload)
		return nil
	}

	instanceGuard, err := acquireInstanceGuard(flags, prepared.profile, identity.mac.String(), identity.hostname)
	if err != nil {
		return err
	}
	defer func() {
		if err := instanceGuard.Close(); err != nil {
			log.Printf("instance guard close failed: %v", err)
		}
	}()

	sshServer, err := startAdoptionSSH(flags, identity.mac, identity.ip, identity.hostname)
	if err != nil {
		return err
	}
	defer func() {
		if err := sshServer.Close(); err != nil {
			log.Printf("adoption ssh close failed: %v", err)
		}
	}()

	return maintainControllerPresence(controllerPresence{
		flags:              flags,
		profile:            prepared.profile,
		mac:                identity.mac,
		ip:                 identity.ip,
		hostname:           identity.hostname,
		portBuildOptions:   identity.portBuildOptions,
		announcement:       ann,
		discoveryPacket:    packet,
		discoverySkipped:   flags.noDiscovery,
		discoveryInterface: effectiveDiscoveryInterface(flags),
		discoveryTargets:   flags.discoveryTargets,
		startedAt:          time.Now(),
	})
}

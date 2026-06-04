package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/konstruktor1/unifi-stubd/internal/observe"
)

// maintainControllerPresence rebuilds discovery and inform state on every
// interval so adoption changes, observed ports, and runtime counters can be
// reflected without restarting the daemon.
func maintainControllerPresence(cfg controllerPresence) error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	ticker := time.NewTicker(cfg.flags.interval)
	defer ticker.Stop()

	packet := cfg.discoveryPacket
	ann := cfg.announcement
	startedAt := cfg.startedAt
	if startedAt.IsZero() {
		startedAt = time.Now()
	}
	var rateTracker *observe.TrafficRateTracker
	if cfg.flags.trafficRatesEnabled {
		rateTracker = observe.NewTrafficRateTracker()
	}

	for {
		uptimeSeconds := runtimeUptime(startedAt)
		store := loadAdoptionState(cfg.flags.sshState)
		informURL := effectiveInformURL(cfg.flags.controller, store)
		flags := cfg.flags
		plt := runtimePlatform(flags)
		if flags.trafficRatesEnabled {
			// Rate reporting is based on current observed counters, so enrich
			// interface-backed overrides on every heartbeat instead of only at
			// startup.
			ctx, cancel := context.WithTimeout(context.Background(), observeTimeout)
			flags.portOverrides = enrichOverrides(ctx, plt, flags.portOverrides)
			cancel()
		}
		ports := portsForRuntime(flags, cfg.profile, cfg.portBuildOptions, plt)
		if flags.trafficRatesEnabled {
			ports = applyTrafficRates(ports, rateTracker, time.Now())
		}
		payload, err := payloadForIdentity(cfg.mac, cfg.ip, cfg.hostname, informURL, store, cfg.flags, cfg.profile, ports, uptimeSeconds)
		if err != nil {
			return err
		}

		sendDiscovery(packet, cfg.hostname, cfg.mac, cfg.discoverySkipped, cfg.discoveryInterface, cfg.discoveryTargets)
		sendInformHeartbeat(cfg.mac, informURL, cfg.flags.sshState, cfg.flags.statusPath, store, payload, informSourceIP(cfg.flags, cfg.ip))

		if cfg.flags.once {
			return nil
		}

		select {
		case <-ticker.C:
			ann.Uptime += uint32(cfg.flags.interval.Seconds())
			ann.Sequence++
			packet, err = ann.MarshalBinary()
			if err != nil {
				return fmt.Errorf("marshal discovery announcement: %w", err)
			}
		case <-stop:
			log.Println("stopping")
			return nil
		}
	}
}

// runtimeUptime converts daemon lifetime into the positive seconds expected by
// UniFi inform payloads.
func runtimeUptime(startedAt time.Time) int {
	if startedAt.IsZero() {
		return 1
	}
	uptime := int(time.Since(startedAt).Seconds()) + 1
	if uptime < 1 {
		return 1
	}
	return uptime
}

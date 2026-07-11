package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/konstruktor1/unifi-stubd/internal/adoption"
	"github.com/konstruktor1/unifi-stubd/internal/inform"
	"github.com/konstruktor1/unifi-stubd/internal/observe"
)

// maintainControllerPresence rebuilds discovery and inform state on every
// interval so adoption changes, observed ports, and runtime counters can be
// reflected without restarting the daemon.
func maintainControllerPresence(cfg controllerPresence) error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(stop)

	var packet []byte
	ann := cfg.announcement
	startedAt := cfg.startedAt
	if startedAt.IsZero() {
		startedAt = time.Now()
	}
	var rateTracker *observe.TrafficRateTracker
	var commands adoption.CommandLifecycle
	schedule, err := inform.NewSchedule(informScheduleConfig(cfg.flags))
	if err != nil {
		return fmt.Errorf("create inform schedule: %w", err)
	}
	if cfg.flags.trafficRatesEnabled {
		rateTracker = observe.NewTrafficRateTracker()
	}

	for {
		uptimeSeconds := runtimeUptime(startedAt)
		store := loadAdoptionState(cfg.flags.sshState)
		informURL := effectiveInformURL(cfg.flags.controller, store)
		ann.Default = store.AuthKey == ""
		ann.Uptime = uint32(uptimeSeconds)
		packet, err = ann.MarshalBinary()
		if err != nil {
			return fmt.Errorf("marshal discovery announcement: %w", err)
		}
		flags := cfg.flags
		plt := runtimePlatform(flags)
		ports := portsForRuntime(flags, cfg.profile, cfg.portBuildOptions, plt)
		if flags.trafficRatesEnabled {
			ports = applyTrafficRates(ports, rateTracker, time.Now())
		}
		payload, err := payloadForIdentity(cfg.mac, cfg.ip, cfg.hostname, informURL, store, cfg.flags, cfg.profile, ports, uptimeSeconds)
		if err != nil {
			return err
		}

		sendDiscovery(packet, cfg.hostname, cfg.mac, cfg.discoverySkipped, cfg.discoveryInterface, cfg.discoveryTargets)
		decision := sendInformHeartbeat(
			cfg.mac,
			informURL,
			cfg.flags.sshState,
			cfg.flags.statusPath,
			store,
			payload,
			informSourceIP(cfg.flags, cfg.ip),
			&commands,
			schedule,
		)

		if cfg.flags.once {
			return nil
		}

		if waitForNextHeartbeat(decision.NextDelay, stop) {
			ann.Sequence++
			continue
		}
		log.Println("stopping")
		return nil
	}
}

func waitForNextHeartbeat(delay time.Duration, stop <-chan os.Signal) bool {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-timer.C:
		return true
	case <-stop:
		return false
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

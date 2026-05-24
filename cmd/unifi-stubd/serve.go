// serveSwitchEmulation orchestrates the daemon lifecycle after configuration is
// resolved. Protocol encoding, profile rendering, observation, and SSH handling
// stay in lower-level packages; this layer wires them together.
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/konstruktor1/unifi-stubd/internal/adoption"
	"github.com/konstruktor1/unifi-stubd/internal/adoptionssh"
	appconfig "github.com/konstruktor1/unifi-stubd/internal/config"
	"github.com/konstruktor1/unifi-stubd/internal/device"
	"github.com/konstruktor1/unifi-stubd/internal/discovery"
	"github.com/konstruktor1/unifi-stubd/internal/observe"
)

// serveSwitchEmulation applies configuration, validates the selected runtime
// mode, builds the initial controller-facing identity, and then either exits
// for one-shot commands or enters the presence loop.
func serveSwitchEmulation() error {
	defaults := appconfig.Default()
	flags, changed := parseRuntimeFlags(defaults)

	// Early exits run before YAML/profile/runtime validation because they do
	// not need host state and must not start any controller-facing traffic.
	if flags.binaryVersion {
		fmt.Println(version)
		return nil
	}
	if strings.TrimSpace(flags.profileTemplate) != "" {
		return printProfileTemplate(flags.profileTemplate)
	}
	if strings.TrimSpace(flags.profileValidate) != "" {
		return validateProfilePath(flags.profileValidate)
	}

	cfg, err := loadConfig(flags.configPath, changed["config"])
	if err != nil {
		if flags.validate {
			return withExitCode(2, err)
		}
		return err
	}
	applyConfig(cfg, changed, &flags)

	registry, err := loadProfileRegistry(flags)
	if err != nil {
		if flags.validate {
			return withExitCode(profileErrorExitCode(err), err)
		}
		return err
	}
	if flags.listProfiles {
		fmt.Print(registry.FormatProfiles())
		return nil
	}
	if strings.TrimSpace(flags.profileExport) != "" {
		return printProfileExport(registry, flags.profileExport)
	}
	if err := validateOperationFlags(&flags); err != nil {
		if flags.validate {
			return withExitCode(1, err)
		}
		return err
	}

	// Profile defaults are applied before mode-specific validation so derived
	// port counts, model IDs, payload kind, and interface names are validated
	// as the daemon will actually use them.
	profile, ok := registry.LookupProfile(flags.profileName)
	if !ok {
		err := fmt.Errorf("unknown profile %q; known profiles: %s", flags.profileName, registry.ProfileNames())
		if flags.validate {
			return withExitCode(1, err)
		}
		return err
	}
	applyProfile(profile, &flags)
	plt := runtimePlatform(flags)
	// Non-live validation catches schema and policy mistakes without touching
	// host interfaces. Live checks are intentionally delayed until -validate or
	// actual runtime, where missing local interfaces should be reported.
	if err := validateManagementLAN(flags, profile, false); err != nil {
		if flags.validate {
			return withExitCode(1, err)
		}
		return err
	}
	if err := validateSourceMappings(flags, false); err != nil {
		if flags.validate {
			return withExitCode(1, err)
		}
		return err
	}
	if flags.validate {
		// -validate performs the same live reachability/name checks that a real
		// run would perform, but exits before discovery, inform, SSH, or status
		// persistence can start.
		if err := validateIdentityFlags(flags); err != nil {
			return withExitCode(1, err)
		}
		if err := validatePortOverrides(flags); err != nil {
			return withExitCode(1, err)
		}
		if err := validateSourceMappings(flags, true); err != nil {
			return withExitCode(1, err)
		}
		if err := validateManagementLAN(flags, profile, true); err != nil {
			return withExitCode(1, err)
		}
		fmt.Printf("configuration valid: profile=%s source=%s payload=%s\n", profile.Name, profile.Source, profile.Payload.Kind)
		return nil
	}
	if !flags.dryRunPlan {
		// Dry-run plans may describe unsupported host-network actions. All other
		// paths require live source checks before any controller-visible payload
		// is built.
		if err := validateSourceMappings(flags, true); err != nil {
			return err
		}
		if err := validateManagementLAN(flags, profile, true); err != nil {
			return err
		}
	}
	enrichCtx, enrichCancel := context.WithTimeout(context.Background(), observeTimeout)
	flags.portOverrides = enrichPortOverridesWithPlatform(enrichCtx, plt, flags.portOverrides)
	enrichCancel()
	// Platform enrichment can add observed MAC/IP/counter fields to operator
	// overrides, so validation runs after enrichment and before payload build.
	if err := validatePortOverrides(flags); err != nil {
		return err
	}

	resolvedHostname := resolveHostname(flags.hostname)
	uplinkPort := effectiveUplinkPort(profile, flags)
	portBuildOptions := resolvePortBuildOptions(flags.portCount, flags.linkSpeed, uplinkPort, effectiveUplinkSpeedMode(flags), flags.controller)
	mac := resolveMAC(flags.macText, resolvedHostname, profile, flags.model, flags.operationMode, flags.observeInterface)
	ip := net.ParseIP(flags.ipText).To4()
	if ip == nil {
		return fmt.Errorf("invalid IPv4 address: %q", flags.ipText)
	}
	ip, err = resolveManagementLANIdentityIP(flags, ip, plt)
	if err != nil {
		return err
	}
	if flags.dryRunPlan {
		printRuntimePlan(flags, profile, mac.String(), ip.String(), resolvedHostname)
		return nil
	}
	if flags.status || flags.statusJSON {
		return printLocalStatus(flags, profile, mac, ip, resolvedHostname, portBuildOptions, plt)
	}

	ann := discovery.Announcement{
		MAC:      mac,
		IP:       ip,
		Model:    flags.model,
		Version:  flags.version,
		Hostname: resolvedHostname,
		Default:  true,
		Uptime:   1,
		Sequence: 1,
	}

	packet, err := ann.MarshalBinary()
	if err != nil {
		return fmt.Errorf("marshal discovery announcement: %w", err)
	}

	ports := portsForRuntime(flags, profile, portBuildOptions, plt)
	if flags.trafficRatesEnabled {
		ports = applyTrafficRates(ports, observe.NewTrafficRateTracker(), time.Now())
	}
	payload, err := payloadForIdentity(mac, ip, resolvedHostname, flags.controller, adoption.Store{}, flags, profile, ports, 1)
	if err != nil {
		return err
	}

	if flags.dryRun {
		printDryRun(packet, payload)
		return nil
	}

	sshServer, err := startAdoptionSSH(flags, mac, ip, resolvedHostname)
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
		profile:            profile,
		mac:                mac,
		ip:                 ip,
		hostname:           resolvedHostname,
		portBuildOptions:   portBuildOptions,
		announcement:       ann,
		discoveryPacket:    packet,
		discoverySkipped:   flags.noDiscovery,
		discoveryInterface: effectiveDiscoveryInterface(flags),
		discoveryTargets:   flags.discoveryTargets,
		startedAt:          time.Now(),
	})
}

// controllerPresence bundles the immutable daemon identity and mutable runtime
// trackers used by the heartbeat loop.
type controllerPresence struct {
	flags              runtimeFlags
	profile            device.Profile
	mac                net.HardwareAddr
	ip                 net.IP
	hostname           string
	portBuildOptions   device.PortBuildOptions
	announcement       discovery.Announcement
	discoveryPacket    []byte
	discoverySkipped   bool
	discoveryInterface string
	discoveryTargets   []string
	startedAt          time.Time
}

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
			flags.portOverrides = enrichPortOverridesWithPlatform(ctx, plt, flags.portOverrides)
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

// sendDiscovery sends the already-encoded UDP announcement and treats send
// failures as heartbeat warnings, not daemon-fatal errors.
func sendDiscovery(packet []byte, hostname string, mac net.HardwareAddr, skip bool, iface string, targets []string) {
	if skip {
		return
	}
	if err := discovery.SendToInterface(packet, targets, iface); err != nil {
		log.Printf("discovery send failed: %v", err)
		return
	}
	log.Printf("sent discovery announcement for %s (%s)", hostname, mac)
}

// sendInformHeartbeat sends one inform packet, persists sanitized exchange
// status, and applies only safe adoption-state updates from decoded controller
// responses.
func sendInformHeartbeat(mac net.HardwareAddr, informURL, statePath, statusPath string, store adoption.Store, payload []byte, sourceIP net.IP) {
	if informURL == "" {
		return
	}
	resp, cipher, err := sendInform(mac, informURL, store, payload, sourceIP)
	if err != nil {
		recordLastInform(statusPath, newLastInformStatus(informURL, store), 0, "", cipher, 0, 0, err)
		log.Printf("inform send failed: %v", err)
		return
	}
	last := newLastInformStatus(informURL, store)
	last.StatusCode = resp.StatusCode
	last.AttemptedAESGCM = cipher.AttemptedAESGCM
	last.UsedAESGCM = cipher.UsedAESGCM
	last.FallbackToCBC = cipher.FallbackToCBC
	last.RawBytes = len(resp.RawBody)
	last.JSONBytes = len(resp.JSONBody)
	if len(resp.JSONBody) > 0 {
		controllerResponse, parseErr := adoption.ParseControllerResponseInfo(resp.JSONBody)
		if parseErr != nil {
			last.Error = parseErr.Error()
			log.Printf("controller response parse failed: %v", parseErr)
		} else {
			store = updateAdoptionState(statePath, store, controllerResponse, cipher.UsedAESGCM)
			if controllerResponse.ResetRequested && store.State == adoption.StateFactory && store.AuthKey == "" {
				controllerResponse.ResetApplied = true
			}
			last.ControllerState = adoptionStateText(store)
			last.CFGVersion = store.CFGVersion
			last.Version = store.Version
			applyControllerResponseStatus(&last, controllerResponse)
			logInformResponse(resp, controllerResponse, store, cipher)
		}
		recordLastInform(statusPath, last, resp.StatusCode, last.ResponseType, cipher, len(resp.RawBody), len(resp.JSONBody), nil)
		return
	}
	recordLastInform(statusPath, last, resp.StatusCode, "", cipher, len(resp.RawBody), 0, nil)
	log.Printf("inform response status=%d raw_bytes=%d cipher=%s", resp.StatusCode, len(resp.RawBody), cipherStatusText(cipher))
}

// applyControllerResponseStatus copies the sanitized controller response into
// persisted status, excluding raw provisioning bodies.
func applyControllerResponseStatus(last *lastInformStatus, response adoption.ControllerResponse) {
	last.ResponseType = response.Type
	last.IntervalSeconds = response.IntervalSeconds
	last.IncludeBlocks = cloneStrings(response.IncludeBlocks)
	last.ResetRequested = response.ResetRequested
	last.ResetApplied = response.ResetApplied
	last.ResetReason = response.ResetReason
	last.HasMgmtCFG = response.HasMgmtCFG
	last.HasSystemCFG = response.HasSystemCFG
	last.SystemCFGBytes = response.SystemCFGBytes
	last.SystemCFGKeys = cloneStrings(response.SystemCFGKeys)
	last.Ignored = response.Ignored
	last.IgnoredReason = response.IgnoredReason
}

// recordLastInform writes the same sanitized controller-exchange summary used
// by --status, keeping auth keys and raw controller payloads out of the status
// file.
func recordLastInform(statusPath string, last lastInformStatus, statusCode int, responseType string, cipher informCipherStatus, rawBytes, jsonBytes int, err error) {
	last.StatusCode = statusCode
	last.ResponseType = responseType
	last.AttemptedAESGCM = cipher.AttemptedAESGCM
	last.UsedAESGCM = cipher.UsedAESGCM
	last.FallbackToCBC = cipher.FallbackToCBC
	last.RawBytes = rawBytes
	last.JSONBytes = jsonBytes
	if err != nil {
		last.Error = err.Error()
	}
	if saveErr := saveLastInformStatus(statusPath, last); saveErr != nil {
		log.Printf("runtime status write failed: %v", saveErr)
	}
}

// startAdoptionSSH wires the optional SSH compatibility shim to the current
// fake identity and adoption store path.
func startAdoptionSSH(flags runtimeFlags, mac net.HardwareAddr, ip net.IP, hostname string) (*adoptionssh.Server, error) {
	sshServer, err := adoptionssh.Start(adoptionssh.Config{
		Listen:      flags.sshListen,
		User:        flags.sshUser,
		Password:    flags.sshPassword,
		HostKeyPath: flags.sshHostKey,
		StatePath:   flags.sshState,
		Identity: adoptionssh.Identity{
			MAC:       mac.String(),
			IP:        ip.String(),
			Hostname:  hostname,
			Model:     flags.model,
			Version:   flags.version,
			InformURL: flags.controller,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("adoption ssh failed: %w", err)
	}
	return sshServer, nil
}

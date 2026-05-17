package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/konstruktor1/unifi-stubd/internal/adoption"
	"github.com/konstruktor1/unifi-stubd/internal/adoptionssh"
	appconfig "github.com/konstruktor1/unifi-stubd/internal/config"
	"github.com/konstruktor1/unifi-stubd/internal/device"
	"github.com/konstruktor1/unifi-stubd/internal/discovery"
)

func serveSwitchEmulation() error {
	defaults := appconfig.Default()
	flags, changed := parseRuntimeFlags(defaults)

	if *flags.binaryVersion {
		fmt.Println(version)
		return nil
	}
	if *flags.listProfiles {
		fmt.Print(device.FormatProfiles())
		return nil
	}

	cfg, err := loadConfig(*flags.configPath, changed["config"])
	if err != nil {
		return err
	}
	applyConfig(cfg, changed, &flags)
	if err := validateOperationFlags(flags); err != nil {
		return err
	}

	profile, ok := device.LookupProfile(*flags.profileName)
	if !ok {
		log.Fatalf("unknown profile %q; known profiles: %s", *flags.profileName, device.ProfileNames())
	}
	applyProfile(profile, flags.model, flags.modelDisplay, flags.version, flags.portCount)
	if err := validatePortOverrides(flags); err != nil {
		return err
	}

	resolvedHostname := resolveHostname(*flags.hostname)
	portOptions := resolvePortOptions(profile, *flags.linkSpeed, *flags.uplinkPort, *flags.uplinkSpeed, *flags.controller)
	mac := resolveMAC(*flags.macText, resolvedHostname, profile, *flags.model, *flags.operationMode, *flags.observeInterface)
	ip := net.ParseIP(*flags.ipText).To4()
	if ip == nil {
		log.Fatalf("invalid IPv4 address: %q", *flags.ipText)
	}
	if *flags.dryRunPlan {
		printRuntimePlan(flags, profile, mac.String(), ip.String(), resolvedHostname)
		return nil
	}
	if *flags.status || *flags.statusJSON {
		return printLocalStatus(flags, profile, mac, ip, resolvedHostname, portOptions)
	}

	ann := discovery.Announcement{
		MAC:      mac,
		IP:       ip,
		Model:    *flags.model,
		Version:  *flags.version,
		Hostname: resolvedHostname,
		Default:  true,
		Uptime:   1,
		Sequence: 1,
	}

	packet, err := ann.MarshalBinary()
	if err != nil {
		return fmt.Errorf("marshal discovery announcement: %w", err)
	}

	ports := portsForRuntime(flags, portOptions)
	payload, err := payloadForIdentity(mac, ip, resolvedHostname, *flags.controller, adoption.Store{}, flags, profile, ports)
	if err != nil {
		return err
	}

	if *flags.dryRun {
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
		flags:            flags,
		profile:          profile,
		mac:              mac,
		ip:               ip,
		hostname:         resolvedHostname,
		portOptions:      portOptions,
		announcement:     ann,
		discoveryPacket:  packet,
		discoverySkipped: flags.noDiscovery,
	})
}

type controllerPresence struct {
	flags            runtimeFlags
	profile          device.Profile
	mac              net.HardwareAddr
	ip               net.IP
	hostname         string
	portOptions      device.PortOptions
	announcement     discovery.Announcement
	discoveryPacket  []byte
	discoverySkipped *bool
}

func maintainControllerPresence(cfg controllerPresence) error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	ticker := time.NewTicker(*cfg.flags.interval)
	defer ticker.Stop()

	packet := cfg.discoveryPacket
	ann := cfg.announcement

	for {
		store := loadAdoptionState(*cfg.flags.sshState)
		informURL := effectiveInformURL(*cfg.flags.controller, store)
		ports := portsForRuntime(cfg.flags, cfg.portOptions)
		payload, err := payloadForIdentity(cfg.mac, cfg.ip, cfg.hostname, informURL, store, cfg.flags, cfg.profile, ports)
		if err != nil {
			return err
		}

		sendDiscovery(packet, cfg.hostname, cfg.mac, *cfg.discoverySkipped)
		sendInformHeartbeat(cfg.mac, informURL, *cfg.flags.sshState, *cfg.flags.statusPath, store, payload)

		if *cfg.flags.once {
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

func sendDiscovery(packet []byte, hostname string, mac net.HardwareAddr, skip bool) {
	if skip {
		return
	}
	if err := discovery.Send(packet); err != nil {
		log.Printf("discovery send failed: %v", err)
		return
	}
	log.Printf("sent discovery announcement for %s (%s)", hostname, mac)
}

func sendInformHeartbeat(mac net.HardwareAddr, informURL, statePath, statusPath string, store adoption.Store, payload []byte) {
	if informURL == "" {
		return
	}
	resp, usedGCM, err := sendInform(mac, informURL, store, payload)
	if err != nil {
		recordLastInform(statusPath, newLastInformStatus(informURL, store), 0, "", usedGCM, 0, 0, err)
		log.Printf("inform send failed: %v", err)
		return
	}
	last := newLastInformStatus(informURL, store)
	last.StatusCode = resp.StatusCode
	last.UsedAESGCM = usedGCM
	last.RawBytes = len(resp.RawBody)
	last.JSONBytes = len(resp.JSONBody)
	if len(resp.JSONBody) > 0 {
		store = updateAdoptionState(statePath, store, resp.JSONBody, usedGCM)
		last.ControllerState = adoptionStateText(store)
		last.CFGVersion = store.CFGVersion
		last.Version = store.Version
		if _, kind, ok, _ := adoption.ParseControllerResponse(resp.JSONBody); ok {
			last.ResponseType = kind
		}
		recordLastInform(statusPath, last, resp.StatusCode, last.ResponseType, usedGCM, len(resp.RawBody), len(resp.JSONBody), nil)
		logInformResponse(resp, store)
		return
	}
	recordLastInform(statusPath, last, resp.StatusCode, "", usedGCM, len(resp.RawBody), 0, nil)
	log.Printf("inform response status=%d raw_bytes=%d", resp.StatusCode, len(resp.RawBody))
}

func recordLastInform(statusPath string, last lastInformStatus, statusCode int, responseType string, usedGCM bool, rawBytes, jsonBytes int, err error) {
	last.StatusCode = statusCode
	last.ResponseType = responseType
	last.UsedAESGCM = usedGCM
	last.RawBytes = rawBytes
	last.JSONBytes = jsonBytes
	if err != nil {
		last.Error = err.Error()
	}
	if saveErr := saveLastInformStatus(statusPath, last); saveErr != nil {
		log.Printf("runtime status write failed: %v", saveErr)
	}
}

func startAdoptionSSH(flags runtimeFlags, mac net.HardwareAddr, ip net.IP, hostname string) (*adoptionssh.Server, error) {
	sshServer, err := adoptionssh.Start(adoptionssh.Config{
		Listen:      *flags.sshListen,
		User:        *flags.sshUser,
		Password:    *flags.sshPassword,
		HostKeyPath: *flags.sshHostKey,
		StatePath:   *flags.sshState,
		Identity: adoptionssh.Identity{
			MAC:       mac.String(),
			IP:        ip.String(),
			Hostname:  hostname,
			Model:     *flags.model,
			Version:   *flags.version,
			InformURL: *flags.controller,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("adoption ssh failed: %w", err)
	}
	return sshServer, nil
}

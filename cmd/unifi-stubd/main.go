package main

import (
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/corspi/unifi-stubd/internal/adoption"
	"github.com/corspi/unifi-stubd/internal/adoptionssh"
	"github.com/corspi/unifi-stubd/internal/device"
	"github.com/corspi/unifi-stubd/internal/discovery"
	"github.com/corspi/unifi-stubd/internal/inform"
)

func main() {
	var (
		profileName  = flag.String("profile", "us16p150", "device profile to emulate; use -list-profiles to show options")
		listProfiles = flag.Bool("list-profiles", false, "list known device profiles and exit")
		macText      = flag.String("mac", "auto", "fake device MAC address, or auto to derive one from hostname and profile")
		ipText       = flag.String("ip", "192.168.1.50", "fake device IPv4 address")
		hostname     = flag.String("hostname", "auto", "fake device hostname, or auto to use the OS hostname")
		model        = flag.String("model", "", "override UniFi model identifier from the selected profile")
		modelDisplay = flag.String("model-display", "", "override display name from the selected profile")
		version      = flag.String("version", "", "override firmware version from the selected profile")
		portCount    = flag.Int("ports", 0, "override number of switch ports from the selected profile")
		linkSpeed    = flag.Int("link-speed", 0, "override default switch port speed in Mbps; 0 uses selected profile")
		uplinkSpeed  = flag.String("uplink-speed", "auto", "uplink speed in Mbps, auto, or profile")
		controller   = flag.String("controller", "", "optional UniFi inform URL, for example http://192.168.1.10:8080/inform")
		interval     = flag.Duration("interval", 10*time.Second, "announcement interval")
		dryRun       = flag.Bool("dry-run", false, "print payloads without sending packets")
		once         = flag.Bool("once", false, "send one discovery/inform batch and exit")
		noDiscovery  = flag.Bool("no-discovery", false, "skip UDP discovery and only send inform when -controller is set")
		sshListen    = flag.String("ssh-listen", "", "optional built-in adoption SSH listen address, for example 0.0.0.0:22")
		sshUser      = flag.String("ssh-user", "ubnt", "built-in adoption SSH username")
		sshPassword  = flag.String("ssh-password", "ubnt", "built-in adoption SSH password")
		sshHostKey   = flag.String("ssh-host-key", "/etc/unifi-stubd/ssh_host_rsa_key", "built-in adoption SSH host key path")
		sshState     = flag.String("ssh-state", "/var/lib/unifi-stubd/adoption.env", "built-in adoption SSH state file path")
	)
	flag.Parse()

	if *listProfiles {
		fmt.Print(device.FormatProfiles())
		return
	}
	profile, ok := device.LookupProfile(*profileName)
	if !ok {
		log.Fatalf("unknown profile %q; known profiles: %s", *profileName, device.ProfileNames())
	}
	applyProfile(profile, model, modelDisplay, version, portCount)
	resolvedHostname := resolveHostname(*hostname)
	portOptions := profile.PortOptions()
	if *linkSpeed > 0 {
		portOptions.Speed = *linkSpeed
		portOptions.UplinkSpeed = *linkSpeed
		portOptions.Media = ""
		portOptions.UplinkMedia = ""
	}
	portOptions = resolveUplinkSpeed(portOptions, *uplinkSpeed, *controller)

	mac := resolveMAC(*macText, resolvedHostname, profile, *model)
	ip := net.ParseIP(*ipText).To4()
	if ip == nil {
		log.Fatalf("invalid IPv4 address: %q", *ipText)
	}

	ann := discovery.Announcement{
		MAC:      mac,
		IP:       ip,
		Model:    *model,
		Version:  *version,
		Hostname: resolvedHostname,
		Default:  true,
		Uptime:   1,
		Sequence: 1,
	}

	packet, err := ann.MarshalBinary()
	if err != nil {
		log.Fatal(err)
	}

	payload, err := buildPayload(device.Identity{
		MAC:          mac.String(),
		IP:           ip.String(),
		Hostname:     resolvedHostname,
		Model:        *model,
		ModelDisplay: *modelDisplay,
		Version:      *version,
		Serial:       serialFromMAC(mac),
		InformURL:    *controller,
	}, adoption.Store{}, *portCount, portOptions)
	if err != nil {
		log.Fatal(err)
	}

	if *dryRun {
		fmt.Println("discovery_packet_hex:")
		fmt.Println(hex.EncodeToString(packet))
		fmt.Println()
		fmt.Println("minimal_inform_payload_json:")
		fmt.Println(string(payload))
		return
	}

	sshServer, err := adoptionssh.Start(adoptionssh.Config{
		Listen:      *sshListen,
		User:        *sshUser,
		Password:    *sshPassword,
		HostKeyPath: *sshHostKey,
		StatePath:   *sshState,
		Identity: adoptionssh.Identity{
			MAC:       mac.String(),
			IP:        ip.String(),
			Hostname:  resolvedHostname,
			Model:     *model,
			Version:   *version,
			InformURL: *controller,
		},
	})
	if err != nil {
		log.Fatalf("adoption ssh failed: %v", err)
	}
	defer sshServer.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	for {
		store := loadAdoptionState(*sshState)
		informURL := effectiveInformURL(*controller, store)
		payload, err := buildPayload(device.Identity{
			MAC:          mac.String(),
			IP:           ip.String(),
			Hostname:     resolvedHostname,
			Model:        *model,
			ModelDisplay: *modelDisplay,
			Version:      *version,
			Serial:       serialFromMAC(mac),
			InformURL:    informURL,
		}, store, *portCount, portOptions)
		if err != nil {
			log.Fatal(err)
		}

		if !*noDiscovery {
			if err := discovery.Send(packet); err != nil {
				log.Printf("discovery send failed: %v", err)
			} else {
				log.Printf("sent discovery announcement for %s (%s)", resolvedHostname, mac)
			}
		}
		if informURL != "" {
			resp, usedGCM, err := sendInform(mac, informURL, store, payload)
			if err != nil {
				log.Printf("inform send failed: %v", err)
			} else if len(resp.JSONBody) > 0 {
				store = updateAdoptionState(*sshState, store, resp.JSONBody, usedGCM)
				logInformResponse(resp, store)
			} else {
				log.Printf("inform response status=%d raw_bytes=%d", resp.StatusCode, len(resp.RawBody))
			}
		}
		if *once {
			return
		}

		select {
		case <-ticker.C:
			ann.Uptime += uint32(interval.Seconds())
			ann.Sequence++
			packet, err = ann.MarshalBinary()
			if err != nil {
				log.Fatal(err)
			}
		case <-stop:
			log.Println("stopping")
			return
		}
	}
}

func buildPayload(id device.Identity, store adoption.Store, portCount int, portOptions device.PortOptions) ([]byte, error) {
	id.CFGVersion = store.CFGVersion
	id.Adopted = store.AuthKey != ""
	if store.Version != "" {
		id.Version = store.Version
	}
	return device.MinimalSwitchPayload(id, device.SwitchPortsWithOptions(portCount, portOptions))
}

func resolveHostname(value string) string {
	value = strings.TrimSpace(value)
	if value != "" && strings.ToLower(value) != "auto" {
		return value
	}
	host, err := os.Hostname()
	if err == nil && strings.TrimSpace(host) != "" {
		return strings.TrimSpace(host)
	}
	return "unifi-stubd"
}

func resolveMAC(value, hostname string, profile device.Profile, model string) net.HardwareAddr {
	value = strings.TrimSpace(value)
	if value == "" || strings.EqualFold(value, "auto") {
		seed := strings.Join([]string{"unifi-stubd", hostname, profile.Name, model}, "|")
		mac := device.AutoMAC(seed)
		log.Printf("auto MAC resolved: %s seed=%q", mac, seed)
		return mac
	}
	mac, err := net.ParseMAC(value)
	if err != nil {
		log.Fatalf("invalid MAC address: %v", err)
	}
	return mac
}

func resolveUplinkSpeed(options device.PortOptions, value, target string) device.PortOptions {
	value = strings.TrimSpace(strings.ToLower(value))
	switch value {
	case "", "profile":
		return options
	case "auto":
		info, err := device.DetectEgressLink(target)
		if err != nil {
			log.Printf("uplink speed auto-detect failed: %v; using profile speed %d Mbps", err, options.UplinkSpeed)
			return options
		}
		options.UplinkSpeed = info.SpeedMbps
		if options.UplinkMedia == "" || options.UplinkMedia == options.Media {
			options.UplinkMedia = ""
		}
		log.Printf("uplink speed auto-detected: interface=%s local_ip=%s speed=%d Mbps", info.Interface, info.LocalIP, info.SpeedMbps)
		return options
	default:
		speed, err := strconv.Atoi(value)
		if err != nil || speed <= 0 {
			log.Fatalf("invalid -uplink-speed %q; use auto, profile, or a positive Mbps value", value)
		}
		options.UplinkSpeed = speed
		if options.UplinkMedia == "" || options.UplinkMedia == options.Media {
			options.UplinkMedia = ""
		}
		return options
	}
}

func applyProfile(profile device.Profile, model, modelDisplay, version *string, portCount *int) {
	if *model == "" {
		*model = profile.Model
	}
	if *modelDisplay == "" {
		*modelDisplay = profile.ModelDisplay
	}
	if *version == "" {
		*version = profile.Version
	}
	if *portCount == 0 {
		*portCount = profile.Ports
	}
}

func loadAdoptionState(path string) adoption.Store {
	store, err := adoption.LoadEnv(path)
	if err == nil {
		return store
	}
	if !errors.Is(err, os.ErrNotExist) {
		log.Printf("adoption state read failed: %v", err)
	}
	return adoption.Store{}
}

func effectiveInformURL(fallback string, store adoption.Store) string {
	if store.InformURL != "" {
		return store.InformURL
	}
	return fallback
}

func sendInform(mac net.HardwareAddr, url string, store adoption.Store, payload []byte) (*inform.Response, bool, error) {
	key, err := authKeyBytes(store.AuthKey)
	if err != nil {
		log.Printf("invalid adoption authkey, falling back to default key: %v", err)
		key = nil
	}

	options := []inform.Options{{Zlib: true}}
	if store.AuthKey != "" {
		if store.UseAESGCM {
			options = []inform.Options{{Zlib: true, GCM: true}}
		} else {
			options = []inform.Options{
				{Zlib: true, GCM: true},
				{Zlib: true},
			}
		}
	}

	var lastErr error
	var lastResp *inform.Response
	var lastUsedGCM bool
	for _, opts := range options {
		resp, err := inform.Client{
			URL:     url,
			MAC:     mac,
			Key:     key,
			Options: opts,
		}.Send(payload)
		if err == nil {
			lastResp = resp
			lastUsedGCM = opts.GCM
			if resp.StatusCode == http.StatusOK {
				return resp, opts.GCM, nil
			}
			continue
		}
		lastErr = err
	}
	if lastResp != nil {
		return lastResp, lastUsedGCM, nil
	}
	return nil, false, lastErr
}

func authKeyBytes(authKey string) ([]byte, error) {
	if authKey == "" {
		return nil, nil
	}
	if len(authKey) == 16 {
		return []byte(authKey), nil
	}
	key, err := hex.DecodeString(authKey)
	if err != nil {
		return nil, err
	}
	if len(key) != 16 {
		return nil, fmt.Errorf("decoded authkey has %d bytes, want 16", len(key))
	}
	return key, nil
}

func updateAdoptionState(path string, store adoption.Store, body []byte, usedGCM bool) adoption.Store {
	update, kind, ok, err := adoption.ParseControllerResponse(body)
	if err != nil {
		log.Printf("controller response parse failed: %v", err)
		return store
	}
	if !ok {
		if usedGCM && store.AuthKey != "" && !store.UseAESGCM {
			store.UseAESGCM = true
			if err := adoption.SaveEnv(path, store); err != nil {
				log.Printf("adoption state write failed: %v", err)
			}
		}
		return store
	}
	if usedGCM {
		update.UseAESGCM = true
	}
	store, changed := adoption.Merge(store, update)
	if changed {
		if err := adoption.SaveEnv(path, store); err != nil {
			log.Printf("adoption state write failed: %v", err)
		}
	}
	if kind == "upgrade" {
		log.Printf("controller requested firmware version %q; reporting it from next inform", store.Version)
	}
	return store
}

func logInformResponse(resp *inform.Response, store adoption.Store) {
	if _, kind, ok, _ := adoption.ParseControllerResponse(resp.JSONBody); ok {
		if kind == "setparam" {
			log.Printf(
				"inform response status=%d setparam cfgversion=%q inform_url=%q use_aes_gcm=%t authkey_set=%t",
				resp.StatusCode,
				store.CFGVersion,
				store.InformURL,
				store.UseAESGCM,
				store.AuthKey != "",
			)
			return
		}
		log.Printf(
			"inform response status=%d type=%s state=%q version=%q",
			resp.StatusCode,
			kind,
			store.State,
			store.Version,
		)
		return
	}
	log.Printf("inform response status=%d body=%s", resp.StatusCode, string(resp.JSONBody))
}

func serialFromMAC(mac net.HardwareAddr) string {
	out := make([]byte, hex.EncodedLen(len(mac)))
	hex.Encode(out, mac)
	for i := range out {
		if out[i] >= 'a' && out[i] <= 'f' {
			out[i] -= 'a' - 'A'
		}
	}
	return string(out)
}

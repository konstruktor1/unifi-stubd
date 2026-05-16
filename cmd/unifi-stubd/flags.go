package main

import (
	"flag"
	"time"

	appconfig "github.com/konstruktor1/unifi-stubd/internal/config"
	"github.com/konstruktor1/unifi-stubd/internal/device"
)

type runtimeFlags struct {
	configPath       *string
	operationMode    *string
	profileName      *string
	listProfiles     *bool
	macText          *string
	ipText           *string
	hostname         *string
	model            *string
	modelDisplay     *string
	version          *string
	binaryVersion    *bool
	portCount        *int
	linkSpeed        *int
	uplinkSpeed      *string
	uplinkPort       *int
	uplinkNeighbor   *device.MacTableEntry
	portOverrides    []device.PortOverride
	observeInterface *string
	observeBridge    *string
	lldpSource       *string
	trafficSource    *string
	controller       *string
	interval         *time.Duration
	dryRun           *bool
	dryRunPlan       *bool
	status           *bool
	statusJSON       *bool
	once             *bool
	noDiscovery      *bool
	sshListen        *string
	sshUser          *string
	sshPassword      *string
	sshHostKey       *string
	sshState         *string
	statusPath       *string
}

func parseRuntimeFlags(defaults appconfig.Config) (runtimeFlags, map[string]bool) {
	flags := runtimeFlags{
		configPath:       flag.String("config", appconfig.DefaultPath, "YAML config file path; default path is optional when absent"),
		operationMode:    flag.String("operation-mode", defaults.OperationMode, "runtime mode: stub, observe, host-direct, or macvlan"),
		profileName:      flag.String("profile", defaults.Profile, "device profile to emulate; use -list-profiles to show options"),
		listProfiles:     flag.Bool("list-profiles", false, "list known device profiles and exit"),
		macText:          flag.String("mac", defaults.MAC, "fake device MAC address, or auto to derive one from hostname and profile"),
		ipText:           flag.String("ip", defaults.IP, "fake device IPv4 address"),
		hostname:         flag.String("hostname", defaults.Hostname, "fake device hostname, or auto to use the OS hostname"),
		model:            flag.String("model", defaults.Model, "override UniFi model identifier from the selected profile"),
		modelDisplay:     flag.String("model-display", defaults.ModelDisplay, "override display name from the selected profile"),
		version:          flag.String("firmware-version", defaults.Version, "override firmware version from the selected profile"),
		binaryVersion:    flag.Bool("version", false, "print unifi-stubd version and exit"),
		portCount:        flag.Int("ports", defaults.Ports, "override number of switch ports from the selected profile"),
		linkSpeed:        flag.Int("link-speed", defaults.LinkSpeed, "override default switch port speed in Mbps; 0 uses selected profile"),
		uplinkSpeed:      flag.String("uplink-speed", defaults.UplinkSpeed, "uplink speed in Mbps, auto, or profile"),
		uplinkPort:       flag.Int("uplink-port", defaults.UplinkPort, "override uplink port index; 0 uses selected profile"),
		uplinkNeighbor:   configUplinkNeighbor(defaults.UplinkNeighbor),
		portOverrides:    configPortOverrides(defaults.PortOverrides),
		observeInterface: flag.String("observe-interface", defaults.ObserveInterface, "host interface used for passive link counters and speed"),
		observeBridge:    flag.String("observe-bridge", defaults.ObserveBridge, "Linux bridge used for passive FDB MAC table data"),
		lldpSource:       flag.String("lldp-source", defaults.LLDPSource, "passive LLDP source: off or lldpd"),
		trafficSource:    flag.String("traffic-source", defaults.TrafficSource, "traffic metadata source: off"),
		controller:       flag.String("controller", defaults.ControllerURL, "optional UniFi inform URL, for example http://192.168.1.10:8080/inform"),
		interval:         flag.Duration("interval", time.Duration(defaults.IntervalSeconds)*time.Second, "announcement interval"),
		dryRun:           flag.Bool("dry-run", false, "print payloads without sending packets"),
		dryRunPlan:       flag.Bool("dry-run-plan", false, "print the planned runtime actions without starting the stub"),
		status:           flag.Bool("status", false, "print local runtime status and exit"),
		statusJSON:       flag.Bool("status-json", false, "print local runtime status as JSON and exit"),
		once:             flag.Bool("once", false, "send one discovery/inform batch and exit"),
		noDiscovery:      flag.Bool("no-discovery", defaults.NoDiscovery, "skip UDP discovery and only send inform when -controller is set"),
		sshListen:        flag.String("ssh-listen", defaults.SSHListen, "optional built-in adoption SSH listen address, for example 0.0.0.0:22"),
		sshUser:          flag.String("ssh-user", defaults.SSHUser, "built-in adoption SSH username"),
		sshPassword:      flag.String("ssh-password", defaults.SSHPassword, "built-in adoption SSH password"),
		sshHostKey:       flag.String("ssh-host-key", defaults.SSHHostKeyPath, "built-in adoption SSH host key path"),
		sshState:         flag.String("ssh-state", defaults.StatePath, "built-in adoption SSH state file path"),
		statusPath:       flag.String("status-path", defaults.StatusPath, "non-sensitive runtime status file path"),
	}
	flag.Parse()
	return flags, changedFlags()
}

func changedFlags() map[string]bool {
	out := map[string]bool{}
	flag.Visit(func(f *flag.Flag) {
		out[f.Name] = true
	})
	return out
}

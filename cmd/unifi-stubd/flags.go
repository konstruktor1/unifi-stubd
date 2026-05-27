// Runtime flags define the operator-facing command surface for unifi-stubd.
// Defaults originate in the config package, then YAML and explicit CLI values
// are layered onto the same runtime structure.
package main

import (
	"flag"
	"time"

	appconfig "github.com/konstruktor1/unifi-stubd/internal/config"
	"github.com/konstruktor1/unifi-stubd/internal/config/flagvalue"
	"github.com/konstruktor1/unifi-stubd/internal/device"
)

// runtimeFlags holds the fully layered CLI/config runtime settings before they
// are validated and applied.
type runtimeFlags struct {
	configPath          string
	operationMode       string
	profileName         string
	profileFile         string
	profileDir          string
	listProfiles        bool
	validate            bool
	profileValidate     string
	profileExport       string
	profileTemplate     string
	macText             string
	ipText              string
	hostname            string
	model               string
	modelDisplay        string
	version             string
	binaryVersion       bool
	portCount           int
	linkSpeed           int
	uplinkSpeed         string
	uplinkPort          int
	uplinkNeighbor      *device.MacTableEntry
	portNeighbors       []device.PortNeighbor
	portOverrides       []device.PortOverride
	wanHealth           appconfig.WANHealthConfig
	bridgeObserve       appconfig.BridgeObserve
	portMappings        []appconfig.PortMapping
	observeInterface    string
	observeBridge       string
	lldpSource          string
	trafficSource       string
	trafficRatesEnabled bool
	logSource           string
	procSource          string
	dbusEnabled         bool
	dbusBus             string
	syslogPath          string
	controller          string
	interval            time.Duration
	dryRun              bool
	dryRunPlan          bool
	status              bool
	statusJSON          bool
	once                bool
	noDiscovery         bool
	discoveryInterface  string
	discoveryTargets    []string
	managementLAN       appconfig.ManagementLAN
	sshListen           string
	sshUser             string
	sshPassword         string
	sshHostKey          string
	sshState            string
	statusPath          string
}

// parseRuntimeFlags registers all CLI flags against defaults and records which
// flags were explicitly set so YAML config cannot override them later.
func parseRuntimeFlags(defaults appconfig.Config) (runtimeFlags, map[string]bool) {
	flags := runtimeFlags{
		uplinkNeighbor:   configUplinkNeighbor(defaults.UplinkNeighbor),
		portNeighbors:    configPortNeighbors(defaults.PortNeighbors),
		portOverrides:    configPortOverrides(defaults.PortOverrides),
		wanHealth:        cloneWANHealth(defaults.WANHealth),
		bridgeObserve:    cloneBridgeObserve(defaults.BridgeObserve),
		portMappings:     clonePortMappings(defaults.PortMappings),
		managementLAN:    defaults.ManagementLAN,
		discoveryTargets: cloneStrings(defaults.DiscoveryTargets),
	}
	flag.StringVar(&flags.configPath, "config", appconfig.DefaultPath, "YAML config file path; default path is optional when absent")
	flag.BoolVar(&flags.listProfiles, "list-profiles", false, "list known device profiles and exit")
	flag.BoolVar(&flags.validate, "validate", false, "validate config, profiles, and runtime constraints without starting the stub")
	flag.StringVar(&flags.profileValidate, "profile-validate", "", "validate one external profile YAML file or directory and exit")
	flag.StringVar(&flags.profileExport, "profile-export", "", "export a built-in or loaded profile as canonical YAML and exit")
	flag.StringVar(&flags.profileTemplate, "profile-template", "", "print a starter profile YAML template: switch or gateway")
	flag.BoolVar(&flags.binaryVersion, "version", false, "print unifi-stubd version and exit")
	flag.BoolVar(&flags.dryRun, "dry-run", false, "print payloads without sending packets")
	flag.BoolVar(&flags.dryRunPlan, "dry-run-plan", false, "print the planned runtime actions without starting the stub")
	flag.BoolVar(&flags.status, "status", false, "print local runtime status and exit")
	flag.BoolVar(&flags.statusJSON, "status-json", false, "print local runtime status as JSON and exit")
	flag.BoolVar(&flags.once, "once", false, "send one discovery/inform batch and exit")
	flag.Var((*flagvalue.StringList)(&flags.bridgeObserve.IgnoredMembers), "bridge-ignore-member", "exclude bridge member from port mapping; repeatable")
	flag.Var((*flagvalue.BridgeMemberPortMap)(&flags.bridgeObserve.MemberPortMap), "bridge-member-port", "pin bridge member to port, syntax member=PORT")
	flag.Var((*flagvalue.PortMapping)(&flags.portMappings), "port-map", "map port source, syntax port=N,interface=eth0 or port=N,disabled=true or port=N,unmapped=true")
	registerRuntimeSettings(&flags, defaults)
	flag.Parse()
	return flags, changedFlags()
}

// changedFlags records which CLI flags the operator set so YAML config cannot
// silently override them later.
func changedFlags() map[string]bool {
	out := map[string]bool{}
	flag.Visit(func(f *flag.Flag) {
		out[f.Name] = true
	})
	return out
}

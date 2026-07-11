// Command assert-events verifies captured discovery and inform lifecycle events.
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/inform"
)

const (
	eventKindInform    = "inform"
	eventKindDiscovery = "discovery"
	eventKindSetParam  = "setparam"
	eventKindGCM       = "gcm"
)

type options struct {
	eventsPath string
	startLine  int
	minCount   int
	kind       string
	macs       []string
}

type captureEvent struct {
	Event              string                  `json:"event"`
	TNBU               tnbuEvent               `json:"tnbu"`
	Discovery          discoveryEvent          `json:"discovery"`
	ControllerResponse controllerResponseEvent `json:"controller_response"`
}

type tnbuEvent struct {
	Present bool   `json:"present"`
	MAC     string `json:"mac"`
	Flags   uint16 `json:"flags"`
}

type discoveryEvent struct {
	Present bool   `json:"present"`
	MAC     string `json:"mac"`
}

type controllerResponseEvent struct {
	DecodedWithDefaultKey bool   `json:"decoded_with_default_key"`
	Recognized            bool   `json:"recognized"`
	Type                  string `json:"type"`
	HasMgmtCFG            bool   `json:"has_mgmt_cfg"`
	AuthKeySet            bool   `json:"authkey_set"`
	CFGVersionSet         bool   `json:"cfgversion_set"`
	UseAESGCM             bool   `json:"use_aes_gcm"`
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	opts, err := parseOptions(args)
	if err != nil {
		return err
	}
	found := make(map[string]int, len(opts.macs))
	for _, mac := range opts.macs {
		found[strings.ToLower(mac)] = 0
	}
	if err := countEvents(opts, found); err != nil {
		return err
	}

	missing := make([]string, 0)
	for mac, count := range found {
		if count < opts.minCount {
			missing = append(missing, mac)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		details := make([]string, 0, len(missing))
		for _, mac := range missing {
			details = append(details, fmt.Sprintf("%s=%d/%d", mac, found[mac], opts.minCount))
		}
		return fmt.Errorf("missing %s events for MACs: %s", opts.kind, strings.Join(details, ", "))
	}

	orderedMACs := make([]string, 0, len(found))
	for mac := range found {
		orderedMACs = append(orderedMACs, mac)
	}
	sort.Strings(orderedMACs)
	details := make([]string, 0, len(orderedMACs))
	for _, mac := range orderedMACs {
		details = append(details, fmt.Sprintf("%s=%d", mac, found[mac]))
	}
	fmt.Printf("found %s events for MACs: %s\n", opts.kind, strings.Join(details, ", "))
	return nil
}

func parseOptions(args []string) (options, error) {
	if len(args) < 3 {
		return options{}, errors.New("usage: assert-events <events.jsonl> <start-line> [--kind inform|discovery|setparam|gcm] [--min-count N] <mac> [<mac> ...]")
	}
	startLine, err := strconv.Atoi(args[1])
	if err != nil {
		return options{}, fmt.Errorf("invalid start-line %q: %w", args[1], err)
	}
	opts := options{
		eventsPath: args[0],
		startLine:  startLine,
		minCount:   1,
		kind:       eventKindInform,
	}
	remaining := args[2:]
	for len(remaining) > 0 && strings.HasPrefix(remaining[0], "--") {
		if len(remaining) < 2 {
			return options{}, fmt.Errorf("missing value for %s", remaining[0])
		}
		switch remaining[0] {
		case "--kind":
			opts.kind = remaining[1]
		case "--min-count":
			opts.minCount, err = strconv.Atoi(remaining[1])
			if err != nil {
				return options{}, fmt.Errorf("invalid min-count %q: %w", remaining[1], err)
			}
		default:
			return options{}, fmt.Errorf("unknown option %s", remaining[0])
		}
		remaining = remaining[2:]
	}
	if opts.minCount <= 0 {
		return options{}, errors.New("min-count must be positive")
	}
	switch opts.kind {
	case eventKindInform, eventKindDiscovery, eventKindSetParam, eventKindGCM:
	default:
		return options{}, fmt.Errorf("unknown event kind %q", opts.kind)
	}
	if len(remaining) == 0 {
		return options{}, errors.New("no MAC addresses supplied")
	}
	opts.macs = remaining
	return opts, nil
}

func countEvents(opts options, found map[string]int) error {
	file, err := os.Open(opts.eventsPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("open events: %w", err)
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if lineNumber <= opts.startLine || line == "" {
			continue
		}
		var event captureEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			return fmt.Errorf("decode event line %d: %w", lineNumber, err)
		}
		mac, matches := matchingMAC(event, opts.kind)
		if !matches {
			continue
		}
		mac = strings.ToLower(mac)
		if _, ok := found[mac]; ok {
			found[mac]++
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan events: %w", err)
	}
	return nil
}

func matchingMAC(event captureEvent, kind string) (string, bool) {
	switch kind {
	case eventKindInform:
		return event.TNBU.MAC, event.Event == "request" && event.TNBU.Present
	case eventKindDiscovery:
		return event.Discovery.MAC, event.Event == "discovery" && event.Discovery.Present
	case eventKindSetParam:
		response := event.ControllerResponse
		return event.TNBU.MAC, event.Event == "response" &&
			event.TNBU.Present &&
			response.DecodedWithDefaultKey &&
			response.Recognized &&
			response.Type == "setparam" &&
			response.HasMgmtCFG &&
			response.AuthKeySet &&
			response.CFGVersionSet &&
			response.UseAESGCM
	case eventKindGCM:
		return event.TNBU.MAC, event.Event == "request" &&
			event.TNBU.Present &&
			event.TNBU.Flags&inform.FlagEncryptedGCM != 0
	default:
		return "", false
	}
}

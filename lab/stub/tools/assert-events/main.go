// Command assert-events verifies captured inform request events.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: assert-events <events.jsonl> <start-line> [--min-count N] <mac> [<mac> ...]")
		os.Exit(2)
	}
	eventsPath := args[0]
	startLine, err := strconv.Atoi(args[1])
	if err != nil {
		return fmt.Errorf("invalid start-line %q: %w", args[1], err)
	}
	minCount := 1
	macs := args[2:]
	if len(macs) >= 2 && macs[0] == "--min-count" {
		minCount, err = strconv.Atoi(macs[1])
		if err != nil {
			return fmt.Errorf("invalid min-count %q: %w", macs[1], err)
		}
		macs = macs[2:]
	}
	if len(macs) == 0 {
		fmt.Fprintln(os.Stderr, "no MAC addresses supplied")
		os.Exit(2)
	}

	found := map[string]int{}
	for _, mac := range macs {
		found[strings.ToLower(mac)] = 0
	}
	if err := countEvents(eventsPath, startLine, found); err != nil {
		return err
	}

	missing := make([]string, 0)
	for mac, count := range found {
		if count < minCount {
			missing = append(missing, mac)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		details := make([]string, 0, len(missing))
		for _, mac := range missing {
			details = append(details, fmt.Sprintf("%s=%d/%d", mac, found[mac], minCount))
		}
		return fmt.Errorf("missing inform request events for MACs: %s", strings.Join(details, ", "))
	}

	macs = make([]string, 0, len(found))
	for mac := range found {
		macs = append(macs, mac)
	}
	sort.Strings(macs)
	details := make([]string, 0, len(macs))
	for _, mac := range macs {
		details = append(details, fmt.Sprintf("%s=%d", mac, found[mac]))
	}
	fmt.Printf("found inform request events for MACs: %s\n", strings.Join(details, ", "))
	return nil
}

func countEvents(path string, startLine int, found map[string]int) error {
	file, err := os.Open(path)
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
		if lineNumber <= startLine || line == "" {
			continue
		}
		var event map[string]any
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			return fmt.Errorf("decode event line %d: %w", lineNumber, err)
		}
		if event["event"] != "request" {
			continue
		}
		tnbu, ok := event["tnbu"].(map[string]any)
		if !ok || tnbu["present"] != true {
			continue
		}
		mac := strings.ToLower(fmt.Sprint(tnbu["mac"]))
		if _, ok := found[mac]; ok {
			found[mac]++
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan events: %w", err)
	}
	return nil
}

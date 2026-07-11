// Command assert-status verifies sanitized runtime status invariants used by
// the Docker controller gate.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"
)

type runtimeStatus struct {
	Config struct {
		ControllerIntervalMin string `json:"controller_interval_min"`
		ControllerIntervalMax string `json:"controller_interval_max"`
	} `json:"config"`
	Runtime struct {
		LastInform struct {
			IntervalSeconds int `json:"interval_seconds"`
			Schedule        struct {
				EffectiveInterval     string `json:"effective_interval"`
				NextDelay             string `json:"next_delay"`
				ConsecutiveFailures   int    `json:"consecutive_failures"`
				Backoff               bool   `json:"backoff"`
				ControllerHintApplied bool   `json:"controller_hint_applied"`
			} `json:"schedule"`
			ControllerCommands controllerCommandStatus `json:"controller_commands"`
		} `json:"last_inform"`
	} `json:"runtime"`
}

type controllerCommandStatus struct {
	LastCommand     string `json:"last_command"`
	HandledCount    int    `json:"handled_count"`
	ResultSentCount int    `json:"result_sent_count"`
	DuplicateCount  int    `json:"duplicate_count"`
	Pending         bool   `json:"pending"`
}

func main() {
	if err := run(os.Args[1:], os.Stdin); err != nil {
		fmt.Fprintf(os.Stderr, "status assertion failed: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, input io.Reader) error {
	if len(args) != 1 || (args[0] != "wg-genkeypair-complete" && args[0] != "no-command-stable") {
		return errors.New("usage: assert-status wg-genkeypair-complete|no-command-stable")
	}
	var status runtimeStatus
	if err := json.NewDecoder(io.LimitReader(input, 4*1024*1024)).Decode(&status); err != nil {
		return fmt.Errorf("decode runtime status: %w", err)
	}
	commands := status.Runtime.LastInform.ControllerCommands
	switch args[0] {
	case "wg-genkeypair-complete":
		if err := validateCompletedCommand(commands); err != nil {
			return err
		}
	case "no-command-stable":
		if err := validateNoCommand(commands); err != nil {
			return err
		}
	}
	last := status.Runtime.LastInform
	if last.IntervalSeconds <= 0 {
		return fmt.Errorf("controller interval = %d, want a positive hint", last.IntervalSeconds)
	}
	minimum, err := time.ParseDuration(status.Config.ControllerIntervalMin)
	if err != nil {
		return fmt.Errorf("parse controller interval minimum: %w", err)
	}
	maximum, err := time.ParseDuration(status.Config.ControllerIntervalMax)
	if err != nil {
		return fmt.Errorf("parse controller interval maximum: %w", err)
	}
	expected := time.Duration(last.IntervalSeconds) * time.Second
	if expected < minimum {
		expected = minimum
	}
	if expected > maximum {
		expected = maximum
	}
	schedule := last.Schedule
	if schedule.EffectiveInterval != expected.String() || schedule.NextDelay != expected.String() {
		return fmt.Errorf("inform schedule = effective %q next %q, want %s from controller hint %ds", schedule.EffectiveInterval, schedule.NextDelay, expected, last.IntervalSeconds)
	}
	if schedule.ConsecutiveFailures != 0 || schedule.Backoff {
		return fmt.Errorf("inform schedule did not reset: failures=%d backoff=%t", schedule.ConsecutiveFailures, schedule.Backoff)
	}
	if !schedule.ControllerHintApplied {
		return errors.New("controller interval hint was not applied")
	}
	fmt.Printf("controller lifecycle: expectation=%s command=%q handled=%d result_sent=%d duplicates=%d pending=%t controller_interval=%ds effective=%s failures=0 backoff=false\n",
		args[0], commands.LastCommand, commands.HandledCount, commands.ResultSentCount, commands.DuplicateCount, commands.Pending, last.IntervalSeconds, expected)
	return nil
}

func validateCompletedCommand(commands controllerCommandStatus) error {
	if commands.LastCommand != "wg-genkeypair" {
		return fmt.Errorf("last command = %q, want wg-genkeypair", commands.LastCommand)
	}
	if commands.HandledCount != 1 {
		return fmt.Errorf("handled command count = %d, want 1", commands.HandledCount)
	}
	if commands.ResultSentCount != 1 {
		return fmt.Errorf("sent result count = %d, want 1", commands.ResultSentCount)
	}
	if commands.DuplicateCount != 0 {
		return fmt.Errorf("duplicate command count = %d, want 0", commands.DuplicateCount)
	}
	if commands.Pending {
		return errors.New("controller command result is still pending")
	}
	return nil
}

func validateNoCommand(commands controllerCommandStatus) error {
	if commands.LastCommand != "" || commands.HandledCount != 0 || commands.ResultSentCount != 0 || commands.DuplicateCount != 0 || commands.Pending {
		return fmt.Errorf("unexpected controller command state: last=%q handled=%d sent=%d duplicates=%d pending=%t",
			commands.LastCommand, commands.HandledCount, commands.ResultSentCount, commands.DuplicateCount, commands.Pending)
	}
	return nil
}

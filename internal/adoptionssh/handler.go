package adoptionssh

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"path"
	"strings"
)

// Execute runs one or more shell-like adoption commands.
func (h *Handler) Execute(command string) (string, int) {
	var out strings.Builder
	status := 0
	for _, part := range splitCommands(command) {
		text, code := h.executeOne(part)
		out.WriteString(text)
		if code != 0 {
			status = code
		}
	}
	return out.String(), status
}

// Shell serves a minimal interactive CLI over rw.
func (h *Handler) Shell(rw io.ReadWriter) {
	_, _ = io.WriteString(rw, "UniFi CLI shim\n")
	scanner := bufio.NewScanner(rw)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if line == "exit" || line == "quit" {
			return
		}
		output, _ := h.Execute(line)
		_, _ = io.WriteString(rw, output)
	}
}

// executeOne handles one parsed adoption command and intentionally maps unknown
// commands to accepted no-ops rather than host execution.
func (h *Handler) executeOne(command string) (string, int) {
	args := CommandFields(strings.TrimSpace(command))
	if len(args) == 0 {
		return "", 0
	}
	if args[0] == "sudo" {
		args = args[1:]
		if len(args) == 0 {
			return okOutput, 0
		}
	}
	if (args[0] == "sh" || args[0] == "/bin/sh") && len(args) >= 3 && args[1] == "-c" {
		// Controller SSH adoption often wraps commands in a shell. Re-parse the
		// quoted command inside the shim instead of invoking a host shell.
		return h.Execute(strings.Join(args[2:], " "))
	}

	name := path.Base(args[0])
	log.Printf("adoption ssh command: %s", strings.Join(args, " "))

	switch name {
	case "syswrapper.sh":
		return h.handleSyswrapper(args)
	case "mca-cli-op":
		if len(args) > 1 && args[1] == commandInfo {
			return h.info(), 0
		}
		return h.handleSetInform(args)
	case commandInfo:
		return h.info(), 0
	case commandSetInform:
		return h.handleSetInform(args)
	case "mca-cli":
		return h.handleMCA(args)
	case "ubntbox":
		h.saveState("", "", strings.Join(args, " "))
		return okOutput, 0
	case "reset2defaults", "restore-default":
		// Factory-reset commands clear only the stub adoption file. They must not
		// reset the host, services, users, interfaces, or firewall.
		h.resetState(strings.Join(args, " "))
		return "Factory reset accepted\n", 0
	case "hostname":
		return h.config.Identity.Hostname + "\n", 0
	case "uname":
		return "Linux unifi-stubd-lab 6.18.22-0-virt #1-Alpine SMP aarch64 GNU/Linux\n", 0
	case "cat":
		return h.handleCat(args)
	case "echo":
		return strings.Join(args[1:], " ") + "\n", 0
	case "true", "exit", "quit":
		return "", 0
	default:
		if url := findInformURL(args); url != "" {
			h.saveState(url, "", strings.Join(args, " "))
			return fmt.Sprintf("Adoption request accepted: %s\n", url), 0
		}
		// Unsupported commands are acknowledged for compatibility but are not
		// executed. Persisting the command gives operators an audit trail.
		h.saveState("", "", strings.Join(args, " "))
		return okOutput, 0
	}
}

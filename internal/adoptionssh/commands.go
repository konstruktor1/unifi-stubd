package adoptionssh

import (
	"fmt"
	"strings"
)

// handleSyswrapper emulates the syswrapper subcommands used by advanced
// adoption while keeping reset and adopt effects inside the stub state file.
func (h *Handler) handleSyswrapper(args []string) (string, int) {
	if len(args) < 2 {
		return h.info(), 0
	}
	switch args[1] {
	case "set-adopt":
		url := ""
		if len(args) > 2 {
			url = args[2]
		}
		authKey := ""
		if len(args) > 3 {
			authKey = args[3]
		}
		h.saveState(url, authKey, strings.Join(args, " "))
		if url == "" {
			return "Adoption command accepted\n", 0
		}
		return fmt.Sprintf("Adoption request accepted: %s\n", url), 0
	case commandSetInform:
		return h.handleSetInform(args)
	case commandInfo, "status", "get-info":
		return h.info(), 0
	case "restore-default", "reset2defaults":
		h.resetState(strings.Join(args, " "))
		return "Factory reset accepted\n", 0
	default:
		h.saveState("", "", strings.Join(args, " "))
		return okOutput, 0
	}
}

// handleMCA emulates the mca-cli command forms that controllers commonly try
// during SSH adoption.
func (h *Handler) handleMCA(args []string) (string, int) {
	if len(args) < 2 {
		return "UniFi CLI shim\n", 0
	}
	if args[1] == "op" && len(args) > 2 {
		return h.handleSetInform(args[1:])
	}
	switch args[1] {
	case commandSetInform:
		return h.handleSetInform(args)
	case commandInfo, "status":
		return h.info(), 0
	default:
		h.saveState("", "", strings.Join(args, " "))
		return okOutput, 0
	}
}

// handleSetInform persists the controller-supplied inform URL as adoption
// state, not as a command to run on the host.
func (h *Handler) handleSetInform(args []string) (string, int) {
	url := findInformURL(args)
	h.saveState(url, "", strings.Join(args, " "))
	if url == "" {
		return "Inform command accepted\n", 0
	}
	return fmt.Sprintf("Inform URL set: %s\n", url), 0
}

// handleCat returns only safe identity/version files that adoption clients
// probe; arbitrary file reads are not supported.
func (h *Handler) handleCat(args []string) (string, int) {
	if len(args) < 2 {
		return "", 0
	}
	switch args[1] {
	case "/etc/version", "/usr/lib/version":
		return h.config.Identity.Version + "\n", 0
	case "/proc/ubnthal/system.info":
		return h.info(), 0
	default:
		return "", 0
	}
}

// info returns a small firmware-like identity report used by controller SSH
// adoption checks.
func (h *Handler) info() string {
	id := h.config.Identity
	if id.Model == "" {
		id.Model = "US8"
	}
	if id.Version == "" {
		id.Version = "6.6.0"
	}
	if id.Hostname == "" {
		id.Hostname = "unifi-stubd"
	}
	informURL := id.InformURL
	if informURL == "" {
		informURL = "http://unifi:8080/inform"
	}
	return fmt.Sprintf(
		"Model:       %s\nVersion:     %s\nMAC Address: %s\nIP Address:  %s\nHostname:    %s\nUptime:      1 seconds\nStatus:      Not Adopted (%s)\n",
		id.Model,
		id.Version,
		id.MAC,
		id.IP,
		id.Hostname,
		informURL,
	)
}

package ifsource

import (
	"os/exec"
	"strings"
)

// hostInterfaceDetails carries link state parsed from platform network tools.
type hostInterfaceDetails struct {
	// Up is the optional parsed carrier state.
	Up *bool
	// Speed is the parsed link speed in Mbps.
	Speed int
	// Media is the UniFi media label inferred from platform output.
	Media string
}

// readHostInterfaceDetails reads platform link details through ifconfig.
func readHostInterfaceDetails(ifaceName string) hostInterfaceDetails {
	out, err := exec.Command("ifconfig", ifaceName).Output()
	if err != nil {
		return hostInterfaceDetails{}
	}
	return parseIfconfigDetails(string(out))
}

// parseIfconfigDetails extracts link state, speed, and media from ifconfig output.
func parseIfconfigDetails(output string) hostInterfaceDetails {
	var details hostInterfaceDetails
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		lower := strings.ToLower(line)
		switch {
		case strings.HasPrefix(lower, "status:"):
			up := strings.Contains(lower, "active") || strings.Contains(lower, "up")
			details.Up = &up
		case strings.HasPrefix(lower, "media:"):
			details.Speed = speedFromMediaLine(lower)
			details.Media = mediaFromMediaLine(lower)
		}
	}
	return details
}

// speedFromMediaLine maps ifconfig media text to Mbps.
func speedFromMediaLine(line string) int {
	switch {
	case strings.Contains(line, "25g"):
		return 25000
	case strings.Contains(line, "10g"):
		return 10000
	case strings.Contains(line, "5g"):
		return 5000
	case strings.Contains(line, "2.5g"), strings.Contains(line, "2500base"):
		return 2500
	case strings.Contains(line, "1000base"), strings.Contains(line, "1g"):
		return 1000
	case strings.Contains(line, "100base"):
		return 100
	case strings.Contains(line, "10base"):
		return 10
	default:
		return 0
	}
}

// mediaFromMediaLine maps ifconfig media text to UniFi media labels.
func mediaFromMediaLine(line string) string {
	switch {
	case strings.Contains(line, "sfp28"), strings.Contains(line, "25gbase"):
		return "SFP28"
	case strings.Contains(line, "sfp+"),
		strings.Contains(line, "twinax"),
		strings.Contains(line, "10gbase-sr"),
		strings.Contains(line, "10gbase-lr"),
		strings.Contains(line, "10gbase-lrm"),
		strings.Contains(line, "10gbase-er"),
		strings.Contains(line, "10gbase-cr"),
		strings.Contains(line, "10gbase-cu"):
		return "SFP+"
	case speedFromMediaLine(line) > 0:
		return "GE"
	default:
		return ""
	}
}

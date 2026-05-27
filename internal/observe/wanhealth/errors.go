package wanhealth

import "strings"

func sanitizeError(err error) string {
	message := strings.TrimSpace(err.Error())
	if message == "" {
		return "ping failed"
	}
	message = strings.Join(strings.Fields(message), " ")
	if len(message) > maxErrorLength {
		return strings.TrimSpace(message[:maxErrorLength]) + "..."
	}
	return message
}

func firstNonEmptyLine(value string) string {
	for _, line := range strings.Split(value, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return ""
}

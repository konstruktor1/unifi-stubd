package adoptionssh

import "strings"

// splitCommands handles the small command chains controllers send over SSH
// while routing each segment through the safe shim.
func splitCommands(command string) []string {
	raw := strings.NewReplacer("&&", ";", "\n", ";").Replace(command)
	parts := strings.Split(raw, ";")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

// CommandFields splits a shell-like command line for the adoption command shim.
func CommandFields(input string) []string {
	var fields []string
	var current strings.Builder
	var quote rune
	escaped := false

	flush := func() {
		if current.Len() == 0 {
			return
		}
		fields = append(fields, current.String())
		current.Reset()
	}

	for _, r := range input {
		switch {
		case escaped:
			current.WriteRune(r)
			escaped = false
		case r == '\\':
			escaped = true
		case quote != 0:
			if r == quote {
				quote = 0
			} else {
				current.WriteRune(r)
			}
		case r == '\'' || r == '"':
			quote = r
		case r == ' ' || r == '\t' || r == '\r' || r == '\n':
			flush()
		default:
			current.WriteRune(r)
		}
	}
	if escaped {
		current.WriteRune('\\')
	}
	flush()
	return fields
}

// findInformURL extracts only inform endpoints from SSH command arguments.
func findInformURL(args []string) string {
	for _, arg := range args {
		if (strings.HasPrefix(arg, "http://") || strings.HasPrefix(arg, "https://")) && strings.Contains(arg, "/inform") {
			return arg
		}
	}
	return ""
}

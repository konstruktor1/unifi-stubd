package adoption

// knownControllerResponseType reports whether the response discriminator has
// an evidence-bound top-level branch. Every other value is a dispatcher error;
// it must not become a successful no-op or trigger a local reset.
func knownControllerResponseType(value string) bool {
	switch value {
	case "noop", "setparam", "setdefault", "reboot", "upgrade", "cmd", "geo_ip_update":
		return true
	default:
		return false
	}
}

// isUnsafeControllerCommand identifies known response types that imply host or
// firmware changes and must remain metadata-only in the stub.
func isUnsafeControllerCommand(responseType string) bool {
	switch responseType {
	case "reboot", "upgrade", "geo_ip_update":
		return true
	default:
		return false
	}
}

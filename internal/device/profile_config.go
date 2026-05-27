// Package device loads and stores embedded device profile data.
package device

// Embedded profile config decoding turns checked-in YAML documents into
// validated built-in profiles during init registration.

// Profile schema defaults describe the current YAML version and renderer
// fallback values.
const (
	schemaVersion          = 1
	payloadKindSwitch      = "switch"
	payloadKindGateway     = "gateway"
	defaultRequiredVersion = "5.0.0"
	defaultMgmtInterface   = "eth0"
	defaultGatewayPrefix   = "eth"
)

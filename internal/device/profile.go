// Package device exposes the canonical profile model without requiring callers
// to depend on profilemodel internals. AutoMAC provides deterministic local
// identity generation for lab defaults.
package device

import (
	"crypto/sha256"
	"net"
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/device/profilemodel"
)

// Profile defines a built-in UniFi device profile.
type Profile = profilemodel.Profile

// PayloadProfile contains profile-driven inform payload rendering metadata.
type PayloadProfile = profilemodel.PayloadProfile

// AutoMAC derives a stable locally administered MAC address from seed.
func AutoMAC(seed string) net.HardwareAddr {
	sum := sha256.Sum256([]byte(strings.TrimSpace(seed)))
	mac := net.HardwareAddr{sum[0], sum[1], sum[2], sum[3], sum[4], sum[5]}
	mac[0] = (mac[0] | 0x02) & 0xfe
	return mac
}

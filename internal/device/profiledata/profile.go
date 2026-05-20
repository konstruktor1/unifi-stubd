// Package profiledata aliases the canonical profile model so the loader,
// registry, and renderer use the same typed schema while keeping package
// boundaries clear.
package profiledata

import "github.com/konstruktor1/unifi-stubd/internal/device/profilemodel"

// Profile defines a built-in UniFi device profile.
type Profile = profilemodel.Profile

// PortGroup describes one contiguous block in a profile port layout.
type PortGroup = profilemodel.PortGroup

// PayloadProfile contains profile-driven inform payload rendering metadata.
type PayloadProfile = profilemodel.PayloadProfile

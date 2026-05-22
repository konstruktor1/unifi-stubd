// Package profile registers the embedded device profile.
package profile

import (
	_ "embed"

	"github.com/konstruktor1/unifi-stubd/internal/device/profiledata"
)

//go:embed profile.yaml
var config []byte

// init registers the embedded US-8 profile with the global registry.
func init() {
	profiledata.RegisterConfig("profiles/us8/profile.yaml", config)
}

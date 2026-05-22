// Package profile registers the embedded device profile.
package profile

import (
	_ "embed"

	"github.com/konstruktor1/unifi-stubd/internal/device/profiledata"
)

//go:embed profile.yaml
var config []byte

// init registers the embedded UGW3 profile with the global registry.
func init() {
	profiledata.RegisterConfig("profiles/ugw3/profile.yaml", config)
}

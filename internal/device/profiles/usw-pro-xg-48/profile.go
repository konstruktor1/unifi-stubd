// Package profile registers the embedded device profile.
package profile

import (
	_ "embed"

	"github.com/konstruktor1/unifi-stubd/internal/device/profiledata"
)

//go:embed profile.yaml
var config []byte

// init registers the embedded USW-Pro-XG-48 profile with the global registry.
func init() {
	profiledata.RegisterConfig("profiles/usw-pro-xg-48/profile.yaml", config)
}

// Package device registers checked-in YAML profiles through blank imports at
// init time. Runtime lookup stays data-driven without hard-coding every profile
// in one switch statement.
package device

import (
	_ "github.com/konstruktor1/unifi-stubd/internal/device/profiles/ucg-fiber"     // register embedded profile
	_ "github.com/konstruktor1/unifi-stubd/internal/device/profiles/ugw3"          // register embedded profile
	_ "github.com/konstruktor1/unifi-stubd/internal/device/profiles/us16p150"      // register embedded profile
	_ "github.com/konstruktor1/unifi-stubd/internal/device/profiles/us16xg"        // register embedded profile
	_ "github.com/konstruktor1/unifi-stubd/internal/device/profiles/us24p250"      // register embedded profile
	_ "github.com/konstruktor1/unifi-stubd/internal/device/profiles/us48p500"      // register embedded profile
	_ "github.com/konstruktor1/unifi-stubd/internal/device/profiles/us8"           // register embedded profile
	_ "github.com/konstruktor1/unifi-stubd/internal/device/profiles/us8p60"        // register embedded profile
	_ "github.com/konstruktor1/unifi-stubd/internal/device/profiles/usaggpro"      // register embedded profile
	_ "github.com/konstruktor1/unifi-stubd/internal/device/profiles/usw-pro-xg-48" // register embedded profile
	_ "github.com/konstruktor1/unifi-stubd/internal/device/profiles/uxg-lite"      // register embedded profile
	_ "github.com/konstruktor1/unifi-stubd/internal/device/profiles/uxgpro"        // register embedded profile
)

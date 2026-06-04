package main

import "github.com/konstruktor1/unifi-stubd/internal/device"

// applyProfile fills CLI defaults from the selected profile after external
// profile loading and before identity and payload construction.
func applyProfile(profile device.Profile, flags *runtimeFlags) {
	for _, field := range []struct {
		target *string
		value  string
	}{
		{target: &flags.model, value: profile.Model},
		{target: &flags.modelDisplay, value: profile.ModelDisplay},
		{target: &flags.version, value: profile.Version},
	} {
		setDefaultString(field.target, field.value)
	}
	setDefaultInt(&flags.portCount, profile.Ports)
}

// setDefaultString fills profile-derived string defaults only when the operator
// did not set a value.
func setDefaultString(target *string, value string) {
	if *target == "" {
		*target = value
	}
}

// setDefaultInt fills profile-derived numeric defaults only when the operator
// did not set a value.
func setDefaultInt(target *int, value int) {
	if *target == 0 {
		*target = value
	}
}

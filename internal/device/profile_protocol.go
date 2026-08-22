package device

const profileProtocolNoBinding = "no controller policy for the exact built-in profile, model, and version"

type userProfileProtocolKey struct {
	sourceType string
	name       string
	model      string
	version    string
}

type userProfileProtocolPolicy struct {
	firmwareMinimumSeconds int
	blockSelection         bool
	controllerClock        bool
	idleSpeedtest          bool
}

// userProfileProtocolPolicies contains only controller behaviors the product
// actually applies. Missing entries and identity overrides remain inert.
var userProfileProtocolPolicies = map[userProfileProtocolKey]userProfileProtocolPolicy{
	{sourceTypeBuiltIn, "us16p150", "US16P150", defaultSwitchVersion}: mainlineSwitchProtocolPolicy,
	{sourceTypeBuiltIn, "us24p250", "US24P250", defaultSwitchVersion}: mainlineSwitchProtocolPolicy,
	{sourceTypeBuiltIn, "us48p500", "US48P500", defaultSwitchVersion}: mainlineSwitchProtocolPolicy,
	{sourceTypeBuiltIn, "us8", "US8", defaultSwitchVersion}:           mainlineSwitchProtocolPolicy,
	{sourceTypeBuiltIn, "us8p60", "US8P60", defaultSwitchVersion}:     mainlineSwitchProtocolPolicy,
	{sourceTypeBuiltIn, "uxgpro", "UXGPRO", defaultGatewayVersion}: {
		firmwareMinimumSeconds: 1,
		blockSelection:         true,
		controllerClock:        true,
		idleSpeedtest:          true,
	},
}

var mainlineSwitchProtocolPolicy = userProfileProtocolPolicy{
	firmwareMinimumSeconds: 30,
	controllerClock:        true,
}

func resolveUserProfileProtocol(profile Profile, effectiveModel, effectiveVersion string) (userProfileProtocolPolicy, bool) {
	policy, ok := userProfileProtocolPolicies[userProfileProtocolKey{
		sourceType: profile.SourceType,
		name:       profile.Name,
		model:      effectiveModel,
		version:    effectiveVersion,
	}]
	if !ok {
		return userProfileProtocolPolicy{}, false
	}
	return policy, true
}

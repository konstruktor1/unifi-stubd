package device

// Setter order is policy: text metadata is normalized first, speed may set a
// default media label, explicit media can override that label, link-down can
// clear speed, and disabled finally clears all live port state.
var portOverrideSetters = []portOverrideSetter{
	setPortOverrideStrings,
	setPortOverrideAssignment,
	setPortOverrideSpeed,
	setPortOverrideWANHealth,
	setPortOverrideCounters,
	setPortOverrideRates,
	setPortOverrideMedia,
	setPortOverrideLinkState,
	setPortOverrideDisabled,
}

// portCounterOverrides lists counter fields that can be supplied by config or
// observation.
var portCounterOverrides = []portCounterOverride{
	{key: portCounterRXBytes},
	{key: portCounterTXBytes},
	{key: portCounterRXPackets},
	{key: portCounterTXPackets},
	{key: portCounterRXErrors},
	{key: portCounterTXErrors},
}

// portOverrideStringFields is ordered so fields that depend on speed can be
// applied after speed-derived defaults.
var portOverrideStringFields = []portOverrideStringField{
	{key: portOverrideTextName, normalizer: portOverrideTrimSpace},
	{key: portOverrideTextInterface, normalizer: portOverrideTrimSpace, validator: portOverrideValidateInterface},
	{key: portOverrideTextMAC, normalizer: portOverrideLowerTrimmed, validator: portOverrideValidateMAC},
	{key: portOverrideTextIP, normalizer: portOverrideTrimSpace, validator: portOverrideValidateIPv4},
	{key: portOverrideTextNetmask, normalizer: portOverrideTrimSpace, validator: portOverrideValidateNetmask},
	{key: portOverrideTextRole, normalizer: portOverrideNormalizeRole, validator: portOverrideValidateRole},
	{key: portOverrideTextNetworkGroup, normalizer: portOverrideNormalizeNetworkGroup, validator: portOverrideValidateNetworkGroup},
	{key: portOverrideTextMedia, normalizer: portOverrideTrimSpace, applyAfterSpeed: true},
	{key: portOverrideTextPortConfID, normalizer: portOverrideTrimSpace},
	{key: portOverrideTextNetworkConfID, normalizer: portOverrideTrimSpace},
	{key: portOverrideTextNativeNetworkConfID, normalizer: portOverrideTrimSpace},
	{key: portOverrideTextNetworkName, normalizer: portOverrideTrimSpace},
}

package device

const (
	profileProtocolNoBinding    = "no exact built-in profile, model, version, and firmware evidence binding"
	protocolBehaviorUnsupported = "unsupported"
)

type protocolBehavior uint8

const (
	protocolBehaviorNone protocolBehavior = iota
	protocolBehaviorGatewayUXG
	protocolBehaviorSwitchMainlineBCM
)

func (behavior protocolBehavior) id() string {
	switch behavior {
	case protocolBehaviorGatewayUXG:
		return "gateway-uxg"
	case protocolBehaviorSwitchMainlineBCM:
		return "switch-mainline-bcm"
	default:
		return protocolBehaviorUnsupported
	}
}

func (behavior protocolBehavior) supported() bool {
	return behavior == protocolBehaviorGatewayUXG || behavior == protocolBehaviorSwitchMainlineBCM
}

type userProfileProtocolKey struct {
	sourceType string
	name       string
	model      string
	version    string
}

type userProfileProtocolBinding struct {
	evidenceID             string
	behavior               protocolBehavior
	firmwareMinimumSeconds int
	blockSelection         bool
	controllerClock        bool
}

// userProfileProtocolBindings binds user-facing built-ins to exact evidence
// records. Missing entries, external profiles, and identity overrides remain
// unsupported instead of inheriting behavior from a related product family.
var userProfileProtocolBindings = map[userProfileProtocolKey]userProfileProtocolBinding{
	{sourceTypeBuiltIn, "ucg-fiber", "UCGF", "5.0.16"}: {
		evidenceID: "DT-004",
	},
	{sourceTypeBuiltIn, "ugw3", "UGW3", "4.4.57.5578372"}: {
		evidenceID: "DT-007",
	},
	{sourceTypeBuiltIn, "us16p150", "US16P150", defaultSwitchVersion}: mainlineSwitchProtocolBinding,
	{sourceTypeBuiltIn, "us24p250", "US24P250", defaultSwitchVersion}: mainlineSwitchProtocolBinding,
	{sourceTypeBuiltIn, "us48p500", "US48P500", defaultSwitchVersion}: mainlineSwitchProtocolBinding,
	{sourceTypeBuiltIn, "us8", "US8", defaultSwitchVersion}:           mainlineSwitchProtocolBinding,
	{sourceTypeBuiltIn, "us8p60", "US8P60", defaultSwitchVersion}:     mainlineSwitchProtocolBinding,
	{sourceTypeBuiltIn, "uxg-lite", "UXG", defaultGatewayVersion}: {
		evidenceID: "DT-006",
	},
	{sourceTypeBuiltIn, "uxgpro", "UXGPRO", defaultGatewayVersion}: {
		evidenceID:             "DT-001",
		behavior:               protocolBehaviorGatewayUXG,
		firmwareMinimumSeconds: 1,
		blockSelection:         true,
		controllerClock:        true,
	},
}

var mainlineSwitchProtocolBinding = userProfileProtocolBinding{
	evidenceID:             "DT-008",
	behavior:               protocolBehaviorSwitchMainlineBCM,
	firmwareMinimumSeconds: 30,
	controllerClock:        true,
}

func resolveUserProfileProtocol(profile Profile, effectiveModel, effectiveVersion string) (userProfileProtocolBinding, bool) {
	binding, ok := userProfileProtocolBindings[userProfileProtocolKey{
		sourceType: profile.SourceType,
		name:       profile.Name,
		model:      effectiveModel,
		version:    effectiveVersion,
	}]
	if !ok {
		return userProfileProtocolBinding{}, false
	}
	return binding, true
}

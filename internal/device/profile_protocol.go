package device

const profileProtocolNoBinding = "no exact built-in profile, model, version, and firmware evidence binding"

type userProfileProtocolKey struct {
	sourceType string
	name       string
	model      string
	version    string
}

type userProfileProtocolBinding struct {
	evidenceID     string
	minimumSeconds int
}

// userProfileProtocolBindings binds user-facing built-ins to exact evidence
// records. Missing entries, external profiles, and identity overrides remain
// unsupported instead of inheriting behavior from a related product family.
var userProfileProtocolBindings = map[userProfileProtocolKey]userProfileProtocolBinding{
	{sourceTypeBuiltIn, "ucg-fiber", "UCGF", "5.0.16"}:                {evidenceID: "DT-004"},
	{sourceTypeBuiltIn, "ugw3", "UGW3", "4.4.57.5578372"}:             {evidenceID: "DT-007"},
	{sourceTypeBuiltIn, "us16p150", "US16P150", defaultSwitchVersion}: {evidenceID: protocolEvidenceDT008, minimumSeconds: 30},
	{sourceTypeBuiltIn, "us24p250", "US24P250", defaultSwitchVersion}: {evidenceID: protocolEvidenceDT008, minimumSeconds: 30},
	{sourceTypeBuiltIn, "us48p500", "US48P500", defaultSwitchVersion}: {evidenceID: protocolEvidenceDT008, minimumSeconds: 30},
	{sourceTypeBuiltIn, "us8", "US8", defaultSwitchVersion}:           {evidenceID: protocolEvidenceDT008, minimumSeconds: 30},
	{sourceTypeBuiltIn, "us8p60", "US8P60", defaultSwitchVersion}:     {evidenceID: protocolEvidenceDT008, minimumSeconds: 30},
	{sourceTypeBuiltIn, "uxg-lite", "UXG", defaultGatewayVersion}:     {evidenceID: "DT-006"},
	{sourceTypeBuiltIn, "uxgpro", "UXGPRO", defaultGatewayVersion}:    {evidenceID: protocolEvidenceDT001, minimumSeconds: 1},
}

func resolveUserProfileProtocol(profile Profile, effectiveModel, effectiveVersion string) (ProtocolProfile, userProfileProtocolBinding, bool) {
	binding, ok := userProfileProtocolBindings[userProfileProtocolKey{
		sourceType: profile.SourceType,
		name:       profile.Name,
		model:      effectiveModel,
		version:    effectiveVersion,
	}]
	if !ok {
		return ProtocolProfile{}, userProfileProtocolBinding{}, false
	}
	protocolProfile, ok := lookupProtocolProfileByEvidenceID(binding.evidenceID)
	if !ok {
		return ProtocolProfile{}, binding, false
	}
	return protocolProfile, binding, true
}

func lookupProtocolProfileByEvidenceID(evidenceID string) (ProtocolProfile, bool) {
	for _, profile := range protocolProfileDefinitions {
		if profile.EvidenceID == evidenceID {
			return profile, true
		}
	}
	return ProtocolProfile{}, false
}

package device

const (
	protocolVersionGateway5016 = "5.0.16"
	protocolVersionCurrent218  = "2.1.8+971"
	protocolUSRBINMCAD         = "/usr/bin/mcad"
	protocolBINMCAD            = "/bin/mcad"
	protocolXtensaAgent        = "mcagent-current-app"
	protocolMainlineFirmware   = "50c167b60f9f1f575f95ccb5f193bdc02ae62643c1eb720c9a41959b80fd8e07"
	protocolFlexMiniFirmware   = "40c585ee96e10eff782e37f033b8f8124e070e5b761d3e73931aa546a2bd25da"
)

const (
	protocolEvidenceDT001       = "DT-001"
	protocolEvidenceDT002       = "DT-002"
	protocolEvidenceDT008       = "DT-008"
	protocolEvidenceDT009       = "DT-009"
	protocolEvidenceDT010       = "DT-010"
	protocolEvidenceDT012       = "DT-012"
	protocolEvidenceDT013       = "DT-013"
	protocolEvidenceDT014       = "DT-014"
	protocolEvidenceDT015       = "DT-015"
	protocolEvidenceUnsupported = "DT-003,DT-004,DT-005,DT-006,DT-007,DT-011"
	protocolEvidenceGateways    = "DT-001,DT-002"
	protocolEvidenceZlib        = "DT-001,DT-002,DT-008,DT-009,DT-010"
	protocolEvidenceMainline    = "DT-008,DT-009"
	protocolEvidenceCurrentESP  = "DT-012,DT-013,DT-014,DT-015"
	protocolEvidenceFallback    = "DT-013,DT-014,DT-015"
	protocolEvidencePSE         = "DT-013,DT-015"
	protocolEvidenceNoDiscovery = "DT-001,DT-002,DT-003,DT-004,DT-005,DT-006,DT-007,DT-011"
)

const protocolPrincipleUnknown = "UNKNOWN protocol semantics select no active component"

var protocolPolicyComponentDefinitions = []ProtocolPolicyComponent{
	policyComponent(ProtocolPolicyAxisWire, string(WireUnsupported), protocolEvidenceUnsupported, protocolPrincipleUnknown),
	policyComponent(ProtocolPolicyAxisWire, string(WireTNBUZlib), protocolEvidenceZlib, "Use only the common safe zlib TNBU subset; do not enable adjacent codecs"),
	policyComponent(ProtocolPolicyAxisWire, string(WireTNBURaw), protocolEvidenceDT012, "Current usmini2 sends raw JSON before TNBU encryption"),
	policyComponent(ProtocolPolicyAxisWire, string(WireTNBUZlibRawFallback), protocolEvidenceFallback, "Current Ultra and M25G apps share zlib with raw fallback"),

	policyComponent(ProtocolPolicyAxisDiscovery, string(DiscoveryUnsupported), protocolEvidenceNoDiscovery, "Exact gateway discovery remains UNKNOWN and unsupported identities activate none"),
	policyComponent(ProtocolPolicyAxisDiscovery, string(DiscoveryMainlineV2), protocolEvidenceMainline, "BCM and MSCC share the exact mainline v2 builder contract"),
	policyComponent(ProtocolPolicyAxisDiscovery, string(DiscoverySwitchWANV2), protocolEvidenceDT010, "Switch-WAN keeps its distinct alternate-builder contract"),
	policyComponent(ProtocolPolicyAxisDiscovery, string(DiscoveryLowResourceV2), protocolEvidenceCurrentESP, "Current ESP apps use v2 subnet-directed broadcast at startup"),

	policyComponent(ProtocolPolicyAxisFieldGroups, string(FieldGroupsUnsupported), protocolEvidenceUnsupported, protocolPrincipleUnknown),
	policyComponent(ProtocolPolicyAxisFieldGroups, string(FieldGroupsGatewayUXG), protocolEvidenceDT001, "UXG producer telemetry extends the gateway identity and management groups"),
	policyComponent(ProtocolPolicyAxisFieldGroups, string(FieldGroupsGatewayUDM), protocolEvidenceDT002, "UDM saved-config and probe fields extend its gateway groups"),
	policyComponent(ProtocolPolicyAxisFieldGroups, string(FieldGroupsSwitchMainline), protocolEvidenceMainline, "BCM and MSCC share the evidenced mainline switch field denominator"),
	policyComponent(ProtocolPolicyAxisFieldGroups, string(FieldGroupsSwitchWAN), protocolEvidenceDT010, "Switch-WAN has a separate WAN and RPS field vocabulary"),
	policyComponent(ProtocolPolicyAxisFieldGroups, string(FieldGroupsSwitchLowResource), protocolEvidenceCurrentESP, "Current ESP apps share a low-resource switch field denominator"),

	policyComponent(ProtocolPolicyAxisCommands, string(CommandsUnsupported), protocolEvidenceUnsupported, protocolPrincipleUnknown),
	policyComponent(ProtocolPolicyAxisCommands, string(CommandsLinuxTop7Direct70), protocolEvidenceDT001, "UXG exposes seven response types and 70 direct names with host effects rejected"),
	policyComponent(ProtocolPolicyAxisCommands, string(CommandsLinuxTop7Direct65), protocolEvidenceDT002, "UDM exposes seven response types and its distinct 65-name set"),
	policyComponent(ProtocolPolicyAxisCommands, string(CommandsLinuxTop7Direct40), protocolEvidenceMainline, "BCM and MSCC share the exact 40-name mainline set"),
	policyComponent(ProtocolPolicyAxisCommands, string(CommandsLinuxTop7Direct41), protocolEvidenceDT010, "Switch-WAN adds one evidenced name while host mutation remains rejected"),
	policyComponent(ProtocolPolicyAxisCommands, string(CommandsLowResourceTop5), protocolEvidenceCurrentESP, "Only the five proven top-level MCU response types are enabled; direct cmd stays unsupported"),

	policyComponent(ProtocolPolicyAxisCapabilities, protocolUnsupportedValue, protocolEvidenceUnsupported, protocolPrincipleUnknown),
	policyComponent(ProtocolPolicyAxisCapabilities, capabilitySelectorGatewayInform, protocolEvidenceGateways, "The two analysed gateways share management and Inform vocabulary"),
	policyComponent(ProtocolPolicyAxisCapabilities, capabilitySelectorGatewayProducerTelemetry, protocolEvidenceDT001, "UXG alone carries the reviewed producer-bound telemetry overlay"),
	policyComponent(ProtocolPolicyAxisCapabilities, capabilitySelectorGatewaySavedConfigProbes, protocolEvidenceDT002, "UDM alone carries the reviewed saved-config and probe overlay"),
	policyComponent(ProtocolPolicyAxisCapabilities, capabilitySelectorControllerClock, "DT-001,DT-008,DT-009", "Only exact reviewed UXG Pro and mainline switch builds apply server time as a payload-only offset"),
	policyComponent(ProtocolPolicyAxisCapabilities, capabilitySelectorSwitchInform, "DT-008,DT-009,DT-010,DT-012,DT-013,DT-014,DT-015", "Analysed switch agents share bounded switch management vocabulary"),
	policyComponent(ProtocolPolicyAxisCapabilities, capabilitySelectorSwitchMainline, protocolEvidenceMainline, "BCM and MSCC share a positive mainline capability base"),
	policyComponent(ProtocolPolicyAxisCapabilities, capabilitySelectorSwitchBCMTables, protocolEvidenceDT008, "Only BCM has positive network and IGMP table vocabulary"),
	policyComponent(ProtocolPolicyAxisCapabilities, capabilitySelectorSwitchMSCCPlatform, protocolEvidenceDT009, "MSCC retains its distinct positive platform contract"),
	policyComponent(ProtocolPolicyAxisCapabilities, capabilitySelectorSwitchWANVocabulary, protocolEvidenceDT010, "WAN and RPS vocabulary is source-gated and grants no action authority"),
	policyComponent(ProtocolPolicyAxisCapabilities, capabilitySelectorSwitchLowResource, protocolEvidenceCurrentESP, "Current ESP apps share a positive low-resource switch base"),
	policyComponent(ProtocolPolicyAxisCapabilities, capabilitySelectorSwitchCurrentESP, protocolEvidenceCurrentESP, "Only the four reviewed current ESP apps select current-ESP behavior"),
	policyComponent(ProtocolPolicyAxisCapabilities, capabilitySelectorSwitchPSEVocabulary, protocolEvidencePSE, "Only Ultra and M25G8 have positive reviewed PSE vocabulary"),
	policyComponent(ProtocolPolicyAxisCapabilities, capabilitySelectorSwitchRTL8238BVocabulary, protocolEvidenceDT013, "Ultra carries source-gated RTL8238B state vocabulary"),
	policyComponent(ProtocolPolicyAxisCapabilities, capabilitySelectorSwitchRTL8239Vocabulary, protocolEvidenceDT015, "M25G8 carries source-gated RTL8239MCU vocabulary"),

	policyComponent(ProtocolPolicyAxisPresence, string(PresenceUnsupported), protocolEvidenceUnsupported, protocolPrincipleUnknown),
	policyComponent(ProtocolPolicyAxisPresence, string(PresenceGatewayUXG), protocolEvidenceDT001, "UXG fields remain producer and request gated"),
	policyComponent(ProtocolPolicyAxisPresence, string(PresenceGatewayUDM), protocolEvidenceDT002, "UDM saved-config and probe fields remain state and source gated"),
	policyComponent(ProtocolPolicyAxisPresence, string(PresenceSwitchMainlineBCM), protocolEvidenceDT008, "BCM-only tables require their typed local sources"),
	policyComponent(ProtocolPolicyAxisPresence, string(PresenceSwitchMainlineMSCC), protocolEvidenceDT009, "MSCC omits the unproven BCM-only table overlay"),
	policyComponent(ProtocolPolicyAxisPresence, string(PresenceSwitchWAN), protocolEvidenceDT010, "WAN and RPS fields require model and local source gates"),
	policyComponent(ProtocolPolicyAxisPresence, string(PresenceLowResourceUSMini2), protocolEvidenceDT012, "usmini2 fields require their reviewed typed sources"),
	policyComponent(ProtocolPolicyAxisPresence, string(PresenceLowResourceUltra), protocolEvidenceDT013, "Ultra PSE and RTL8238B fields require model and provider gates"),
	policyComponent(ProtocolPolicyAxisPresence, string(PresenceLowResourceM25G5), protocolEvidenceDT014, "M25G5 exposes only sourced fields and makes no hardware-absence claim"),
	policyComponent(ProtocolPolicyAxisPresence, string(PresenceLowResourceM25G8), protocolEvidenceDT015, "M25G8 PSE and RTL8239MCU fields require model and provider gates"),
}

var protocolPolicyComponentsByKey = buildProtocolPolicyComponentRegistry(protocolPolicyComponentDefinitions)

var protocolBehaviorDefinitions = []ProtocolBehaviorProfile{
	behaviorDefinition(
		ProtocolBehaviorUnsupported,
		protocolEvidenceUnsupported,
		"Exact identities with UNKNOWN protocol semantics remain wholly unsupported",
		WireUnsupported,
		DiscoveryUnsupported,
		FieldGroupsUnsupported,
		CommandsUnsupported,
		CapabilityUnsupported,
		PresenceUnsupported,
	),
	behaviorDefinition(
		ProtocolBehaviorGatewayUXG,
		protocolEvidenceDT001,
		"Compose only the exact reviewed UXG gateway contracts",
		WireTNBUZlib,
		DiscoveryUnsupported,
		FieldGroupsGatewayUXG,
		CommandsLinuxTop7Direct70,
		CapabilityGatewayInform|CapabilityGatewayProducerTelemetry|CapabilityControllerClock,
		PresenceGatewayUXG,
	),
	behaviorDefinition(
		ProtocolBehaviorGatewayUDM,
		protocolEvidenceDT002,
		"Compose only the exact reviewed UDM gateway contracts",
		WireTNBUZlib,
		DiscoveryUnsupported,
		FieldGroupsGatewayUDM,
		CommandsLinuxTop7Direct65,
		CapabilityGatewayInform|CapabilityGatewaySavedConfigProbes,
		PresenceGatewayUDM,
	),
	behaviorDefinition(
		ProtocolBehaviorSwitchMainlineBCM,
		protocolEvidenceDT008,
		"Reuse mainline semantics and add only positive BCM overlays",
		WireTNBUZlib,
		DiscoveryMainlineV2,
		FieldGroupsSwitchMainline,
		CommandsLinuxTop7Direct40,
		CapabilitySwitchInform|CapabilitySwitchMainline|CapabilitySwitchBCMTables|CapabilityControllerClock,
		PresenceSwitchMainlineBCM,
	),
	behaviorDefinition(
		ProtocolBehaviorSwitchMainlineMSCC,
		protocolEvidenceDT009,
		"Reuse mainline semantics and add only positive MSCC overlays",
		WireTNBUZlib,
		DiscoveryMainlineV2,
		FieldGroupsSwitchMainline,
		CommandsLinuxTop7Direct40,
		CapabilitySwitchInform|CapabilitySwitchMainline|CapabilitySwitchMSCCPlatform|CapabilityControllerClock,
		PresenceSwitchMainlineMSCC,
	),
	behaviorDefinition(
		ProtocolBehaviorSwitchWAN,
		protocolEvidenceDT010,
		"Keep the evidenced WAN builder and vocabulary distinct from mainline",
		WireTNBUZlib,
		DiscoverySwitchWANV2,
		FieldGroupsSwitchWAN,
		CommandsLinuxTop7Direct41,
		CapabilitySwitchInform|CapabilitySwitchWANVocabulary,
		PresenceSwitchWAN,
	),
	behaviorDefinition(
		ProtocolBehaviorLowResourceUSMini2,
		protocolEvidenceDT012,
		"Reuse current ESP components while retaining raw usmini2 Inform",
		WireTNBURaw,
		DiscoveryLowResourceV2,
		FieldGroupsSwitchLowResource,
		CommandsLowResourceTop5,
		CapabilitySwitchInform|CapabilitySwitchLowResource|CapabilitySwitchCurrentESP,
		PresenceLowResourceUSMini2,
	),
	behaviorDefinition(
		ProtocolBehaviorLowResourceUltra,
		protocolEvidenceDT013,
		"Reuse current ESP components with only Ultra positive overlays",
		WireTNBUZlibRawFallback,
		DiscoveryLowResourceV2,
		FieldGroupsSwitchLowResource,
		CommandsLowResourceTop5,
		CapabilitySwitchInform|CapabilitySwitchLowResource|CapabilitySwitchCurrentESP|
			CapabilitySwitchPSEVocabulary|CapabilitySwitchRTL8238BVocabulary,
		PresenceLowResourceUltra,
	),
	behaviorDefinition(
		ProtocolBehaviorLowResourceM25G5,
		protocolEvidenceDT014,
		"Reuse current ESP components without inferring unknown PSE hardware",
		WireTNBUZlibRawFallback,
		DiscoveryLowResourceV2,
		FieldGroupsSwitchLowResource,
		CommandsLowResourceTop5,
		CapabilitySwitchInform|CapabilitySwitchLowResource|CapabilitySwitchCurrentESP,
		PresenceLowResourceM25G5,
	),
	behaviorDefinition(
		ProtocolBehaviorLowResourceM25G8,
		protocolEvidenceDT015,
		"Reuse current ESP components with only M25G8 positive overlays",
		WireTNBUZlibRawFallback,
		DiscoveryLowResourceV2,
		FieldGroupsSwitchLowResource,
		CommandsLowResourceTop5,
		CapabilitySwitchInform|CapabilitySwitchLowResource|CapabilitySwitchCurrentESP|
			CapabilitySwitchPSEVocabulary|CapabilitySwitchRTL8239MCUVocabulary,
		PresenceLowResourceM25G8,
	),
}

var protocolBehaviorsByID = buildProtocolBehaviorRegistry(
	protocolBehaviorDefinitions,
	protocolPolicyComponentsByKey,
)

var protocolProfileDefinitions = []ProtocolProfile{
	protocolDefinition("DT-001", ProtocolFamilyUXG, ProtocolPlatformAArch64LE, protocolUSRBINMCAD, protocolVersionGateway5016,
		"UXGPRO.al324.v5.0.16.9d45777.260226.1635",
		"18a7f198f71edc0161365114356239b0b370b4b90f664bb90f253b33f8b5658c",
		"f16afc048fa0085478648b64f154faa129485b153a0d93d4f590d07aad286455", ProtocolDigestBinarySHA256,
		ProtocolSupported, ProtocolBehaviorGatewayUXG),
	protocolDefinition("DT-002", ProtocolFamilyUDM, ProtocolPlatformAArch64LE, protocolUSRBINMCAD, "5.1.19",
		"UDMPROSE.al324.v5.1.19.3fbc1da.260613.0944",
		"6253fe20ba75d159a113715f3a2080bc4f30d0208dac2528c6030f01c8d89305",
		"8ad4076fe2475926ed5d2fe21dec66982a4f432d08bf2cac709b84a6c00fa92b", ProtocolDigestBinarySHA256,
		ProtocolSupported, ProtocolBehaviorGatewayUDM),
	unsupportedProtocolDefinition("DT-003", ProtocolFamilyUDM, ProtocolPlatformAArch64LE, protocolUSRBINMCAD, protocolVersionGateway5016,
		"UDMPROSE.al324.v5.0.16.238fde6.260227.0037",
		"7cf58f4563522220716f5025a7b2954b070df6be9364d7f60af0bc644512bce4",
		"35852806c912bd3cbe4f7a66ace7f8c3eff538de5d629acd04a59cb6812a505b", ProtocolDigestBinarySHA256),
	unsupportedProtocolDefinition("DT-004", ProtocolFamilyUDM, ProtocolPlatformAArch64LE, protocolUSRBINMCAD, protocolVersionGateway5016,
		"UCGF.ipq9574.v5.0.16.238fde6.260227.0038",
		"7c8635974513413f19b4542c85b188d93a2fef38707ad53d237ce6a657e88ce5",
		"72fbc46198b49a9b376be4797e48816b0430fdf3c3666234cd27214b0ec7515c", ProtocolDigestBinarySHA256),
	unsupportedProtocolDefinition("DT-005", ProtocolFamilyUDM, ProtocolPlatformUnknown, "firmware-artifact", "5.1.19",
		"0b7a-UCGF-5.1.19-aa9acd86-1e16-479b-9a84-cf7e21e3662a.bin",
		"c671c41d6b9d769b6633974d8823803bfe6aa151f917833628ac0d1d8f2b9f8b", "", ProtocolDigestUnknown),
	unsupportedProtocolDefinition("DT-006", ProtocolFamilyUXG, ProtocolPlatformAArch64LE, protocolUSRBINMCAD, protocolVersionGateway5016,
		"UXG.ipq5018.v5.0.16.9d45777.260226.1635",
		"e2e361fc9b4296628f1b4fa10280449695c9df8fb8da95a440a7462044a4765c",
		"ffcd46b9fc1baee16b15982a8592be1db0438879007fe39e4721da51e09b75cd", ProtocolDigestBinarySHA256),
	unsupportedProtocolDefinition("DT-007", ProtocolFamilyLegacyGateway, ProtocolPlatformMIPS32BE, "legacy-agent-unknown", "4.4.57",
		"v4.4.57.5578372.230112.0823",
		"08a35a626e9733018b2e49af92aa3474255136bb0178e0697f44c6d8042cdd74", "", ProtocolDigestUnknown),
	protocolDefinition("DT-008", ProtocolFamilyMainlineBCM, ProtocolPlatformARM32LEEABI5, protocolBINMCAD, "7.4.1+16850.260311.1748",
		"UBNTUS.bcm5334x.v7.4.1+16850.260311.1748",
		protocolMainlineFirmware,
		"bc2185d192d0a01504285b9d7c931f86c91bc7ddf0dc624fd6edf7d18cb30a8d", ProtocolDigestBinarySHA256,
		ProtocolSupported, ProtocolBehaviorSwitchMainlineBCM),
	protocolDefinition("DT-009", ProtocolFamilyMainlineMSCC, ProtocolPlatformMIPS32R2LE, protocolBINMCAD, "7.4.1+16850.260311.1155",
		"UBNTUS.mscc7514.v7.4.1+16850.260311.1155",
		protocolMainlineFirmware,
		"8d6fb75cc0365625d43a8ed4be913806c349e6e88d52af5979b4847176f233e8", ProtocolDigestBinarySHA256,
		ProtocolSupported, ProtocolBehaviorSwitchMainlineMSCC),
	protocolDefinition("DT-010", ProtocolFamilySwitchWAN, ProtocolPlatformMIPS32R2BE, protocolBINMCAD, "7.3.109+16640",
		"UBNTWRS.rtl93xx.v7.3.109+16640.250909.1235",
		"2426c467d54188b936a712ac5746f026def653a0547e4feefe3254cf769fb77e",
		"49760ff4d26dcea80180865c009059976653da595ff88f24f0fd9c1235a62725", ProtocolDigestBinarySHA256,
		ProtocolSupported, ProtocolBehaviorSwitchWAN),
	unsupportedProtocolDefinition("DT-011", ProtocolFamilyLowResource, ProtocolPlatformCortexMThumbLE, "m487-application", "2.1.6+1012",
		"USMINI.multi_2.1.6+1012.250807.1806.bin",
		protocolFlexMiniFirmware,
		"8b4ad113d507eb5d47750ed4c7bb125cd1675d047d97b2489ebacfef88ee5a2f", ProtocolDigestBinarySHA256),
	protocolDefinition("DT-012", ProtocolFamilyLowResource, ProtocolPlatformXtensaESP32LE, protocolXtensaAgent, "2.1.6+1012",
		"v2.1.6-0-g2a76",
		protocolFlexMiniFirmware,
		"3f42cb51f525d8cb2c96f0463da7a9bdf83f2ab38bb39490156a409ef3b997d4", ProtocolDigestEmbeddedAppSHA256,
		ProtocolSupported, ProtocolBehaviorLowResourceUSMini2),
	protocolDefinition("DT-013", ProtocolFamilyLowResource, ProtocolPlatformXtensaESP32LE, protocolXtensaAgent, protocolVersionCurrent218,
		"USM8P.esp32.v2.1.8.g7d7e.250320.1505",
		"fee50cb3ba72e9129100549d4d61945e6410407b6e25171a847f6acb768e41cb",
		"6d85ae4168e4923fc4bd34091f2627fb0e8dbce7e0b0047375da96d40128e491", ProtocolDigestEmbeddedAppSHA256,
		ProtocolSupported, ProtocolBehaviorLowResourceUltra),
	protocolDefinition("DT-014", ProtocolFamilyLowResource, ProtocolPlatformXtensaESP32LE, protocolXtensaAgent, protocolVersionCurrent218,
		"USM25G5.esp32.v2.1.8.g7d7e.250320.1507",
		"8c6c46b790c899acb489313e79baec4b3a54110c7722a353e18a1e2cd004b59f",
		"78427392e9d612db6d41a317c060cebb3e28e4b3e122c90f4710db7e94540a57", ProtocolDigestEmbeddedAppSHA256,
		ProtocolSupported, ProtocolBehaviorLowResourceM25G5),
	protocolDefinition("DT-015", ProtocolFamilyLowResource, ProtocolPlatformXtensaESP32LE, protocolXtensaAgent, protocolVersionCurrent218,
		"USM25G8.esp32.v2.1.8.g7d7e.250320.1510",
		"c858dd090c1a76be4a3ab1db2a2fcf0b8384ef83d5be336a6f45cbd0176c387d",
		"73fb7e3a20dc4ed52ac7d4815c49fb6b62d2a621b3084f974b1f07560cd1b71a", ProtocolDigestEmbeddedAppSHA256,
		ProtocolSupported, ProtocolBehaviorLowResourceM25G8),
}

var protocolProfilesByBuild = buildProtocolRegistry(protocolProfileDefinitions, protocolBehaviorsByID)

func policyComponent(
	axis ProtocolPolicyAxis,
	selector string,
	evidenceRecords string,
	principle string,
) ProtocolPolicyComponent {
	return ProtocolPolicyComponent{
		Axis:            axis,
		Selector:        selector,
		EvidenceRecords: evidenceRecords,
		Principle:       principle,
	}
}

func behaviorDefinition(
	id ProtocolBehaviorID,
	evidenceRecords string,
	principle string,
	wire WirePolicy,
	discovery DiscoveryPolicy,
	fieldGroups FieldGroupsPolicy,
	commands CommandPolicy,
	capabilities CapabilityPolicy,
	presence PresencePolicy,
) ProtocolBehaviorProfile {
	return ProtocolBehaviorProfile{
		ID:              id,
		EvidenceRecords: evidenceRecords,
		Principle:       principle,
		Policies: ProtocolPolicies{
			Wire:         wire,
			Discovery:    discovery,
			FieldGroups:  fieldGroups,
			Commands:     commands,
			Capabilities: capabilities,
			Presence:     presence,
		},
	}
}

func protocolDefinition(
	evidenceID string,
	family ProtocolFamily,
	platform ProtocolPlatform,
	componentRole string,
	version string,
	buildMarker string,
	firmwareSHA256 string,
	componentSHA256 string,
	digestKind ProtocolDigestKind,
	support ProtocolSupport,
	behaviorID ProtocolBehaviorID,
) ProtocolProfile {
	return ProtocolProfile{
		EvidenceID: evidenceID,
		Build: ProtocolBuild{
			Family:          family,
			Platform:        platform,
			ComponentRole:   componentRole,
			Version:         version,
			BuildMarker:     buildMarker,
			FirmwareSHA256:  firmwareSHA256,
			ComponentSHA256: componentSHA256,
			DigestKind:      digestKind,
		},
		Support:    support,
		BehaviorID: behaviorID,
	}
}

func unsupportedProtocolDefinition(
	evidenceID string,
	family ProtocolFamily,
	platform ProtocolPlatform,
	componentRole string,
	version string,
	buildMarker string,
	firmwareSHA256 string,
	componentSHA256 string,
	digestKind ProtocolDigestKind,
) ProtocolProfile {
	return protocolDefinition(
		evidenceID,
		family,
		platform,
		componentRole,
		version,
		buildMarker,
		firmwareSHA256,
		componentSHA256,
		digestKind,
		ProtocolUnsupported,
		ProtocolBehaviorUnsupported,
	)
}

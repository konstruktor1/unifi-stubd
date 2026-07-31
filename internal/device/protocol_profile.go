package device

import (
	"encoding/hex"
	"fmt"
	"strings"
)

// ProtocolFamily identifies one reverse-engineered device implementation
// family. It is deliberately narrower than the controller-facing device type.
type ProtocolFamily string

// ProtocolPlatform is a canonical architecture class taken from the static
// device evidence matrix.
type ProtocolPlatform string

// ProtocolDigestKind describes what the component digest authenticates.
type ProtocolDigestKind string

// ProtocolSupport is the fail-closed support decision for one exact build.
type ProtocolSupport string

// ProtocolBehaviorID identifies a reusable, version-independent composition.
type ProtocolBehaviorID string

// WirePolicy identifies the TNBU and request-compression contract.
type WirePolicy string

// DiscoveryPolicy identifies one evidence-bounded discovery contract.
type DiscoveryPolicy string

// FieldGroupsPolicy identifies a reusable Inform field-group composition.
type FieldGroupsPolicy string

// CommandPolicy identifies the response and command vocabulary contract.
type CommandPolicy string

// CapabilityPolicy composes narrow, positive protocol capabilities as flags.
// Absence of a flag is not a claim that the physical hardware lacks a feature.
type CapabilityPolicy uint32

// PresencePolicy identifies source- and state-dependent field-presence rules.
type PresencePolicy string

// ProtocolPolicyAxis names one independently composed behavior axis.
type ProtocolPolicyAxis string

const protocolUnsupportedValue = "unsupported"

const (
	// ProtocolFamilyUXG is the standalone UXG implementation family.
	ProtocolFamilyUXG ProtocolFamily = "uxg"
	// ProtocolFamilyUDM is the integrated UDM/UCG implementation family.
	ProtocolFamilyUDM ProtocolFamily = "udm"
	// ProtocolFamilyMainlineBCM is the BCM mainline switch implementation.
	ProtocolFamilyMainlineBCM ProtocolFamily = "mainline-bcm"
	// ProtocolFamilyMainlineMSCC is the MSCC mainline switch implementation.
	ProtocolFamilyMainlineMSCC ProtocolFamily = "mainline-mscc"
	// ProtocolFamilySwitchWAN is the WRS/RTL93xx switch-WAN implementation.
	ProtocolFamilySwitchWAN ProtocolFamily = "switch-wan"
	// ProtocolFamilyLowResource is the M487 or current Xtensa switch family.
	ProtocolFamilyLowResource ProtocolFamily = "low-resource"
	// ProtocolFamilyLegacyGateway is the separate legacy UGW implementation.
	ProtocolFamilyLegacyGateway ProtocolFamily = "legacy-gateway"
)

const (
	// ProtocolPlatformUnknown records an artifact whose architecture is unknown.
	ProtocolPlatformUnknown ProtocolPlatform = "unknown"
	// ProtocolPlatformAArch64LE is the matrix's AArch64 little-endian class.
	ProtocolPlatformAArch64LE ProtocolPlatform = "aarch64-little-endian"
	// ProtocolPlatformMIPS32BE is the matrix's legacy MIPS32 big-endian class.
	ProtocolPlatformMIPS32BE ProtocolPlatform = "mips32-big-endian"
	// ProtocolPlatformARM32LEEABI5 is the mainline ARM32 LE EABI5 class.
	ProtocolPlatformARM32LEEABI5 ProtocolPlatform = "arm32-little-endian-eabi5"
	// ProtocolPlatformMIPS32R2LE is the mainline MIPS32r2 LE class.
	ProtocolPlatformMIPS32R2LE ProtocolPlatform = "mips32r2-little-endian"
	// ProtocolPlatformMIPS32R2BE is the switch-WAN MIPS32r2 BE class.
	ProtocolPlatformMIPS32R2BE ProtocolPlatform = "mips32r2-big-endian"
	// ProtocolPlatformCortexMThumbLE is the M487 Cortex-M/Thumb LE class.
	ProtocolPlatformCortexMThumbLE ProtocolPlatform = "arm-cortex-m-thumb-little-endian"
	// ProtocolPlatformXtensaESP32LE is the current ESP32 Xtensa LE class.
	ProtocolPlatformXtensaESP32LE ProtocolPlatform = "xtensa-esp32-little-endian"
)

const (
	// ProtocolPolicyAxisWire is the TNBU/request-compression axis.
	ProtocolPolicyAxisWire ProtocolPolicyAxis = "wire"
	// ProtocolPolicyAxisDiscovery is the discovery axis.
	ProtocolPolicyAxisDiscovery ProtocolPolicyAxis = "discovery"
	// ProtocolPolicyAxisFieldGroups is the Inform field-group axis.
	ProtocolPolicyAxisFieldGroups ProtocolPolicyAxis = "field-groups"
	// ProtocolPolicyAxisCommands is the response/command axis.
	ProtocolPolicyAxisCommands ProtocolPolicyAxis = "commands"
	// ProtocolPolicyAxisCapabilities is the positive capability axis.
	ProtocolPolicyAxisCapabilities ProtocolPolicyAxis = "capabilities"
	// ProtocolPolicyAxisPresence is the source/state presence axis.
	ProtocolPolicyAxisPresence ProtocolPolicyAxis = "presence"
)

const (
	// ProtocolDigestBinarySHA256 is a SHA-256 over exact component bytes.
	ProtocolDigestBinarySHA256 ProtocolDigestKind = "binary-sha256"
	// ProtocolDigestEmbeddedAppSHA256 is the digest stored in a current ESP app.
	ProtocolDigestEmbeddedAppSHA256 ProtocolDigestKind = "embedded-app-sha256"
	// ProtocolDigestUnknown records that no agent/component digest was established.
	ProtocolDigestUnknown ProtocolDigestKind = "unknown"
)

const (
	// ProtocolSupported marks a build with an evidence-backed behavior selector.
	ProtocolSupported ProtocolSupport = "supported"
	// ProtocolUnsupported denies protocol use without a neighboring fallback.
	ProtocolUnsupported ProtocolSupport = protocolUnsupportedValue
)

const (
	// ProtocolBehaviorUnsupported selects the closed six-axis composition.
	ProtocolBehaviorUnsupported ProtocolBehaviorID = protocolUnsupportedValue
	// ProtocolBehaviorGatewayUXG selects the proven standalone gateway contract.
	ProtocolBehaviorGatewayUXG ProtocolBehaviorID = "gateway-uxg"
	// ProtocolBehaviorGatewayUDM selects the proven integrated gateway contract.
	ProtocolBehaviorGatewayUDM ProtocolBehaviorID = "gateway-udm"
	// ProtocolBehaviorSwitchMainlineBCM selects mainline plus BCM overlays.
	ProtocolBehaviorSwitchMainlineBCM ProtocolBehaviorID = "switch-mainline-bcm"
	// ProtocolBehaviorSwitchMainlineMSCC selects mainline plus MSCC overlays.
	ProtocolBehaviorSwitchMainlineMSCC ProtocolBehaviorID = "switch-mainline-mscc"
	// ProtocolBehaviorSwitchWAN selects the proven WRS/RTL93xx contract.
	ProtocolBehaviorSwitchWAN ProtocolBehaviorID = "switch-wan"
	// ProtocolBehaviorLowResourceUSMini2 selects the raw-JSON ESP contract.
	ProtocolBehaviorLowResourceUSMini2 ProtocolBehaviorID = "low-resource-usmini2"
	// ProtocolBehaviorLowResourceUltra selects the Ultra ESP overlay.
	ProtocolBehaviorLowResourceUltra ProtocolBehaviorID = "low-resource-ultra"
	// ProtocolBehaviorLowResourceM25G5 selects the M25G5 ESP overlay.
	ProtocolBehaviorLowResourceM25G5 ProtocolBehaviorID = "low-resource-m25g5"
	// ProtocolBehaviorLowResourceM25G8 selects the M25G8 ESP overlay.
	ProtocolBehaviorLowResourceM25G8 ProtocolBehaviorID = "low-resource-m25g8"
)

const (
	// WireUnsupported disables Inform wire behavior for the build.
	WireUnsupported WirePolicy = protocolUnsupportedValue
	// WireTNBUZlib selects the common safe zlib/CBC/GCM subset. A firmware's
	// additional codec branches, such as BCM Snappy, are not enabled by it.
	WireTNBUZlib WirePolicy = "tnbu-zlib"
	// WireTNBURaw selects raw JSON before CBC or GCM TNBU encryption.
	WireTNBURaw WirePolicy = "tnbu-raw"
	// WireTNBUZlibRawFallback selects zlib with a proven raw fallback.
	WireTNBUZlibRawFallback WirePolicy = "tnbu-zlib-raw-fallback"
)

const (
	// DiscoveryUnsupported disables discovery for the behavior composition.
	DiscoveryUnsupported DiscoveryPolicy = protocolUnsupportedValue
	// DiscoveryMainlineV2 selects the shared BCM/MSCC builder and destinations.
	DiscoveryMainlineV2 DiscoveryPolicy = "mainline-v2-multicast-broadcast"
	// DiscoverySwitchWANV2 selects the distinct WAN alternate-builder contract.
	DiscoverySwitchWANV2 DiscoveryPolicy = "switch-wan-v2-multicast-broadcast"
	// DiscoveryLowResourceV2 selects the current ESP startup builder and target.
	DiscoveryLowResourceV2 DiscoveryPolicy = "low-resource-v2-directed-broadcast"
)

const (
	// FieldGroupsUnsupported disables build-specific Inform field groups.
	FieldGroupsUnsupported FieldGroupsPolicy = protocolUnsupportedValue
	// FieldGroupsGatewayUXG selects gateway identity, management, and UXG telemetry.
	FieldGroupsGatewayUXG FieldGroupsPolicy = "gateway-identity-management-uxg-telemetry"
	// FieldGroupsGatewayUDM selects gateway identity, management, and UDM probes.
	FieldGroupsGatewayUDM FieldGroupsPolicy = "gateway-identity-management-udm-probes"
	// FieldGroupsSwitchMainline selects the common mainline switch vocabulary.
	FieldGroupsSwitchMainline FieldGroupsPolicy = "switch-mainline-common"
	// FieldGroupsSwitchWAN selects the WRS switch and WAN/RPS vocabulary.
	FieldGroupsSwitchWAN FieldGroupsPolicy = "switch-wan-rps"
	// FieldGroupsSwitchLowResource selects the common current ESP vocabulary.
	FieldGroupsSwitchLowResource FieldGroupsPolicy = "switch-low-resource-common"
)

const (
	// CommandsUnsupported disables response and command vocabulary selection.
	CommandsUnsupported CommandPolicy = protocolUnsupportedValue
	// CommandsLinuxTop7Direct70 selects the seven response types and 70 names.
	CommandsLinuxTop7Direct70 CommandPolicy = "linux-top7-direct70"
	// CommandsLinuxTop7Direct65 selects the seven response types and 65 names.
	CommandsLinuxTop7Direct65 CommandPolicy = "linux-top7-direct65"
	// CommandsLinuxTop7Direct40 selects the mainline namespace with safe rejects.
	CommandsLinuxTop7Direct40 CommandPolicy = "linux-top7-direct40-reject-host-mutation"
	// CommandsLinuxTop7Direct41 selects the WAN namespace with safe rejects.
	CommandsLinuxTop7Direct41 CommandPolicy = "linux-top7-direct41-reject-host-mutation"
	// CommandsLowResourceTop5 selects the proven top-level response namespace;
	// direct cmd remains unsupported rather than asserted absent in firmware.
	CommandsLowResourceTop5 CommandPolicy = "low-resource-top5-direct-cmd-unsupported"
)

// CapabilityUnsupported enables no positive protocol capability.
const CapabilityUnsupported CapabilityPolicy = 0

const (
	// CapabilityGatewayInform marks the shared gateway management vocabulary.
	CapabilityGatewayInform CapabilityPolicy = 1 << iota
	// CapabilityGatewayProducerTelemetry marks UXG producer-bound telemetry.
	CapabilityGatewayProducerTelemetry
	// CapabilityGatewaySavedConfigProbes marks UDM saved-config/probe semantics.
	CapabilityGatewaySavedConfigProbes
	// CapabilitySwitchInform marks the shared switch management vocabulary.
	CapabilitySwitchInform
	// CapabilitySwitchMainline marks the shared BCM/MSCC mainline contract.
	CapabilitySwitchMainline
	// CapabilitySwitchBCMTables marks BCM-only network and IGMP table vocabulary.
	CapabilitySwitchBCMTables
	// CapabilitySwitchMSCCPlatform marks the positive MSCC platform contract.
	CapabilitySwitchMSCCPlatform
	// CapabilitySwitchWANVocabulary marks source-gated WAN/RPS vocabulary only.
	CapabilitySwitchWANVocabulary
	// CapabilitySwitchLowResource marks the shared low-resource switch contract.
	CapabilitySwitchLowResource
	// CapabilitySwitchCurrentESP marks the four statically reviewed ESP apps.
	CapabilitySwitchCurrentESP
	// CapabilitySwitchPSEVocabulary marks source-gated PSE vocabulary.
	CapabilitySwitchPSEVocabulary
	// CapabilitySwitchRTL8238BVocabulary marks the reviewed RTL8238B state table.
	CapabilitySwitchRTL8238BVocabulary
	// CapabilitySwitchRTL8239MCUVocabulary marks the reviewed RTL8239MCU table.
	CapabilitySwitchRTL8239MCUVocabulary
	// CapabilityControllerClock marks exact builds whose noop server time is
	// evidenced as a process-local reporting offset rather than a host mutation.
	CapabilityControllerClock
)

const (
	capabilitySelectorGatewayInform            = "gateway-inform"
	capabilitySelectorGatewayProducerTelemetry = "gateway-producer-telemetry"
	capabilitySelectorGatewaySavedConfigProbes = "gateway-saved-config-probes"
	capabilitySelectorSwitchInform             = "switch-inform"
	capabilitySelectorSwitchMainline           = "switch-mainline"
	capabilitySelectorSwitchBCMTables          = "switch-bcm-tables"
	capabilitySelectorSwitchMSCCPlatform       = "switch-mscc-platform"
	capabilitySelectorSwitchWANVocabulary      = "switch-wan-vocabulary"
	capabilitySelectorSwitchLowResource        = "switch-low-resource"
	capabilitySelectorSwitchCurrentESP         = "switch-current-esp"
	capabilitySelectorSwitchPSEVocabulary      = "switch-pse-vocabulary"
	capabilitySelectorSwitchRTL8238BVocabulary = "switch-rtl8238b-vocabulary"
	capabilitySelectorSwitchRTL8239Vocabulary  = "switch-rtl8239mcu-vocabulary"
	capabilitySelectorControllerClock          = "controller-reporting-clock"
)

type protocolCapabilitySelector struct {
	Policy   CapabilityPolicy
	Selector string
}

var protocolCapabilitySelectors = []protocolCapabilitySelector{
	{Policy: CapabilityControllerClock, Selector: capabilitySelectorControllerClock},
	{Policy: CapabilitySwitchRTL8239MCUVocabulary, Selector: capabilitySelectorSwitchRTL8239Vocabulary},
	{Policy: CapabilitySwitchRTL8238BVocabulary, Selector: capabilitySelectorSwitchRTL8238BVocabulary},
	{Policy: CapabilitySwitchPSEVocabulary, Selector: capabilitySelectorSwitchPSEVocabulary},
	{Policy: CapabilitySwitchCurrentESP, Selector: capabilitySelectorSwitchCurrentESP},
	{Policy: CapabilitySwitchLowResource, Selector: capabilitySelectorSwitchLowResource},
	{Policy: CapabilitySwitchWANVocabulary, Selector: capabilitySelectorSwitchWANVocabulary},
	{Policy: CapabilitySwitchMSCCPlatform, Selector: capabilitySelectorSwitchMSCCPlatform},
	{Policy: CapabilitySwitchBCMTables, Selector: capabilitySelectorSwitchBCMTables},
	{Policy: CapabilitySwitchMainline, Selector: capabilitySelectorSwitchMainline},
	{Policy: CapabilitySwitchInform, Selector: capabilitySelectorSwitchInform},
	{Policy: CapabilityGatewaySavedConfigProbes, Selector: capabilitySelectorGatewaySavedConfigProbes},
	{Policy: CapabilityGatewayProducerTelemetry, Selector: capabilitySelectorGatewayProducerTelemetry},
	{Policy: CapabilityGatewayInform, Selector: capabilitySelectorGatewayInform},
}

const (
	// PresenceUnsupported disables source- and state-dependent presence rules.
	PresenceUnsupported PresencePolicy = protocolUnsupportedValue
	// PresenceGatewayUXG selects UXG producer and request presence gates.
	PresenceGatewayUXG PresencePolicy = "gateway-uxg-producer-request"
	// PresenceGatewayUDM selects UDM saved-config and probe presence gates.
	PresenceGatewayUDM PresencePolicy = "gateway-udm-saved-config-probes"
	// PresenceSwitchMainlineBCM selects BCM-only and common switch gates.
	PresenceSwitchMainlineBCM PresencePolicy = "switch-mainline-bcm-tables"
	// PresenceSwitchMainlineMSCC selects MSCC common switch gates.
	PresenceSwitchMainlineMSCC PresencePolicy = "switch-mainline-mscc-tables"
	// PresenceSwitchWAN selects switch-WAN model and source gates.
	PresenceSwitchWAN PresencePolicy = "switch-wan-model-sources"
	// PresenceLowResourceUSMini2 selects current usmini2 source gates.
	PresenceLowResourceUSMini2 PresencePolicy = "low-resource-usmini2-sources"
	// PresenceLowResourceUltra selects current Ultra and RTL8238B gates.
	PresenceLowResourceUltra PresencePolicy = "low-resource-ultra-rtl8238b"
	// PresenceLowResourceM25G5 selects current M25G5 model/source gates. It does
	// not infer absence of external PSE hardware from missing app vocabulary.
	PresenceLowResourceM25G5 PresencePolicy = "low-resource-m25g5-model-sources"
	// PresenceLowResourceM25G8 selects current M25G8 and RTL8239MCU gates.
	PresenceLowResourceM25G8 PresencePolicy = "low-resource-m25g8-rtl8239mcu"
)

// ProtocolBuild is the complete immutable artifact key for one evidence row.
// Firmware digests alone are insufficient because some containers hold
// multiple agents. BuildMarker is an observed version or artifact marker, not
// a normalized product label.
type ProtocolBuild struct {
	Family          ProtocolFamily
	Platform        ProtocolPlatform
	ComponentRole   string
	Version         string
	BuildMarker     string
	FirmwareSHA256  string
	ComponentSHA256 string
	DigestKind      ProtocolDigestKind
}

// ProtocolPolicies contains the six independent behavior axes. A policy ID is
// a reusable semantic contract; it is not an artifact version or build name.
type ProtocolPolicies struct {
	Wire         WirePolicy
	Discovery    DiscoveryPolicy
	FieldGroups  FieldGroupsPolicy
	Commands     CommandPolicy
	Capabilities CapabilityPolicy
	Presence     PresencePolicy
}

// Has reports whether all requested positive capability flags are present.
func (p CapabilityPolicy) Has(requested CapabilityPolicy) bool {
	return requested != CapabilityUnsupported && p&requested == requested
}

// ProtocolPolicyComponent provides the traceability chain for one reusable
// selector or positive capability flag.
type ProtocolPolicyComponent struct {
	Axis            ProtocolPolicyAxis
	Selector        string
	EvidenceRecords string
	Principle       string
}

// ProtocolBehaviorProfile composes six reusable policies independently of
// exact artifact identity. Equal complete compositions have exactly one ID.
type ProtocolBehaviorProfile struct {
	ID              ProtocolBehaviorID
	EvidenceRecords string
	Principle       string
	Policies        ProtocolPolicies
}

// ProtocolProfile binds one exact evidence build to a reusable behavior
// composition. Merely selecting metadata does not activate network or host
// behavior; downstream implementations consume each axis explicitly.
type ProtocolProfile struct {
	EvidenceID string
	Build      ProtocolBuild
	Support    ProtocolSupport
	BehaviorID ProtocolBehaviorID
}

// IsSupported reports whether both the build and behavior composition resolve.
func (p ProtocolProfile) IsSupported() bool {
	return p.Support == ProtocolSupported && p.Behavior().ID != ProtocolBehaviorUnsupported
}

// Behavior resolves the selected composition and fails closed for unknown IDs.
func (p ProtocolProfile) Behavior() ProtocolBehaviorProfile {
	return LookupProtocolBehaviorProfile(p.BehaviorID)
}

// Policies resolves the six selected axes and fails closed for unknown IDs.
func (p ProtocolProfile) Policies() ProtocolPolicies {
	return p.Behavior().Policies
}

// PolicyComponents resolves the traceability record for every selected
// policy component. Invalid or incomplete compositions return no records.
func (p ProtocolBehaviorProfile) PolicyComponents() []ProtocolPolicyComponent {
	keys, err := protocolPolicyComponentKeys(p.Policies)
	if err != nil {
		return nil
	}
	components := make([]ProtocolPolicyComponent, 0, len(keys))
	for _, key := range keys {
		component, ok := protocolPolicyComponentsByKey[key]
		if !ok {
			return nil
		}
		components = append(components, component)
	}
	return components
}

// ProtocolProfiles returns a detached copy of every exact evidence binding.
func ProtocolProfiles() []ProtocolProfile {
	out := make([]ProtocolProfile, len(protocolProfileDefinitions))
	copy(out, protocolProfileDefinitions)
	return out
}

// ProtocolBehaviorProfiles returns a detached copy of the behavior catalog.
func ProtocolBehaviorProfiles() []ProtocolBehaviorProfile {
	out := make([]ProtocolBehaviorProfile, len(protocolBehaviorDefinitions))
	copy(out, protocolBehaviorDefinitions)
	return out
}

// ProtocolPolicyComponents returns a detached copy of each evidence reference
// used by the reusable policy catalog.
func ProtocolPolicyComponents() []ProtocolPolicyComponent {
	out := make([]ProtocolPolicyComponent, len(protocolPolicyComponentDefinitions))
	copy(out, protocolPolicyComponentDefinitions)
	return out
}

// LookupProtocolProfile performs an exact all-field build lookup. A miss
// returns an explicit unsupported binding and never searches by family,
// version prefix, model similarity, or a neighboring digest.
func LookupProtocolProfile(build ProtocolBuild) ProtocolProfile {
	build = normalizeProtocolBuild(build)
	if profile, ok := protocolProfilesByBuild[build]; ok {
		return profile
	}
	return unsupportedProtocolProfile(build)
}

// LookupProtocolBehaviorProfile resolves an exact behavior ID. Unknown IDs
// select the all-unsupported composition instead of a similar profile.
func LookupProtocolBehaviorProfile(id ProtocolBehaviorID) ProtocolBehaviorProfile {
	id = normalizeProtocolBehaviorID(id)
	if behavior, ok := protocolBehaviorsByID[id]; ok {
		return behavior
	}
	return protocolBehaviorsByID[ProtocolBehaviorUnsupported]
}

func unsupportedProtocolPolicies() ProtocolPolicies {
	return ProtocolPolicies{
		Wire:         WireUnsupported,
		Discovery:    DiscoveryUnsupported,
		FieldGroups:  FieldGroupsUnsupported,
		Commands:     CommandsUnsupported,
		Capabilities: CapabilityUnsupported,
		Presence:     PresenceUnsupported,
	}
}

func unsupportedProtocolProfile(build ProtocolBuild) ProtocolProfile {
	return ProtocolProfile{
		Build:      normalizeProtocolBuild(build),
		Support:    ProtocolUnsupported,
		BehaviorID: ProtocolBehaviorUnsupported,
	}
}

func normalizeProtocolBuild(build ProtocolBuild) ProtocolBuild {
	build.Family = ProtocolFamily(strings.TrimSpace(string(build.Family)))
	build.Platform = ProtocolPlatform(strings.TrimSpace(string(build.Platform)))
	build.ComponentRole = strings.TrimSpace(build.ComponentRole)
	build.Version = strings.TrimSpace(build.Version)
	build.BuildMarker = strings.TrimSpace(build.BuildMarker)
	build.FirmwareSHA256 = strings.ToLower(strings.TrimSpace(build.FirmwareSHA256))
	build.ComponentSHA256 = strings.ToLower(strings.TrimSpace(build.ComponentSHA256))
	build.DigestKind = ProtocolDigestKind(strings.TrimSpace(string(build.DigestKind)))
	return build
}

func normalizeProtocolBehaviorID(id ProtocolBehaviorID) ProtocolBehaviorID {
	return ProtocolBehaviorID(strings.TrimSpace(string(id)))
}

type protocolPolicyComponentKey struct {
	Axis     ProtocolPolicyAxis
	Selector string
}

func buildProtocolPolicyComponentRegistry(
	definitions []ProtocolPolicyComponent,
) map[protocolPolicyComponentKey]ProtocolPolicyComponent {
	byKey := make(map[protocolPolicyComponentKey]ProtocolPolicyComponent, len(definitions))
	for index, definition := range definitions {
		definition.Axis = ProtocolPolicyAxis(strings.TrimSpace(string(definition.Axis)))
		definition.Selector = strings.TrimSpace(definition.Selector)
		definition.EvidenceRecords = strings.TrimSpace(definition.EvidenceRecords)
		definition.Principle = strings.TrimSpace(definition.Principle)
		if err := validateProtocolPolicyComponent(definition); err != nil {
			panic(fmt.Sprintf("protocol policy component %d: %v", index, err))
		}
		key := protocolPolicyComponentKey{Axis: definition.Axis, Selector: definition.Selector}
		if _, exists := byKey[key]; exists {
			panic(fmt.Sprintf("duplicate protocol policy component %s/%s", key.Axis, key.Selector))
		}
		byKey[key] = definition
	}
	return byKey
}

func protocolPolicyComponentKeys(policies ProtocolPolicies) ([]protocolPolicyComponentKey, error) {
	keys := []protocolPolicyComponentKey{
		{Axis: ProtocolPolicyAxisWire, Selector: string(policies.Wire)},
		{Axis: ProtocolPolicyAxisDiscovery, Selector: string(policies.Discovery)},
		{Axis: ProtocolPolicyAxisFieldGroups, Selector: string(policies.FieldGroups)},
		{Axis: ProtocolPolicyAxisCommands, Selector: string(policies.Commands)},
	}
	if policies.Capabilities == CapabilityUnsupported {
		keys = append(keys, protocolPolicyComponentKey{
			Axis: ProtocolPolicyAxisCapabilities, Selector: protocolUnsupportedValue,
		})
	} else {
		remaining := policies.Capabilities
		for _, capability := range protocolCapabilitySelectors {
			if !remaining.Has(capability.Policy) {
				continue
			}
			keys = append(keys, protocolPolicyComponentKey{
				Axis: ProtocolPolicyAxisCapabilities, Selector: capability.Selector,
			})
			remaining &^= capability.Policy
		}
		if remaining != CapabilityUnsupported {
			return nil, fmt.Errorf("unknown capability flags 0x%x", uint32(remaining))
		}
	}
	keys = append(keys, protocolPolicyComponentKey{
		Axis: ProtocolPolicyAxisPresence, Selector: string(policies.Presence),
	})
	return keys, nil
}

func buildProtocolBehaviorRegistry(
	definitions []ProtocolBehaviorProfile,
	components map[protocolPolicyComponentKey]ProtocolPolicyComponent,
) map[ProtocolBehaviorID]ProtocolBehaviorProfile {
	byID := make(map[ProtocolBehaviorID]ProtocolBehaviorProfile, len(definitions))
	byComposition := make(map[ProtocolPolicies]ProtocolBehaviorID, len(definitions))
	for index, definition := range definitions {
		definition.ID = normalizeProtocolBehaviorID(definition.ID)
		definition.EvidenceRecords = strings.TrimSpace(definition.EvidenceRecords)
		definition.Principle = strings.TrimSpace(definition.Principle)
		if err := validateProtocolBehaviorDefinition(definition, components); err != nil {
			panic(fmt.Sprintf("protocol behavior definition %d: %v", index, err))
		}
		if _, exists := byID[definition.ID]; exists {
			panic(fmt.Sprintf("duplicate protocol behavior id %q", definition.ID))
		}
		if existing, exists := byComposition[definition.Policies]; exists {
			panic(fmt.Sprintf("protocol behaviors %q and %q duplicate one composition", existing, definition.ID))
		}
		byID[definition.ID] = definition
		byComposition[definition.Policies] = definition.ID
	}
	if _, ok := byID[ProtocolBehaviorUnsupported]; !ok {
		panic("protocol behavior registry has no unsupported composition")
	}
	return byID
}

func buildProtocolRegistry(
	definitions []ProtocolProfile,
	behaviors map[ProtocolBehaviorID]ProtocolBehaviorProfile,
) map[ProtocolBuild]ProtocolProfile {
	byBuild := make(map[ProtocolBuild]ProtocolProfile, len(definitions))
	evidenceIDs := make(map[string]struct{}, len(definitions))
	for index, definition := range definitions {
		definition.EvidenceID = strings.TrimSpace(definition.EvidenceID)
		definition.Build = normalizeProtocolBuild(definition.Build)
		definition.BehaviorID = normalizeProtocolBehaviorID(definition.BehaviorID)
		if err := validateProtocolProfileDefinition(definition, behaviors); err != nil {
			panic(fmt.Sprintf("protocol profile definition %d: %v", index, err))
		}
		if _, exists := evidenceIDs[definition.EvidenceID]; exists {
			panic(fmt.Sprintf("duplicate protocol evidence id %q", definition.EvidenceID))
		}
		if _, exists := byBuild[definition.Build]; exists {
			panic(fmt.Sprintf("duplicate protocol build %q", definition.EvidenceID))
		}
		evidenceIDs[definition.EvidenceID] = struct{}{}
		byBuild[definition.Build] = definition
	}
	return byBuild
}

func validateProtocolPolicyComponent(component ProtocolPolicyComponent) error {
	if !isProtocolPolicyAxis(component.Axis) {
		return fmt.Errorf("axis %q is invalid", component.Axis)
	}
	if component.Selector == "" {
		return fmt.Errorf("selector is required for %s", component.Axis)
	}
	if component.EvidenceRecords == "" || !strings.Contains(component.EvidenceRecords, "DT-") {
		return fmt.Errorf("evidence records are required for %s/%s", component.Axis, component.Selector)
	}
	if component.Principle == "" {
		return fmt.Errorf("principle is required for %s/%s", component.Axis, component.Selector)
	}
	return nil
}

func validateProtocolBehaviorDefinition(
	behavior ProtocolBehaviorProfile,
	components map[protocolPolicyComponentKey]ProtocolPolicyComponent,
) error {
	if behavior.ID == "" {
		return fmt.Errorf("behavior id is required")
	}
	if behavior.EvidenceRecords == "" || !strings.Contains(behavior.EvidenceRecords, "DT-") {
		return fmt.Errorf("evidence records are required for behavior %s", behavior.ID)
	}
	if behavior.Principle == "" {
		return fmt.Errorf("principle is required for behavior %s", behavior.ID)
	}
	policies := behavior.Policies
	if policies.Wire == "" || policies.Discovery == "" || policies.FieldGroups == "" ||
		policies.Commands == "" || policies.Presence == "" {
		return fmt.Errorf("all six policies are required for %s", behavior.ID)
	}
	if behavior.ID == ProtocolBehaviorUnsupported && policies != unsupportedProtocolPolicies() {
		return fmt.Errorf("unsupported behavior carries an active policy")
	}
	if behavior.ID != ProtocolBehaviorUnsupported && policies.Capabilities == CapabilityUnsupported {
		return fmt.Errorf("active behavior %s has no positive capability", behavior.ID)
	}
	keys, err := protocolPolicyComponentKeys(policies)
	if err != nil {
		return fmt.Errorf("behavior %s: %w", behavior.ID, err)
	}
	for _, key := range keys {
		if _, ok := components[key]; !ok {
			return fmt.Errorf("behavior %s lacks traceability for %s/%s", behavior.ID, key.Axis, key.Selector)
		}
	}
	return nil
}

func validateProtocolProfileDefinition(
	profile ProtocolProfile,
	behaviors map[ProtocolBehaviorID]ProtocolBehaviorProfile,
) error {
	build := profile.Build
	for name, value := range map[string]string{
		"evidence id":    profile.EvidenceID,
		"family":         string(build.Family),
		"platform":       string(build.Platform),
		"component role": build.ComponentRole,
		"version":        build.Version,
		"build marker":   build.BuildMarker,
	} {
		if value == "" {
			return fmt.Errorf("%s is required", name)
		}
	}
	if !isEvidenceProtocolPlatform(build.Platform) {
		return fmt.Errorf("platform %q is outside the evidence taxonomy for %s", build.Platform, profile.EvidenceID)
	}
	if profile.Support == ProtocolSupported && build.Platform == ProtocolPlatformUnknown {
		return fmt.Errorf("supported build %s has an unknown platform", profile.EvidenceID)
	}
	if !isProtocolSHA256(build.FirmwareSHA256) {
		return fmt.Errorf("firmware SHA-256 is invalid for %s", profile.EvidenceID)
	}
	if build.DigestKind == ProtocolDigestUnknown {
		if build.ComponentSHA256 != "" {
			return fmt.Errorf("unknown component digest kind has a digest for %s", profile.EvidenceID)
		}
	} else if !isProtocolSHA256(build.ComponentSHA256) {
		return fmt.Errorf("component SHA-256 is invalid for %s", profile.EvidenceID)
	}
	if profile.Support != ProtocolSupported && profile.Support != ProtocolUnsupported {
		return fmt.Errorf("support %q is invalid for %s", profile.Support, profile.EvidenceID)
	}
	if _, ok := behaviors[profile.BehaviorID]; !ok {
		return fmt.Errorf("behavior %q is unknown for %s", profile.BehaviorID, profile.EvidenceID)
	}
	if profile.Support == ProtocolUnsupported && profile.BehaviorID != ProtocolBehaviorUnsupported {
		return fmt.Errorf("unsupported build %s carries an active behavior", profile.EvidenceID)
	}
	if profile.Support == ProtocolSupported && profile.BehaviorID == ProtocolBehaviorUnsupported {
		return fmt.Errorf("supported build %s carries no behavior", profile.EvidenceID)
	}
	return nil
}

func isProtocolPolicyAxis(axis ProtocolPolicyAxis) bool {
	switch axis {
	case ProtocolPolicyAxisWire,
		ProtocolPolicyAxisDiscovery,
		ProtocolPolicyAxisFieldGroups,
		ProtocolPolicyAxisCommands,
		ProtocolPolicyAxisCapabilities,
		ProtocolPolicyAxisPresence:
		return true
	default:
		return false
	}
}

func isEvidenceProtocolPlatform(platform ProtocolPlatform) bool {
	switch platform {
	case ProtocolPlatformUnknown,
		ProtocolPlatformAArch64LE,
		ProtocolPlatformMIPS32BE,
		ProtocolPlatformARM32LEEABI5,
		ProtocolPlatformMIPS32R2LE,
		ProtocolPlatformMIPS32R2BE,
		ProtocolPlatformCortexMThumbLE,
		ProtocolPlatformXtensaESP32LE:
		return true
	default:
		return false
	}
}

func isProtocolSHA256(value string) bool {
	if len(value) != 64 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

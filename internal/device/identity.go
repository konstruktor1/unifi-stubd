package device

// InformState is the controller-facing numeric lifecycle value carried by an
// Inform payload. The complete vocabulary is typed even though the safe stub
// actively transitions only through Default, Unknown, and Managed.
type InformState int

const (
	// InformStateUnknown is the transient key-switch/config-request state.
	InformStateUnknown InformState = iota
	// InformStateDefault is the pending, public-default-key state.
	InformStateDefault
	// InformStateManaged is the adopted steady state.
	InformStateManaged
	// InformStateSelfrun is a vendor status value; it grants no host action.
	InformStateSelfrun
	// InformStateUpgrade is a vendor status value; it grants no upgrade action.
	InformStateUpgrade
)

// Valid reports whether state belongs to the observed five-value vocabulary.
func (state InformState) Valid() bool {
	return state >= InformStateUnknown && state <= InformStateUpgrade
}

// String returns the fixed, non-secret status label for one wire state.
func (state InformState) String() string {
	switch state {
	case InformStateUnknown:
		return "unknown"
	case InformStateDefault:
		return "default"
	case InformStateManaged:
		return "managed"
	case InformStateSelfrun:
		return "selfrun"
	case InformStateUpgrade:
		return "upgrade"
	default:
		return "invalid"
	}
}

// Identity contains the device attributes reported in inform payloads.
type Identity struct {
	// MAC is the fake device MAC address in controller-facing text form.
	MAC string
	// IP is the device management IP address reported to UniFi.
	IP string
	// Hostname is the device name reported to UniFi.
	Hostname string
	// Model is the UniFi model identifier.
	Model string
	// ModelDisplay is the human-readable UniFi model name.
	ModelDisplay string
	// DeviceType is the controller-facing UniFi device family.
	DeviceType string
	// Version is the firmware version reported by the stub.
	Version string
	// Serial is the serial number reported by the stub.
	Serial string
	// InformURL is the controller inform URL currently known by the device.
	InformURL string
	// InformIP is the numeric controller inform endpoint address reported to UniFi.
	InformIP string
	// CFGVersion is the controller configuration version applied to the device.
	CFGVersion string
	// ManagementVLAN is the optional controller-facing management VLAN ID.
	ManagementVLAN int
	// UptimeSeconds is the monotonic runtime uptime reported in inform payloads.
	UptimeSeconds int
	// ControllerTimeOffsetSeconds shifts reported wall-clock fields only. It is
	// process-local controller metadata and never changes the host clock.
	ControllerTimeOffsetSeconds int64
	// The block flags are producer-local gates learned from the latest noop.
	// They do not form a generic payload filter and never authorize host I/O.
	IPv4ActiveLeasesRequested bool
	GatewayMACIPv6Requested   bool
	GWCapabilitiesRequested   bool
	UDAPIVersionExcluded      bool
	DPIStatsExcluded          bool
	FingerprintsExcluded      bool
	// InformState is an optional explicit controller-facing lifecycle value.
	// InformStateSet keeps zero meaningful during the adopted-key switch.
	InformState    InformState
	InformStateSet bool
	// Adopted reports whether the stub should present itself as adopted.
	Adopted bool
}

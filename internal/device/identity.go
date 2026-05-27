package device

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
	// Adopted reports whether the stub should present itself as adopted.
	Adopted bool
}

// Management LAN handling keeps switch management VLAN behavior explicit. The
// daemon may report metadata or bind to a preexisting VLAN interface, but it
// never creates host VLAN devices or applies controller provisioning locally.
package main

// Management-LAN modes and policies are explicit so the daemon can report VLAN
// intent without creating host VLAN interfaces.
const (
	managementLANModeMetadataOnly         = "metadata-only"
	managementLANModePreexistingInterface = "preexisting-interface"
	managementLANModePlannedHostVLAN      = "planned-host-vlan"

	managementLANReachOff      = "off"
	managementLANReachWarn     = "warn"
	managementLANReachRequired = "required"

	managementLANAdoptUntaggedFirst = "untagged-first"
	managementLANAdoptTaggedOnly    = "tagged-only"
)

package platform

import (
	"context"
	"net"
	"time"

	"github.com/konstruktor1/unifi-stubd/internal/observe"
)

// Source constants select optional host integrations.
const (
	// SourceOff disables an optional platform source.
	SourceOff = "off"
	// LLDPSourceLLDPD reads passive LLDP neighbors from lldpd through lldpcli.
	LLDPSourceLLDPD = "lldpd"
	// ProcSourceProcFS reads Linux procfs counters.
	ProcSourceProcFS = "procfs"
)

// Capability state strings are stable status values, not dependency management
// actions.
const (
	capabilityDisabled    = "disabled"
	capabilityAvailable   = "available"
	capabilityMissing     = "missing"
	capabilityUnsupported = "unsupported"
	capabilityLLDP        = "lldp"
	capabilityProc        = "proc"
	goosFreeBSD           = "freebsd"
	goosLinux             = "linux"
)

// Config contains optional platform integration settings.
type Config struct {
	LLDPSource     string
	ProcSource     string
	CommandTimeout time.Duration
}

// Platform groups read-only OS integration behind one runtime boundary.
type Platform interface {
	observe.ObservationSource
	InterfaceReader
	EgressLinkReader
	AddressReader
	LLDPReader
	ProcReader
	Capabilities(Config) CapabilityReport
}

// InterfaceReader reads a single host interface.
type InterfaceReader interface {
	Interface(context.Context, string) (observe.PortObservation, []error)
}

// EgressLinkReader resolves the local link used to reach one target.
type EgressLinkReader interface {
	EgressLink(context.Context, string) (EgressLink, error)
}

// AddressReader reads host-interface identity and checks IPv4 assignment.
type AddressReader interface {
	InterfaceIdentity(context.Context, string) (InterfaceIdentity, error)
	IPv4Assigned(context.Context, string, net.IP) (bool, error)
}

// InterfaceIdentity contains host-interface addresses without optional counter,
// media, or command-backed enrichment.
type InterfaceIdentity struct {
	MAC  net.HardwareAddr
	IPv4 []net.IP
}

// LLDPReader reads passive LLDP neighbors.
type LLDPReader interface {
	LLDP(context.Context, LLDPConfig) ([]LLDPNeighbor, []error)
}

// ProcReader reads optional proc-style counters.
type ProcReader interface {
	Proc(context.Context, ProcConfig) (ProcSnapshot, []error)
}

// CapabilityReport describes which platform sources are usable at runtime.
type CapabilityReport struct {
	GOOS         string       `json:"goos"`
	Capabilities []Capability `json:"capabilities,omitempty"`
}

// Capability describes one optional runtime source.
type Capability struct {
	Name   string `json:"name"`
	Source string `json:"source,omitempty"`
	State  string `json:"state"`
	Detail string `json:"detail,omitempty"`
}

// LLDPConfig selects a passive LLDP backend.
type LLDPConfig struct {
	Source  string
	Timeout time.Duration
}

// LLDPNeighbor is one normalized neighbor learned from a passive source.
type LLDPNeighbor struct {
	Interface       string   `json:"interface"`
	ChassisID       string   `json:"chassis_id,omitempty"`
	ChassisMAC      string   `json:"chassis_mac,omitempty"`
	SystemName      string   `json:"system_name,omitempty"`
	PortID          string   `json:"port_id,omitempty"`
	PortDescription string   `json:"port_description,omitempty"`
	Capabilities    []string `json:"capabilities,omitempty"`
	ManagementIP    string   `json:"management_ip,omitempty"`
}

// ProcConfig selects proc-style counters.
type ProcConfig struct {
	Source string
	Root   string
}

// ProcSnapshot contains optional host counters read from proc-style sources.
type ProcSnapshot struct {
	Interfaces map[string]observe.InterfaceStats `json:"interfaces,omitempty"`
}

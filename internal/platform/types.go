// Package platform defines the small contracts used by cmd/unifi-stubd and the
// observe package. Keeping these structs plain makes it clear which host facts
// are read-only inputs and prevents adapter-specific details from leaking into
// payload rendering.
package platform

import (
	"context"
	"time"

	"github.com/konstruktor1/unifi-stubd/internal/observe"
)

// Source constants select optional host integrations.
const (
	// SourceOff disables an optional platform source.
	SourceOff = "off"
	// LLDPSourceLLDPD reads passive LLDP neighbors from lldpd through lldpcli.
	LLDPSourceLLDPD = "lldpd"
	// LogSourceJournalctl reads Linux journal entries through journalctl.
	LogSourceJournalctl = "journalctl"
	// LogSourceSyslog reads syslog-style log files.
	LogSourceSyslog = "syslog"
	// ProcSourceProcFS reads Linux procfs counters.
	ProcSourceProcFS = "procfs"
	// DBusBusSystem selects the system D-Bus.
	DBusBusSystem = "system"
	// DBusBusSession selects the session D-Bus.
	DBusBusSession = "session"
)

// Capability state strings are stable status values, not dependency management
// actions.
const (
	capabilityDisabled    = "disabled"
	capabilityAvailable   = "available"
	capabilityMissing     = "missing"
	capabilityUnsupported = "unsupported"
	capabilityDBus        = "dbus"
	capabilityLLDP        = "lldp"
	capabilityLogs        = "logs"
	capabilityProc        = "proc"
	capabilityTraffic     = "traffic"
	goosFreeBSD           = "freebsd"
	goosLinux             = "linux"
)

// Config contains optional platform integration settings.
type Config struct {
	LLDPSource     string
	TrafficSource  string
	LogSource      string
	ProcSource     string
	DBusEnabled    bool
	DBusBus        string
	SyslogPath     string
	JournalUnit    string
	CommandTimeout time.Duration
}

// Platform groups read-only OS integration behind one runtime boundary.
type Platform interface {
	observe.ObservationSource
	InterfaceReader
	LLDPReader
	LogReader
	ProcReader
	ServiceBus
	Capabilities(context.Context, Config) CapabilityReport
}

// InterfaceReader reads a single host interface.
type InterfaceReader interface {
	Interface(context.Context, string) (observe.PortObservation, []error)
}

// LLDPReader reads passive LLDP neighbors.
type LLDPReader interface {
	LLDP(context.Context, LLDPConfig) ([]LLDPNeighbor, []error)
}

// LogReader reads recent platform logs.
type LogReader interface {
	Logs(context.Context, LogConfig) ([]LogEntry, []error)
}

// ProcReader reads optional proc-style counters.
type ProcReader interface {
	Proc(context.Context, ProcConfig) (ProcSnapshot, []error)
}

// ServiceBus checks optional D-Bus connectivity.
type ServiceBus interface {
	ServiceBus(context.Context, DBusConfig) (ServiceBusStatus, error)
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

// LogConfig selects the platform log reader.
type LogConfig struct {
	Source string
	Path   string
	Unit   string
	Lines  int
}

// LogEntry is one normalized journal or syslog entry.
type LogEntry struct {
	Time     string `json:"time,omitempty"`
	Unit     string `json:"unit,omitempty"`
	Priority string `json:"priority,omitempty"`
	Message  string `json:"message,omitempty"`
	Raw      string `json:"raw,omitempty"`
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

// DBusConfig selects optional D-Bus connectivity.
type DBusConfig struct {
	Enabled bool
	Bus     string
}

// ServiceBusStatus reports optional D-Bus availability.
type ServiceBusStatus struct {
	Enabled bool   `json:"enabled"`
	Bus     string `json:"bus,omitempty"`
	State   string `json:"state"`
	Detail  string `json:"detail,omitempty"`
}

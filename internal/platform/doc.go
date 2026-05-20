// Package platform hides OS-specific read-only host integration behind small
// adapters. Runtime code asks for observations, capabilities, logs, and service
// bus status without knowing whether the data came from Linux sysfs/procfs,
// FreeBSD ifconfig/syslog, lldpd, or a disabled no-op source.
package platform

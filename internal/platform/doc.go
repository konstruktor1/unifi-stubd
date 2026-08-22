// Package platform hides OS-specific read-only host integration behind small
// adapters. Runtime code asks for interface, bridge, procfs, LLDP, and capability
// observations without depending on Linux or FreeBSD host details.
package platform

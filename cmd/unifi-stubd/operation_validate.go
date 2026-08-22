package main

import (
	"fmt"
	"strings"

	"github.com/konstruktor1/unifi-stubd/internal/platform"
)

// validateOperationFlags normalizes operator-selected modes and optional host
// sources before later validation can decide whether live checks are required.
func validateOperationFlags(flags *runtimeFlags) error {
	mode := normalizeMode(flags.operationMode)
	flags.operationMode = mode
	switch mode {
	case operationModeStub, operationModeBridgeObserve, operationModePortMap, operationModeHostDirect, operationModeMacvlan:
	default:
		return fmt.Errorf("invalid -operation-mode %q; use stub, bridge-observe, observe, port-map, host-direct, or macvlan", mode)
	}

	lldpSource := strings.ToLower(strings.TrimSpace(flags.lldpSource))
	if lldpSource == "" {
		lldpSource = platform.SourceOff
	}
	flags.lldpSource = lldpSource
	switch lldpSource {
	case platform.SourceOff, platform.LLDPSourceLLDPD:
	default:
		return fmt.Errorf("invalid -lldp-source %q; use off or lldpd", lldpSource)
	}

	flags.procSource = strings.ToLower(strings.TrimSpace(flags.procSource))
	if flags.procSource == "" {
		flags.procSource = platform.SourceOff
	}
	if flags.procSource != platform.SourceOff && flags.procSource != platform.ProcSourceProcFS {
		return fmt.Errorf("invalid -proc-source %q; use off or procfs", flags.procSource)
	}
	if err := validateInstanceGuard(flags); err != nil {
		return err
	}
	if strings.EqualFold(strings.TrimSpace(flags.macText), "host") && mode != operationModeHostDirect {
		return fmt.Errorf("mac: host is only allowed with -operation-mode host-direct")
	}
	normalizeWANHealthConfig(flags)
	// Planned host-networking modes remain review-only. The daemon may print
	// the intended macvlan commands, but it must not create interfaces itself.
	if mode == operationModeMacvlan && !flags.dryRunPlan {
		return fmt.Errorf("operation-mode macvlan is planned only; use -dry-run-plan to inspect the non-mutating plan")
	}
	if iface := strings.TrimSpace(flags.discoveryInterface); strings.Contains(iface, "/") {
		return fmt.Errorf("invalid -discovery-interface %q", iface)
	}
	return nil
}

// normalizeMode keeps the legacy observe alias while making stub mode the
// default safety posture.
func normalizeMode(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return operationModeStub
	}
	if value == operationModeObserve {
		return operationModeBridgeObserve
	}
	return value
}

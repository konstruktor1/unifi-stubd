// Package platform checks D-Bus only as an optional capability probe, not a
// runtime dependency. The adapter only checks whether the configured bus is
// reachable so deployments can expose that fact in status without making Linux
// or FreeBSD require D-Bus.
package platform

import (
	"context"
	"fmt"

	"github.com/godbus/dbus/v5"
)

// ServiceBus probes the configured D-Bus only when enabled and reports
// availability as status metadata.
func (p hostPlatform) ServiceBus(_ context.Context, cfg DBusConfig) (ServiceBusStatus, error) {
	cfg.Bus = normalizedDBusBus(cfg.Bus)
	if !cfg.Enabled {
		return ServiceBusStatus{Enabled: false, Bus: cfg.Bus, State: capabilityDisabled}, nil
	}
	conn, err := connectDBus(cfg.Bus)
	if err != nil {
		return ServiceBusStatus{Enabled: true, Bus: cfg.Bus, State: capabilityMissing, Detail: err.Error()}, err
	}
	defer func() {
		_ = conn.Close()
	}()
	return ServiceBusStatus{Enabled: true, Bus: cfg.Bus, State: capabilityAvailable}, nil
}

// dbusCapability adapts the D-Bus probe into the generic capability report used
// by --status.
func (p hostPlatform) dbusCapability(ctx context.Context, cfg Config) Capability {
	status, err := p.ServiceBus(ctx, DBusConfig{Enabled: cfg.DBusEnabled, Bus: cfg.DBusBus})
	if !status.Enabled {
		return Capability{Name: capabilityDBus, Source: status.Bus, State: capabilityDisabled}
	}
	if err != nil {
		return Capability{Name: capabilityDBus, Source: status.Bus, State: capabilityMissing, Detail: status.Detail}
	}
	return Capability{Name: capabilityDBus, Source: status.Bus, State: capabilityAvailable}
}

// connectDBus opens only the configured bus for a capability probe.
func connectDBus(bus string) (*dbus.Conn, error) {
	switch normalizedDBusBus(bus) {
	case DBusBusSession:
		conn, err := dbus.ConnectSessionBus()
		if err != nil {
			return nil, fmt.Errorf("connect session dbus: %w", err)
		}
		return conn, nil
	default:
		conn, err := dbus.ConnectSystemBus()
		if err != nil {
			return nil, fmt.Errorf("connect system dbus: %w", err)
		}
		return conn, nil
	}
}

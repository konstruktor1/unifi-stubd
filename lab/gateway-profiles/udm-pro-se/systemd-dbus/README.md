# systemd DBus Facade

UniFi Core expects to inspect `org.freedesktop.systemd1`, but the Docker
webportal path does not run firmware systemd as PID 1. This facade provides the
small DBus surface Core needs to decide whether known applications and services
are present, enabled, or running.

The implementation is deliberately fixture-driven. `units.cjs` defines the
known units, `interfaces.cjs` exposes only the manager/unit/service properties
used in this lab, and the DBus/server modules publish that state. It should not
grow into a general systemd replacement.

Lifecycle actions that would mutate the host are handled separately by the
lab `systemctl` wrapper. Keep this DBus facade read-oriented and deterministic.

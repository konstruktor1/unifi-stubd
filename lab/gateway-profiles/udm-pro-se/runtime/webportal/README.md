# Webportal Runtime Modules

This runtime extends the reduced Docker firmware path with just enough UniFi OS
support to load the setup web UI. It starts PostgreSQL, DBus, the systemd DBus
facade, the Network facade, `ulp-go`, nginx, UniFi Core, and the firmware
management services.

The subdirectories are split by the kind of compatibility being installed:
wrappers replace host-mutating commands with logged lab behavior, HTTP assets
patch generated nginx/Core routes, templates render Core config and sudoers,
and data files provide deterministic identity payloads.

This is not the native boot reference. It is a controlled UI/API inspection
surface. Keep additions narrow and explain which UniFi Core or setup request
needs them.

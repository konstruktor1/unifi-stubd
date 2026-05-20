# Firmware Runtime Modules

This runtime starts the smallest Docker process chain needed to inspect UDM Pro
SE firmware management behavior: `ubios-udapi-server`, `udapi-bridge`, `mcad`,
and local socket checks. It is networkless by default and is useful for testing
the C mock, UDAPI socket creation, and `mca-ctrl -t dump`.

It intentionally does not start UniFi Core, PostgreSQL, nginx setup routes, or
the Network facade. Those belong to `runtime/webportal/`.

When changing this runtime, keep it focused on firmware service startup and
inspection. Do not add controller adoption or host networking behavior here.

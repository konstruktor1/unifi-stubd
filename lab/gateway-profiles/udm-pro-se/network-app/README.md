# Network App Facade

The Docker webportal path starts UniFi Core without the real UniFi Network
backend. This CommonJS facade supplies only the Network endpoints that Core and
the setup UI need for the lab: app readiness, manifest discovery, setup
queries, feature checks, health, and controlled no-op commands.

The code is split so new behavior has an obvious home: configuration, logging,
HTTP helpers, deterministic payloads, route handling, websocket bootstrap, and
the process entry point. Keep endpoint responses explicit and fixture-like; do
not add broad proxying or controller-derived behavior.

This process is intentionally Docker-only. In the UTM VM path, the real
firmware Network backend is allowed to expose its own state so the VM can teach
us what the stub or mocks still need to emulate.

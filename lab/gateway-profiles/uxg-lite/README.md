# UXG-Lite Simulation

This folder contains project-owned helper files for the UXG-Lite real firmware
profile. The current wrapper starts enough of the ARM64 UbiOS userspace to see
the early UDAPI process behavior, but it is not a complete adoption lab yet.

Use `docker-howto.md` for the reproducible setup.

Current state:

- `ubios-udapi-server` starts with mocked board data.
- `udapi-bridge` starts.
- `mcad` starts but does not expose `/tmp/.mcad`.
- `mca-ctrl -t dump` therefore fails in this profile today.

Keep the container networkless unless you intentionally attach it to a
disposable controller lab.

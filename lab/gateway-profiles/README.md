# Gateway Firmware Simulation Profiles

This directory is for real firmware simulation containers. Do not place
`internal/device` stub profile data here.

Only `lab/stub/` is the generic `unifi-stubd` stub lab. The gateway profile
directories below are wrappers around locally imported or extracted vendor
firmware artifacts, with those artifacts kept out of Git.

Current profile directories:

- `ugw3/`: QEMU-MIPS runner for an extracted UGW3 rootfs.
- `uxg-lite/`: ARM64 UbiOS userspace wrapper; partial simulation.
- `uxgpro/`: ARM64 UbiOS userspace wrapper; includes controller lab helpers.
- `ucg-fiber/`: ARM64 UbiOS userspace wrapper prepared for startup analysis.

Typical per-profile files:

- `Dockerfile`: project-owned wrapper around a local firmware rootfs or runner.
- `compose.yaml`: isolated simulation startup.
- `docker-howto.md`: local firmware import and runtime steps.
- `start-*.sh`: project-owned process startup wrapper.
- `firmware.md`: safe firmware inventory and findings.
- `source-inventory.md`: attribution and source boundary notes.

Firmware images, extracted rootfs trees, raw captures, keys, tokens,
certificates, and private controller data must not be committed.

Run a profile through its own compose file, for example:

```sh
docker compose -f lab/gateway-profiles/ugw3/compose.yaml up -d --build
```

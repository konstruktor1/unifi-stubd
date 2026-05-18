# Gateway Firmware Simulation Profiles

This directory is for real firmware simulation profiles. Do not place
`internal/device` stub profile data here.

Only `lab/stub/` is the generic `unifi-stubd` stub lab. The gateway profile
directories below are wrappers around locally imported or extracted vendor
firmware artifacts, with those artifacts kept out of Git.

Current profile directories:

- `ugw3/`: QEMU-MIPS runner for an extracted UGW3 rootfs.
- `uxg-lite/`: ARM64 UbiOS userspace wrapper; partial simulation.
- `uxgpro/`: ARM64 UbiOS userspace wrapper; includes controller lab helpers.
- `ucg-fiber/`: ARM64 UbiOS userspace wrapper prepared for startup analysis.
- `udm-pro-se/`: ARM64 UbiOS userspace wrapper; reaches UDAPI and
  `mca-ctrl` through a deterministic RTL8370-style switch mock. Its Docker
  webportal override exposes a partial UniFi OS setup surface using modular
  CommonJS facades under `network-app/` and `systemd-dbus/`.
- `udm-pro-se-vm/`: real `qemu-system-aarch64` VM boot profile using copied
  local UDM Pro SE firmware artifacts. The direct vendor kernel hangs before
  serial output on QEMU `virt`; the foreign-kernel `udm-systemd` path reaches
  UDM firmware `systemd`, applies userspace hardware mocks, completes
  `network-init.service`, starts `ubios-udapi-server`, exercises
  `/firewall/nat`, and reaches a serial login prompt.

Current project status is summarized in
`../../docs/en/project-status.md` and `../../docs/de/project-status.md`.

Typical per-profile files:

- `Dockerfile`: project-owned wrapper around a local firmware rootfs or runner.
- `compose.yaml`: isolated simulation startup for Compose-backed profiles.
- `docker-howto.md`: local firmware import and runtime steps.
- `start-*.sh`: project-owned process startup wrapper.
- `firmware.md`: safe firmware inventory and findings.
- `source-inventory.md`: attribution and source boundary notes.
- `scripts/`: extraction, preparation, and VM runner helpers for profiles that
  do not use Compose.

Firmware images, extracted rootfs trees, raw captures, keys, tokens,
certificates, and private controller data must not be committed.

Run a Compose-backed profile through its own compose file, for example:

```sh
docker compose -f lab/gateway-profiles/ugw3/compose.yaml up -d --build
```

Run the UDM Pro SE VM profile through its scripts:

```sh
lab/gateway-profiles/udm-pro-se-vm/scripts/prepare-vm.sh
lab/gateway-profiles/udm-pro-se-vm/scripts/run-direct-kernel.sh
```

Run the UDM Pro SE VM path that reaches firmware `systemd`:

```sh
lab/gateway-profiles/udm-pro-se-vm/scripts/prepare-vm.sh
lab/gateway-profiles/udm-pro-se-vm/scripts/fetch-foreign-kernel.sh
lab/gateway-profiles/udm-pro-se-vm/scripts/prepare-mocks.sh
lab/gateway-profiles/udm-pro-se-vm/scripts/build-lab-initramfs.sh
lab/gateway-profiles/udm-pro-se-vm/scripts/deploy-kernel-artifacts.sh
UDM_PRO_SE_FOREIGN_MODE=udm-systemd \
  lab/gateway-profiles/udm-pro-se-vm/scripts/run-foreign-kernel.sh
```

For `udm-systemd`, the direct QEMU runner can use a transparent `vmnet-bridged`
LAN NIC. The UTM profile uses the closer two-port mapping: UTM `Shared` / NAT
as `eth9` for the first SFP+ WAN role, and UTM `Host` as `eth8` for the 2.5G
RJ45 LAN role attached to `br0`. The guest keeps `br0` on `192.168.1.1/24` and
adds a host-only access alias `192.168.128.2/24`. The latest UTM full test also
reached the setup UI directly through the UTM Shared/NAT guest address and
reported `/api/system` with `hasInternet=true`. The profile writes a UTM
localhost forward for `https://127.0.0.1:10443/`, but the latest observed UTM
CLI run did not bind that host port without an explicit local TCP helper. Other
UDM-facing lab devices remain dummy devices inside the guest. The lab initramfs
uses the vendor setup nginx template for this QEMU path and keeps WAN ingress
explicit.

The deploy script stages the shared ignored kernel payload under
`udm-pro-se-vm/artifacts/deploy/kernel/`. UTM prefers that payload for its
kernel/initramfs inputs, and the Docker UDM Pro SE profile mounts it read-only
under `/opt/unifi-fw-sim/kernel` for comparison with the VM reference.

The UDM Pro SE profile documentation is split by scope:

- `udm-pro-se/docker-howto.md`: Docker rootfs import, mock preparation,
  webportal startup, and local checks.
- `udm-pro-se/firmware.md`: safe firmware observations and Docker simulation
  status.
- `udm-pro-se/source-inventory.md`: project-owned helper files and vendor
  boundary.
- `udm-pro-se-vm/README.md`: QEMU/UTM VM runbook.
- `udm-pro-se-vm/source-inventory.md`: VM scripts, initramfs payloads, kernel
  deployment, and UTM installer modules.

To recreate the local Docker inputs and run every committed lab container stack
together, use `run-all-containers.md`. A German runbook is available as
`run-all-containers.de.md`.

# Firmware Research Profiles

This directory tracks real firmware research separately from synthetic
`unifi-stubd` profiles. The machine-readable catalog is
`research/firmware/profiles.yaml`; the sections below are the human working
notes per profile.

Do not commit firmware images, extracted root filesystems, raw captures,
adoption keys, controller tokens, certificates, password hashes, SSH keys, or
private lab addresses. Keep those inputs under ignored `artifacts/`, `rootfs/`,
or `captures/` paths and commit only checksums, safe structural summaries, and
project-owned helper code.

## uxgpro-5.0.16

Status: adopted in the Docker controller lab.

- Stub profile: `uxgpro`
- Device type: `uxg`
- Model: `UXGPRO`
- Firmware: `5.0.16.30689`
- Architecture: `arm64`

Research folder:

```text
research/firmware/uxgpro-5.0.16/
```

Committed profile artifacts:

- `README.md`: firmware image inventory, service chain, simulation snapshot,
  adoption result, and source availability notes.
- `source-inventory.md`: project-owned helper source vs observed vendor files.
- `simulation/Dockerfile`: wrapper around the imported local firmware rootfs
  image.
- `simulation/start-firmware-processes.sh`: starts
  `ubios-udapi-server`, `udapi-bridge`, `mcad`, and optional Dropbear.
- `simulation/controller-lab.compose.yaml`: firmware + UniFi controller +
  MongoDB + inform MITM lab.
- `simulation/fixtures/adoption-mitm-timeline.json`: sanitized adoption
  timeline.
- `simulation/fixtures/adopted-system-config-summary.json`: sanitized
  structural summary of the adopted `system_cfg`.

Current finding summary:

- `mcad` is the controller-facing management agent.
- `udapi-bridge` and `ubios-udapi-server` provide local gateway state and
  configuration data to `mcad`.
- Portal adoption succeeds once the controller accepts a default-key inform and
  returns `setparam` with `mgmt_cfg`.
- The firmware immediately switches to the adopted inform key and AES-GCM.
- Adopted `system_cfg` is provisioning data, not safe host instructions for
  `unifi-stubd`.

Next research steps:

- Compare controller responses across UniFi Network versions.
- Extract additional safe gateway status fields from adopted informs.
- Keep raw adopted captures local only.

## uxg-lite-5.0.16

Status: partial UbiOS userspace simulation; blocked before `mcad` control
socket.

- Stub profile: `uxg-lite`
- Device type: `uxg`
- Model: `UXG`
- Firmware: `5.0.16.30689`
- Architecture: `arm64`

Research folder:

```text
research/firmware/uxg-lite-5.0.16/
```

Committed profile artifacts:

- `README.md`: firmware image inventory, rootfs metadata, service chain,
  current simulation result, and source availability notes.
- `source-inventory.md`: project-owned helper source vs observed vendor files.
- `simulation/Dockerfile`: wrapper around the imported local firmware rootfs
  image.
- `simulation/start-firmware-processes.sh`: starts the UbiOS processes while
  allowing the currently partial startup path to remain inspectable.
- `simulation/compose.yaml`: networkless ARM64 simulation.
- `simulation/docker-howto.md`: rootfs import, mock hardware, shim build, and
  startup instructions.

Current finding summary:

- UXG-Lite uses the same broad UbiOS `ubios-udapi-server` ->
  `udapi-bridge` -> `mcad` architecture as UXG-Pro.
- The local wrapper can start `ubios-udapi-server`, `udapi-bridge`, and
  `mcad`.
- In the current containerized run, `ubios-udapi-server` creates the bridge
  event notifier socket but not `/var/run/ubnt-udapi-server.sock`.
- Because that socket is missing, `mcad` does not expose `/tmp/.mcad` yet, so
  this profile is not ready for controller adoption.

Next research steps:

- Add ARM64 `strace` or broader shim tracing to identify the missing runtime
  dependency before the UDAPI server socket bind.
- Extend the mock hardware/sysctl set only with deterministic lab values.
- Do not connect this profile to a controller until `mca-ctrl -t dump` works.

## ucg-fiber-5.0.16

Status: partial UbiOS userspace simulation; blocked before `mcad` control
socket.

- Stub profile: `ucg-fiber`
- Device type: `udm`
- Model: `UCGF`
- Firmware: `5.0.16`
- Architecture: `arm64`

Research folder:

```text
research/firmware/ucg-fiber-5.0.16/
```

Committed profile artifacts:

- `README.md`: firmware image inventory, rootfs metadata, service chain,
  current simulation status, and source availability notes.
- `source-inventory.md`: project-owned helper source vs observed vendor files.
- `lab/gateway-profiles/ucg-fiber/Dockerfile`: wrapper around the imported
  local firmware rootfs image.
- `lab/gateway-profiles/ucg-fiber/start-firmware-processes.sh`: starts
  `ubios-udapi-server`, `udapi-bridge`, and `mcad`.
- `lab/gateway-profiles/ucg-fiber/compose.yaml`: networkless ARM64 simulation.
- `lab/gateway-profiles/ucg-fiber/docker-howto.md`: rootfs import, mock
  hardware, shim build, and startup instructions.

Current finding summary:

- UCG-Fiber uses the UbiOS `ubios-udapi-server` -> `udapi-bridge` -> `mcad`
  process shape seen in the other ARM64 gateway profiles.
- The firmware rootfs is Debian GNU/Linux 11 (`bullseye`) for the `ipq9574`
  platform.
- The board config identifies product `ucg-fiber`, board ID `a6a8`, model
  `UCGF`, SFP+ interfaces on `eth5` and `eth6`, and WAN mappings `wan0` on
  `eth4`, `wan1` on `eth6`.
- The local wrapper can start `ubios-udapi-server`, `udapi-bridge`, and
  `mcad`.
- In the current containerized run, `ubios-udapi-server` creates the bridge
  event notifier socket but not `/var/run/ubnt-udapi-server.sock`.
- Because that socket is missing, `mcad` does not expose `/tmp/.mcad` yet, so
  this profile is not ready for controller adoption.

Next research steps:

- Add ARM64 `strace` or broader shim tracing to identify the missing runtime
  dependency before the UDAPI server socket bind.
- Add deterministic mock paths only as the firmware startup logs require them.
- Keep the profile away from a controller until local `mca-ctrl` access works.

## udm-pro-se-5.0.16

Status: Docker wrapper reaches the UbiOS UDAPI socket and `mca-ctrl -t dump`
with a deterministic RTL8370-style switch mock.

- Device type: `udm`
- Model: `UDMPROSE`
- Firmware: `5.0.16`
- Image version string: `UDMPROSE.al324.v5.0.16.238fde6.260227.0037`
- Architecture: `arm64`

Research folder:

```text
research/firmware/udm-pro-se-5.0.16/
```

Committed profile artifacts:

- `README.md`: firmware image source, local artifact path, checksum, and
  current extraction status.
- `source-inventory.md`: project-owned notes vs ignored local vendor artifact
  boundary.
- `lab/gateway-profiles/udm-pro-se/Dockerfile`: wrapper around the imported
  local firmware rootfs image.
- `lab/gateway-profiles/udm-pro-se/start-firmware-processes.sh`: starts
  `ubios-udapi-server`, `udapi-bridge`, and `mcad`.
- `lab/gateway-profiles/udm-pro-se/compose.yaml`: networkless ARM64
  simulation.
- `lab/gateway-profiles/udm-pro-se/docker-howto.md`: rootfs import, mock
  hardware, shim build, and startup instructions.
- `lab/gateway-profiles/udm-pro-se/fixtures/mca-dump-summary.json`: sanitized
  local `mca-ctrl -t dump` summary.

Current finding summary:

- The official firmware image was downloaded from the Ubiquiti firmware CDN.
- The local artifact and extracted `rootfs.squashfs` are ignored under
  `artifacts/`.
- The image header reports
  `UBNTUDMPROSE.al324.v5.0.16.238fde6.260227.0037`.
- The file type is currently reported as `HIT archive data`.
- The rootfs is SquashFS 4.0 with zstd compression and Debian GNU/Linux 11
  (`bullseye`) userspace.
- The board config identifies product `udm-pro-se`, board ID `ea2c`, model
  `UDM-SE`, WAN mappings `wan0` on `eth8`, `wan1` on `eth9`, and switch
  driver `RTL8370`.
- The same UbiOS `ubios-udapi-server` -> `udapi-bridge` -> `mcad` process
  shape is present.
- The networkless wrapper redirects board, sysctl, MTD, sysfs, and persistent
  paths into `/mock`.
- The RTL8370 is not emulated at register level; the wrapper exposes a
  deterministic userspace `libsw.so`/OpenWrt `swconfig` ABI mock.
- `ubios-udapi-server` creates `/var/run/ubnt-udapi-server.sock`, and
  `udapi-bridge` exchanges internal UDAPI requests through it.
- `mcad` creates `/tmp/.mcad`, and `mca-ctrl -t dump` returns a usable local
  management dump with the mocked identity.
- `if_table` and `network_table` are still empty until the lab provides
  deterministic `switch0` and `eth0` through `eth10` netdevs.
- No controller adoption lab has been completed for this profile yet.

Next research steps:

- Add deterministic Linux netdev and netlink behavior for `switch0` and `eth0`
  through `eth10`.
- Attach the profile to a controller/MITM lab and sanitize the adoption
  findings before committing them.

## ugw3-4.4.57

Status: QEMU-MIPS chroot simulation starts `mcad` and supports `mca-ctrl`.

- Stub profile: `ugw3`
- Device type: `ugw`
- Model: `UGW3`
- Firmware: `4.4.57.5578372`
- Architecture: `mips` userspace under QEMU

Research folder:

```text
research/firmware/ugw3-4.4.57/
```

Committed profile artifacts:

- `README.md`: firmware tar inventory, rootfs metadata, legacy service chain,
  current simulation result, and source availability notes.
- `source-inventory.md`: project-owned helper source vs observed vendor files.
- `simulation/Dockerfile`: Debian/QEMU-MIPS runner.
- `simulation/start-ugw3-qemu.sh`: chroots into the extracted rootfs and starts
  legacy `mcad`.
- `simulation/compose.yaml`: runner using the external
  `unifi-ugw3-rootfs` Docker volume.
- `simulation/docker-howto.md`: extraction and runtime commands.

Current finding summary:

- UGW3 is not UbiOS. It uses the legacy USG/EdgeOS `mcad`, `mca-monitor`,
  `linkcheck`, `syswrapper.sh`, and Vyatta stack.
- The imported rootfs cannot be launched directly as a Docker image on the
  current host because the userspace is 32-bit big-endian MIPS.
- The committed Debian/QEMU runner starts `mcad` and creates `/tmp/.mcad`.
- `mca-ctrl -t dump` works and shows default inform behavior toward
  `http://unifi:8080/inform`.
- Hardware identity is still placeholder data until a legacy board and
  interface mock is added.

Next research steps:

- Add deterministic board, serial, MAC, and interface mocks for UGW3.
- Attach the runner to the disposable controller/MITM lab only after identity
  fields are stable.
- Capture only sanitized inform summaries.

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

## uxg-lite-real-firmware

Status: planned.

- Stub profile: `uxg-lite`
- Device type: `uxg`
- Model: `UXG`

Research folder target:

```text
research/firmware/uxg-lite-<version>/
```

Before simulation:

- Download the matching public firmware image into an ignored `artifacts/`
  directory.
- Record product, firmware filename, header, SHA-256, architecture, and rootfs
  format.
- Inventory the boot/service chain and confirm whether it uses the same
  UbiOS `mcad`/UDAPI process model as UXG-Pro.
- If compatible, reuse the generic firmware wrapper variables:
  `UNIFI_FW_SIM_MODEL`, `UNIFI_FW_SIM_MAC`,
  `UNIFI_FW_SIM_STATIC_ADDRESS`, and `UNIFI_FW_SIM_DUMMY_INTERFACES`.
- If not compatible, create a profile-specific wrapper instead of forcing the
  UXG-Pro start flow.

Expected committed artifacts:

- `README.md`: firmware image inventory and findings.
- `source-inventory.md`: observed vendor files and project-owned helpers.
- `simulation/`: only project-owned wrapper files and docs.
- `simulation/fixtures/`: sanitized JSON summaries only.

## ugw3-real-firmware

Status: planned.

- Stub profile: `ugw3`
- Device type: `ugw`
- Model: `UGW3`

Research folder target:

```text
research/firmware/ugw3-<version>/
```

Before simulation:

- Download the matching public firmware image into an ignored `artifacts/`
  directory.
- Record product, firmware filename, header, SHA-256, architecture, and rootfs
  format.
- Inventory the agent stack. This profile may use a legacy USG/EdgeOS flow
  rather than the newer UbiOS `mcad`/UDAPI chain.
- Identify the inform/adoption process names and local control sockets before
  writing any wrapper.
- Keep extracted configs and raw rootfs files out of Git.

Expected committed artifacts:

- `README.md`: firmware image inventory and findings.
- `source-inventory.md`: observed vendor files and project-owned helpers.
- `simulation/`: profile-specific wrapper only after the service chain is
  known.
- `simulation/fixtures/`: sanitized discovery, inform, and adoption summaries.

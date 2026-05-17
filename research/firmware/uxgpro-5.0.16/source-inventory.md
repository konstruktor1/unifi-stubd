# Source Inventory

This inventory separates project-owned helper source from vendor artifacts.

## Project-Owned Helper Source

| Path | Purpose |
| --- | --- |
| `simulation/ubnthal_redirect.c` | LD_PRELOAD shim used to redirect selected firmware reads from `/proc/ubnthal` and `/proc/sys` into mock files during isolated Docker simulation. |

## Vendor Files Observed, Not Copied

The following files were inspected locally from the mounted rootfs, but are not
copied into this repository:

| Vendor path | Role |
| --- | --- |
| `/usr/bin/mcad` | Management Console Agent; also target of `mca-ctrl`, `mca-cli`, and `mca-cli-op` symlinks. |
| `/usr/bin/udapi-bridge` | Local REST bridge for UDAPI. |
| `/usr/bin/ubios-udapi-server` | Local gateway configuration engine. |
| `/usr/bin/syswrapper.sh` | Command dispatcher used by management actions and adoption helpers. |
| `/usr/bin/ubnt-shadow-mode-adopt` | Shadow-mode adoption helper script. |
| `/usr/share/ubios-udapi-server/config-board/uxg-pro-ea19.json` | Board-specific UDAPI config. |
| `/usr/share/ubios-udapi-server/uxg-pro-ea19.default` | Default gateway configuration. |
| `/usr/share/ubios-udapi-server/uxg-pro-ea19.fallback` | Fallback gateway configuration. |
| `/lib/systemd/system/mcagent.service` | Starts `mcad`. |
| `/lib/systemd/system/udapi-server.service` | Starts `ubios-udapi-server`. |
| `/lib/systemd/system/udapi-bridge.service` | Starts `udapi-bridge`. |

## Safe Reuse Decision

- Do not copy vendor firmware source, scripts, binaries, or structured default
  configurations into the project.
- Keep protocol observations, command names, paths, hashes, and own helper code.
- If external source is ever copied intentionally, update `CREDITS.md`,
  `NOTICE.md`, and the license decision first.

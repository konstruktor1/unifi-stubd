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

## Online Source Search

Checked on 2026-05-17:

- Ubiquiti's public GitHub organization does not publish a UXG-Pro firmware
  source tree or the `mcad`, `udapi-bridge`, or `ubios-udapi-server` sources.
- Ubiquiti's public firmware update docs and release flow expose firmware
  download links, not complete firmware source.
- Community support threads point GPL-covered source requests to Ubiquiti
  support channels, but those requests do not make the proprietary UniFi agent
  sources public.

References:

- https://github.com/ubiquiti
- https://help.ui.com/hc/en-us/articles/204910064-UniFi-Advanced-Updating-Techniques
- https://community.ui.com/questions/Source-Code-UniFi/967271b5-bbd0-4c4a-84af-1c01ddb95a8c
- https://community.ui.com/questions/GPL-Source-Code-UDM/c05784b8-1fe5-4055-9a3a-68a18f559ea7

## Safe Reuse Decision

- Do not copy vendor firmware source, scripts, binaries, or structured default
  configurations into the project.
- Keep protocol observations, command names, paths, hashes, and own helper code.
- If external source is ever copied intentionally, update `CREDITS.md`,
  `NOTICE.md`, and the license decision first.

# UDM Pro SE Source Inventory

This inventory separates project-owned helper code from observed vendor
firmware paths. It is intentionally structural only.

## Project-Owned Helpers

| Path | Purpose |
| --- | --- |
| `README.md` | Profile overview, current state, and source layout. |
| `docker-howto.md` | Local firmware import, mock preparation, kernel payload staging, Docker startup, and webportal checks. |
| `firmware.md` | Safe firmware findings, simulation status, and current limitations. |
| `Dockerfile` | Wrapper around a locally imported ARM64 firmware rootfs. |
| `compose.yaml` | Networkless firmware process simulation and shim-builder profile. |
| `webportal.compose.yaml` | Localhost-only webportal override for setup UI inspection. |
| `start-firmware-processes.sh` | Thin orchestrator for the reduced firmware process modules under `runtime/firmware/`. |
| `start-webportal-processes.sh` | Thin orchestrator for the webportal modules and assets under `runtime/webportal/`. |
| `udapi-lab-shim.cjs` | Read-only wrapper for UDAPI and MCA queries that need deterministic Docker WAN metadata. |
| `runtime/README.md` | Runtime helper layout and source boundary. |
| `runtime/common/kernel-artifacts.sh` | Shared kernel payload manifest logging. |
| `runtime/firmware/config.sh` | Environment-driven configuration for the reduced firmware process chain. |
| `runtime/firmware/processes.sh` | Firmware process startup, readiness waiting, tracing, and shutdown helpers. |
| `runtime/webportal/config.sh` | Environment-driven configuration for the webportal path. |
| `runtime/webportal/install-wrappers.sh` | Installs runtime wrapper files into firmware paths. |
| `runtime/webportal/http.sh` | Applies nginx config snippets and generated-config filters. |
| `runtime/webportal/services.sh` | Starts PostgreSQL, DBus, facades, `ulp-go`, nginx, and `unifi-core`. |
| `runtime/webportal/data/ubnt-tools-id.txt` | Deterministic `ubnt-tools id` payload. |
| `runtime/webportal/http/shared-runnable-lab.conf` | Lab-only nginx locations for optional UniFi Core sidecar endpoints. |
| `runtime/webportal/http/site-local-ip-preview.awk` | Filter that keeps the HTTP preview on host port `9080` usable. |
| `runtime/webportal/http/site-setup-api-guard.awk` | Filter that blocks reset/reboot setup routes. |
| `runtime/webportal/templates/unifi-core-default.yaml.in` | UniFi Core override template. |
| `runtime/webportal/templates/unifi-core-lab.sudoers` | Narrow sudoers rule for the lab `systemctl` wrapper. |
| `runtime/webportal/wrappers/systemctl` | Lab systemctl implementation for known UniFi service lifecycle calls. |
| `runtime/webportal/wrappers/systemd-run` | Lab systemd-run wrapper for approved transient support commands. |
| `runtime/webportal/wrappers/tar` | Support-bundle tar wrapper that tolerates readable archives with changing logs. |
| `runtime/webportal/wrappers/timedatectl` | Minimal non-systemd timedatectl compatibility wrapper. |
| `runtime/webportal/wrappers/ubnt-systool` | Lab ubnt-systool compatibility wrapper. |
| `runtime/webportal/wrappers/ubnt-tools` | Lab ubnt-tools identity wrapper. |
| `runtime/webportal/wrappers/udapi-tool` | Shared entry wrapper for `mca-ctrl`, `mca-dump`, and `ubios-udapi-client`. |
| `network-app/config.cjs` | Tunables for the local Network facade. |
| `network-app/http.cjs` | JSON request/response helpers. |
| `network-app/index.cjs` | Network facade entry point and HTTP/websocket transport. |
| `network-app/logger.cjs` | Append-only request logging. |
| `network-app/payloads.cjs` | Deterministic Network/Core payload fixtures. |
| `network-app/routes.cjs` | Observed Core-facing Network endpoint table. |
| `network-app/websocket.cjs` | Minimal websocket handshake support for frontend bootstrap. |
| `systemd-dbus/config.cjs` | DBus service name, manager path, and keepalive settings. |
| `systemd-dbus/dbus.cjs` | Firmware-bundled `@jellybrick/dbus-next` binding wrapper. |
| `systemd-dbus/index.cjs` | systemd DBus facade entry point. |
| `systemd-dbus/interfaces.cjs` | Manager, Unit, and Service DBus interfaces. |
| `systemd-dbus/server.cjs` | DBus name claim and object export loop. |
| `systemd-dbus/units.cjs` | Deterministic active/inactive systemd unit fixtures. |
| `mock/README.md` | Mock tree overview. |
| `mock/files/ubnthal/board` | Deterministic board identity data. |
| `mock/files/ubnthal/system.info` | Deterministic system identity data. |
| `mock/ldpreload/auth.c` | Narrow lab root-user compatibility for UDAPI checks. |
| `mock/ldpreload/common.c` | Environment-gated feature flags. |
| `mock/ldpreload/fs_paths.c` | Path rewriting and file-descriptor tracking policy. |
| `mock/ldpreload/fs_open.c` | `open`, `fopen`, `access`, `stat`, and `lstat` interposition. |
| `mock/ldpreload/fs_io.c` | `write`, `send`, `sendmsg`, `writev`, `read`, `fread`, and `ioctl` interposition. |
| `mock/ldpreload/process_control.c` | Containment for host-management command execution. |
| `mock/ldpreload/response_patch.c` | Byte-preserving setup/readiness response patches. |
| `mock/ldpreload/socket_trace.c` | Optional socket boundary tracing. |
| `mock/ldpreload/swconfig.c` | Deterministic RTL8370-style `libsw.so`/OpenWrt `swconfig` ABI. |
| `mock/ldpreload/ubnthal_redirect.h` | Shared C ABI declarations and helper prototypes. |
| `fixtures/mca-dump-summary.json` | Sanitized local `mca-ctrl -t dump` summary. |

## Observed Vendor Components

Observed inside the extracted rootfs:

- `/usr/bin/mcad`
- `/usr/bin/mca-ctrl -> mcad`
- `/usr/bin/mca-cli -> mcad`
- `/usr/bin/mca-cli-op -> mcad`
- `/usr/bin/mca-monitor -> mcad`
- `/usr/bin/mca.sh`
- `/usr/bin/syswrapper.sh`
- `/usr/bin/ubios-udapi-server`
- `/usr/bin/ubios-udapi-client`
- `/usr/bin/udapi-bridge`
- `/usr/bin/udapic`
- `/usr/lib/libsw.so`
- `/lib/systemd/system/udapi-server.service`
- `/lib/systemd/system/udapi-bridge.service`
- `/lib/systemd/system/mcagent.service`
- `/usr/share/ubios-udapi-server/config-board/udm-pro-se-ea2c.json`
- `/usr/share/ubios-udapi-server/udm-pro-se-ea2c.default`
- `/usr/share/ubios-udapi-server/udm-pro-se-ea2c.fallback`
- `/etc/lagd/configs/udmse-ea2c.json`
- `/lib/modules/4.19.152-ui-alpine/`

These files are not copied into Git. Keep all extracted vendor content under
ignored `artifacts/`, `rootfs/`, or Docker volume storage.

## License Boundary

Use the vendor binaries only as local research inputs. If a future change needs
to copy source code, structured data, scripts, or configuration from external
projects or firmware content, update `CREDITS.md`, `NOTICE.md`, and the license
decision before merging.

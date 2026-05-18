# Project Status

Last updated: 2026-05-18.

`unifi-stubd` has two related but separate lab tracks:

- The main Go service is a safe UniFi device stub for isolated lab or
  management networks. It sends discovery and inform traffic, supports adoption
  state persistence, and must not let a controller blindly provision the host.
- `lab/gateway-profiles/` contains firmware research profiles. Those profiles
  are not product code and they do not add full gateway behavior to
  `unifi-stubd`; they are controlled lab references for understanding what the
  real firmware expects.

## Main Service

The main service currently provides:

- Discovery and inform framing for synthetic UniFi device identities.
- Built-in switch and experimental gateway identity profiles.
- YAML configuration, OpenRC/systemd units, and package builds.
- Stub, observe, and packaging workflows documented in this repository.
- An adoption SSH shim that accepts the minimal advanced-adoption command shape
  without executing arbitrary controller shell commands.

The safety boundary remains unchanged: controller-triggered restart, upgrade,
shell, and host-networking changes must stay explicit, lab-scoped, and
reviewable.

## UDM Pro SE VM Reference

The UDM Pro SE QEMU/UTM work is a firmware reference path, not a replacement
for the Go stub.

What is working:

- The direct vendor kernel and vendor U-Boot payload have been tried on QEMU
  `virt`; they do not produce useful serial output there.
- A foreign QEMU-virt-capable ARM64 kernel can boot the UDM initramfs/rootfs
  path far enough to reach the UDM firmware `systemd`.
- The lab initramfs keeps the UDM initramfs, UDM rootfs, overlay layout, and
  final `/sbin/init` path, but patches the VM-only hardware boundary.
- The mocked boot reaches UDM firmware `systemd`, installs userspace hardware
  mocks, starts UDM-facing services such as `uhwd`, `usd`, `rpsd`, nginx, SSH,
  UniFi OS Agent, and `ubios-udapi-server`, and reaches the serial login
  prompt.
- `network-init.service` completes in the mocked path and creates the expected
  UDM-style switch/LAN devices.
- The web setup surface can be reached through the lab network configuration
  when the VM is running.
- The latest UTM full test applied the profile to `UDM-Pro-SE-QEMU`, booted
  the serial-only 4 GiB VM, reached the web setup surface on the UTM Shared/NAT
  guest address, and returned `/api/system` with `hasInternet=true`.
- In that UTM run, Network-facing API state exposed `eth9` as the SFP+ WAN
  role, kept `eth8` attached to `br0`, and kept `br0` on `192.168.1.1/24` plus
  the host-only `192.168.128.2/24` alias.
- Browser login and refresh were stable when the same host name was used
  consistently. Switching between `localhost` and `127.0.0.1` can create a
  separate cookie scope and look like a login loop.

Current UTM mapping:

- `eth9`: first 10G SFP+ WAN role, backed by UTM `Shared` / NAT.
- `eth8`: 2.5G RJ45 LAN role, backed by UTM `Host` networking and kept on
  `br0`.
- The VM uses 4 GiB RAM and a serial-only display path.
- The installer writes a UTM `Network:0:PortForward` entry for host
  `127.0.0.1:10443` to guest `443`, but the latest observed UTM CLI run did
  not bind that host port by itself. Use the direct UTM Shared/NAT guest HTTPS
  address when available, or an explicit local TCP forwarder, until the native
  UTM forward behavior is verified.

How to read the test results:

- The Docker test is the fastest way to inspect UniFi OS setup UI behavior,
  Core/API expectations, support-bundle generation, and the project-owned
  facades. It is weaker for boot validation because it uses the host kernel,
  Docker networking, Docker process supervision, and lab wrappers instead of a
  real VM boot path.
- The UTM test is the better reference for "does this look like a booting UDM?"
  because it runs a full ARM64 VM, enters the UDM initramfs/rootfs handoff,
  starts firmware `systemd`, exercises the firmware network initialization, and
  gives the guest NICs stable UDM-like roles.
- The direct QEMU-only tests failed at the native vendor hardware boundary: the
  vendor kernel and vendor U-Boot payload do not produce useful serial output
  on QEMU `virt` because that machine is not the Annapurna Labs AL324 board.
  The foreign-kernel QEMU path works as a lab boot path, but it is less useful
  than UTM for the Mac-side browser/network validation because the UTM profile
  owns the two-NIC Shared/NAT plus Host networking shape used in the full test.

Important limitation:

- The native Network self-provisioning/self-inform path is still under
  investigation. A host-side `unifi-stubd` inform was used only as a diagnostic
  proof that the Network application can adopt the device payload. It is not
  the target design for the VM reference.

Known remaining gaps:

- QEMU `virt` is not an Annapurna Labs AL324 board model.
- Vendor-specific kernel modules such as `xt_dpi` and `xt_dyn_random` are not
  available in the foreign kernel.
- Some late services still fail or restart, including Bluetooth, logging,
  directory, and `mcagent` paths.
- Dummy VM devices still produce some `ethtool` warnings.
- `/api/setup/support/generate` is reliable in the Docker webportal path, but
  timed out during the latest UTM VM run and still needs VM-side debugging.
- Native UTM localhost forwarding for `127.0.0.1:10443` is configured by the
  profile but not proven by the latest test. Do not confuse a helper TCP
  forwarder with UTM's own port-forward implementation.

## UDM Pro SE Docker Webportal

The Docker UDM Pro SE profile is a setup/API inspection path, not the native VM
boot reference.

What is working:

- The networkless Docker path starts `ubios-udapi-server`, `udapi-bridge`, and
  `mcad` against deterministic `/mock` hardware inputs.
- The modular C `LD_PRELOAD` shim under `mock/ldpreload/` redirects selected
  `/proc`, `/sys`, MTD, and persistence paths, exposes the RTL8370-style
  `swconfig` ABI, and keeps unsafe process actions contained.
- `webportal.compose.yaml` exposes the setup UI on
  `https://127.0.0.1:9443/` and an HTTP preview on `http://127.0.0.1:9080/`.
- The modular CommonJS Network facade under `network-app/` reports Network as
  installed/running, publishes the packaged UI manifest, and returns
  deterministic setup payloads.
- The modular CommonJS systemd DBus facade under `systemd-dbus/` lets UniFi
  Core inspect known service units without running systemd as PID 1.
- `udapi-lab-shim.cjs` maps Docker `eth0` to a UDM-style WAN view and returns
  deterministic DNS/ISP metadata for internet-readiness checks.
- The latest Docker full test rebuilt the ARM64 C shim, rebuilt the
  `udm-pro-se-fw-sim:5.0.16` image, started both the networkless firmware path
  and the webportal path, and confirmed `unifi-core`, `ulp-go`, nginx,
  PostgreSQL, the DBus facade, the Network facade, UDAPI, `udapi-bridge`, and
  `mcad` were running.
- Docker HTTPS on `https://127.0.0.1:9443/` returned the UniFi OS setup HTML
  with `UNIFI_OS_MANIFEST`; `/api/system` reported internet readiness after
  the Core/UDAPI readiness state settled.
- Docker support-bundle generation through `/api/setup/support/generate`
  returned a local archive containing lab system metadata and `unifi-core`
  logs.

Important limitation:

- Docker still runs on the host kernel and does not boot a UDM kernel. It
  mounts the shared deployable kernel payload only for comparison and logging.
  Use the QEMU/UTM profile when the question is native firmware boot behavior.
- Docker success therefore means "the UI/API compatibility layer is plausible";
  it does not prove kernel boot, initramfs handoff, systemd unit ordering,
  VM network enumeration, or the native Network self-view.

## Where The Work Lives

Primary entry points:

- `lab/gateway-profiles/udm-pro-se-vm/README.md`: QEMU/UTM runbook.
- `lab/gateway-profiles/udm-pro-se-vm/firmware.md`: firmware findings and boot
  status.
- `lab/gateway-profiles/udm-pro-se-vm/source-inventory.md`: VM profile source
  boundary.
- `lab/gateway-profiles/udm-pro-se/source-inventory.md`: UDM Pro SE Docker and
  mock source boundary.

Project-owned VM payloads:

- `lab/gateway-profiles/udm-pro-se-vm/initramfs/`: initramfs hooks, module
  lists, rootfs payload files, systemd units, drop-ins, and HTTP/SSH templates
  injected into the VM path.
- `lab/gateway-profiles/udm-pro-se-vm/utm/`: versioned UTM profile inputs,
  including default VM/network values, kernel boot arguments, and notes for
  the generated UTM bundle files kept outside Git.
- `lab/gateway-profiles/udm-pro-se-vm/kernel/`: notes for the shared ignored
  kernel deployment payload under `artifacts/deploy/kernel/`, used by both UTM
  and Docker.
- `lab/gateway-profiles/udm-pro-se/udapi-lab-shim.cjs`: Docker webportal UDAPI
  read wrapper for WAN, DNS, and ISP metadata.
- `lab/gateway-profiles/udm-pro-se/runtime/`: sourced Docker startup modules,
  wrapper scripts, nginx snippets, AWK filters, templates, and deterministic
  lab data used by the firmware and webportal entry points.
- `lab/gateway-profiles/udm-pro-se/network-app/`: modular CommonJS Network
  facade used by the Docker webportal path, split into configuration, HTTP,
  logging, payload, route, and websocket modules.
- `lab/gateway-profiles/udm-pro-se/systemd-dbus/`: modular CommonJS
  `org.freedesktop.systemd1` facade for the Docker webportal path, split into
  DBus binding, unit fixture, interface, and server modules.
- `lab/gateway-profiles/udm-pro-se/mock/files/`: deterministic mock filesystem
  inputs copied to `/mock`.
- `lab/gateway-profiles/udm-pro-se/mock/ldpreload/`: modular C `LD_PRELOAD`
  shim source:
  - `common.c`: feature flags.
  - `auth.c`: narrow lab root-user compatibility.
  - `response_patch.c`: byte-preserving setup/readiness response patches.
  - `swconfig.c`: RTL8370-style `libsw.so`/OpenWrt `swconfig` ABI.
  - `fs_paths.c`, `fs_open.c`, `fs_io.c`, `process_control.c`, and
    `socket_trace.c`: `/proc`, `/sys`, MTD, process, socket, and syscall
    interposition.

Ignored generated artifacts:

- `lab/gateway-profiles/udm-pro-se-vm/artifacts/`: copied firmware, extracted
  boot artifacts, VM disks, fetched toolchains, built mock roots, and generated
  initramfs images.

## Rebuild The Current VM Path

```sh
lab/gateway-profiles/udm-pro-se-vm/scripts/prepare-vm.sh
lab/gateway-profiles/udm-pro-se-vm/scripts/fetch-foreign-kernel.sh
lab/gateway-profiles/udm-pro-se-vm/scripts/prepare-mocks.sh
lab/gateway-profiles/udm-pro-se-vm/scripts/build-lab-initramfs.sh
lab/gateway-profiles/udm-pro-se-vm/scripts/deploy-kernel-artifacts.sh
```

Run directly with QEMU:

```sh
UDM_PRO_SE_FOREIGN_MODE=udm-systemd \
  lab/gateway-profiles/udm-pro-se-vm/scripts/run-foreign-kernel.sh
```

Configure a cloned UTM profile:

```sh
lab/gateway-profiles/udm-pro-se-vm/scripts/install-utm-profile.sh
```

Keep firmware images, extracted firmware files, captures, keys, tokens,
certificates, controller URLs, and private lab data out of Git.

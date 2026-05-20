# UDM Pro SE QEMU VM Profile

This profile is for real `qemu-system-aarch64` VM boot attempts using the UDM
Pro SE firmware image. It does not use Docker, Compose, chroot, or
`qemu-aarch64-static`.

Current status is summarized in `../../../docs/en/project-status.md` and
`../../../docs/de/project-status.md`. In short: the direct vendor-kernel path
does not boot on QEMU `virt`, but the foreign-kernel `udm-systemd` path reaches
UDM firmware `systemd`, starts the web setup surface, and is useful as a
reference for native UDM behavior.

The local firmware image and extracted VM artifacts live under ignored
`artifacts/` paths. The profile scripts copy the firmware image into this
directory before extracting:

- `uboot.bin`: vendor U-Boot payload from the HIT image.
- `kernel.fit`: vendor FIT image.
- `kernel.Image`: decompressed ARM64 Linux kernel from the FIT.
- `udm-pro-se.dtb`: vendor AL324 UDM Pro SE device tree from the FIT.
- `initramfs.cpio.gz`: vendor initramfs from the FIT.
- `rootfs.squashfs`: vendor root filesystem.
- `vm-disk.raw`: GPT disk with UDM-style partition labels.
- `lab-initramfs.cpio.gz`: UDM initramfs with the smallest QEMU VM lab patch
  needed for a systemd boot attempt.
- `mock-root/`: deterministic userspace hardware files and an ARM64
  `LD_PRELOAD` shim, staged into the UDM root by the lab initramfs.
- `deploy/kernel/`: deployable local kernel payload shared by UTM and Docker.

Prepare artifacts:

```sh
lab/gateway-profiles/udm-pro-se-vm/scripts/prepare-vm.sh
```

Run the direct kernel VM attempt:

```sh
lab/gateway-profiles/udm-pro-se-vm/scripts/run-direct-kernel.sh
```

Run the direct kernel with the SquashFS as a single virtio root block:

```sh
UDM_PRO_SE_VM_MODE=rootfs-block \
  lab/gateway-profiles/udm-pro-se-vm/scripts/run-direct-kernel.sh
```

Try the vendor U-Boot payload as QEMU firmware:

```sh
lab/gateway-profiles/udm-pro-se-vm/scripts/run-vendor-uboot.sh
```

Fetch and run a QEMU-virt-capable foreign ARM64 kernel for comparison. This
also downloads the matching Debian kernel module package into ignored
artifacts:

```sh
lab/gateway-profiles/udm-pro-se-vm/scripts/fetch-foreign-kernel.sh
lab/gateway-profiles/udm-pro-se-vm/scripts/run-foreign-kernel.sh
```

Run the foreign kernel with the UDM initramfs and VM disk:

```sh
UDM_PRO_SE_FOREIGN_MODE=udm-initramfs \
  lab/gateway-profiles/udm-pro-se-vm/scripts/run-foreign-kernel.sh
```

Build the userspace hardware mocks, then build the lab initramfs and try to
reach the UDM firmware `systemd`:

```sh
lab/gateway-profiles/udm-pro-se-vm/scripts/fetch-foreign-kernel.sh
lab/gateway-profiles/udm-pro-se-vm/scripts/prepare-mocks.sh
lab/gateway-profiles/udm-pro-se-vm/scripts/build-lab-initramfs.sh
lab/gateway-profiles/udm-pro-se-vm/scripts/deploy-kernel-artifacts.sh
UDM_PRO_SE_FOREIGN_MODE=udm-systemd \
  lab/gateway-profiles/udm-pro-se-vm/scripts/run-foreign-kernel.sh
```

The files injected by `build-lab-initramfs.sh` live under `initramfs/`:
early initramfs hooks, the module list shared by the builder and guest, and the
rootfs payload that becomes systemd units, helper scripts, and HTTP/SSH
templates inside the VM.

The project-owned UTM profile inputs live under `utm/`. `utm/defaults.env`
contains the default UTM VM and network values, `utm/bootargs.txt` contains the
kernel command line injected by the installer, `utm/install/` contains the
split installer modules, and `utm/README.md` describes which generated UTM
bundle files stay outside Git.

The project-owned kernel deployment note lives under `kernel/`. The deploy
script writes ignored artifacts under `artifacts/deploy/kernel/`: vendor kernel
pieces, the QEMU-virt-capable foreign kernel and module tree, and the lab
initramfs. `install-utm-profile.sh` prefers that staged kernel payload when it
exists, and the Docker profile mounts the same directory read-only.

## Project-Owned Source Layout

This VM profile is split by responsibility:

- `artifacts/`: ignored local firmware, extracted boot pieces, VM disks, logs,
  generated initramfs images, fetched kernels/modules, and UTM build output.
- `scripts/prepare-vm.sh`: copies the firmware image into ignored artifacts and
  extracts boot/rootfs inputs.
- `scripts/fetch-foreign-kernel.sh`: fetches the QEMU-virt-capable ARM64 kernel
  and matching module package used for the mixed boot path.
- `scripts/prepare-mocks.sh`: builds the ARM64 C shim from
  `../udm-pro-se/mock/ldpreload/` and stages deterministic `/mock` files.
- `scripts/build-lab-initramfs.sh`: builds the UDM-based initramfs that patches
  only the VM hardware boundary before handing off to firmware `systemd`.
- `scripts/deploy-kernel-artifacts.sh`: stages the shared ignored payload used
  by UTM boot inputs and Docker inspection.
- `scripts/install-utm-profile.sh`: orchestrates the split UTM installer modules
  in `utm/install/`.
- `qemu/`: direct-QEMU environment presets for transparent LAN and older
  user-mode forwarding experiments.
- `initramfs/`: hooks, module lists, static initramfs metadata, and
  `rootfs-payload/` files injected into the mounted UDM root. Its internal
  target-root layout is described in `initramfs/README.md`.
- `utm/`: reproducible UTM defaults, boot arguments, and installer modules. The
  split installer files under `utm/install/` are described in `utm/README.md`.
- `kernel/`: documentation for the shared ignored deployable kernel payload.

For the UTM profile, the lab uses two virtual NICs that map to physical UDM SE
port roles:

- UTM `Shared` / NAT maps to `eth9`, the first 10G SFP+ WAN.
- UTM `Host` maps to `eth8`, the 2.5G RJ45 port, which the lab keeps attached
  to `br0` with `192.168.1.1/24` plus a host-only access alias
  `192.168.128.2/24` for macOS UTM's default host bridge.

The other UDM-facing lab devices remain internal dummy devices in the guest.

```text
host-only <-> guest eth8 <-> guest br0 192.168.1.1/24, 192.168.128.2/24
shared NAT <-> guest eth9 <-> SFP+ WAN
```

After the `udm-systemd` boot reaches nginx, test the transparent LAN web
surface from the host or isolated lab segment:

```sh
curl -k https://192.168.1.1/
```

For the UTM profile, also test the Shared/NAT guest address assigned by UTM.
The exact address is assigned by UTM and can change between runs:

```sh
curl -k https://<utm-shared-guest-ip>/
curl -k https://<utm-shared-guest-ip>/api/system
```

The lab initramfs keeps the QEMU-backed `eth8` LAN member attached to `br0`
and installs the vendor `site-setup.conf` nginx template when UniFi Core has
not yet produced dynamic HTTP config. This keeps the web path on guest port
`443`; it does not open WAN ingress.

The UTM installer writes a plist port-forward entry intended to expose guest
`443` as `https://127.0.0.1:10443/`, but the latest observed UTM CLI run did
not bind that host port by itself. In that run, direct guest HTTPS worked, and
localhost `10443` access was only reliable when an explicit local TCP forwarder
was running. Verify the listener before treating `10443` as native UTM
behavior.

The default host interface is `en0`; override it with
`UDM_PRO_SE_VM_VMNET_IFNAME=<interface>`. macOS `vmnet-bridged` commonly
requires QEMU to run with the required privileges. Set `UDM_PRO_SE_VM_NET=none`
to disable networking or `UDM_PRO_SE_VM_NET=default` to return to QEMU's
default NIC behavior. The older QEMU user-mode forwarding paths remain
available explicitly with `UDM_PRO_SE_VM_NET=user-lan` or
`UDM_PRO_SE_VM_NET=user-wan`, but they are not the default web path.

## Choosing A Test Path

Use the Docker webportal path when the question is about UniFi OS setup UI,
Core API calls, support-bundle generation, or the project-owned facade code. It
starts quickly and is easy to inspect, but it is not a firmware boot proof: the
kernel is the host kernel, networking is Docker networking, and several UniFi OS
expectations are satisfied by lab wrappers.

Use the UTM profile when the question is native boot behavior. UTM still uses
QEMU under the hood, but the profile persists the VM shape, boot inputs, serial
console, RAM, disk, and two NICs in one place. That made it the useful full-test
target for the web portal because the guest sees `eth9` as SFP+ WAN through
Shared/NAT and `eth8` as the 2.5G LAN on `br0`.

Use the direct QEMU scripts for narrow boot experiments and fast kernel/initrd
iteration. The direct native attempt failed at the board boundary:
`run-direct-kernel.sh` and `run-vendor-uboot.sh` did not produce useful serial
output with the vendor kernel or vendor U-Boot on QEMU `virt`. That is the
reason the working VM reference uses a foreign QEMU-virt-capable kernel while
keeping the UDM initramfs, rootfs, overlay layout, and final firmware
`/sbin/init`.

`prepare-mocks.sh` builds the C shim modules from
`lab/gateway-profiles/udm-pro-se/mock/ldpreload/` so the Docker and QEMU/UTM
paths share one mock implementation. It uses a local `zig` when available. If
none is found, it downloads a Zig toolchain into ignored `artifacts/toolchains/`
and cross-builds the shim for `aarch64-linux-gnu.2.31`.

Current limitation: QEMU does not provide an Annapurna Labs AL324 machine, so
the vendor UDM kernel has to be tried on QEMU `virt`. That is a real VM, but it
is not yet a faithful board model.

For the `udm-systemd` path, the kernel is foreign because it must support QEMU
`virt`. The initramfs, rootfs, overlay layout, and final `/sbin/init` are still
from the UDM firmware. The lab initramfs replaces the hardware-only
`/dev/mtdblock5` config store with a QEMU GPT partition labeled `config`, then
copies the mock hardware tree into the UDM root before `systemd` starts. The VM
disk is attached through QEMU USB storage in this mode because the current
foreign initrd carries USB storage modules but not `virtio_blk`.

Current result with mocks: `udm-systemd` reaches UDM firmware `systemd`, logs
`unifi-stubd-vm-mock: installed userspace hardware mocks`, starts `uhwd`,
`usd`, `rpsd`, nginx, SSH, UniFi OS Agent, and other UniFi OS services, and
reaches the serial login prompt. The lab initramfs prepares QEMU network devices
for the UDM board profile, then `network-init.service` succeeds on attempt 1:
it creates `eth0` through `eth7` as 802.1ad interfaces on `switch0`, applies
MAC addresses, and applies the UDM queue/link setup far enough for the lab
LAN/WAN mapping. The explicit
`udapi-server.service` command from the userspace lab starts
`ubios-udapi-server`, it listens on `/var/run/ubnt-udapi-server.sock`, and the
foreign-kernel netfilter modules are enough for `/firewall/nat` to report
`MASQUERADE` and `DNAT` rules. The later firewall path also reaches
`firewall/filter` and begins `firewall/mangle`. The previous board identity
`libubnt.get_board()` failure was not observed in the mocked boot log. The
current default network shape is transparent LAN bridging, not localhost
forwarding.

Latest UTM web result: the serial-only 4 GiB UTM VM reached the UniFi OS setup
surface, `/api/system` returned `hasInternet=true`, the Network-facing state
reported `eth9` as the SFP+ WAN role, and `br0` carried `192.168.1.1/24` plus
`192.168.128.2/24`. The Network SPA routes, including
`/network/default/integrations`, returned the UniFi OS/Network application HTML.
After login, reload stayed on the Network dashboard when the browser kept the
same host name. `/api/setup/support/generate` timed out in the VM path and
remains a debugging target; the same support-bundle endpoint is only known-good
in the Docker webportal path.

The VM is also a reference for what `unifi-stubd` may need to emulate later. It
does not embed or run the Go stub as the solution. A host-side inform test was
used only to prove that the Network application can adopt the payload shape.

Remaining failures are now later QEMU hardware/runtime gaps:
`svc-dpi-service` cannot load the vendor-specific `xt_dpi` module, WAN failover
cannot load the vendor-specific `xt_dyn_random` module,
`bluetooth-controller@hci0.service` fails, `ulogd2.service`,
`unifi-directory.service`, and `mcagent.service` restart or fail, and dummy QEMU
network devices still emit some `ethtool` warnings.

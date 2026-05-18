# UDM Pro SE QEMU VM Research

Status: real QEMU system profile prepared. The direct vendor-kernel path has
been tested, and the foreign-kernel UDM systemd path now reaches the firmware
userspace, applies userspace hardware mocks, starts UDAPI, and reaches the login
prompt.

Local firmware observations:

- Firmware image: `473c-UDMPROSE-5.0.16-511eddc1-cb19-476d-a02d-fcaf3dbddc29.bin`
- SHA-256: `7cf58f4563522220716f5025a7b2954b070df6be9364d7f60af0bc644512bce4`
- Image container: HIT archive.
- U-Boot payload starts at byte offset `352`.
- Kernel FIT starts at byte offset `1499296`.
- Rootfs SquashFS starts at byte offset `16141142`.
- FIT description: `AL324 UniFi Dream Machine Pro SE FIT image`.
- Kernel: Linux `4.19.152-ui-alpine`, ARM64, gzip-compressed in FIT.
- Vendor FDT compatible string: `annapurna-labs,alpine`.
- Vendor FDT model: `Annapurna Labs Alpine V2 UBNT`.
- Initramfs contains Ubiquiti boot scripts for UDM-style eMMC partition labels
  and `/dev/mtdblock5` configuration storage.

Initial boot result:

- `qemu-system-aarch64 -M virt` is available on the host.
- The extracted ARM64 kernel, initramfs, rootfs, and vendor U-Boot payload were
  tried under QEMU `virt`.
- The direct kernel attempts did not produce serial output before timeout.
- The vendor U-Boot payload also did not produce serial output on QEMU `virt`.
- A foreign Debian ARM64 installer kernel from `deb.debian.org` boots on QEMU
  `virt` and reaches `/init`, proving the VM and serial path are working.
- The foreign kernel with the UDM initramfs reaches the Ubiquiti initramfs
  scripts, then waits for `/dev/mtdblock5`.
- That mixed path also reports missing Debian kernel modules inside the UDM
  initramfs: `modprobe: can't change directory to '6.12.86+deb13-arm64'`.
- A lab initramfs path is prepared for systemd boot attempts. It keeps the UDM
  initramfs as the base, adds the foreign-kernel module tree, and redirects the
  hardware-only config device from `/dev/mtdblock5` to the QEMU GPT partition
  labeled `config`.
- The `udm-systemd` runner attaches the VM disk through QEMU USB storage,
  because the current foreign initrd carries USB storage modules but not the
  `virtio_blk` module needed for a direct VirtIO block disk.
- The fetch helper also downloads the matching Debian `linux-image` module
  package so the lab initramfs can load `ext4`, `loop`, `squashfs`, `overlay`,
  and storage drivers for the foreign kernel.
- The unmocked `udm-systemd` path reaches UDM firmware `systemd`:
  `systemd 247.3-7+deb11u7 running in system mode`, detects `qemu`, sets the
  hostname to `UDM-SE`, starts many UniFi OS services, and reaches a serial
  login prompt on `ttyAMA0`.
- The mocked `udm-systemd` path stages `/mock` and
  `/usr/local/lib/unifi-stubd-vm/libubnthal_redirect.so` into the UDM root from
  initramfs `init-bottom`, then applies targeted systemd drop-ins with
  `LD_PRELOAD` for UDM hardware-facing services.
- With the mocks applied, the boot log contains
  `unifi-stubd-vm-mock: installed userspace hardware mocks`; `uhwd.service`,
  `usd.service`, `rpsd.service`, nginx, SSH, UniFi OS Agent, and other UniFi OS
  services start, and the serial login prompt is reached.
- The foreign-kernel runner now uses explicit transparent QEMU networking for
  `udm-systemd` by default. The UTM profile uses two NICs: UTM `Shared` / NAT
  maps to guest `eth9` for the first SFP+ WAN role, and UTM `Host` maps to
  guest `eth8` for the 2.5G RJ45 LAN role attached to `br0`. The direct QEMU
  runner can still use a single bridged LAN NIC for comparison.
- The lab initramfs removes stale VM-only service entries from earlier test
  overlays, keeps the QEMU-backed LAN member attached to `br0`, and installs
  the vendor `site-setup.conf` nginx template when UniFi Core has not yet
  produced dynamic HTTP config. The intended transparent checks are
  `curl -k https://192.168.1.1/`, host-only `https://192.168.128.2/`, or UTM's
  Shared/NAT guest address. The UTM installer
  also writes a plist entry for `https://127.0.0.1:10443/`, but the latest UTM
  CLI run did not bind that host port natively; a separate local TCP helper was
  needed when localhost access worked.
- The previous board-identify failure such as `libubnt.get_board()` was not
  observed in the mocked boot log.
- The lab initramfs now loads the foreign-kernel `dummy`, `8021q`, and bridge
  modules and prepares `switch0`, `eth8`, `eth9`, and `eth10` before
  `network-init.service`. In the latest mocked boot, `network-init.service`
  succeeds on attempt 1 after creating `eth0` through `eth7` as 802.1ad
  interfaces on `switch0` and applying MAC and queue settings.
- The explicit `udapi-server.service` command from the userspace lab starts
  `ubios-udapi-server`; it listens on `/var/run/ubnt-udapi-server.sock`, and
  the foreign-kernel netfilter modules are sufficient for `/firewall/nat` to
  return `MASQUERADE` and `DNAT` rules. In the latest mocked boot,
  `firewall/filter` also completes and `firewall/mangle` is reached after adding
  the `xt_connmark` module and related conntrack/NAT helpers to the early module
  load set.
- In the latest UTM full test, the serial-only 4 GiB VM reached the UniFi OS
  setup surface directly through the UTM Shared/NAT guest address. `/api/system`
  reported `hasInternet=true`, the SFP+ WAN role was associated with `eth9`,
  `eth8` stayed attached to `br0`, and `br0` exposed `192.168.1.1/24` plus
  `192.168.128.2/24`. Browser login and refresh were stable when the same host
  name was used consistently. `/api/setup/support/generate` timed out in this
  VM path and remains unresolved.
- Expected hardware-dependent failures remain after network initialization:
  `svc-dpi-service` cannot load the vendor-specific `xt_dpi` module,
  WAN failover cannot load the vendor-specific `xt_dyn_random` module,
  `bluetooth-controller@hci0.service` fails, `ulogd2.service`,
  `unifi-directory.service`, and `mcagent.service` restart or fail, and dummy
  QEMU network devices still emit some `ethtool` warnings.

Working hypothesis:

The kernel and U-Boot are tightly coupled to Annapurna Labs Alpine V2/AL324
early hardware, interrupt, UART, and board-description assumptions. QEMU `virt`
can host an ARM64 VM, but it does not emulate that board. The closest practical
VM path is therefore: QEMU-virt-capable ARM64 kernel, UDM initramfs with a
minimal config-device lab patch, UDM rootfs, and then the firmware's own
`/sbin/init`/`systemd`.

The C hardware mock modules live under
`lab/gateway-profiles/udm-pro-se/mock/ldpreload/` and now run in the VM path as
an ARM64 Linux `LD_PRELOAD` shim for selected systemd services. They still
cannot affect early kernel boot; they only apply after userspace starts. The
next useful debug step is to inspect whether the vendor-specific `xt_dpi` and
`xt_dyn_random` behavior should stay documented as absent in QEMU or be replaced
by narrow lab-only no-op compatibility.

The native Network self-inform path is still a research target. A host-side
`unifi-stubd` inform was used only as a diagnostic adoption check and is not the
intended VM design.

Keep all firmware images, extracted artifacts, disks, logs, captures, keys,
tokens, certificates, and private lab data out of Git.

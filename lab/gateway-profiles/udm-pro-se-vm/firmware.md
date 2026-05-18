# UDM Pro SE QEMU VM Research

Status: real QEMU system profile prepared. The direct vendor-kernel path has
been tested, and the foreign-kernel UDM systemd path now reaches the firmware
userspace and login prompt.

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
- The `udm-systemd` path reaches UDM firmware `systemd`:
  `systemd 247.3-7+deb11u7 running in system mode`, detects `qemu`, sets the
  hostname to `UDM-SE`, starts many UniFi OS services, and reaches a serial
  login prompt on `ttyAMA0`.
- Expected hardware-dependent failures remain after systemd starts:
  `uhwd.service`, `usdbd.service`, `network-init.service`, and board-identify
  paths such as `libubnt.get_board()` fail because QEMU does not expose the UDM
  AL324 board devices or `/proc/ubnthal` data.

Working hypothesis:

The kernel and U-Boot are tightly coupled to Annapurna Labs Alpine V2/AL324
early hardware, interrupt, UART, and board-description assumptions. QEMU `virt`
can host an ARM64 VM, but it does not emulate that board. The closest practical
VM path is therefore: QEMU-virt-capable ARM64 kernel, UDM initramfs with a
minimal config-device lab patch, UDM rootfs, and then the firmware's own
`/sbin/init`/`systemd`.

The next useful debug step is to carry the existing UDM Pro SE C hardware mock
into this VM path as an ARM64 Linux `LD_PRELOAD` shim for selected systemd
services, starting with board identity and `/proc/ubnthal` consumers. That mock
still cannot affect early kernel boot; it only applies after userspace starts.

Keep all firmware images, extracted artifacts, disks, logs, captures, keys,
tokens, certificates, and private lab data out of Git.

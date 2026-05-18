# UDM Pro SE QEMU VM Profile

This profile is for real `qemu-system-aarch64` VM boot attempts using the UDM
Pro SE firmware image. It does not use Docker, Compose, chroot, or
`qemu-aarch64-static`.

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

Build the lab initramfs and try to reach the UDM firmware `systemd`:

```sh
lab/gateway-profiles/udm-pro-se-vm/scripts/fetch-foreign-kernel.sh
lab/gateway-profiles/udm-pro-se-vm/scripts/build-lab-initramfs.sh
UDM_PRO_SE_FOREIGN_MODE=udm-systemd \
  lab/gateway-profiles/udm-pro-se-vm/scripts/run-foreign-kernel.sh
```

Current limitation: QEMU does not provide an Annapurna Labs AL324 machine, so
the vendor UDM kernel has to be tried on QEMU `virt`. That is a real VM, but it
is not yet a faithful board model.

For the `udm-systemd` path, the kernel is foreign because it must support QEMU
`virt`. The initramfs, rootfs, overlay layout, and final `/sbin/init` are still
from the UDM firmware. The lab initramfs replaces the hardware-only
`/dev/mtdblock5` config store with a QEMU GPT partition labeled `config`. The
VM disk is attached through QEMU USB storage in this mode because the current
foreign initrd carries USB storage modules but not `virtio_blk`.

Current result: `udm-systemd` reaches UDM firmware `systemd`, starts many UniFi
OS services, and reaches the serial login prompt. Remaining failures are
hardware-dependent services such as `uhwd`, `usdbd`, `network-init`, and board
identity consumers; those are the places where the C hardware mock from the
userspace profile becomes relevant.

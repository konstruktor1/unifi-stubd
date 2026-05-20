# UDM Pro SE QEMU VM Source Inventory

This directory contains only project-owned VM wrapper scripts and research
notes.

| File | Purpose |
| --- | --- |
| `initramfs/` | Project-owned initramfs hooks, module lists, rootfs payload files, systemd units, and config templates staged by the VM builder. |
| `kernel/` | Project-owned notes for the ignored deployable kernel payload shared by UTM and Docker. |
| `utm/` | Project-owned UTM profile inputs: default VM/network values, kernel boot arguments, split installer modules, and notes for applying them to a cloned UTM VM. |
| `utm/defaults.env` | Default UTM VM shape, memory, serial, and network values. |
| `utm/bootargs.txt` | Kernel command line injected into the generated UTM DTB and QEMU append args. |
| `utm/install/common.sh` | Shared installer input/path/tool checks. |
| `utm/install/plist.sh` | PlistBuddy read/write helpers. |
| `utm/install/boot.sh` | Kernel, initramfs, bootargs, and generated DTB deployment. |
| `utm/install/drives.sh` | Writable disk conversion and UTM drive entries. |
| `utm/install/network.sh` | SFP-WAN Shared/NAT, LAN Host networking, port-forward intent, and QEMU args. |
| `utm/install/system.sh` | Serial-only display, memory, CPU, sharing/input cleanup, and summary output. |
| `scripts/extract-fit.py` | Extracts U-Boot, FIT, kernel, FDT, initramfs, and rootfs from the local firmware image. |
| `scripts/build-vm-disk.py` | Builds a GPT disk with UDM-style partition labels for VM boot attempts. |
| `scripts/build-lab-initramfs.sh` | Builds the UDM-based lab initramfs used to reach the firmware `systemd` path under QEMU `virt`. |
| `scripts/fetch-zig.sh` | Downloads a local Zig toolchain into ignored artifacts when no suitable local `zig` is available for ARM64 Linux shim builds. |
| `scripts/deploy-kernel-artifacts.sh` | Stages vendor, foreign, and lab kernel/initramfs artifacts into ignored `artifacts/deploy/kernel/` for UTM and Docker. |
| `scripts/prepare-mocks.sh` | Builds the ARM64 Linux userspace hardware shim from `../udm-pro-se/mock/ldpreload/` and stages deterministic `/mock` hardware files into ignored artifacts for VM boot attempts. |
| `scripts/prepare-vm.sh` | Copies the local firmware image into ignored artifacts and prepares VM images. |
| `scripts/install-utm-profile.sh` | Configures a cloned UTM VM with the foreign kernel, lab initramfs, VM disk, serial-only display, 4 GiB RAM, SFP-WAN Shared/NAT, and 2.5G LAN Host networking. |
| `scripts/run-direct-kernel.sh` | Starts a real `qemu-system-aarch64` VM with the extracted kernel. |
| `scripts/run-vendor-uboot.sh` | Starts a real `qemu-system-aarch64` VM with the vendor U-Boot payload as firmware. |
| `scripts/fetch-foreign-kernel.sh` | Downloads a QEMU-virt-capable ARM64 Debian installer kernel, initrd, and matching kernel module package into ignored artifacts for comparison and systemd boot attempts. |
| `scripts/run-foreign-kernel.sh` | Starts a real `qemu-system-aarch64` VM with the foreign kernel as a smoke test, with the UDM initramfs, or with the UDM-based lab initramfs for systemd boot attempts. |
| `README.md` | Usage and scope. |
| `firmware.md` | Safe firmware findings and current boot status. |

Ignored paths under `artifacts/` contain the copied vendor firmware and derived
VM artifacts. Do not commit those files.

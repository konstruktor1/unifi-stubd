# UDM Pro SE QEMU VM Source Inventory

This directory contains only project-owned VM wrapper scripts and research
notes.

| File | Purpose |
| --- | --- |
| `scripts/extract-fit.py` | Extracts U-Boot, FIT, kernel, FDT, initramfs, and rootfs from the local firmware image. |
| `scripts/build-vm-disk.py` | Builds a GPT disk with UDM-style partition labels for VM boot attempts. |
| `scripts/build-lab-initramfs.sh` | Builds the UDM-based lab initramfs used to reach the firmware `systemd` path under QEMU `virt`. |
| `scripts/prepare-vm.sh` | Copies the local firmware image into ignored artifacts and prepares VM images. |
| `scripts/run-direct-kernel.sh` | Starts a real `qemu-system-aarch64` VM with the extracted kernel. |
| `scripts/run-vendor-uboot.sh` | Starts a real `qemu-system-aarch64` VM with the vendor U-Boot payload as firmware. |
| `scripts/fetch-foreign-kernel.sh` | Downloads a QEMU-virt-capable ARM64 Debian installer kernel, initrd, and matching kernel module package into ignored artifacts for comparison and systemd boot attempts. |
| `scripts/run-foreign-kernel.sh` | Starts a real `qemu-system-aarch64` VM with the foreign kernel as a smoke test, with the UDM initramfs, or with the UDM-based lab initramfs for systemd boot attempts. |
| `README.md` | Usage and scope. |
| `firmware.md` | Safe firmware findings and current boot status. |

Ignored paths under `artifacts/` contain the copied vendor firmware and derived
VM artifacts. Do not commit those files.

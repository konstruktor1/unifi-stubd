#!/bin/sh
# Run QEMU-virt with a foreign ARM64 kernel for VM boundary checks.
set -eu

profile_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
artifacts="${UDM_PRO_SE_VM_ARTIFACTS:-$profile_dir/artifacts}"
foreign_dir="$artifacts/foreign-kernel"
mode="${UDM_PRO_SE_FOREIGN_MODE:-smoke}"
qemu="${UDM_PRO_SE_QEMU_SYSTEM:-qemu-system-aarch64}"

foreign_kernel="$foreign_dir/debian-arm64-linux"
foreign_initrd="$foreign_dir/debian-arm64-initrd.gz"
udm_initrd="$artifacts/initramfs.cpio.gz"
lab_initrd="$artifacts/lab-initramfs.cpio.gz"
disk="$artifacts/vm-disk.raw"

if [ ! -f "$foreign_kernel" ] || [ ! -f "$foreign_initrd" ]; then
    echo "missing foreign kernel artifacts; run $profile_dir/scripts/fetch-foreign-kernel.sh first" >&2
    exit 1
fi

common_args="
  -M virt,gic-version=3,highmem=off
  -cpu max
  -smp ${UDM_PRO_SE_VM_SMP:-4}
  -m ${UDM_PRO_SE_VM_MEMORY:-2048}
  -nographic
  -no-reboot
  -serial mon:stdio
"

case "$mode" in
smoke)
    append="${UDM_PRO_SE_VM_APPEND:-console=ttyAMA0,115200n8 earlycon=pl011,mmio32,0x09000000 DEBIAN_FRONTEND=text priority=low}"
    exec "$qemu" \
        $common_args \
        -kernel "$foreign_kernel" \
        -initrd "$foreign_initrd" \
        -append "$append"
    ;;
udm-initramfs)
    if [ ! -f "$udm_initrd" ] || [ ! -f "$disk" ]; then
        echo "missing UDM artifacts; run $profile_dir/scripts/prepare-vm.sh first" >&2
        exit 1
    fi
    append="${UDM_PRO_SE_VM_APPEND:-earlycon=pl011,mmio32,0x09000000 console=ttyAMA0,115200n8 loglevel=8 ignore_loglevel keep_bootcon boot=ubnt sysid=ea2c root=rootfs no_reboot panic=-1}"
    exec "$qemu" \
        $common_args \
        -kernel "$foreign_kernel" \
        -initrd "$udm_initrd" \
        -drive "file=$disk,format=raw,if=virtio" \
        -append "$append"
    ;;
udm-systemd)
    if [ ! -f "$lab_initrd" ] || [ ! -f "$disk" ]; then
        echo "missing lab UDM artifacts; run prepare-vm.sh, fetch-foreign-kernel.sh, then build-lab-initramfs.sh" >&2
        exit 1
    fi
    append="${UDM_PRO_SE_VM_APPEND:-earlycon=pl011,mmio32,0x09000000 console=ttyAMA0,115200n8 loglevel=8 ignore_loglevel keep_bootcon boot=ubnt sysid=ea2c root=rootfs rootdelay=2 no_reboot panic=-1 systemd.log_target=console systemd.show_status=1}"
    exec "$qemu" \
        $common_args \
        -kernel "$foreign_kernel" \
        -initrd "$lab_initrd" \
        -device qemu-xhci,id=udm_xhci \
        -drive "if=none,id=udm_disk,file=$disk,format=raw" \
        -device usb-storage,drive=udm_disk \
        -append "$append"
    ;;
*)
    echo "unknown UDM_PRO_SE_FOREIGN_MODE: $mode" >&2
    echo "supported modes: smoke, udm-initramfs, udm-systemd" >&2
    exit 1
    ;;
esac

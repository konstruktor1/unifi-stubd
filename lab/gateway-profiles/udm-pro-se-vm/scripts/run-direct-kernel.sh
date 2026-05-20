#!/bin/sh
# Run the extracted UDM Pro SE kernel as a qemu-system-aarch64 VM.
set -eu

profile_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
artifacts="${UDM_PRO_SE_VM_ARTIFACTS:-$profile_dir/artifacts}"
mode="${UDM_PRO_SE_VM_MODE:-initramfs-disk}"
qemu="${UDM_PRO_SE_QEMU_SYSTEM:-qemu-system-aarch64}"

kernel="$artifacts/kernel.Image"
initramfs="$artifacts/initramfs.cpio.gz"
disk="$artifacts/vm-disk.raw"
rootfs="$artifacts/rootfs.squashfs"

if [ ! -f "$kernel" ] || [ ! -f "$initramfs" ] || [ ! -f "$disk" ]; then
    echo "missing VM artifacts; run $profile_dir/scripts/prepare-vm.sh first" >&2
    exit 1
fi

common_args="
  -M virt,gic-version=3,highmem=off
  -cpu max
  -smp ${UDM_PRO_SE_VM_SMP:-4}
  -m ${UDM_PRO_SE_VM_MEMORY:-4096}
  -nographic
  -no-reboot
  -serial mon:stdio
"

case "$mode" in
initramfs-disk)
    append="${UDM_PRO_SE_VM_APPEND:-earlycon=pl011,mmio32,0x09000000 console=ttyAMA0,115200n8 loglevel=8 ignore_loglevel keep_bootcon boot=ubnt sysid=ea2c root=rootfs no_reboot panic=-1}"
    exec "$qemu" \
        $common_args \
        -kernel "$kernel" \
        -initrd "$initramfs" \
        -drive "file=$disk,format=raw,if=virtio" \
        -append "$append"
    ;;
rootfs-block)
    append="${UDM_PRO_SE_VM_APPEND:-earlycon=pl011,mmio32,0x09000000 console=ttyAMA0,115200n8 loglevel=8 ignore_loglevel keep_bootcon root=/dev/vda rootfstype=squashfs ro init=/sbin/init panic=-1}"
    exec "$qemu" \
        $common_args \
        -kernel "$kernel" \
        -drive "file=$rootfs,format=raw,if=virtio,readonly=on" \
        -append "$append"
    ;;
*)
    echo "unknown UDM_PRO_SE_VM_MODE: $mode" >&2
    echo "supported modes: initramfs-disk, rootfs-block" >&2
    exit 1
    ;;
esac

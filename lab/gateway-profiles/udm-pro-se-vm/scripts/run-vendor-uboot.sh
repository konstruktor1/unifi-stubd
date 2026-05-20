#!/bin/sh
# Try the vendor U-Boot payload as qemu-system-aarch64 firmware.
set -eu

profile_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
artifacts="${UDM_PRO_SE_VM_ARTIFACTS:-$profile_dir/artifacts}"
qemu="${UDM_PRO_SE_QEMU_SYSTEM:-qemu-system-aarch64}"
uboot="$artifacts/uboot.bin"

if [ ! -f "$uboot" ]; then
    echo "missing U-Boot artifact; run $profile_dir/scripts/prepare-vm.sh first" >&2
    exit 1
fi

exec "$qemu" \
    -M virt,gic-version=3,highmem=off \
    -cpu cortex-a57 \
    -smp 4 \
    -m 1024 \
    -nographic \
    -no-reboot \
    -serial mon:stdio \
    -bios "$uboot"

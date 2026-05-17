#!/bin/sh
# Start the extracted UGW3 MIPS firmware rootfs under qemu-mips-static.
set -eu

rootfs="${UGW3_ROOTFS:-/firmware-rootfs}"
instance="${UGW3_MCAD_INSTANCE:-0}"
debug="${UGW3_MCAD_DEBUG:-1}"
qemu_src="/usr/bin/qemu-mips-static"
qemu_dst="$rootfs/usr/bin/qemu-mips-static"

if [ ! -d "$rootfs" ]; then
    echo "missing UGW3 rootfs mount: $rootfs" >&2
    exit 1
fi

if [ ! -x "$rootfs/usr/bin/mcad" ]; then
    echo "missing UGW3 mcad binary in rootfs: $rootfs/usr/bin/mcad" >&2
    exit 1
fi

mkdir -p \
    "$rootfs/tmp" \
    "$rootfs/run" \
    "$rootfs/var/log" \
    "$rootfs/var/etc"

cp "$qemu_src" "$qemu_dst"

if [ "$debug" = "1" ]; then
    exec chroot "$rootfs" /usr/bin/qemu-mips-static /usr/bin/mcad -n "$instance" -d -v
fi

exec chroot "$rootfs" /usr/bin/qemu-mips-static /usr/bin/mcad -n "$instance" -v

#!/bin/sh
# Build a UDM initramfs variant for QEMU-virt systemd boot attempts.
set -eu

profile_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
artifacts="${UDM_PRO_SE_VM_ARTIFACTS:-$profile_dir/artifacts}"
udm_initramfs="$artifacts/initramfs.cpio.gz"
foreign_dir="$artifacts/foreign-kernel"
foreign_initrd="$artifacts/foreign-kernel/debian-arm64-initrd.gz"
tree="$artifacts/lab-initramfs-tree"
foreign_tree="$artifacts/lab-foreign-initrd-tree"
out="$artifacts/lab-initramfs.cpio.gz"

if [ ! -f "$udm_initramfs" ]; then
    echo "missing UDM initramfs; run $profile_dir/scripts/prepare-vm.sh first" >&2
    exit 1
fi

if [ ! -f "$foreign_initrd" ]; then
    echo "missing foreign initrd; run $profile_dir/scripts/fetch-foreign-kernel.sh first" >&2
    exit 1
fi

rm -rf "$tree" "$foreign_tree"
mkdir -p "$tree" "$foreign_tree"

gzip -dc "$udm_initramfs" | (cd "$tree" && cpio -idmu)
gzip -dc "$foreign_initrd" | (cd "$foreign_tree" && cpio -idmu)

if [ -d "$foreign_tree/usr/lib/modules" ]; then
    mkdir -p "$tree/usr/lib/modules"
    cp -R "$foreign_tree/usr/lib/modules/." "$tree/usr/lib/modules/"
fi

if [ -d "$foreign_dir/modules" ]; then
    mkdir -p "$tree/usr/lib/modules"
    cp -R "$foreign_dir/modules/." "$tree/usr/lib/modules/"
fi

foreign_modules=$(find "$tree/usr/lib/modules" -mindepth 1 -maxdepth 1 -type d | sed -n '/deb[0-9][0-9]-arm64$/p' | sed -n '1p')
if [ -n "$foreign_modules" ]; then
    kernel_version=${foreign_modules##*/}
    module_base="$tree/usr/lib/modules/$kernel_version"
    : > "$module_base/modules.dep"

    for module in \
        kernel/crypto/xxhash_generic.ko \
        kernel/crypto/zstd.ko \
        kernel/lib/crc16.ko \
        kernel/drivers/block/loop.ko \
        kernel/drivers/block/virtio_blk.ko \
        kernel/drivers/usb/common/usb-common.ko \
        kernel/drivers/usb/core/usbcore.ko \
        kernel/drivers/usb/host/xhci-hcd.ko \
        kernel/drivers/usb/host/xhci-pci.ko \
        kernel/drivers/scsi/scsi_common.ko \
        kernel/drivers/scsi/scsi_mod.ko \
        kernel/drivers/scsi/sd_mod.ko \
        kernel/drivers/usb/storage/usb-storage.ko \
        kernel/drivers/usb/storage/uas.ko \
        kernel/fs/jbd2/jbd2.ko \
        kernel/fs/mbcache.ko \
        kernel/fs/ext4/ext4.ko \
        kernel/fs/squashfs/squashfs.ko \
        kernel/fs/overlayfs/overlay.ko
    do
        if [ -f "$module_base/$module.xz" ] && [ ! -f "$module_base/$module" ]; then
            xz -dk "$module_base/$module.xz"
        fi
    done
fi

patch_configdev() {
    file=$1
    [ -f "$file" ] || return 0
    tmp="$file.tmp"
    sed 's|^CONFIGDEV=.*|CONFIGDEV="/dev/disk/by-partlabel/config"|' "$file" > "$tmp"
    mv "$tmp" "$file"
}

patch_configdev "$tree/scripts/product-override"
patch_configdev "$tree/board-define"

cat > "$tree/scripts/init-top/qemu-storage" <<'SCRIPT'
#!/bin/sh
set +e

kver=$(uname -r)
base="/usr/lib/modules/$kver/kernel"

load_module() {
    module=$1
    [ -f "$module" ] || return 0
    insmod "$module" >/dev/null 2>&1 || true
}

for module in \
    crypto/xxhash_generic.ko \
    crypto/zstd.ko \
    lib/crc16.ko \
    drivers/usb/common/usb-common.ko \
    drivers/usb/core/usbcore.ko \
    drivers/scsi/scsi_common.ko \
    drivers/scsi/scsi_mod.ko \
    drivers/scsi/sd_mod.ko \
    drivers/block/loop.ko \
    drivers/block/virtio_blk.ko \
    fs/jbd2/jbd2.ko \
    fs/mbcache.ko \
    fs/ext4/ext4.ko \
    fs/squashfs/squashfs.ko \
    fs/overlayfs/overlay.ko \
    drivers/usb/host/xhci-hcd.ko \
    drivers/usb/host/xhci-pci.ko \
    drivers/usb/storage/usb-storage.ko \
    drivers/usb/storage/uas.ko
do
    load_module "$base/$module"
done

udevadm trigger --type=devices --action=add >/dev/null 2>&1 || true
udevadm settle --timeout=10 >/dev/null 2>&1 || true
SCRIPT
chmod +x "$tree/scripts/init-top/qemu-storage"

if ! grep -q qemu-storage "$tree/scripts/init-top/ORDER"; then
    {
        echo '/scripts/init-top/qemu-storage "$@"'
        echo '[ -e /conf/param.conf ] && . /conf/param.conf'
    } >> "$tree/scripts/init-top/ORDER"
fi

mkdir -p "$tree/etc/unifi-stubd-vm"
cat > "$tree/etc/unifi-stubd-vm/lab-initramfs-note" <<'NOTE'
UDM Pro SE QEMU VM lab initramfs.

Changes from the vendor initramfs:
- CONFIGDEV points to /dev/disk/by-partlabel/config instead of /dev/mtdblock5.
- Foreign-kernel modules from the comparison initrd are available for QEMU virt.
- QEMU USB-storage modules are loaded early so the VM disk appears before the
  UDM mount scripts wait for GPT partition labels.

The root filesystem, overlay setup, and final /sbin/init path remain the UDM
firmware boot path.
NOTE

rm -f "$out"
(cd "$tree" && LC_ALL=C find . -print | cpio -o --format=newc --owner 0:0 | gzip -9 > "$out")
rm -rf "$foreign_tree"

echo "wrote $out"

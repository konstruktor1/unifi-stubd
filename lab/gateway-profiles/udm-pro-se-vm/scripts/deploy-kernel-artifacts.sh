#!/bin/sh
# Stage deployable UDM Pro SE kernel artifacts for UTM and Docker lab profiles.
set -eu

profile_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
artifacts="${UDM_PRO_SE_VM_ARTIFACTS:-$profile_dir/artifacts}"
deploy_dir="${UDM_PRO_SE_KERNEL_DEPLOY_DIR:-$artifacts/deploy/kernel}"
foreign_dir="$artifacts/foreign-kernel"

require_file() {
    if [ ! -f "$1" ]; then
        echo "missing kernel deployment input: $1" >&2
        exit 1
    fi
}

require_dir() {
    if [ ! -d "$1" ]; then
        echo "missing kernel deployment input directory: $1" >&2
        exit 1
    fi
}

copy_file() {
    src=$1
    dst=$2

    mkdir -p "$(dirname "$dst")"
    cp "$src" "$dst"
}

copy_optional_file() {
    src=$1
    dst=$2

    [ -f "$src" ] || return 0
    copy_file "$src" "$dst"
}

copy_tree() {
    src=$1
    dst=$2

    rm -rf "$dst"
    mkdir -p "$dst"
    cp -R "$src/." "$dst/"
}

hash_file() {
    shasum -a 256 "$1" | awk '{print $1}'
}

require_file "$artifacts/kernel.Image"
require_file "$artifacts/kernel.fit"
require_file "$artifacts/udm-pro-se.dtb"
require_file "$artifacts/initramfs.cpio.gz"
require_file "$foreign_dir/debian-arm64-linux"
require_file "$foreign_dir/debian-arm64-initrd.gz"
require_dir "$foreign_dir/modules"
require_file "$artifacts/lab-initramfs.cpio.gz"

# Rebuild the deploy tree rather than updating it in place. This prevents stale
# modules or DTBs from an older firmware/kernel run from being mounted by Docker
# or copied into UTM.
rm -rf "$deploy_dir"
mkdir -p "$deploy_dir/vendor" "$deploy_dir/foreign" "$deploy_dir/lab"

# Vendor artifacts are kept for comparison and forensic checks. They are not
# used to boot the current UTM systemd path because the vendor kernel is tied to
# Annapurna Labs AL324 hardware that QEMU virt does not emulate.
copy_file "$artifacts/kernel.Image" "$deploy_dir/vendor/kernel.Image"
copy_file "$artifacts/kernel.fit" "$deploy_dir/vendor/kernel.fit"
copy_file "$artifacts/udm-pro-se.dtb" "$deploy_dir/vendor/udm-pro-se.dtb"
copy_file "$artifacts/initramfs.cpio.gz" "$deploy_dir/vendor/initramfs.cpio.gz"
copy_optional_file "$artifacts/kernel.Image.gz" "$deploy_dir/vendor/kernel.Image.gz"
copy_optional_file "$artifacts/uboot.bin" "$deploy_dir/vendor/uboot.bin"
copy_optional_file "$artifacts/extract-summary.json" "$deploy_dir/vendor/extract-summary.json"
copy_optional_file "$artifacts/firmware.sha256" "$deploy_dir/vendor/firmware.sha256"

# The foreign kernel is the bootable QEMU virt boundary. Its module tree is
# staged beside it so the lab initramfs build and later Docker inspection point
# at the exact same kernel payload.
copy_file "$foreign_dir/debian-arm64-linux" "$deploy_dir/foreign/debian-arm64-linux"
copy_file "$foreign_dir/debian-arm64-initrd.gz" "$deploy_dir/foreign/debian-arm64-initrd.gz"
copy_optional_file "$foreign_dir/debian-arm64-linux.url" "$deploy_dir/foreign/debian-arm64-linux.url"
copy_optional_file "$foreign_dir/debian-arm64-initrd.gz.url" "$deploy_dir/foreign/debian-arm64-initrd.gz.url"
copy_optional_file "$foreign_dir/Packages.xz.url" "$deploy_dir/foreign/Packages.xz.url"
copy_tree "$foreign_dir/modules" "$deploy_dir/foreign/modules"

# Lab artifacts are generated from project-owned initramfs hooks plus local
# firmware inputs. The optional DTB appears after the UTM installer has run.
copy_file "$artifacts/lab-initramfs.cpio.gz" "$deploy_dir/lab/lab-initramfs.cpio.gz"
copy_optional_file "$artifacts/utm/virt-udm-bootargs.dtb" "$deploy_dir/lab/virt-udm-bootargs.dtb"
copy_optional_file "$artifacts/utm/virt-udm-bootargs.dts" "$deploy_dir/lab/virt-udm-bootargs.dts"

manifest="$deploy_dir/MANIFEST.txt"
{
    printf 'UDM Pro SE kernel deployment artifacts\n'
    printf 'source_artifacts=%s\n' "$artifacts"
    printf 'deploy_dir=%s\n' "$deploy_dir"
    printf '\n'
    # Hash every staged file so Docker logs, UTM bundle contents, and local test
    # results can be compared without committing the binary artifacts.
    find "$deploy_dir" -type f ! -name MANIFEST.txt | sort | while IFS= read -r file; do
        rel=${file#"$deploy_dir"/}
        size=$(wc -c < "$file" | tr -d ' ')
        hash=$(hash_file "$file")
        printf '%s  %s  %s\n' "$hash" "$size" "$rel"
    done
} > "$manifest"

echo "deployed kernel artifacts: $deploy_dir"
echo "manifest: $manifest"

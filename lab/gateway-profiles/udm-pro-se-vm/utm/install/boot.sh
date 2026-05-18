# shellcheck shell=sh
# Kernel, initramfs, and generated device-tree deployment for UTM.

utm_deploy_boot_inputs() {
    mkdir -p "$utm_artifacts"

    # UTM's Linux boot drive entries do not expose an obvious place to edit
    # /chosen/bootargs directly, so generate a QEMU virt DTB and inject the
    # same command line that the direct QEMU runner uses.
    "${UDM_PRO_SE_QEMU_SYSTEM:-qemu-system-aarch64}" \
        -M virt,gic-version=3,highmem=off,dumpdtb="$dtb.base" \
        -cpu max \
        -display none \
        -nodefaults >/dev/null 2>&1 || true

    dtc -I dtb -O dts -o "$dtb.dts.tmp" "$dtb.base" 2>/dev/null
    # Some QEMU builds already emit an empty bootargs property; others only
    # emit /chosen. Handle both forms so the generated DTB is reproducible.
    if grep -q 'bootargs = ' "$dtb.dts.tmp"; then
        perl -0pi -e 's/bootargs = "[^"]*";/bootargs = "'"$cmdline"'";/' "$dtb.dts.tmp"
    else
        perl -0pi -e 's/(chosen \{\n)/$1\t\tbootargs = "'"$cmdline"'";\n/' "$dtb.dts.tmp"
    fi
    mv "$dtb.dts.tmp" "$utm_artifacts/virt-udm-bootargs.dts"
    dtc -I dts -O dtb -o "$dtb" "$utm_artifacts/virt-udm-bootargs.dts" 2>/dev/null

    # The paths stored in config.plist are bundle-relative ImageName values.
    # Copy the actual payloads into Data/ before the Drive entries are rebuilt.
    copy_into_bundle "$kernel" "$utm_kernel"
    copy_into_bundle "$initrd" "$utm_initrd"
    copy_into_bundle "$dtb" "$utm_dtb"
}

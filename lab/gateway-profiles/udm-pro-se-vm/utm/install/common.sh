# shellcheck shell=sh
# Shared input and filesystem helpers for the UTM profile installer.

require_file() {
    if [ ! -f "$1" ]; then
        echo "missing UTM input: $1" >&2
        exit 1
    fi
}

choose_file() {
    # Prefer the normalized deployment tree when it exists, but keep the raw
    # artifact paths as a fallback so older local checkouts still run.
    for candidate do
        if [ -f "$candidate" ]; then
            printf '%s\n' "$candidate"
            return 0
        fi
    done
    printf '%s\n' "$1"
}

copy_into_bundle() {
    src=$1
    dst=$2
    # APFS clone copies are fast for large kernel/initramfs images. Fall back
    # to a normal copy on filesystems that do not support clonefile(2).
    if ! cp -c "$src" "$dst" 2>/dev/null; then
        cp "$src" "$dst"
    fi
}

utm_resolve_artifact_paths() {
    # UTM consumes files from the bundle's Data/ directory, while QEMU scripts
    # consume files from artifacts/. Resolve both sides here so later modules
    # only talk about their local responsibility.
    kernel=$(choose_file \
        "$kernel_deploy_dir/foreign/debian-arm64-linux" \
        "$artifacts/foreign-kernel/debian-arm64-linux")
    initrd=$(choose_file \
        "$kernel_deploy_dir/lab/lab-initramfs.cpio.gz" \
        "$artifacts/lab-initramfs.cpio.gz")
    disk="$artifacts/vm-disk.raw"
    utm_artifacts="$artifacts/utm"
    dtb="$utm_artifacts/virt-udm-bootargs.dtb"
    utm_kernel="$utm_bundle/Data/udm-foreign-kernel"
    utm_initrd="$utm_bundle/Data/udm-lab-initramfs.cpio.gz"
    utm_dtb="$utm_bundle/Data/virt-udm-bootargs.dtb"
}

utm_resolve_cmdline() {
    # Keep bootargs in a text file because both the generated DTB and QEMU
    # -append need the same command line. The fallback matches the committed
    # bootargs file for callers that source only the shell modules.
    if [ -f "$utm_bootargs_file" ]; then
        cmdline_default=$(tr '\n' ' ' < "$utm_bootargs_file" | sed 's/[[:space:]]*$//')
    else
        cmdline_default="earlycon=pl011,mmio32,0x09000000 console=ttyAMA0,115200n8 loglevel=8 ignore_loglevel keep_bootcon boot=ubnt sysid=ea2c root=rootfs rootdelay=2 no_reboot panic=-1 systemd.log_target=console systemd.show_status=1"
    fi
    cmdline="${UDM_PRO_SE_UTM_BOOTARGS:-$cmdline_default}"
}

utm_require_inputs() {
    # Do the file checks before touching config.plist so a missing kernel or
    # disk cannot leave a cloned UTM bundle half-mutated.
    for required in "$kernel" "$initrd" "$disk"; do
        require_file "$required"
    done

    if [ ! -f "$config" ]; then
        echo "missing registered UTM clone: $config" >&2
        echo "create it first with:" >&2
        echo "  utmctl clone unifi-stubd-lab --name '$utm_name'" >&2
        exit 1
    fi
}

utm_require_tools() {
    if ! command -v dtc >/dev/null 2>&1; then
        echo "missing dtc; install the device tree compiler before configuring UTM" >&2
        exit 1
    fi
    if ! command -v qemu-img >/dev/null 2>&1; then
        echo "missing qemu-img; install QEMU before configuring UTM" >&2
        exit 1
    fi
}

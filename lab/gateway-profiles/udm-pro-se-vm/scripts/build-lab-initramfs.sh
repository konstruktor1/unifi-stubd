#!/bin/sh
# Build a UDM initramfs variant for QEMU-virt systemd boot attempts.
set -eu

profile_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
artifacts="${UDM_PRO_SE_VM_ARTIFACTS:-$profile_dir/artifacts}"
initramfs_src="$profile_dir/initramfs"
module_list="$initramfs_src/module-lists/qemu-storage-modules.txt"
udm_initramfs="$artifacts/initramfs.cpio.gz"
foreign_dir="$artifacts/foreign-kernel"
foreign_initrd="$artifacts/foreign-kernel/debian-arm64-initrd.gz"
tree="$artifacts/lab-initramfs-tree"
foreign_tree="$artifacts/lab-foreign-initrd-tree"
out="$artifacts/lab-initramfs.cpio.gz"
sfp_wan_iface="${UDM_PRO_SE_VM_SFP_WAN_IFACE:-eth9}"
lan_iface="${UDM_PRO_SE_VM_LAN_IFACE:-eth8}"
lan_cidr="${UDM_PRO_SE_VM_LAN_CIDR:-192.168.1.1/24}"
lan_cidr_addr=${lan_cidr%/*}
lan_host_cidr="${UDM_PRO_SE_VM_LAN_HOST_CIDR:-192.168.128.2/24}"
lan_host_cidr_addr=${lan_host_cidr%/*}
web_ingress_ifaces="${UDM_PRO_SE_VM_WEB_INGRESS_IFACES:-${UDM_PRO_SE_VM_WEB_INGRESS_IFACE:-$lan_iface $sfp_wan_iface}}"
debug_ssh_pubkey="${UDM_PRO_SE_VM_DEBUG_SSH_PUBKEY:-}"
debug_ssh_pubkey_b64=$(printf '%s' "$debug_ssh_pubkey" | base64 | tr -d '\n')

if [ ! -f "$udm_initramfs" ]; then
    echo "missing UDM initramfs; run $profile_dir/scripts/prepare-vm.sh first" >&2
    exit 1
fi

if [ ! -f "$foreign_initrd" ]; then
    echo "missing foreign initrd; run $profile_dir/scripts/fetch-foreign-kernel.sh first" >&2
    exit 1
fi

if [ ! -f "$module_list" ]; then
    echo "missing QEMU module list: $module_list" >&2
    exit 1
fi

copy_file() {
    src=$1
    dst=$2
    mode=${3:-0644}

    mkdir -p "$(dirname "$dst")"
    cp "$src" "$dst"
    chmod "$mode" "$dst"
}

copy_tree() {
    src=$1
    dst=$2

    rm -rf "$dst"
    mkdir -p "$dst"
    cp -R "$src/." "$dst/"
}

render_template() {
    src=$1
    dst=$2
    mode=${3:-0644}

    mkdir -p "$(dirname "$dst")"
    LAN_IFACE="$lan_iface" \
    LAN_CIDR="$lan_cidr" \
    LAN_CIDR_ADDR="$lan_cidr_addr" \
    LAN_HOST_CIDR="$lan_host_cidr" \
    LAN_HOST_CIDR_ADDR="$lan_host_cidr_addr" \
    SFP_WAN_IFACE="$sfp_wan_iface" \
    WEB_INGRESS_IFACES="$web_ingress_ifaces" \
    DEBUG_SSH_PUBKEY_B64="$debug_ssh_pubkey_b64" \
        perl -pe '
            s/\@UNIFI_STUBD_VM_LAN_IFACE\@/$ENV{LAN_IFACE}/g;
            s/\@UNIFI_STUBD_VM_LAN_CIDR\@/$ENV{LAN_CIDR}/g;
            s/\@UNIFI_STUBD_VM_LAN_CIDR_ADDR\@/$ENV{LAN_CIDR_ADDR}/g;
            s/\@UNIFI_STUBD_VM_LAN_HOST_CIDR\@/$ENV{LAN_HOST_CIDR}/g;
            s/\@UNIFI_STUBD_VM_LAN_HOST_CIDR_ADDR\@/$ENV{LAN_HOST_CIDR_ADDR}/g;
            s/\@UNIFI_STUBD_VM_SFP_WAN_IFACE\@/$ENV{SFP_WAN_IFACE}/g;
            s/\@UNIFI_STUBD_VM_WEB_INGRESS_IFACES\@/$ENV{WEB_INGRESS_IFACES}/g;
            s/\@UNIFI_STUBD_VM_DEBUG_SSH_PUBKEY_B64\@/$ENV{DEBUG_SSH_PUBKEY_B64}/g;
        ' "$src" > "$dst"
    chmod "$mode" "$dst"
}

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

    # Keep the module set in one file. The builder decompresses the modules
    # here, and the init-top hook uses the same list at boot time.
    while IFS= read -r module; do
        case "$module" in
            ""|\#*) continue ;;
        esac
        if [ -f "$module_base/kernel/$module.xz" ] && [ ! -f "$module_base/kernel/$module" ]; then
            xz -dk "$module_base/kernel/$module.xz"
        fi
    done < "$module_list"
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

for net_rules in \
    "$tree/usr/lib/udev/rules.d/70-ui-persistent-net.rules" \
    "$tree/lib/udev/rules.d/70-ui-persistent-net.rules"
do
    [ -f "$net_rules" ] || continue
    tmp="$net_rules.tmp"
    sed \
        -e 's/KERNELS=="0000:00:01.0", NAME="eth8"/KERNELS=="0000:00:01.0", NAME="eth9"/' \
        -e 's/KERNELS=="0000:00:02.0", NAME="eth10"/KERNELS=="0000:00:02.0", NAME="eth8"/' \
        "$net_rules" > "$tmp"
    mv "$tmp" "$net_rules"
done

# The remaining steps stage project-owned initramfs hooks and rootfs payload
# files from initramfs/. The shell fragments are kept as separate files so the
# builder stays readable and the guest-side behavior can be reviewed directly.
copy_file "$initramfs_src/init-top/qemu-storage" "$tree/scripts/init-top/qemu-storage" 0755

if ! grep -q qemu-storage "$tree/scripts/init-top/ORDER"; then
    {
        echo '/scripts/init-top/qemu-storage "$@"'
        echo '[ -e /conf/param.conf ] && . /conf/param.conf'
    } >> "$tree/scripts/init-top/ORDER"
fi

mkdir -p "$tree/etc/unifi-stubd-vm"
copy_file "$module_list" "$tree/etc/unifi-stubd-vm/qemu-storage-modules.txt" 0644
copy_file \
    "$initramfs_src/etc/unifi-stubd-vm/lab-initramfs-note" \
    "$tree/etc/unifi-stubd-vm/lab-initramfs-note" \
    0644

mock_root="$artifacts/mock-root"
if [ -d "$mock_root" ]; then
    mkdir -p "$tree/etc/unifi-stubd-vm/mock-root"
    cp -R "$mock_root/." "$tree/etc/unifi-stubd-vm/mock-root/"
fi

payload_dir="$tree/etc/unifi-stubd-vm/rootfs-payload"
copy_tree "$initramfs_src/rootfs-payload" "$payload_dir"
# README files document the source tree only. They must not become guest files.
find "$payload_dir" -name README.md -type f -delete
render_template \
    "$payload_dir/usr/local/lib/unifi-stubd-vm/keep-lan-forwarding-link.sh.in" \
    "$payload_dir/usr/local/lib/unifi-stubd-vm/keep-lan-forwarding-link.sh" \
    0755
render_template \
    "$payload_dir/usr/local/lib/unifi-stubd-vm/keep-sfp-wan-link.sh.in" \
    "$payload_dir/usr/local/lib/unifi-stubd-vm/keep-sfp-wan-link.sh" \
    0755
render_template \
    "$payload_dir/usr/local/lib/unifi-stubd-vm/open-web-ingress.sh.in" \
    "$payload_dir/usr/local/lib/unifi-stubd-vm/open-web-ingress.sh" \
    0755
render_template \
    "$payload_dir/usr/local/lib/unifi-stubd-vm/install-debug-ssh.sh.in" \
    "$payload_dir/usr/local/lib/unifi-stubd-vm/install-debug-ssh.sh" \
    0755
rm -f \
    "$payload_dir/usr/local/lib/unifi-stubd-vm/keep-lan-forwarding-link.sh.in" \
    "$payload_dir/usr/local/lib/unifi-stubd-vm/keep-sfp-wan-link.sh.in" \
    "$payload_dir/usr/local/lib/unifi-stubd-vm/open-web-ingress.sh.in" \
    "$payload_dir/usr/local/lib/unifi-stubd-vm/install-debug-ssh.sh.in"

for guest_script in \
    sbin/ubnt-tools \
    sbin/ubnt-systool \
    usr/local/lib/unifi-stubd-vm/prepare-netdevs.sh \
    usr/local/lib/unifi-stubd-vm/prepare-web-config.sh \
    usr/local/lib/unifi-stubd-vm/dump-web-state.sh
do
    chmod 0755 "$payload_dir/$guest_script"
done

copy_file "$initramfs_src/init-bottom/unifi-stubd-vm-mock" "$tree/scripts/init-bottom/unifi-stubd-vm-mock" 0755

if ! grep -q unifi-stubd-vm-mock "$tree/scripts/init-bottom/ORDER"; then
    echo '/scripts/init-bottom/unifi-stubd-vm-mock "$@"' >> "$tree/scripts/init-bottom/ORDER"
fi

rm -f "$out"
(cd "$tree" && LC_ALL=C find . -print | cpio -o --format=newc --owner 0:0 | gzip -9 > "$out")
rm -rf "$foreign_tree"

echo "wrote $out"

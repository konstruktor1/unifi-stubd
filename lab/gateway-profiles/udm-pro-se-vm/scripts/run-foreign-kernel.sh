#!/bin/sh
# Run QEMU-virt with a foreign ARM64 kernel for VM boundary checks.
set -eu

profile_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
artifacts="${UDM_PRO_SE_VM_ARTIFACTS:-$profile_dir/artifacts}"
foreign_dir="$artifacts/foreign-kernel"
mode="${UDM_PRO_SE_FOREIGN_MODE:-smoke}"
qemu="${UDM_PRO_SE_QEMU_SYSTEM:-qemu-system-aarch64}"
network_mode="${UDM_PRO_SE_VM_NET:-}"
hostfwd_addr="${UDM_PRO_SE_VM_HOSTFWD_ADDR:-127.0.0.1}"
http_port="${UDM_PRO_SE_VM_HTTP_PORT:-10080}"
https_port="${UDM_PRO_SE_VM_HTTPS_PORT:-10443}"
vmnet_ifname="${UDM_PRO_SE_VM_VMNET_IFNAME:-en0}"
wan_mac="${UDM_PRO_SE_VM_WAN_MAC:-02:15:6d:00:ea:35}"
lan_mac="${UDM_PRO_SE_VM_LAN_MAC:-02:15:6d:00:ea:34}"
lan_guest="${UDM_PRO_SE_VM_LAN_GUEST:-192.168.1.1}"
lan_net="${UDM_PRO_SE_VM_LAN_NET:-192.168.1.0/24}"
lan_host="${UDM_PRO_SE_VM_LAN_HOST:-192.168.1.254}"
lan_dhcpstart="${UDM_PRO_SE_VM_LAN_DHCPSTART:-192.168.1.100}"

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
  -m ${UDM_PRO_SE_VM_MEMORY:-4096}
  -nographic
  -no-reboot
  -serial mon:stdio
"

if [ -z "$network_mode" ]; then
    case "$mode" in
    udm-systemd)
        network_mode="vmnet-bridged"
        ;;
    *)
        network_mode="user"
        ;;
    esac
fi

case "$network_mode" in
vmnet-bridged)
    network_args="
      -netdev vmnet-bridged,id=udm_lan,ifname=$vmnet_ifname
      -device virtio-net-pci,netdev=udm_lan,mac=$lan_mac,addr=0x2
    "
    echo "QEMU vmnet bridged LAN networking: host interface $vmnet_ifname -> guest eth8/br0 ($lan_guest); no hostfwd or NAT" >&2
    ;;
user-lan)
    network_args="
      -netdev user,id=udm_lan,net=$lan_net,host=$lan_host,dhcpstart=$lan_dhcpstart,hostfwd=tcp:$hostfwd_addr:$http_port-$lan_guest:80,hostfwd=tcp:$hostfwd_addr:$https_port-$lan_guest:443
      -device virtio-net-pci,netdev=udm_lan,mac=$lan_mac,addr=0x2
    "
    echo "QEMU user LAN networking: host $hostfwd_addr:$http_port -> guest $lan_guest:80, host $hostfwd_addr:$https_port -> guest $lan_guest:443" >&2
    ;;
user)
    network_args="
      -netdev user,id=udm_wan,hostfwd=tcp:$hostfwd_addr:$http_port-:80,hostfwd=tcp:$hostfwd_addr:$https_port-:443
      -device virtio-net-pci,netdev=udm_wan,mac=$wan_mac
    "
    echo "QEMU generic user networking: host $hostfwd_addr:$http_port -> guest :80, host $hostfwd_addr:$https_port -> guest :443" >&2
    ;;
user-wan)
    network_args="
      -netdev user,id=udm_wan,hostfwd=tcp:$hostfwd_addr:$http_port-:80,hostfwd=tcp:$hostfwd_addr:$https_port-:443
      -device virtio-net-pci,netdev=udm_wan,mac=$wan_mac,addr=0x1
    "
    echo "QEMU user WAN networking: host $hostfwd_addr:$http_port -> guest :80, host $hostfwd_addr:$https_port -> guest :443" >&2
    ;;
none)
    network_args="-nic none"
    echo "QEMU networking disabled by UDM_PRO_SE_VM_NET=none" >&2
    ;;
default)
    network_args=""
    echo "QEMU networking left to emulator defaults by UDM_PRO_SE_VM_NET=default" >&2
    ;;
*)
    echo "unknown UDM_PRO_SE_VM_NET: $network_mode" >&2
    echo "supported values: vmnet-bridged, user-lan, user, user-wan, none, default" >&2
    exit 1
    ;;
esac

case "$mode" in
smoke)
    append="${UDM_PRO_SE_VM_APPEND:-console=ttyAMA0,115200n8 earlycon=pl011,mmio32,0x09000000 DEBIAN_FRONTEND=text priority=low}"
    exec "$qemu" \
        $common_args \
        $network_args \
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
        $network_args \
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
        $network_args \
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

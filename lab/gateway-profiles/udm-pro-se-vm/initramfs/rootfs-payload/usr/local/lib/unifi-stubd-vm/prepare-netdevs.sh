#!/bin/sh
# Create the lab-only interfaces that QEMU does not provide as real UDM ports.
set +e

PATH=/usr/sbin:/usr/bin:/sbin:/bin

# The foreign kernel does not provide the UDM SE switch ASIC or all front-panel
# ports. Load generic link drivers first; failure is non-fatal because some
# modules may already be built in.
modprobe dummy >/dev/null 2>&1 || true
modprobe 8021q >/dev/null 2>&1 || true

create_lab_link() {
    dev=$1
    [ -d "/sys/class/net/$dev" ] && return 0

    # Prefer dummy devices because Network only needs stable carrier-like
    # interfaces here. A bridge fallback keeps the boot moving on kernels
    # without dummy support.
    if ip link add "$dev" type dummy >/dev/null 2>&1; then
        echo "unifi-stubd-vm-netdevs: created dummy $dev"
    elif ip link add "$dev" type bridge >/dev/null 2>&1; then
        echo "unifi-stubd-vm-netdevs: created bridge fallback $dev"
    else
        echo "unifi-stubd-vm-netdevs: failed to create $dev" >&2
        return 0
    fi

    ip link set "$dev" up >/dev/null 2>&1 || true
}

# switch0 is the CPU-facing switch device expected by network-init. eth10 is
# kept as an internal role so persistent UDM rules that mention it do not fail.
create_lab_link switch0
create_lab_link eth10

# eth9 and eth8 are created only when QEMU/UTM did not provide real virtio NICs.
# In the UTM profile those names map to SFP+ WAN and 2.5G LAN respectively.
if [ ! -d /sys/class/net/eth9 ]; then
    create_lab_link eth9
fi

if [ ! -d /sys/class/net/eth8 ]; then
    create_lab_link eth8
fi

exit 0

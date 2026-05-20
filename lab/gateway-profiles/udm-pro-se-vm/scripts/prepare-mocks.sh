#!/bin/sh
# Build and stage userspace hardware mocks for the QEMU VM boot path.
set -eu

profile_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
artifacts="${UDM_PRO_SE_VM_ARTIFACTS:-$profile_dir/artifacts}"
source_profile="$profile_dir/../udm-pro-se"
shim_source="$source_profile/mock/ldpreload"
mock_files="$source_profile/mock/files"
mock_src="$artifacts/mock-src"
mock_root="$artifacts/mock-root"
shim_dir="$mock_root/usr/local/lib/unifi-stubd-vm"
shim_out="$shim_dir/libubnthal_redirect.so"
target="${UDM_PRO_SE_SHIM_TARGET:-aarch64-linux-gnu.2.31}"

if [ ! -d "$shim_source" ]; then
    echo "missing userspace shim source: $shim_source" >&2
    exit 1
fi

if [ ! -d "$mock_files" ]; then
    echo "missing userspace mock files: $mock_files" >&2
    exit 1
fi

rm -rf "$mock_src" "$mock_root"

# Static identity files are project-owned fixtures under the UDM Pro SE profile.
# Runtime-only files are generated below because they are easier to audit as
# small scalar values than as a large copied tree.
mkdir -p \
    "$mock_src" \
    "$shim_dir" \
    "$mock_root/mock/mtd" \
    "$mock_root/mock/persistent" \
    "$mock_root/mock/sys/class/hwmon/hwmon0/device" \
    "$mock_root/mock/sys/class/mtd/mtd5" \
    "$mock_root/mock/sys/class/thermal/thermal_zone0" \
    "$mock_root/mock/ubnthal/poe" \
    "$mock_root/mock/ubnthal/status" \
    "$mock_root/mock/proc/sys/crypto" \
    "$mock_root/mock/proc/sys/kernel" \
    "$mock_root/mock/proc/sys/net/core" \
    "$mock_root/mock/proc/sys/net/ipv4" \
    "$mock_root/mock/proc/sys/net/ipv4/neigh/default" \
    "$mock_root/mock/proc/sys/net/ipv6/conf/all" \
    "$mock_root/mock/proc/sys/net/ipv6/neigh/default" \
    "$mock_root/mock/proc/sys/net/netfilter"

mkdir -p "$mock_src/ldpreload"
cp -R "$shim_source/." "$mock_src/ldpreload/"
cp -R "$mock_files/." "$mock_root/mock/"

# UBNTHAL status and proc/sys values that firmware services probe during early
# userspace startup. They are deliberately deterministic and lab-only.
printf 'false\n' > "$mock_root/mock/ubnthal/status/IsLocated"
printf '0\n' > "$mock_root/mock/proc/sys/crypto/fips_enabled"
printf 'UDM-Pro-SE\n' > "$mock_root/mock/proc/sys/kernel/hostname"
printf '(none)\n' > "$mock_root/mock/proc/sys/kernel/domainname"
printf '212992\n' > "$mock_root/mock/proc/sys/net/core/rmem_max"
printf '212992\n' > "$mock_root/mock/proc/sys/net/core/wmem_max"
printf '4096\n' > "$mock_root/mock/proc/sys/net/core/somaxconn"
printf '0\n' > "$mock_root/mock/proc/sys/net/ipv4/ip_forward"
printf '0\n' > "$mock_root/mock/proc/sys/net/ipv6/conf/all/forwarding"
printf '0\n' > "$mock_root/mock/proc/sys/net/netfilter/nf_conntrack_helper"

for family in ipv4 ipv6; do
    neigh_dir="$mock_root/mock/proc/sys/net/$family/neigh/default"
    printf '30000\n' > "$neigh_dir/base_reachable_time_ms"
    printf '60\n' > "$neigh_dir/gc_stale_time"
    printf '128\n' > "$neigh_dir/gc_thresh1"
    printf '512\n' > "$neigh_dir/gc_thresh2"
    printf '1024\n' > "$neigh_dir/gc_thresh3"
    printf '1000\n' > "$neigh_dir/retrans_time_ms"
    printf '5\n' > "$neigh_dir/delay_first_probe_time"
    printf '100\n' > "$neigh_dir/anycast_delay"
    printf '0\n' > "$neigh_dir/app_solicit"
    printf '100\n' > "$neigh_dir/locktime"
    printf '3\n' > "$neigh_dir/mcast_solicit"
    printf '80\n' > "$neigh_dir/proxy_delay"
    printf '64\n' > "$neigh_dir/proxy_qlen"
    printf '3\n' > "$neigh_dir/ucast_solicit"
    printf '101\n' > "$neigh_dir/unres_qlen"
    printf '212992\n' > "$neigh_dir/unres_qlen_bytes"
done

printf 'dev:    size   erasesize  name\nmtd5: 00010000 00010000 "eeprom"\n' > "$mock_root/mock/mtd/proc_mtd"
dd if=/dev/zero of="$mock_root/mock/mtd/mtd5" bs=65536 count=1 >/dev/null 2>&1
cp "$mock_root/mock/mtd/mtd5" "$mock_root/mock/mtd/mtdblock5"
printf 'c2 20 18\n' > "$mock_root/mock/sys/class/mtd/mtd5/jedec_id"
printf '50000\n' > "$mock_root/mock/sys/class/hwmon/hwmon0/device/temp1_input"
printf '42000\n' > "$mock_root/mock/sys/class/hwmon/hwmon0/device/temp2_input"
printf '43000\n' > "$mock_root/mock/sys/class/hwmon/hwmon0/device/temp3_input"
printf '1800\n' > "$mock_root/mock/sys/class/hwmon/hwmon0/device/fan1_input"
printf '1600\n' > "$mock_root/mock/sys/class/hwmon/hwmon0/device/fan2_input"
printf '50000\n' > "$mock_root/mock/sys/class/thermal/thermal_zone0/temp"

for port in 0 1 2 3 4 5 6 7; do
    mkdir -p "$mock_root/mock/ubnthal/poe/port-$port"
    printf 'off\n' > "$mock_root/mock/ubnthal/poe/port-$port/mode"
    printf 'disabled\n' > "$mock_root/mock/ubnthal/poe/port-$port/status"
    printf '0\n' > "$mock_root/mock/ubnthal/poe/port-$port/power"
    printf '0\n' > "$mock_root/mock/ubnthal/poe/port-$port/power_mw"
    printf '0\n' > "$mock_root/mock/ubnthal/poe/port-$port/current_ma"
    printf '0\n' > "$mock_root/mock/ubnthal/poe/port-$port/voltage_mv"
    printf '0\n' > "$mock_root/mock/ubnthal/poe/port-$port/pd_class"
done

zig_bin="${UDM_PRO_SE_ZIG:-}"
if [ -z "$zig_bin" ]; then
    if command -v zig >/dev/null 2>&1; then
        zig_bin=$(command -v zig)
    else
        zig_bin=$("$profile_dir/scripts/fetch-zig.sh")
    fi
fi

"$zig_bin" cc \
    -target "$target" \
    -shared \
    -fPIC \
    -O2 \
    -g0 \
    -Wall \
    -Wextra \
    -Wl,-soname,libubnthal_redirect.so \
    -I "$mock_src/ldpreload" \
    -o "$shim_out" \
    "$mock_src/ldpreload/"*.c \
    -ldl

{
    printf 'source=%s\n' "$shim_source"
    printf 'target=%s\n' "$target"
    printf 'zig=%s\n' "$zig_bin"
    printf 'output=%s\n' "$shim_out"
} > "$mock_src/build-info.txt"

file "$shim_out"
printf 'wrote %s\n' "$mock_root"

#!/bin/sh
# Fetch a QEMU-virt-capable ARM64 kernel for comparison boot attempts.
set -eu

profile_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
artifacts="${UDM_PRO_SE_VM_ARTIFACTS:-$profile_dir/artifacts}"
foreign_dir="$artifacts/foreign-kernel"

kernel_url="${UDM_PRO_SE_FOREIGN_KERNEL_URL:-https://deb.debian.org/debian/dists/stable/main/installer-arm64/current/images/netboot/debian-installer/arm64/linux}"
initrd_url="${UDM_PRO_SE_FOREIGN_INITRD_URL:-https://deb.debian.org/debian/dists/stable/main/installer-arm64/current/images/netboot/debian-installer/arm64/initrd.gz}"
packages_url="${UDM_PRO_SE_FOREIGN_PACKAGES_URL:-https://deb.debian.org/debian/dists/stable/main/binary-arm64/Packages.xz}"
debian_base_url="${UDM_PRO_SE_FOREIGN_DEBIAN_BASE_URL:-https://deb.debian.org/debian}"

mkdir -p "$foreign_dir"

curl -fL --retry 2 -o "$foreign_dir/debian-arm64-linux" "$kernel_url"
curl -fL --retry 2 -o "$foreign_dir/debian-arm64-initrd.gz" "$initrd_url"
curl -fL --retry 2 -o "$foreign_dir/Packages.xz" "$packages_url"

printf '%s\n' "$kernel_url" > "$foreign_dir/debian-arm64-linux.url"
printf '%s\n' "$initrd_url" > "$foreign_dir/debian-arm64-initrd.gz.url"
printf '%s\n' "$packages_url" > "$foreign_dir/Packages.xz.url"

kernel_release=$(strings -a "$foreign_dir/debian-arm64-linux" | sed -n 's/^Linux version \([^ ]*\).*/\1/p' | head -1)
if [ -z "$kernel_release" ]; then
    echo "could not determine foreign kernel release" >&2
    exit 1
fi

package_name="linux-image-$kernel_release"
package_file=$(xz -dc "$foreign_dir/Packages.xz" | awk -v pkg="$package_name" '
BEGIN { RS=""; FS="\n" }
{
    hit=0
    file=""
    for (i = 1; i <= NF; i++) {
        if ($i == "Package: " pkg) {
            hit=1
        }
        if ($i ~ /^Filename: /) {
            file=substr($i, 11)
        }
    }
    if (hit && file != "") {
        print file
        exit
    }
}
')

if [ -z "$package_file" ]; then
    echo "could not find $package_name in $packages_url" >&2
    exit 1
fi

package_deb="$foreign_dir/${package_file##*/}"
curl -fL --retry 2 -o "$package_deb" "$debian_base_url/$package_file"

extract_dir="$foreign_dir/linux-image-extract"
rm -rf "$extract_dir" "$foreign_dir/modules"
mkdir -p "$extract_dir" "$foreign_dir/modules"

(
    cd "$extract_dir"
    ar x "$package_deb"
    for data in data.tar.*; do
        case "$data" in
        *.zst)
            zstd -dc "$data" | tar -xf - ./usr/lib/modules
            ;;
        *.xz)
            xz -dc "$data" | tar -xf - ./usr/lib/modules
            ;;
        *.gz)
            gzip -dc "$data" | tar -xf - ./usr/lib/modules
            ;;
        *)
            tar -xf "$data" ./usr/lib/modules
            ;;
        esac
    done
)

cp -R "$extract_dir/usr/lib/modules/." "$foreign_dir/modules/"

file "$foreign_dir/debian-arm64-linux" "$foreign_dir/debian-arm64-initrd.gz"
printf 'foreign kernel release: %s\n' "$kernel_release"
printf 'foreign module package: %s\n' "$package_file"

#!/bin/sh
# Prepare local artifacts for the UDM Pro SE qemu-system-aarch64 VM profile.
set -eu

profile_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
repo_dir=$(CDPATH= cd -- "$profile_dir/../../.." && pwd)
artifacts="${UDM_PRO_SE_VM_ARTIFACTS:-$profile_dir/artifacts}"
firmware_src="${UDM_PRO_SE_VM_FIRMWARE:-$repo_dir/research/firmware/udm-pro-se-5.0.16/artifacts/473c-UDMPROSE-5.0.16-511eddc1-cb19-476d-a02d-fcaf3dbddc29.bin}"
firmware_dst="$artifacts/firmware.bin"

mkdir -p "$artifacts"

if [ ! -f "$firmware_dst" ]; then
    cp "$firmware_src" "$firmware_dst"
fi

python3 "$profile_dir/scripts/extract-fit.py" \
    --firmware "$firmware_dst" \
    --out "$artifacts"

python3 "$profile_dir/scripts/build-vm-disk.py" \
    --artifacts "$artifacts"

#!/bin/sh
# Fetch a local Zig toolchain for building ARM64 Linux lab shims.
set -eu

profile_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
artifacts="${UDM_PRO_SE_VM_ARTIFACTS:-$profile_dir/artifacts}"
toolchain_dir="$artifacts/toolchains"
version="${UDM_PRO_SE_ZIG_VERSION:-0.14.1}"
index_url="${UDM_PRO_SE_ZIG_INDEX_URL:-https://ziglang.org/download/index.json}"
index_file="$toolchain_dir/zig-index.json"

case "$(uname -s):$(uname -m)" in
Darwin:arm64|Darwin:aarch64)
    platform="aarch64-macos"
    ;;
Darwin:x86_64)
    platform="x86_64-macos"
    ;;
Linux:aarch64|Linux:arm64)
    platform="aarch64-linux"
    ;;
Linux:x86_64)
    platform="x86_64-linux"
    ;;
*)
    echo "unsupported Zig host platform: $(uname -s) $(uname -m)" >&2
    exit 1
    ;;
esac

mkdir -p "$toolchain_dir"

preferred="$toolchain_dir/zig-$platform-$version/zig"
if [ -x "$preferred" ] && "$preferred" version >/dev/null 2>&1; then
    printf '%s\n' "$preferred"
    exit 0
fi

existing=$(find "$toolchain_dir" -maxdepth 2 -type f -name zig 2>/dev/null | sort | sed -n '1p')
if [ -n "$existing" ] && "$existing" version >/dev/null 2>&1; then
    printf '%s\n' "$existing"
    exit 0
fi

curl -fL --retry 2 -o "$index_file" "$index_url"

metadata=$(python3 - "$index_file" "$version" "$platform" <<'PY'
import json
import sys

index_path, version, platform = sys.argv[1:]
with open(index_path, "r", encoding="utf-8") as handle:
    data = json.load(handle)

try:
    entry = data[version][platform]
except KeyError as exc:
    raise SystemExit(f"missing Zig download metadata for {version} {platform}") from exc

print(entry["tarball"])
print(entry.get("shasum", ""))
PY
)

tarball_url=$(printf '%s\n' "$metadata" | sed -n '1p')
tarball_sha=$(printf '%s\n' "$metadata" | sed -n '2p')
tarball="$toolchain_dir/${tarball_url##*/}"
extract_name=${tarball##*/}
extract_name=${extract_name%.tar.xz}
extract_name=${extract_name%.tar.gz}
extract_path="$toolchain_dir/$extract_name"

curl -fL --retry 2 -o "$tarball" "$tarball_url"

if [ -n "$tarball_sha" ] && command -v shasum >/dev/null 2>&1; then
    printf '%s  %s\n' "$tarball_sha" "$tarball" | shasum -a 256 -c - >&2
elif [ -n "$tarball_sha" ] && command -v sha256sum >/dev/null 2>&1; then
    printf '%s  %s\n' "$tarball_sha" "$tarball" | sha256sum -c - >&2
fi

rm -rf "$extract_path"
tar -xf "$tarball" -C "$toolchain_dir"

zig_bin="$extract_path/zig"
if [ ! -x "$zig_bin" ]; then
    zig_bin=$(find "$toolchain_dir" -maxdepth 2 -type f -name zig | sort | sed -n '1p')
fi

if [ -z "$zig_bin" ] || [ ! -x "$zig_bin" ]; then
    echo "could not find Zig executable after extracting $tarball" >&2
    exit 1
fi

printf '%s\n' "$zig_bin"

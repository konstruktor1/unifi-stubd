#!/bin/sh
set -eu

cd "$(dirname "$0")/.."

PKG_VERSION="${PKG_VERSION:-0.1.0}"
PKG_RELEASE="${PKG_RELEASE:-1}"
PKG_LICENSE="${PKG_LICENSE:-AGPL-3.0-or-later}"
PKG_MAINTAINER="${PKG_MAINTAINER:-unifi-stubd maintainers <info@spinas.org>}"
BUILD_LDFLAGS="${BUILD_LDFLAGS:--s -w -X main.version=$PKG_VERSION}"
FREEBSD_PKG_REMOTE="${FREEBSD_PKG_REMOTE:-unifi-stubd-freebsd-builder}"
FREEBSD_PKG_REMOTE_DIR="${FREEBSD_PKG_REMOTE_DIR:-/tmp/unifi-stubd-freebsd-pkg}"
DIST_DIR="${DIST_DIR:-dist}"
WORK_DIR="${WORK_DIR:-$DIST_DIR/freebsd-pkg-work}"
OUT_DIR="${OUT_DIR:-$DIST_DIR/freebsd-pkg-repos}"
ABIS="${FREEBSD_PKG_ABIS:-FreeBSD:14:amd64 FreeBSD:14:aarch64 FreeBSD:14:armv7 FreeBSD:15:amd64 FreeBSD:15:aarch64 FreeBSD:15:armv7}"

fail() {
  printf '%s\n' "$1" >&2
  exit 1
}

need_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    fail "missing required command: $1"
  fi
}

file_sha256() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | awk '{print $1}'
  elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$1" | awk '{print $1}'
  else
    fail "missing required command: sha256sum or shasum"
  fi
}

freebsd_pkg_version() {
  version="$(printf '%s' "$PKG_VERSION" | sed 's/^[vV]//; s/[-~]/./g')"
  printf '%s_%s' "$version" "$PKG_RELEASE"
}

pkg_arch() {
  case "$1" in
    FreeBSD:*:amd64)
      major="$(printf '%s' "$1" | awk -F: '{print $2}')"
      printf 'freebsd:%s:x86:64' "$major"
      ;;
    FreeBSD:*:aarch64)
      major="$(printf '%s' "$1" | awk -F: '{print $2}')"
      printf 'freebsd:%s:aarch64' "$major"
      ;;
    FreeBSD:*:armv7)
      major="$(printf '%s' "$1" | awk -F: '{print $2}')"
      printf 'freebsd:%s:armv7' "$major"
      ;;
    *)
      fail "unsupported FreeBSD ABI: $1"
      ;;
  esac
}

binary_label_for_abi() {
  case "$1" in
    FreeBSD:*:amd64)
      printf 'amd64'
      ;;
    FreeBSD:*:aarch64)
      printf 'aarch64'
      ;;
    FreeBSD:*:armv7)
      printf 'armv7'
      ;;
    *)
      fail "unsupported FreeBSD ABI: $1"
      ;;
  esac
}

build_binary() {
  label="$1"
  goarch="$2"
  goarm="$3"
  out="$WORK_DIR/bin/unifi-stubd_freebsd_$label"

  printf '== build freebsd/%s ==\n' "$label"
  if [ -n "$goarm" ]; then
    CGO_ENABLED=0 GOOS=freebsd GOARCH="$goarch" GOARM="$goarm" \
      go build -trimpath -ldflags="$BUILD_LDFLAGS" -o "$out" ./cmd/unifi-stubd
  else
    CGO_ENABLED=0 GOOS=freebsd GOARCH="$goarch" \
      go build -trimpath -ldflags="$BUILD_LDFLAGS" -o "$out" ./cmd/unifi-stubd
  fi
}

write_stage() {
  abi="$1"
  arch="$2"
  label="$(binary_label_for_abi "$abi")"
  pkg_version="$(freebsd_pkg_version)"
  stage="$WORK_DIR/stage/$abi"

  mkdir -p \
    "$stage/pkgroot/usr/local/bin" \
    "$stage/pkgroot/usr/local/etc/unifi-stubd" \
    "$stage/pkgroot/usr/local/etc/rc.d" \
    "$stage/pkgroot/usr/local/share/doc/unifi-stubd" \
    "$stage/pkgroot/var/db/unifi-stubd"

  install -m 0755 "$WORK_DIR/bin/unifi-stubd_freebsd_$label" "$stage/pkgroot/usr/local/bin/unifi-stubd"
  install -m 0600 packaging/freebsd/usr/local/etc/unifi-stubd/config.yaml "$stage/pkgroot/usr/local/etc/unifi-stubd/config.yaml"
  install -m 0755 packaging/freebsd/usr/local/etc/rc.d/unifi-stubd "$stage/pkgroot/usr/local/etc/rc.d/unifi-stubd"
  install -m 0644 LICENSE "$stage/pkgroot/usr/local/share/doc/unifi-stubd/LICENSE"
  install -m 0644 NOTICE.md "$stage/pkgroot/usr/local/share/doc/unifi-stubd/NOTICE.md"
  install -m 0644 CREDITS.md "$stage/pkgroot/usr/local/share/doc/unifi-stubd/CREDITS.md"

  binary_sum="$(file_sha256 "$stage/pkgroot/usr/local/bin/unifi-stubd")"
  rc_sum="$(file_sha256 "$stage/pkgroot/usr/local/etc/rc.d/unifi-stubd")"
  config_sum="$(file_sha256 "$stage/pkgroot/usr/local/etc/unifi-stubd/config.yaml")"
  license_sum="$(file_sha256 "$stage/pkgroot/usr/local/share/doc/unifi-stubd/LICENSE")"
  notice_sum="$(file_sha256 "$stage/pkgroot/usr/local/share/doc/unifi-stubd/NOTICE.md")"
  credits_sum="$(file_sha256 "$stage/pkgroot/usr/local/share/doc/unifi-stubd/CREDITS.md")"

  # Write the package manifest directly instead of deriving it from a plist.
  # Newer pkg-create versions can emit per-file owner/mode/mtime objects from a
  # plist. pkg 2.3.1 on OPNsense 26.1 crashes when such a package overwrites
  # unregistered files from an earlier tarball install. Simple checksum entries
  # match the older, migration-safe package manifest shape.
  cat >"$stage/manifest.ucl" <<EOF
name = "unifi-stubd"
origin = "net/unifi-stubd"
version = "$pkg_version"
comment = "Minimal UniFi device stub for isolated lab networks"
maintainer = "$PKG_MAINTAINER"
www = "https://github.com/konstruktor1/unifi-stubd"
prefix = "/usr/local"
abi = "$abi"
arch = "$arch"
licenselogic = "single"
licenses = [ "$PKG_LICENSE" ]
categories = [ "net" ]
desc = <<EOD
unifi-stubd is an experimental lab tool that makes a Linux or FreeBSD host appear as a minimal UniFi device to a UniFi Network Controller. It is intended for isolated lab or management networks and does not execute controller-provided shell, upgrade, restart, or host-networking mutations.
EOD
files = {
  "/usr/local/bin/unifi-stubd" = "1\$$binary_sum"
  "/usr/local/etc/rc.d/unifi-stubd" = "1\$$rc_sum"
  "/usr/local/etc/unifi-stubd/config.yaml" = "1\$$config_sum"
  "/usr/local/share/doc/unifi-stubd/LICENSE" = "1\$$license_sum"
  "/usr/local/share/doc/unifi-stubd/NOTICE.md" = "1\$$notice_sum"
  "/usr/local/share/doc/unifi-stubd/CREDITS.md" = "1\$$credits_sum"
}
directories = {
  "/usr/local/etc/unifi-stubd" = "y"
  "/usr/local/share/doc/unifi-stubd" = "y"
  "/var/db/unifi-stubd" = "y"
}
EOF
}

remote_build_script() {
  cat <<'EOF'
set -eu
cd "$1"
rm -rf stage repo repo.tar.gz
mkdir -p stage repo
tar -xzf stage.tar.gz -C stage
find stage -name '._*' -delete
for abi_dir in stage/*; do
  [ -d "$abi_dir" ] || continue
  abi="$(basename "$abi_dir")"
  printf '== pkg create %s ==\n' "$abi"
  mkdir -p "repo/$abi"
  pkg create -f txz -r "$abi_dir/pkgroot" -M "$abi_dir/manifest.ucl" -o "repo/$abi"
  pkg repo "repo/$abi"
done
tar -czf repo.tar.gz repo
EOF
}

need_cmd go
need_cmd ssh
need_cmd scp
need_cmd tar

rm -rf "$WORK_DIR" "$OUT_DIR"
mkdir -p "$WORK_DIR/bin" "$WORK_DIR/upload" "$OUT_DIR"

build_binary amd64 amd64 ""
build_binary aarch64 arm64 ""
build_binary armv7 arm 7

for abi in $ABIS; do
  write_stage "$abi" "$(pkg_arch "$abi")"
done

COPYFILE_DISABLE=1 tar -C "$WORK_DIR/stage" -czf "$WORK_DIR/upload/stage.tar.gz" .

remote_dir="$FREEBSD_PKG_REMOTE_DIR/$PKG_VERSION-$PKG_RELEASE-$$"
cleanup_remote() {
  ssh "$FREEBSD_PKG_REMOTE" "rm -rf '$remote_dir'; rmdir '$FREEBSD_PKG_REMOTE_DIR' 2>/dev/null || true" >/dev/null 2>&1 || true
}
trap cleanup_remote EXIT INT TERM

ssh "$FREEBSD_PKG_REMOTE" "rm -rf '$remote_dir' && mkdir -p '$remote_dir'"
scp "$WORK_DIR/upload/stage.tar.gz" "$FREEBSD_PKG_REMOTE:$remote_dir/stage.tar.gz"
remote_script="$WORK_DIR/upload/remote-build.sh"
remote_build_script >"$remote_script"
scp "$remote_script" "$FREEBSD_PKG_REMOTE:$remote_dir/remote-build.sh"
ssh "$FREEBSD_PKG_REMOTE" "sh '$remote_dir/remote-build.sh' '$remote_dir'"
scp "$FREEBSD_PKG_REMOTE:$remote_dir/repo.tar.gz" "$OUT_DIR/freebsd-pkg-repos.tar.gz"

tar -xzf "$OUT_DIR/freebsd-pkg-repos.tar.gz" -C "$OUT_DIR"
(
  cd "$OUT_DIR"
  find repo -type f | sort | while IFS= read -r file; do
    printf '%s  %s\n' "$(file_sha256 "$file")" "$file"
  done >checksums.txt
)

printf 'native FreeBSD pkg repositories written to %s/repo\n' "$OUT_DIR"

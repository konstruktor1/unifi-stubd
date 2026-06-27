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
FREEBSD_PKG_BUILD_JAILS="${FREEBSD_PKG_BUILD_JAILS:-}"
FREEBSD_PKG_REQUIRE_JAILS="${FREEBSD_PKG_REQUIRE_JAILS:-0}"
FREEBSD_TGZ_ARCHES="${FREEBSD_TGZ_ARCHES:-amd64 arm64}"
DIST_DIR="${DIST_DIR:-dist}"
WORK_DIR="${WORK_DIR:-$DIST_DIR/freebsd-pkg-work}"
OUT_DIR="${OUT_DIR:-$DIST_DIR/freebsd-pkg-repos}"
PACKAGE_DIR="${DIST_DIR}/packages"
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

file_size() {
  if stat -f %z "$1" >/dev/null 2>&1; then
    stat -f %z "$1"
  elif stat -c %s "$1" >/dev/null 2>&1; then
    stat -c %s "$1"
  else
    wc -c <"$1" | awk '{print $1}'
  fi
}

shell_quote() {
  printf "'%s'" "$(printf '%s' "$1" | sed "s/'/'\\\\''/g")"
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

binary_name_for_abi() {
  printf 'unifi-stubd_%s' "$(printf '%s' "$1" | tr ':[:upper:]' '_[:lower:]')"
}

tarball_arch_for_abi() {
  case "$1" in
    FreeBSD:*:amd64)
      printf 'amd64'
      ;;
    FreeBSD:*:aarch64)
      printf 'arm64'
      ;;
    FreeBSD:*:armv7)
      printf 'armv7'
      ;;
    *)
      fail "unsupported FreeBSD ABI: $1"
      ;;
  esac
}

tarball_enabled() {
  arch="$1"
  for enabled in $FREEBSD_TGZ_ARCHES; do
    if [ "$enabled" = "$arch" ]; then
      return 0
    fi
  done
  return 1
}

write_source_archive() {
  if command -v git >/dev/null 2>&1 && git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    git ls-files -z >"$WORK_DIR/upload/source-files"
    COPYFILE_DISABLE=1 tar --null -T "$WORK_DIR/upload/source-files" -czf "$WORK_DIR/upload/source.tar.gz"
    return 0
  fi

  COPYFILE_DISABLE=1 tar \
    --exclude './.git' \
    --exclude './dist' \
    -czf "$WORK_DIR/upload/source.tar.gz" .
}

write_build_env() {
  {
    printf 'BUILD_LDFLAGS=%s\n' "$(shell_quote "$BUILD_LDFLAGS")"
    printf 'FREEBSD_PKG_ABIS=%s\n' "$(shell_quote "$ABIS")"
    printf 'FREEBSD_PKG_BUILD_JAILS=%s\n' "$(shell_quote "$FREEBSD_PKG_BUILD_JAILS")"
    printf 'FREEBSD_PKG_REQUIRE_JAILS=%s\n' "$(shell_quote "$FREEBSD_PKG_REQUIRE_JAILS")"
  } >"$WORK_DIR/upload/build.env"
}

write_stage() {
  abi="$1"
  arch="$2"
  binary_name="$(binary_name_for_abi "$abi")"
  pkg_version="$(freebsd_pkg_version)"
  stage="$WORK_DIR/stage/$abi"

  mkdir -p \
    "$stage/pkgroot/usr/local/bin" \
    "$stage/pkgroot/usr/local/etc/unifi-stubd" \
    "$stage/pkgroot/usr/local/etc/rc.d" \
    "$stage/pkgroot/usr/local/share/doc/unifi-stubd" \
    "$stage/pkgroot/var/db/unifi-stubd"

  install -m 0755 "$WORK_DIR/bin/$binary_name" "$stage/pkgroot/usr/local/bin/unifi-stubd"
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
  flatsize="$(
    printf '%s\n' \
      "$stage/pkgroot/usr/local/bin/unifi-stubd" \
      "$stage/pkgroot/usr/local/etc/rc.d/unifi-stubd" \
      "$stage/pkgroot/usr/local/etc/unifi-stubd/config.yaml" \
      "$stage/pkgroot/usr/local/share/doc/unifi-stubd/LICENSE" \
      "$stage/pkgroot/usr/local/share/doc/unifi-stubd/NOTICE.md" \
      "$stage/pkgroot/usr/local/share/doc/unifi-stubd/CREDITS.md" |
      while IFS= read -r file; do
        file_size "$file"
      done |
      awk '{sum += $1} END {print sum}'
  )"

  # Write the package manifest directly instead of deriving it from a plist.
  # pkg-create still normalizes file entries into owner/mode/mtime objects when
  # building the archive. The remote build step repacks the generated package
  # with this manifest so OPNsense pkg 2.3.1 sees migration-safe checksum-only
  # file entries when overwriting files from an earlier tarball install. The
  # config list is still required so upgrades preserve local runtime configs
  # and write a .pkgnew on unmergeable changes instead of replacing config.yaml.
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
flatsize = $flatsize
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
config = [
  "/usr/local/etc/unifi-stubd/config.yaml"
]
scripts = {
  post-install = <<EOS
if [ -x /usr/local/bin/unifi-stubd ] && [ -f /usr/local/etc/unifi-stubd/config.yaml ]; then
  /usr/local/bin/unifi-stubd -config-migrate -config /usr/local/etc/unifi-stubd/config.yaml || true
fi
EOS
}
directories = {
  "/usr/local/etc/unifi-stubd" = "y"
  "/usr/local/share/doc/unifi-stubd" = "y"
  "/var/db/unifi-stubd" = "y"
}
EOF
}

write_tgz() {
  abi="$1"
  stage="$WORK_DIR/stage/$abi"
  arch="$(tarball_arch_for_abi "$abi")"
  target="$PACKAGE_DIR/unifi-stubd_${PKG_VERSION}-${PKG_RELEASE}_freebsd_${arch}.tar.gz"

  tarball_enabled "$arch" || return 0
  if [ -f "$target" ]; then
    return 0
  fi

  printf '== package freebsd/%s tgz ==\n' "$arch"
  if tar --version 2>/dev/null | grep -qi 'gnu tar'; then
    tar_owner_flags="--owner=0 --group=0 --numeric-owner"
  else
    tar_owner_flags="--no-xattrs --uid 0 --gid 0 --uname root --gname wheel"
  fi
  # shellcheck disable=SC2086 # tar_owner_flags is a small, controlled option list.
  COPYFILE_DISABLE=1 tar $tar_owner_flags -C "$stage/pkgroot" -czf "$target" .
  printf 'created: %s\n' "$target"
}

remote_binary_build_script() {
  cat <<'EOF'
set -eu
cd "$1"
. ./build.env
rm -rf src bin bin.tar.gz
mkdir -p src bin
tar -xzf source.tar.gz -C src
find src -name '._*' -delete

fail() {
  printf '%s\n' "$1" >&2
  exit 1
}

binary_name_for_abi() {
  printf 'unifi-stubd_%s' "$(printf '%s' "$1" | tr ':[:upper:]' '_[:lower:]')"
}

go_target_for_abi() {
  case "$1" in
    FreeBSD:*:amd64)
      printf 'amd64:'
      ;;
    FreeBSD:*:aarch64)
      printf 'arm64:'
      ;;
    FreeBSD:*:armv7)
      printf 'arm:7'
      ;;
    *)
      fail "unsupported FreeBSD ABI: $1"
      ;;
  esac
}

jail_for_abi() {
  needle="$1"
  for entry in $FREEBSD_PKG_BUILD_JAILS; do
    key="${entry%%=*}"
    value="${entry#*=}"
    if [ "$key" = "$needle" ]; then
      printf '%s' "$value"
      return 0
    fi
  done
}

require_jail_for_abi() {
  abi="$1"
  jail="$(jail_for_abi "$abi")"
  if [ -z "$jail" ] && { [ -n "$FREEBSD_PKG_BUILD_JAILS" ] || [ "$FREEBSD_PKG_REQUIRE_JAILS" = "1" ]; }; then
    fail "missing FreeBSD jail mapping for $abi"
  fi
  printf '%s' "$jail"
}

cat >build-one.sh <<'EOS'
set -eu
src="$1"
out="$2"
goarch="$3"
goarm="$4"
ldflags="$5"
go_bin="$6"
cd "$src"
mkdir -p "$(dirname "$out")"
if [ -n "$goarm" ]; then
  CGO_ENABLED=0 GOOS=freebsd GOARCH="$goarch" GOARM="$goarm" \
    "$go_bin" build -trimpath -ldflags="$ldflags" -o "$out" ./cmd/unifi-stubd
else
  CGO_ENABLED=0 GOOS=freebsd GOARCH="$goarch" \
    "$go_bin" build -trimpath -ldflags="$ldflags" -o "$out" ./cmd/unifi-stubd
fi
EOS

chmod 0755 build-one.sh

go_bin="${GO_CMD:-}"
if [ -n "$go_bin" ] && ! command -v "$go_bin" >/dev/null 2>&1; then
  fail "configured Go command not found on FreeBSD builder: $go_bin"
fi
if [ -z "$go_bin" ]; then
  for candidate in go go125 go126; do
    if command -v "$candidate" >/dev/null 2>&1; then
      go_bin="$candidate"
      break
    fi
  done
fi
[ -n "$go_bin" ] || fail "missing Go command on FreeBSD builder: install go, go125, or go126"
printf '== using Go command: %s ==\n' "$go_bin"

for abi in $FREEBSD_PKG_ABIS; do
  target="$(go_target_for_abi "$abi")"
  goarch="${target%%:*}"
  goarm="${target#*:}"
  binary_name="$(binary_name_for_abi "$abi")"
  out="$(pwd)/bin/$binary_name"
  jail="$(require_jail_for_abi "$abi")"

  if [ -n "$jail" ]; then
    printf '== build %s in jail %s ==\n' "$abi" "$jail"
    jexec "$jail" sh "$(pwd)/build-one.sh" "$(pwd)/src" "$out" "$goarch" "$goarm" "$BUILD_LDFLAGS" "$go_bin"
  else
    printf '== build %s on FreeBSD build host ==\n' "$abi"
    sh "$(pwd)/build-one.sh" "$(pwd)/src" "$out" "$goarch" "$goarm" "$BUILD_LDFLAGS" "$go_bin"
  fi
done

tar -czf bin.tar.gz bin
EOF
}

remote_build_script() {
  cat <<'EOF'
set -eu
cd "$1"
. ./build.env
rm -rf stage repo repo.tar.gz
mkdir -p stage repo
tar -xzf stage.tar.gz -C stage
find stage -name '._*' -delete

fail() {
  printf '%s\n' "$1" >&2
  exit 1
}

repair_manifest() {
  abi_dir="$1"
  pkg_file="$2"
  pkg_dir="$(dirname "$pkg_file")"
  pkg_name="$(basename "$pkg_file")"
  pkg_abs="$(cd "$pkg_dir" && pwd)/$pkg_name"
  repair_dir="$abi_dir/pkg-repair"
  tmp_pkg="$pkg_abs.tmp"

  rm -rf "$repair_dir"
  mkdir -p "$repair_dir"
  tar -xf "$pkg_abs" -C "$repair_dir"
  cp "$abi_dir/manifest.ucl" "$repair_dir/+MANIFEST"
  (
    cd "$repair_dir"
    COPYFILE_DISABLE=1 tar -cJf "$tmp_pkg" -P --no-recursion \
      -s ',^usr/,/usr/,' \
      -s ',^var/,/var/,' \
      +COMPACT_MANIFEST \
      +MANIFEST \
      usr/local/bin/unifi-stubd \
      usr/local/etc/rc.d/unifi-stubd \
      usr/local/etc/unifi-stubd/config.yaml \
      usr/local/share/doc/unifi-stubd/CREDITS.md \
      usr/local/share/doc/unifi-stubd/LICENSE \
      usr/local/share/doc/unifi-stubd/NOTICE.md \
      usr/local/etc/unifi-stubd \
      usr/local/share/doc/unifi-stubd \
      var/db/unifi-stubd
  )
  mv "$tmp_pkg" "$pkg_abs"
  rm -rf "$repair_dir"
}

jail_for_abi() {
  needle="$1"
  for entry in $FREEBSD_PKG_BUILD_JAILS; do
    key="${entry%%=*}"
    value="${entry#*=}"
    if [ "$key" = "$needle" ]; then
      printf '%s' "$value"
      return 0
    fi
  done
}

require_jail_for_abi() {
  abi="$1"
  jail="$(jail_for_abi "$abi")"
  if [ -z "$jail" ] && { [ -n "$FREEBSD_PKG_BUILD_JAILS" ] || [ "$FREEBSD_PKG_REQUIRE_JAILS" = "1" ]; }; then
    fail "missing FreeBSD jail mapping for $abi"
  fi
  printf '%s' "$jail"
}

run_pkg_create() {
  abi="$1"
  abi_dir="$2"
  abi_abs="$(cd "$abi_dir" && pwd)"
  repo_dir="$(pwd)/repo/$abi"
  jail="$(require_jail_for_abi "$abi")"
  if [ -n "$jail" ]; then
    jexec "$jail" pkg create -f txz -r "$abi_abs/pkgroot" -M "$abi_abs/manifest.ucl" -o "$repo_dir"
  else
    pkg create -f txz -r "$abi_abs/pkgroot" -M "$abi_abs/manifest.ucl" -o "$repo_dir"
  fi
}

run_pkg_repo() {
  abi="$1"
  repo_dir="$(pwd)/repo/$abi"
  jail="$(require_jail_for_abi "$abi")"
  if [ -n "$jail" ]; then
    jexec "$jail" pkg repo "$repo_dir"
  else
    pkg repo "$repo_dir"
  fi
}

for abi_dir in stage/*; do
  [ -d "$abi_dir" ] || continue
  abi="$(basename "$abi_dir")"
  printf '== pkg create %s ==\n' "$abi"
  mkdir -p "repo/$abi"
  run_pkg_create "$abi" "$abi_dir"
  pkg_file="$(ls "repo/$abi"/unifi-stubd-*.pkg | sed -n '1p')"
  [ -n "$pkg_file" ] || {
    printf 'missing generated package for %s\n' "$abi" >&2
    exit 1
  }
  repair_manifest "$abi_dir" "$pkg_file"
  run_pkg_repo "$abi"
done
tar -czf repo.tar.gz repo
EOF
}

need_cmd ssh
need_cmd scp
need_cmd tar

rm -rf "$WORK_DIR" "$OUT_DIR"
mkdir -p "$WORK_DIR/bin" "$WORK_DIR/upload" "$OUT_DIR" "$PACKAGE_DIR"

write_source_archive
write_build_env

remote_dir="$FREEBSD_PKG_REMOTE_DIR/$PKG_VERSION-$PKG_RELEASE-$$"
cleanup_remote() {
  ssh "$FREEBSD_PKG_REMOTE" "rm -rf '$remote_dir'; rmdir '$FREEBSD_PKG_REMOTE_DIR' 2>/dev/null || true" >/dev/null 2>&1 || true
}
trap cleanup_remote EXIT INT TERM

ssh "$FREEBSD_PKG_REMOTE" "rm -rf '$remote_dir' && mkdir -p '$remote_dir'"
scp "$WORK_DIR/upload/source.tar.gz" "$FREEBSD_PKG_REMOTE:$remote_dir/source.tar.gz"
scp "$WORK_DIR/upload/build.env" "$FREEBSD_PKG_REMOTE:$remote_dir/build.env"
remote_binary_script="$WORK_DIR/upload/remote-build-binaries.sh"
remote_binary_build_script >"$remote_binary_script"
scp "$remote_binary_script" "$FREEBSD_PKG_REMOTE:$remote_dir/remote-build-binaries.sh"
ssh "$FREEBSD_PKG_REMOTE" "sh '$remote_dir/remote-build-binaries.sh' '$remote_dir'"
scp "$FREEBSD_PKG_REMOTE:$remote_dir/bin.tar.gz" "$WORK_DIR/bin.tar.gz"
tar -xzf "$WORK_DIR/bin.tar.gz" -C "$WORK_DIR"

for abi in $ABIS; do
  write_stage "$abi" "$(pkg_arch "$abi")"
  write_tgz "$abi"
done

COPYFILE_DISABLE=1 tar -C "$WORK_DIR/stage" -czf "$WORK_DIR/upload/stage.tar.gz" .

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

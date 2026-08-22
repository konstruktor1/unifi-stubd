#!/bin/sh
set -eu

cd "$(dirname "$0")/.."

fail() {
  printf '%s\n' "$*" >&2
  exit 1
}

VERSION="${PKG_VERSION:-}"
RELEASE="${PKG_RELEASE:-1}"
LICENSE="${PKG_LICENSE:-AGPL-3.0-or-later}"
MAINTAINER="${PKG_MAINTAINER:-unifi-stubd maintainers <info@spinas.org>}"
GOOS_VALUE="${PKG_GOOS:-}"
GOARCH_VALUE="${PKG_GOARCH:-}"
FORMATS="${*:-${PKG_FORMATS:-deb rpm archlinux tgz}}"
DIST_DIR=dist
LDFLAGS="${BUILD_LDFLAGS:--s -w}"
FINAL_PACKAGE_DIR="$DIST_DIR/packages"
WORK_DIR=
WORK_ROOT=
PACKAGE_DIR=
PUBLISH_LOCK=
PUBLISH_LOCK_OWNED=false
PUBLISH_COMMITTED=false

if [ "$(pwd -P)" != /workspace ] || [ -z "${GOCACHE:-}" ]; then
  fail "package builds must run in the Build Runner workspace"
fi

cleanup_package_work() {
  cleanup_status=$?
  trap - EXIT HUP INT TERM
  if [ "$PUBLISH_COMMITTED" != true ] && [ -n "$PACKAGE_DIR" ] && [ -d "$PACKAGE_DIR" ]; then
    for published_source in "$PACKAGE_DIR"/*; do
      [ -f "$published_source" ] || continue
      published_target="$FINAL_PACKAGE_DIR/${published_source##*/}"
      # shellcheck disable=SC3013 # Build Runner's /bin/sh supports hardlink identity via -ef.
      if [ ! -L "$published_target" ] && [ "$published_source" -ef "$published_target" ]; then
        rm -f -- "$published_target"
      fi
    done
  fi
  if [ "$PUBLISH_LOCK_OWNED" = true ]; then
    rmdir -- "$PUBLISH_LOCK" 2>/dev/null || true
  fi
  if [ -n "$WORK_DIR" ]; then
    rm -rf -- "$WORK_DIR"
  fi
  if [ -n "$WORK_ROOT" ]; then
    rmdir -- "$WORK_ROOT" 2>/dev/null || true
  fi
  exit "$cleanup_status"
}

VERSION="${VERSION:-$(sed -n '1p' VERSION)}"
GOOS_VALUE="${GOOS_VALUE:-linux}"
GOARCH_VALUE="${GOARCH_VALUE:-$(go env GOARCH)}"

case "$GOOS_VALUE" in
linux | freebsd) ;;
*)
  printf 'package target supports linux or freebsd only, got %s\n' "$GOOS_VALUE" >&2
  exit 1
  ;;
esac

WORK_ROOT="$DIST_DIR/package-work"
PUBLISH_LOCK="$DIST_DIR/.packages.publish.lock"

if [ -L "$DIST_DIR" ] || { [ -e "$DIST_DIR" ] && [ ! -d "$DIST_DIR" ]; }; then
  fail "package output root must be a real directory: $DIST_DIR"
fi
mkdir -p "$DIST_DIR"
if [ -e "$WORK_ROOT" ] && { [ ! -d "$WORK_ROOT" ] || [ -L "$WORK_ROOT" ]; }; then
  fail "unsafe package work root: $WORK_ROOT"
fi
trap cleanup_package_work EXIT
trap 'exit 129' HUP
trap 'exit 130' INT
trap 'exit 143' TERM
mkdir -p "$WORK_ROOT"
WORK_DIR="$(mktemp -d "$WORK_ROOT/$GOOS_VALUE-$GOARCH_VALUE.XXXXXXXX")" ||
  fail 'cannot create isolated package work directory'
STAGE_DIR="$WORK_DIR/stage"
PACKAGE_DIR="$WORK_DIR/output"
NFPM_CONFIG="$WORK_DIR/nfpm.yaml"
mkdir "$PACKAGE_DIR"

archlinux_version() {
  printf '%s' "$1" | sed 's/[-~]//g'
}

install_docs() {
  doc_dir="$1"
  install -m 0644 LICENSE "$doc_dir/LICENSE"
  install -m 0644 NOTICE.md "$doc_dir/NOTICE.md"
  install -m 0644 CREDITS.md "$doc_dir/CREDITS.md"
}

mkdir -p "$STAGE_DIR/usr/local/bin"
case "$GOOS_VALUE" in
linux)
  mkdir -p \
    "$STAGE_DIR/etc/unifi-stubd" \
    "$STAGE_DIR/etc/init.d" \
    "$STAGE_DIR/usr/lib/systemd/system" \
    "$STAGE_DIR/usr/share/doc/unifi-stubd" \
    "$STAGE_DIR/var/lib/unifi-stubd"
  install -m 0640 packaging/linux/etc/unifi-stubd/config.yaml "$STAGE_DIR/etc/unifi-stubd/config.yaml"
  install -m 0755 packaging/linux/etc/init.d/unifi-stubd "$STAGE_DIR/etc/init.d/unifi-stubd"
  install -m 0644 packaging/linux/usr/lib/systemd/system/unifi-stubd.service "$STAGE_DIR/usr/lib/systemd/system/unifi-stubd.service"
  install_docs "$STAGE_DIR/usr/share/doc/unifi-stubd"
  ;;
freebsd)
  mkdir -p \
    "$STAGE_DIR/usr/local/etc/unifi-stubd" \
    "$STAGE_DIR/usr/local/etc/rc.d" \
    "$STAGE_DIR/usr/local/share/doc/unifi-stubd" \
    "$STAGE_DIR/var/db/unifi-stubd"
  install -m 0600 packaging/freebsd/usr/local/etc/unifi-stubd/config.yaml "$STAGE_DIR/usr/local/etc/unifi-stubd/config.yaml"
  install -m 0755 packaging/freebsd/usr/local/etc/rc.d/unifi-stubd "$STAGE_DIR/usr/local/etc/rc.d/unifi-stubd"
  install_docs "$STAGE_DIR/usr/local/share/doc/unifi-stubd"
  ;;
esac

printf '== build %s/%s ==\n' "$GOOS_VALUE" "$GOARCH_VALUE"
CGO_ENABLED=0 GOOS="$GOOS_VALUE" GOARCH="$GOARCH_VALUE" \
  go build -trimpath -ldflags="$LDFLAGS" -o "$STAGE_DIR/usr/local/bin/unifi-stubd" ./cmd/unifi-stubd

sed_escape() {
  printf '%s' "$1" | sed 's/[\/&|\\]/\\&/g'
}

write_nfpm_config() {
  format="$1"
  nfpm_version="$VERSION"
  nfpm_version_schema="semver"
  if [ "$format" = "archlinux" ]; then
    nfpm_version="$(archlinux_version "$VERSION")"
    nfpm_version_schema="none"
  fi
  sed \
    -e "s|@PKG_ARCH@|$(sed_escape "$GOARCH_VALUE")|g" \
    -e "s|@PKG_STAGE_DIR@|$(sed_escape "$STAGE_DIR")|g" \
    -e "s|@PKG_VERSION@|$(sed_escape "$nfpm_version")|g" \
    -e "s|@PKG_VERSION_SCHEMA@|$(sed_escape "$nfpm_version_schema")|g" \
    -e "s|@PKG_RELEASE@|$(sed_escape "$RELEASE")|g" \
    -e "s|@PKG_LICENSE@|$(sed_escape "$LICENSE")|g" \
    -e "s|@PKG_MAINTAINER@|$(sed_escape "$MAINTAINER")|g" \
    packaging/nfpm.yaml >"$NFPM_CONFIG"
}

build_nfpm() {
  format="$1"
  if [ "$GOOS_VALUE" != "linux" ]; then
    printf 'format %s is linux-only; use tgz for freebsd packages\n' "$format" >&2
    exit 1
  fi
  write_nfpm_config "$format"
  printf '== package %s ==\n' "$format"
  go tool github.com/goreleaser/nfpm/v2/cmd/nfpm package \
    -f "$NFPM_CONFIG" -p "$format" -t "$PACKAGE_DIR"
}

build_tgz() {
  target="$PACKAGE_DIR/unifi-stubd_${VERSION}-${RELEASE}_${GOOS_VALUE}_${GOARCH_VALUE}.tar.gz"
  printf '== package tgz ==\n'
  if tar --version 2>/dev/null | grep -qi 'gnu tar'; then
    tar_owner_flags="--owner=0 --group=0 --numeric-owner"
  else
    tar_owner_flags="--no-xattrs --uid 0 --gid 0 --uname root --gname wheel"
  fi
  # shellcheck disable=SC2086 # tar_owner_flags is a small, controlled option list.
  COPYFILE_DISABLE=1 tar $tar_owner_flags -C "$STAGE_DIR" -czf "$target" .
  printf 'created: %s\n' "$target"
}

format_count=0
for format in $FORMATS; do
  format_count=$((format_count + 1))
  case "$format" in
  deb | rpm | archlinux)
    build_nfpm "$format"
    ;;
  tgz | tar.gz)
    build_tgz
    ;;
  *)
    printf 'unknown package format: %s\n' "$format" >&2
    exit 1
    ;;
  esac
done
[ "$format_count" -gt 0 ] || fail 'at least one package format is required'

rm -rf -- "$STAGE_DIR"
rm -f -- "$NFPM_CONFIG"

mkdir "$PUBLISH_LOCK" || fail 'another package publication is active'
PUBLISH_LOCK_OWNED=true
if [ -e "$FINAL_PACKAGE_DIR" ]; then
  if [ ! -d "$FINAL_PACKAGE_DIR" ] || [ -L "$FINAL_PACKAGE_DIR" ]; then
    fail "unsafe package output directory: $FINAL_PACKAGE_DIR"
  fi
else
  mkdir "$FINAL_PACKAGE_DIR"
fi
for artifact in "$PACKAGE_DIR"/*; do
  if [ ! -f "$artifact" ] || [ -L "$artifact" ]; then
    fail 'package output is not a regular file'
  fi
  publish_target="$FINAL_PACKAGE_DIR/${artifact##*/}"
  if [ -e "$publish_target" ] || [ -L "$publish_target" ]; then
    fail "package output already exists: $publish_target"
  fi
  ln -- "$artifact" "$publish_target" || fail "failed to publish package output: $publish_target"
done
PUBLISH_COMMITTED=true

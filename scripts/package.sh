#!/bin/sh
set -eu

cd "$(dirname "$0")/.."

VERSION="${PKG_VERSION:-0.1.0}"
RELEASE="${PKG_RELEASE:-1}"
LICENSE="${PKG_LICENSE:-AGPL-3.0-or-later}"
MAINTAINER="${PKG_MAINTAINER:-unifi-stubd maintainers <info@spinas.org>}"
GOOS_VALUE="${PKG_GOOS:-linux}"
GOARCH_VALUE="${PKG_GOARCH:-$(go env GOARCH)}"
FORMATS="${*:-${PKG_FORMATS:-deb rpm archlinux tgz}}"
NFPM_CMD="${NFPM:-go run github.com/goreleaser/nfpm/v2/cmd/nfpm@v2.46.3}"
DIST_DIR="${DIST_DIR:-dist}"
STAGE_DIR="${DIST_DIR}/stage/pkgroot"
PACKAGE_DIR="${DIST_DIR}/packages"
NFPM_CONFIG="${DIST_DIR}/nfpm.yaml"
LDFLAGS="${BUILD_LDFLAGS:--s -w}"

case "$GOOS_VALUE" in
  linux|freebsd)
    ;;
  *)
    printf 'package target supports linux or freebsd only, got %s\n' "$GOOS_VALUE" >&2
    exit 1
    ;;
esac

rm -rf "$STAGE_DIR"
mkdir -p "$PACKAGE_DIR"
rm -f "$PACKAGE_DIR"/unifi-stubd_*.deb \
      "$PACKAGE_DIR"/unifi-stubd_*.rpm \
      "$PACKAGE_DIR"/unifi-stubd_*.pkg.tar.zst \
      "$PACKAGE_DIR"/unifi-stubd_*.tar.gz 2>/dev/null || true

install_docs() {
  doc_dir="$1"
  install -m 0644 LICENSE "$doc_dir/LICENSE"
  install -m 0644 NOTICE.md "$doc_dir/NOTICE.md"
  install -m 0644 CREDITS.md "$doc_dir/CREDITS.md"
}

prepare_linux_stage() {
  mkdir -p \
    "$STAGE_DIR/usr/local/bin" \
    "$STAGE_DIR/etc/unifi-stubd" \
    "$STAGE_DIR/etc/init.d" \
    "$STAGE_DIR/usr/lib/systemd/system" \
    "$STAGE_DIR/usr/share/doc/unifi-stubd" \
    "$STAGE_DIR/var/lib/unifi-stubd"
}

install_linux_stage() {
  install -m 0640 packaging/linux/etc/unifi-stubd/config.yaml "$STAGE_DIR/etc/unifi-stubd/config.yaml"
  install -m 0755 packaging/linux/etc/init.d/unifi-stubd "$STAGE_DIR/etc/init.d/unifi-stubd"
  install -m 0644 packaging/linux/usr/lib/systemd/system/unifi-stubd.service "$STAGE_DIR/usr/lib/systemd/system/unifi-stubd.service"
  install_docs "$STAGE_DIR/usr/share/doc/unifi-stubd"
}

prepare_freebsd_stage() {
  mkdir -p \
    "$STAGE_DIR/usr/local/bin" \
    "$STAGE_DIR/usr/local/etc/unifi-stubd" \
    "$STAGE_DIR/usr/local/etc/rc.d" \
    "$STAGE_DIR/usr/local/share/doc/unifi-stubd" \
    "$STAGE_DIR/var/db/unifi-stubd"
}

install_freebsd_stage() {
  install -m 0600 packaging/freebsd/usr/local/etc/unifi-stubd/config.yaml "$STAGE_DIR/usr/local/etc/unifi-stubd/config.yaml"
  install -m 0755 packaging/freebsd/usr/local/etc/rc.d/unifi-stubd "$STAGE_DIR/usr/local/etc/rc.d/unifi-stubd"
  install_docs "$STAGE_DIR/usr/local/share/doc/unifi-stubd"
}

case "$GOOS_VALUE" in
  linux)
    prepare_linux_stage
    ;;
  freebsd)
    prepare_freebsd_stage
    ;;
esac

printf '== build %s/%s ==\n' "$GOOS_VALUE" "$GOARCH_VALUE"
CGO_ENABLED=0 GOOS="$GOOS_VALUE" GOARCH="$GOARCH_VALUE" \
  go build -trimpath -ldflags="$LDFLAGS" -o "$STAGE_DIR/usr/local/bin/unifi-stubd" ./cmd/unifi-stubd

case "$GOOS_VALUE" in
  linux)
    install_linux_stage
    ;;
  freebsd)
    install_freebsd_stage
    ;;
esac

sed_escape() {
  printf '%s' "$1" | sed 's/[\/&|\\]/\\&/g'
}

write_nfpm_config() {
  sed \
    -e "s|@PKG_ARCH@|$(sed_escape "$GOARCH_VALUE")|g" \
    -e "s|@PKG_VERSION@|$(sed_escape "$VERSION")|g" \
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
  if [ ! -f "$NFPM_CONFIG" ]; then
    write_nfpm_config
  fi
  printf '== package %s ==\n' "$format"
  $NFPM_CMD package -f "$NFPM_CONFIG" -p "$format" -t "$PACKAGE_DIR"
}

build_tgz() {
  target="$PACKAGE_DIR/unifi-stubd_${VERSION}-${RELEASE}_${GOOS_VALUE}_${GOARCH_VALUE}.tar.gz"
  printf '== package tgz ==\n'
  COPYFILE_DISABLE=1 tar --no-xattrs --uid 0 --gid 0 --uname root --gname wheel -C "$STAGE_DIR" -czf "$target" .
  printf 'created: %s\n' "$target"
}

for format in $FORMATS; do
  case "$format" in
    deb|rpm|archlinux)
      build_nfpm "$format"
      ;;
    tgz|tar.gz)
      build_tgz
      ;;
    *)
      printf 'unknown package format: %s\n' "$format" >&2
      exit 1
      ;;
  esac
done

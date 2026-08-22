#!/bin/sh
set -eu

cd "$(dirname "$0")/.."

PACKAGE_DIR="${PACKAGE_DIR:-dist/packages}"
SITE_DIR=dist/package-site
CHANNEL="${CHANNEL:-alpha}"
FREEBSD_PKG_REPO_DIR="${FREEBSD_PKG_REPO_DIR:-dist/freebsd-pkg-repos/repo}"
REQUIRE_FREEBSD_ARTIFACTS="${REQUIRE_FREEBSD_ARTIFACTS:-1}"
TEMPLATE_DIR=scripts/package-site

fail() {
  printf '%s\n' "$1" >&2
  exit 1
}

need_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    fail "missing required command: $1"
  fi
}

find_artifact() {
  pattern="$1"
  for path in $pattern; do
    if [ -f "$path" ]; then
      printf '%s\n' "$path"
      return 0
    fi
  done
  fail "missing package artifact matching: $pattern"
}

find_optional_artifact() {
  pattern="$1"
  for path in $pattern; do
    if [ -f "$path" ]; then
      printf '%s\n' "$path"
      return 0
    fi
  done
  return 1
}

file_md5() {
  if command -v md5sum >/dev/null 2>&1; then
    md5sum "$1" | awk '{print $1}'
  elif command -v md5 >/dev/null 2>&1; then
    md5 -q "$1"
  else
    fail "missing required command: md5sum or md5"
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
  wc -c <"$1" | tr -d ' '
}

copy_artifact() {
  src="$1"
  dst_dir="$2"
  mkdir -p "$dst_dir"
  install -m 0644 "$src" "$dst_dir/"
}

write_apt_release() {
  dist_dir="$1"
  release_file="$dist_dir/Release"
  {
    printf 'Origin: unifi-stubd\n'
    printf 'Label: unifi-stubd\n'
    printf 'Suite: %s\n' "$CHANNEL"
    printf 'Codename: %s\n' "$CHANNEL"
    printf 'Date: %s\n' "$(LC_ALL=C date -u '+%a, %d %b %Y %H:%M:%S +0000')"
    printf 'Architectures: amd64 arm64\n'
    printf 'Components: main\n'
    printf 'Description: unsigned alpha package repository for unifi-stubd\n'
    printf 'MD5Sum:\n'
    find "$dist_dir" -type f ! -name Release | sort | while IFS= read -r file; do
      rel="${file#"$dist_dir"/}"
      printf ' %s %16s %s\n' "$(file_md5 "$file")" "$(file_size "$file")" "$rel"
    done
    printf 'SHA256:\n'
    find "$dist_dir" -type f ! -name Release | sort | while IFS= read -r file; do
      rel="${file#"$dist_dir"/}"
      printf ' %s %16s %s\n' "$(file_sha256 "$file")" "$(file_size "$file")" "$rel"
    done
  } >"$release_file"
}

write_checksums() {
  (
    cd "$SITE_DIR"
    find . -type f ! -name checksums.txt | sort | while IFS= read -r file; do
      clean="${file#./}"
      printf '%s  %s\n' "$(file_sha256 "$clean")" "$clean"
    done
  ) >"$SITE_DIR/checksums.txt"
}

validate_html_token() {
  case "$1" in
  '' | *[!A-Za-z0-9._:+~-]*) fail "unsafe package-site value: $1" ;;
  esac
}

validate_templates() {
  for template in \
    index.html \
    freebsd.html freebsd-unavailable.html \
    freebsd-arch.html freebsd-arch-unavailable.html \
    freebsd-pkg-abi.html freebsd-pkg-unavailable.html \
    freebsd-pkg-index-head.html freebsd-pkg-index-tail.html; do
    path="$TEMPLATE_DIR/$template"
    if [ ! -f "$path" ] || [ -L "$path" ]; then
      fail "missing package-site template: $path"
    fi
  done
}

write_html_page() {
  output="$1"
  title="$2"
  {
    printf '%s\n' '<!doctype html>' '<html lang="en">' '<head>' \
      '  <meta charset="utf-8">' \
      '  <meta name="viewport" content="width=device-width, initial-scale=1">'
    printf '  <title>%s</title>\n</head>\n<body>\n  <main>\n    <h1>%s</h1>\n' "$title" "$title"
    cat
    printf '%s\n' '  </main>' '</body>' '</html>'
  } >"$output"
}

write_index() {
  sed "s|@CHANNEL@|$CHANNEL|g" "$TEMPLATE_DIR/index.html" |
    write_html_page "$SITE_DIR/index.html" 'unifi-stubd alpha package repositories'
}

write_freebsd_index() {
  freebsd_dir="$SITE_DIR/freebsd"
  mkdir -p "$freebsd_dir"
  write_html_page "$freebsd_dir/index.html" 'unifi-stubd FreeBSD and OPNsense tarballs' \
    <"$TEMPLATE_DIR/freebsd.html"
}

write_freebsd_unavailable_index() {
  freebsd_dir="$SITE_DIR/freebsd"
  mkdir -p "$freebsd_dir"
  write_html_page "$freebsd_dir/index.html" 'unifi-stubd FreeBSD and OPNsense artifacts' \
    <"$TEMPLATE_DIR/freebsd-unavailable.html"

  for arch in amd64 arm64; do
    arch_dir="$freebsd_dir/$arch"
    mkdir -p "$arch_dir"
    sed "s|@ARCH@|$arch|g" "$TEMPLATE_DIR/freebsd-arch-unavailable.html" |
      write_html_page "$arch_dir/index.html" "unifi-stubd FreeBSD ${arch} artifacts"
  done
}

write_freebsd_pkg_unavailable_index() {
  pkg_dir="$SITE_DIR/freebsd/pkg"
  mkdir -p "$pkg_dir"
  write_html_page "$pkg_dir/index.html" 'unifi-stubd FreeBSD pkg repositories' \
    <"$TEMPLATE_DIR/freebsd-pkg-unavailable.html"
}

write_freebsd_pkg_index() {
  pkg_dir="$SITE_DIR/freebsd/pkg"
  mkdir -p "$pkg_dir"
  {
    cat "$TEMPLATE_DIR/freebsd-pkg-index-head.html"
    find "$pkg_dir" -mindepth 1 -maxdepth 1 -type d | sort | while IFS= read -r abi_dir; do
      abi="$(basename "$abi_dir")"
      validate_html_token "$abi"
      printf '      <li><a href="%s/">%s</a></li>\n' "$abi" "$abi"
    done
    cat "$TEMPLATE_DIR/freebsd-pkg-index-tail.html"
  } | write_html_page "$pkg_dir/index.html" 'unifi-stubd FreeBSD pkg repositories'
}

write_freebsd_pkg_abi_index() {
  abi="$1"
  abi_dir="$SITE_DIR/freebsd/pkg/$abi"
  artifact="$(basename "$(find_artifact "$abi_dir"/unifi-stubd-*.pkg)")"
  validate_html_token "$abi"
  validate_html_token "$artifact"
  sed -e "s|@ABI@|$abi|g" -e "s|@ARTIFACT@|$artifact|g" \
    "$TEMPLATE_DIR/freebsd-pkg-abi.html" |
    write_html_page "$abi_dir/index.html" "unifi-stubd ${abi} pkg repository"
}

write_freebsd_arch_index() {
  arch="$1"
  arch_dir="$SITE_DIR/freebsd/$arch"
  artifact="$(basename "$(find_artifact "$arch_dir"/unifi-stubd_*_freebsd_"$arch".tar.gz)")"
  validate_html_token "$arch"
  validate_html_token "$artifact"
  sed -e "s|@ARCH@|$arch|g" -e "s|@ARTIFACT@|$artifact|g" \
    "$TEMPLATE_DIR/freebsd-arch.html" |
    write_html_page "$arch_dir/index.html" "unifi-stubd FreeBSD ${arch} tarball"
}

build_apt_repo() {
  need_cmd dpkg-scanpackages
  apt_dir="$SITE_DIR/apt"
  pool_dir="$apt_dir/pool/main/u/unifi-stubd"

  copy_artifact "$(find_artifact "$PACKAGE_DIR"/unifi-stubd_*_amd64.deb)" "$pool_dir"
  copy_artifact "$(find_artifact "$PACKAGE_DIR"/unifi-stubd_*_arm64.deb)" "$pool_dir"

  for arch in amd64 arm64; do
    binary_dir="$apt_dir/dists/$CHANNEL/main/binary-$arch"
    mkdir -p "$binary_dir"
    (
      cd "$apt_dir"
      dpkg-scanpackages --arch "$arch" pool /dev/null >"dists/$CHANNEL/main/binary-$arch/Packages"
      gzip -9 -kf "dists/$CHANNEL/main/binary-$arch/Packages"
    )
  done
  write_apt_release "$apt_dir/dists/$CHANNEL"
}

build_rpm_repo() {
  need_cmd createrepo_c
  rpm_x86_dir="$SITE_DIR/rpm/x86_64"
  rpm_arm_dir="$SITE_DIR/rpm/aarch64"

  copy_artifact "$(find_artifact "$PACKAGE_DIR"/unifi-stubd-*.x86_64.rpm)" "$rpm_x86_dir"
  copy_artifact "$(find_artifact "$PACKAGE_DIR"/unifi-stubd-*.aarch64.rpm)" "$rpm_arm_dir"

  createrepo_c "$rpm_x86_dir"
  createrepo_c "$rpm_arm_dir"
}

build_arch_repo() {
  need_cmd repo-add
  need_cmd bsdtar
  need_cmd zstd
  arch_x86_dir="$SITE_DIR/arch/x86_64"
  arch_arm_dir="$SITE_DIR/arch/aarch64"

  copy_artifact "$(find_artifact "$PACKAGE_DIR"/unifi-stubd-*-x86_64.pkg.tar.zst)" "$arch_x86_dir"
  copy_artifact "$(find_artifact "$PACKAGE_DIR"/unifi-stubd-*-aarch64.pkg.tar.zst)" "$arch_arm_dir"

  for repo_dir in "$arch_x86_dir" "$arch_arm_dir"; do
    (
      cd "$repo_dir"
      rm -f unifi-stubd.db unifi-stubd.db.tar.* unifi-stubd.files unifi-stubd.files.tar.*
      repo-add unifi-stubd.db.tar.gz ./*.pkg.tar.zst
      if [ -L unifi-stubd.db ]; then
        cp -L unifi-stubd.db unifi-stubd.db.copy
        mv unifi-stubd.db.copy unifi-stubd.db
      fi
      if [ -L unifi-stubd.files ]; then
        cp -L unifi-stubd.files unifi-stubd.files.copy
        mv unifi-stubd.files.copy unifi-stubd.files
      fi
    )
  done
}

copy_freebsd_tarballs() {
  amd64_artifact="$(find_optional_artifact "$PACKAGE_DIR"/unifi-stubd_*_freebsd_amd64.tar.gz || true)"
  arm64_artifact="$(find_optional_artifact "$PACKAGE_DIR"/unifi-stubd_*_freebsd_arm64.tar.gz || true)"

  if [ -z "$amd64_artifact" ] && [ -z "$arm64_artifact" ]; then
    if [ "$REQUIRE_FREEBSD_ARTIFACTS" = "1" ]; then
      fail "missing required FreeBSD tarball artifacts in $PACKAGE_DIR"
    fi
    write_freebsd_unavailable_index
    return 0
  fi
  if [ -z "$amd64_artifact" ] || [ -z "$arm64_artifact" ]; then
    fail "FreeBSD tarball set is incomplete in $PACKAGE_DIR"
  fi

  copy_artifact "$amd64_artifact" "$SITE_DIR/freebsd/amd64"
  copy_artifact "$arm64_artifact" "$SITE_DIR/freebsd/arm64"
  write_freebsd_index
  write_freebsd_arch_index amd64
  write_freebsd_arch_index arm64
}

copy_freebsd_pkg_repos() {
  if [ ! -d "$FREEBSD_PKG_REPO_DIR" ]; then
    if [ "$REQUIRE_FREEBSD_ARTIFACTS" = "1" ]; then
      fail "missing required native FreeBSD pkg repositories: $FREEBSD_PKG_REPO_DIR"
    fi
    write_freebsd_pkg_unavailable_index
    return 0
  fi

  pkg_dir="$SITE_DIR/freebsd/pkg"
  mkdir -p "$pkg_dir"
  copied=0
  for src in "$FREEBSD_PKG_REPO_DIR"/FreeBSD:*; do
    [ -d "$src" ] || continue
    abi="$(basename "$src")"
    dst="$pkg_dir/$abi"
    mkdir -p "$dst"
    cp -R "$src"/. "$dst"/
    write_freebsd_pkg_abi_index "$abi"
    copied=1
  done

  if [ "$copied" -eq 0 ]; then
    if [ "$REQUIRE_FREEBSD_ARTIFACTS" = "1" ]; then
      fail "no native FreeBSD pkg repositories found in $FREEBSD_PKG_REPO_DIR"
    fi
    write_freebsd_pkg_unavailable_index
    return 0
  fi
  write_freebsd_pkg_index
}

if [ ! -d "$PACKAGE_DIR" ]; then
  fail "package directory not found: $PACKAGE_DIR"
fi
case "$REQUIRE_FREEBSD_ARTIFACTS" in
0 | 1) ;;
*)
  fail "REQUIRE_FREEBSD_ARTIFACTS must be 0 or 1"
  ;;
esac
validate_html_token "$CHANNEL"
validate_templates

if [ -L dist ] || { [ -e dist ] && [ ! -d dist ]; }; then
  fail "package output root must be a real directory: dist"
fi
mkdir -p dist
rm -rf "$SITE_DIR"
mkdir -p "$SITE_DIR"
touch "$SITE_DIR/.nojekyll"

build_apt_repo
build_rpm_repo
build_arch_repo
copy_freebsd_tarballs
copy_freebsd_pkg_repos
write_index
write_checksums

printf 'package repository site written to %s\n' "$SITE_DIR"

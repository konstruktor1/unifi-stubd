#!/bin/sh
set -eu

cd "$(dirname "$0")/.."

PACKAGE_DIR="${PACKAGE_DIR:-dist/packages}"
SITE_DIR="${SITE_DIR:-dist/package-site}"
CHANNEL="${CHANNEL:-alpha}"

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

write_index() {
  cat >"$SITE_DIR/index.html" <<EOF
<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>unifi-stubd alpha package repositories</title>
</head>
<body>
  <main>
    <h1>unifi-stubd alpha package repositories</h1>
    <p>
      This site hosts unsigned alpha package repositories for
      <a href="https://github.com/konstruktor1/unifi-stubd">unifi-stubd</a>,
      a lab-only UniFi Network device stub for Proxmox, Linux bridges, and
      FreeBSD/OPNsense. Use only in isolated lab or management networks.
    </p>
    <h2>Project Links</h2>
    <ul>
      <li><a href="https://github.com/konstruktor1/unifi-stubd">Source repository</a></li>
      <li><a href="https://github.com/konstruktor1/unifi-stubd/releases">GitHub releases</a></li>
      <li><a href="https://github.com/konstruktor1/unifi-stubd/wiki">GitHub wiki</a></li>
      <li><a href="https://github.com/konstruktor1/unifi-stubd/blob/main/CREDITS.md">Credits and research sources</a></li>
      <li><a href="https://github.com/konstruktor1/unifi-stubd/blob/main/NOTICE.md">Notice and trademark notes</a></li>
    </ul>
    <h2>APT</h2>
    <pre>deb [trusted=yes arch=amd64] https://konstruktor1.github.io/unifi-stubd/apt ${CHANNEL} main</pre>
    <h2>RPM</h2>
    <pre>baseurl=https://konstruktor1.github.io/unifi-stubd/rpm/\$basearch</pre>
    <h2>Arch Linux</h2>
    <pre>Server = https://konstruktor1.github.io/unifi-stubd/arch/\$arch</pre>
    <h2>FreeBSD and OPNsense</h2>
    <p>
      FreeBSD and OPNsense artifacts are tarballs for now, not native
      FreeBSD <code>pkg</code> repositories. Both amd64 and arm64 tarballs are
      published with rc.d service files and neutral defaults.
    </p>
    <ul>
      <li><a href="freebsd/amd64/">FreeBSD amd64 tarballs</a></li>
      <li><a href="freebsd/arm64/">FreeBSD arm64 tarballs</a></li>
    </ul>
    <h2>Checksums</h2>
    <p><a href="checksums.txt">checksums.txt</a></p>
    <h2>Research and Attribution</h2>
    <p>
      The implementation is original project code. Public documentation and
      reverse-engineering projects were used as protocol references, historical
      context, and cross-checks; no source code from unlicensed research
      repositories is copied into unifi-stubd.
    </p>
    <table>
      <thead>
        <tr>
          <th>Project or source</th>
          <th>How it informed unifi-stubd</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td><a href="https://help.ui.com/hc/en-us/articles/218506997-UniFi-Network-Required-Ports-Reference">Ubiquiti required ports</a> and <a href="https://help.ui.com/hc/en-us/articles/204909754-Remote-Adoption-Layer-3">remote adoption docs</a></td>
          <td>Official discovery, inform, adoption, and service-port framing.</td>
        </tr>
        <tr>
          <td><a href="https://techspecs.ui.com/">Ubiquiti tech specs and datasheets</a></td>
          <td>Controller-visible device model and port-layout sanity checks.</td>
        </tr>
        <tr>
          <td><a href="https://jrjparks.github.io/unofficial-unifi-guide/">The unofficial guide to UniFi</a></td>
          <td>Discovery, inform, adoption, encryption, and compression reference notes.</td>
        </tr>
        <tr>
          <td><a href="https://github.com/jeffreykog/unifi-inform-protocol">jeffreykog/unifi-inform-protocol</a> and <a href="https://github.com/fxkr/unifi-protocol-reverse-engineering">fxkr/unifi-protocol-reverse-engineering</a></td>
          <td>Independent UniFi discovery and inform protocol cross-checks.</td>
        </tr>
        <tr>
          <td><a href="https://github.com/mcrute/ubntmfi/blob/master/inform_protocol.md">mcrute inform protocol notes</a>, <a href="https://github.com/jda/pixiedust">jda/pixiedust</a>, <a href="https://github.com/ZAP-Quebec/unifi-inform">ZAP-Quebec/unifi-inform</a>, and <a href="https://github.com/dmke/inform-inspect">dmke/inform-inspect</a></td>
          <td>Packet framing, adoption-state, parser, and lab-comparison references.</td>
        </tr>
        <tr>
          <td><a href="https://github.com/wvengen/unifi-controllable-switch">wvengen/unifi-controllable-switch</a> and <a href="https://github.com/stephanlascar/unifi-gateway">stephanlascar/unifi-gateway</a></td>
          <td>Historical fake switch and gateway-emulation project context.</td>
        </tr>
        <tr>
          <td><a href="https://github.com/openwrt/openwrt/tree/main/package/network/config/swconfig">OpenWrt swconfig</a>, <a href="https://docs.docker.com/engine/network/drivers/macvlan/">Docker macvlan docs</a>, and <a href="https://kernel.org/doc/html/next/networking/bridge.html">Linux bridge docs</a></td>
          <td>Lab simulation, bridge observation, and future network-mode planning references.</td>
        </tr>
      </tbody>
    </table>
    <p>
      See the full attribution matrix in
      <a href="https://github.com/konstruktor1/unifi-stubd/blob/main/CREDITS.md">CREDITS.md</a>.
    </p>
  </main>
</body>
</html>
EOF
}

write_freebsd_index() {
  freebsd_dir="$SITE_DIR/freebsd"
  mkdir -p "$freebsd_dir"
  cat >"$freebsd_dir/index.html" <<EOF
<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>unifi-stubd FreeBSD and OPNsense tarballs</title>
</head>
<body>
  <main>
    <h1>unifi-stubd FreeBSD and OPNsense tarballs</h1>
    <p>
      FreeBSD and OPNsense builds are currently published as tarballs, not as a
      native FreeBSD <code>pkg</code> repository. Use these artifacts only for
      isolated lab or management networks.
    </p>
    <ul>
      <li><a href="amd64/">FreeBSD amd64 tarballs</a></li>
      <li><a href="arm64/">FreeBSD arm64 tarballs</a></li>
    </ul>
    <p><a href="../checksums.txt">checksums.txt</a></p>
    <p><a href="../">Back to package repositories</a></p>
  </main>
</body>
</html>
EOF
}

write_freebsd_arch_index() {
  arch="$1"
  arch_dir="$SITE_DIR/freebsd/$arch"
  artifact="$(basename "$(find_artifact "$arch_dir"/unifi-stubd_*_freebsd_"$arch".tar.gz)")"
  cat >"$arch_dir/index.html" <<EOF
<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>unifi-stubd FreeBSD ${arch} tarball</title>
</head>
<body>
  <main>
    <h1>unifi-stubd FreeBSD ${arch} tarball</h1>
    <p>
      This is a tarball artifact for FreeBSD and OPNsense. It is not a native
      FreeBSD <code>pkg</code> package yet.
    </p>
    <ul>
      <li><a href="${artifact}">${artifact}</a></li>
      <li><a href="../../checksums.txt">checksums.txt</a></li>
    </ul>
    <h2>Fetch and Inspect</h2>
    <pre>fetch https://konstruktor1.github.io/unifi-stubd/freebsd/${arch}/${artifact}
fetch https://konstruktor1.github.io/unifi-stubd/checksums.txt
grep 'freebsd/${arch}/${artifact}' checksums.txt
sha256 ${artifact}
tar -tzf ${artifact}</pre>
    <h2>Extract</h2>
    <pre>sudo tar -xzf ${artifact} -C /
sudo vi /usr/local/etc/unifi-stubd/config.yaml
sudo sysrc unifi_stubd_enable=YES
sudo service unifi-stubd start</pre>
    <p><a href="../">Back to FreeBSD and OPNsense tarballs</a></p>
  </main>
</body>
</html>
EOF
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
  copy_artifact "$(find_artifact "$PACKAGE_DIR"/unifi-stubd_*_freebsd_amd64.tar.gz)" "$SITE_DIR/freebsd/amd64"
  copy_artifact "$(find_artifact "$PACKAGE_DIR"/unifi-stubd_*_freebsd_arm64.tar.gz)" "$SITE_DIR/freebsd/arm64"
  write_freebsd_index
  write_freebsd_arch_index amd64
  write_freebsd_arch_index arm64
}

if [ ! -d "$PACKAGE_DIR" ]; then
  fail "package directory not found: $PACKAGE_DIR"
fi

rm -rf "$SITE_DIR"
mkdir -p "$SITE_DIR"
touch "$SITE_DIR/.nojekyll"

build_apt_repo
build_rpm_repo
build_arch_repo
copy_freebsd_tarballs
write_index
write_checksums

printf 'package repository site written to %s\n' "$SITE_DIR"

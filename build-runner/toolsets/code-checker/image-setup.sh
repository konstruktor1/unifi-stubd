#!/bin/sh
set -eu

export DEBIAN_FRONTEND=noninteractive
js_source=/consumer/build-runner/toolsets/code-checker/javascript
js_root=/opt/unifi-stubd-code-checker-javascript
node_version=24.18.1

apt-get update
apt-get install -y --no-install-recommends ca-certificates curl jq python3 shellcheck

case "$(uname -m)" in
aarch64 | arm64)
  node_platform=arm64
  node_sha256=df224555a083b918e46260cc969838501b9f9a87140c1195e5b9597b56d5dae2
  ;;
x86_64 | amd64)
  node_platform=x64
  node_sha256=9f5eb6ac21845a66c493c91a253b1da32fd684e89e9b7202d4936982336be4ca
  ;;
*)
  echo "unsupported Node build architecture: $(uname -m)" >&2
  exit 2
  ;;
esac
node_directory="node-v${node_version}-linux-${node_platform}"
node_archive="/opt/${node_directory}.tar.gz"
curl -fsSL "https://nodejs.org/dist/v${node_version}/${node_directory}.tar.gz" -o "$node_archive"
printf '%s  %s\n' "$node_sha256" "$node_archive" | sha256sum -c -
tar -xzf "$node_archive" -C /opt
rm -f "$node_archive"
ln -s "$node_directory" /opt/node
export PATH="/opt/node/bin:$PATH"

install -d "$js_root/bin"
cp \
  "$js_source/package.json" \
  "$js_source/package-lock.json" \
  "$js_source/eslint.config.mjs" \
  "$js_root/"
npm --prefix "$js_root" ci \
  --include=dev \
  --engine-strict \
  --ignore-scripts \
  --no-audit \
  --no-fund

install -m 0755 "$js_source/eslint.sh" "$js_root/bin/eslint"
ln -s ../node_modules/.bin/prettier "$js_root/bin/prettier"

GOBIN=/usr/local/bin go install mvdan.cc/sh/v3/cmd/shfmt@v3.12.0
rm -rf /var/lib/apt/lists/*

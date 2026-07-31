#!/bin/sh
set -eu

export DEBIAN_FRONTEND=noninteractive
apt-get update
apt-get install -y --no-install-recommends python3 shellcheck
GOBIN=/usr/local/bin go install mvdan.cc/sh/v3/cmd/shfmt@v3.12.0
rm -rf /var/lib/apt/lists/*

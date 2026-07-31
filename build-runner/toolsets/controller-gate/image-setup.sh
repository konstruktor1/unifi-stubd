#!/bin/sh
set -eu

docker_cli_version='5:29.6.2-1~debian.12~bookworm'
docker_compose_version='5.3.1-1~debian.12~bookworm'
docker_buildx_version='0.35.0-1~debian.12~bookworm'

apt-get update
apt-get install -y --no-install-recommends ca-certificates curl gnupg jq
install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/debian/gpg \
  -o /etc/apt/keyrings/docker.asc
chmod a+r /etc/apt/keyrings/docker.asc

# shellcheck disable=SC1091
. /etc/os-release
architecture="$(dpkg --print-architecture)"
printf '%s\n' \
  "deb [arch=$architecture signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/debian $VERSION_CODENAME stable" \
  >/etc/apt/sources.list.d/docker.list

apt-get update
apt-get install -y --no-install-recommends \
  "docker-ce-cli=$docker_cli_version" \
  "docker-compose-plugin=$docker_compose_version" \
  "docker-buildx-plugin=$docker_buildx_version"
rm -rf /var/lib/apt/lists/*

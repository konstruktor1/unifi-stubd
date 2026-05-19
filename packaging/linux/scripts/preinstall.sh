#!/bin/sh
set -eu

if ! getent group unifi-stubd >/dev/null 2>&1; then
  groupadd --system unifi-stubd
fi

if ! id -u unifi-stubd >/dev/null 2>&1; then
  shell=/usr/sbin/nologin
  if [ ! -x "$shell" ]; then
    shell=/sbin/nologin
  fi
  if [ ! -x "$shell" ]; then
    shell=/bin/false
  fi
  useradd \
    --system \
    --gid unifi-stubd \
    --home-dir /var/lib/unifi-stubd \
    --shell "$shell" \
    --comment "unifi-stubd service user" \
    unifi-stubd
fi

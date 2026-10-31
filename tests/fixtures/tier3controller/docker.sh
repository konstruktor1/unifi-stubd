#!/bin/sh
# Record forbidden Docker access without contacting a daemon or starting a lab.
set -eu
printf '%s\n' "$*" >>"$UNIFI_STUB_TEST_DOCKER_LOG"
case "$1:$2" in
info:*) uname -m ;;
image:ls | network:ls | volume:ls | ps:* | compose:*) ;;
*) exit 97 ;;
esac

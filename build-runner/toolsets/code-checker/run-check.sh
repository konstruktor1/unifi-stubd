#!/bin/sh
set -eu

fail() {
  printf '%s\n' "$1" >&2
  exit 2
}

[ "$#" -eq 2 ] || fail "usage: run-check.sh <bootstrap-launcher> <consumer-project>"
bootstrap=$1
consumer_project=$2

config=$consumer_project/code-checker.toml
[ -f "$config" ] || fail "code-checker.toml is missing from the reported consumer project"
make -C "$consumer_project" check vulncheck
exec "$bootstrap" --config "$config" check "$consumer_project" \
  --use-case cli --require-tools

#!/usr/bin/env bash
# Prepare one GitHub job's native bindings; provider functions keep their owners.
set -euo pipefail

: "${PROVIDER_CONSUMER_ROOT:?Inherit the native GitHub project root}"
: "${PROVIDER_CONSUMER_CONTEXT:?Inherit the native GitHub consumer context}"
: "${PYTHON:?Use the interpreter selected by the native setup action}"
: "${BUILD_RUNNER_SOURCE_URL:?Configure the current Build Runner source}"
: "${GITHUB_OUTPUT:?Use the native Actions output channel}"
: "${GITHUB_ENV:?Use the native Actions environment channel}"
test "${GITHUB_ACTIONS:-}" = true
test "$PROVIDER_CONSUMER_ROOT" = "$(pwd -P)"
test "$PROVIDER_CONSUMER_ROOT" = "$(realpath "$GITHUB_WORKSPACE")"
test "$PROVIDER_CONSUMER_ROOT" = "$(git rev-parse --show-toplevel)"
test -x "$PYTHON"
umask 077

for directory in \
  .consumer-data \
  .consumer-data/dependencies \
  .consumer-data/dependencies/ci-providers \
  .consumer-data/dependencies/unifi-stubd \
  .consumer-data/dependencies/unifi-stubd/pip-cache \
  .provider-provider \
  .provider-provider/runtime \
  .provider-provider/runtime/unifi-stubd \
  .provider-provider/runtime/unifi-stubd/ci-setup \
  .provider-provider/runtime/unifi-stubd/ci-setup/tmp; do
  test ! -L "$directory"
  if test -e "$directory"; then
    test -d "$directory"
    test "$(realpath "$directory")" = "$PROVIDER_CONSUMER_ROOT/$directory"
  else
    mkdir -m 700 "$directory"
  fi
done
export TMPDIR="$PROVIDER_CONSUMER_ROOT/.provider-provider/runtime/unifi-stubd/ci-setup/tmp"
export PIP_CACHE_DIR="$PROVIDER_CONSUMER_ROOT/.consumer-data/dependencies/unifi-stubd/pip-cache"
export PIP_DISABLE_PIP_VERSION_CHECK=1

prepare_source() {
  local provider="$1" source_url="$2" source_directory authorization
  case "$provider" in build-runner | ct) ;; *) return 2 ;; esac
  if [[ ! "$source_url" =~ ^https://[^/@?#[:space:]]+/[^@?#[:space:]]+$ ]]; then
    printf '%s\n' 'Provider sources require an HTTPS URL without embedded credentials' >&2
    return 2
  fi
  source_directory="$PROVIDER_CONSUMER_ROOT/.consumer-data/dependencies/ci-providers/$provider"
  if test -e "$source_directory" || test -L "$source_directory"; then
    printf 'Preserve existing %s source; reconcile the prior setup before retrying\n' "$provider" >&2
    return 2
  fi
  (
    # Git materializes its declared modes inside the private dependency parent.
    umask 022
    if test -n "${PROVIDER_SOURCE_TOKEN:-}"; then
      : "${PROVIDER_SOURCE_USERNAME:?Configure the Forgejo read account}"
      authorization="$(printf '%s:%s' "$PROVIDER_SOURCE_USERNAME" "$PROVIDER_SOURCE_TOKEN" | base64 | tr -d '\n')"
      printf '::add-mask::%s\n' "$authorization"
      GIT_TERMINAL_PROMPT=0 \
        GIT_CONFIG_COUNT=1 \
        GIT_CONFIG_KEY_0="http.$source_url.extraheader" \
        GIT_CONFIG_VALUE_0="Authorization: Basic $authorization" \
        git -c credential.helper= clone --depth=1 --no-tags -- "$source_url" "$source_directory"
    else
      GIT_TERMINAL_PROMPT=0 git -c credential.helper= clone --depth=1 --no-tags -- "$source_url" "$source_directory"
    fi
  )
  test "$(realpath "$source_directory")" = "$source_directory"
  test -x "$source_directory/bin/provider"
  test "$(realpath "$source_directory/bin/provider")" = "$source_directory/bin/provider"
  printf '%s-launcher=%s/bin/provider\n' "$provider" "$source_directory" >>"$GITHUB_OUTPUT"
}

prepare_source build-runner "$BUILD_RUNNER_SOURCE_URL"
if test -n "${CT_SOURCE_URL:-}"; then
  prepare_source ct "$CT_SOURCE_URL"
fi

test ! -e .consumer-data/dependencies/unifi-stubd/mcp
test ! -L .consumer-data/dependencies/unifi-stubd/mcp
"$PYTHON" -m venv .consumer-data/dependencies/unifi-stubd/mcp
.consumer-data/dependencies/unifi-stubd/mcp/bin/python -m pip install -r requirements/unifi-stubd/mcp.txt
printf 'PYTHON=%s\n' "$PYTHON" >>"$GITHUB_ENV"

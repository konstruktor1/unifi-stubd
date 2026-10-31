#!/bin/bash

# Native Core context shared by the single-lane and matrix qualification owners.
qualification_context_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)"
repo_root="$(cd "$qualification_context_dir/../../../.." && pwd -P)"
[[ "${PROVIDER_CONSUMER_ROOT:-}" == "$repo_root" && -n "${PROVIDER_CONSUMER_CONTEXT:-}" ]] || die 'inherit the native Core consumer context'
[[ -z "${UNIFI_STUBD_CORE_ROOT+set}" ]] || die 'UNIFI_STUBD_CORE_ROOT is obsolete; inherit the native Core context'
cd "$repo_root" || die 'cannot enter the native Core root'
load_qualification_binding() {
  local binding="$repo_root/.consumer-data/connections/build-runner/management.json" management_entry entry
  [[ -f "$binding" && ! -L "$binding" && "$(cd "$(dirname "$binding")" && pwd -P)/management.json" == "$binding" ]] || die 'native Build Runner binding must be physical'
  jq -e --arg root "$repo_root" --arg context "$PROVIDER_CONSUMER_CONTEXT" --arg python "${PYTHON:-}" '
  .cwd == $root and .args == ["mcp"] and
  .env.PROVIDER_CONSUMER_ROOT == $root and .env.PROVIDER_CONSUMER_CONTEXT == $context and
  (.env.PYTHON | type == "string" and length > 0) and .env.PYTHON == $python
  ' "$binding" >/dev/null || die 'native Build Runner binding does not match this Core context/interpreter'
  management_entry="$(jq -er '.command | select(startswith("/") and endswith("/bin/provider"))' "$binding")"
  test_runner_entry="${management_entry%/provider}/test-runner"
  monitor_entry="${management_entry%/provider}/runner-monitor"
  for entry in "$management_entry" "$test_runner_entry" "$monitor_entry" "$PYTHON"; do
    [[ -f "$entry" && -x "$entry" && ! -L "$entry" && "$(cd "$(dirname "$entry")" && pwd -P)/${entry##*/}" == "$entry" ]] || die 'native Build Runner entry must be a physical executable'
  done
}

# Resolve the nearest existing ancestor before creating any machine/run state.
physical_directory() {
  local directory="$1"
  [[ "$directory" == /* && "$directory" != *'/../'* && "$directory" != *'/./'* ]] || die 'qualification directories must be canonical'
  while [[ ! -e "$directory" ]]; do
    [[ ! -L "$directory" ]] || die 'qualification directories must not be symlinks'
    directory="${directory%/*}"
    [[ -n "$directory" ]] || directory=/
  done
  [[ -d "$directory" && ! -L "$directory" && "$(cd "$directory" && pwd -P)" == "$directory" ]] || die 'qualification directories must have physical ancestors'
}

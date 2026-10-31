#!/bin/bash
set -euo pipefail
umask 077

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd -- "$script_dir/../../.." && pwd)"

fail() {
  printf 'tier3 controller: %s\n' "$*" >&2
  exit 2
}
# Waiting on an owned background child lets Bash dispatch a signal immediately.
# The trap reaps that client before removing its disposable Compose resources.
docker_pid=
docker() {
  local status
  command docker "$@" <&0 &
  docker_pid=$!
  if wait "$docker_pid"; then status=0; else status=$?; fi
  docker_pid=
  return "$status"
}
stop_docker_client() {
  local elapsed=0
  [[ -n "$docker_pid" ]] || return 0
  kill -TERM "$docker_pid" 2>/dev/null || true
  while kill -0 "$docker_pid" 2>/dev/null && [[ "$elapsed" -lt 10 ]]; do
    sleep 1
    elapsed=$((elapsed + 1))
  done
  if kill -0 "$docker_pid" 2>/dev/null; then kill -KILL "$docker_pid" 2>/dev/null || true; fi
  wait "$docker_pid" 2>/dev/null || true
  docker_pid=
}
get() {
  awk -F= -v key="$1" '$1 == key { count++; value=substr($0,length($1)+2) } END { if (count != 1) exit 1; print value }' "$manifest"
}
compose() {
  if [[ "$schema" == 2 ]]; then
    docker compose -f "$repo_root/lab/stub/compose.tier3-controller-helpers.yaml" "$@"
    return "$?"
  fi
  docker compose -f "$repo_root/lab/stub/compose.yaml" \
    -f "$repo_root/lab/stub/compose.tests.yaml" \
    -f "$repo_root/lab/stub/compose.controller-channel.yaml" \
    -f "$repo_root/lab/stub/compose.tier3-controller.yaml" "$@"
}
controller_api() {
  if [[ "$schema" == 2 ]]; then
    sidecars_match || fail 'controller sidecars changed during the run'
    compose run -T --rm --no-deps -e UNIFI_STUB_CONTROLLER_AUTH_STDIN=1 \
      controller-adoption "$@" --controller-mode unifi-os \
      --controller-url "$api_origin" --auth-file /run/unifi-stubd-controller/auth.json <"$auth_file"
    return
  fi
  compose run -T --rm --no-deps controller-adoption "$@"
}
sidecars_match() {
  [[ "$schema" == 2 ]] || return 0
  [[ -f "$auth_file" && ! -L "$auth_file" && -f "$provenance_file" && ! -L "$provenance_file" ]] || return 1
  [[ "$(sha256sum "$auth_file" | awk '{print $1}')" == "$(get CONTROLLER_AUTH_SHA256)" &&
  "$(sha256sum "$provenance_file" | awk '{print $1}')" == "$(get CONTROLLER_PROVENANCE_SHA256)" ]]
}
write_binding() {
  cat "$manifest"
  printf 'MANIFEST_SHA256=%s\nCONTROLLER_ACTUAL_VERSION=%s\n' "$manifest_sha" "$actual_version"
  if [[ "$schema" == 1 ]]; then printf 'CONTROLLER_RUNTIME_BASE_DIGEST=%s\n' "$runtime_digest"; fi
}
project_is_clean() {
  local label="label=com.docker.compose.project=$COMPOSE_PROJECT_NAME" containers networks volumes
  containers="$(docker ps -aq --filter "$label")" || {
    printf 'cleanup: cannot inspect owned containers\n' >&2
    return 1
  }
  networks="$(docker network ls -q --filter "$label")" || {
    printf 'cleanup: cannot inspect owned networks\n' >&2
    return 1
  }
  volumes="$(docker volume ls -q --filter "$label")" || {
    printf 'cleanup: cannot inspect owned volumes\n' >&2
    return 1
  }
  if [[ -n "$containers" || -n "$networks" || -n "$volumes" ]]; then
    printf 'cleanup: remaining containers=%s networks=%s volumes=%s\n' "$containers" "$networks" "$volumes" >&2
    return 1
  fi
}
remove_resources() {
  local failed=0 alias image_ids
  compose down --volumes --remove-orphans || {
    printf 'cleanup: Compose down failed\n' >&2
    failed=1
  }
  for alias in "${owned_images[@]}"; do
    if ! image_ids="$(docker image ls --quiet --no-trunc "$alias")"; then
      failed=1
      continue
    fi
    if [[ -n "$image_ids" ]]; then
      docker image rm "$alias" >/dev/null || failed=1
    fi
    if ! image_ids="$(docker image ls --quiet --no-trunc "$alias")"; then
      failed=1
      continue
    fi
    [[ -z "$image_ids" ]] || {
      printf 'cleanup: remaining owned image alias=%s\n' "$alias" >&2
      failed=1
    }
  done
  project_is_clean || failed=1
  return "$failed"
}
finish() {
  local status=$? cleanup=passed
  trap - EXIT HUP INT TERM
  set +e
  stop_docker_client
  sidecars_match || status=2
  if [[ "$status" -ne 0 && "$schema" == 1 ]]; then compose logs --no-color --tail 100 unifi unifi-db >"$run_dir/controller.log" 2>&1; fi
  remove_resources >"$run_dir/cleanup.log" 2>&1 || cleanup=failed
  [[ "$cleanup" == passed ]] || status=2
  [[ "$qualified" == yes || "$status" -ne 0 ]] || status=1
  {
    write_binding
    printf 'RESULT=%s\nADOPTED=%s\nCONNECTED=%s\nTELEMETRY=%s\nCLEANUP=%s\n' \
      "$([[ "$status" -eq 0 && "$qualified" == yes ]] && printf passed || printf failed)" \
      "$adopted" "$connected" "$telemetry" "$cleanup"
    if [[ "$schema" == 2 ]]; then printf 'CONTROLLER_CLEANUP=external-owner-required\n'; fi
  } >"$run_dir/result.env.incoming"
  mv "$run_dir/result.env.incoming" "$run_dir/result.env" || status=2
  printf 'tier3 controller: cleanup=%s evidence=%s\n' "$cleanup" "$run_dir"
  exit "$status"
}

[[ "$repo_root" == /workspace ]] || fail 'run through the Core Make entry'
[[ $# -eq 0 || ($# -eq 1 && "$1" == --cleanup) ]] || fail 'use no arguments or --cleanup'
relative="${UNIFI_STUB_TIER3_GATE_MANIFEST:?set UNIFI_STUB_TIER3_GATE_MANIFEST}"
[[ "$relative" =~ ^dist/tier3-controller-input/[a-z0-9][a-z0-9_-]{0,31}\.env$ ]] || fail 'unsafe input manifest path'
manifest="$repo_root/$relative"
run_id="${relative##*/}"
run_id="${run_id%.env}"
run_dir="$repo_root/dist/tier3-controller/$run_id"
for directory in "$repo_root/dist" "$repo_root/dist/tier3-controller-input" "$repo_root/dist/tier3-controller" "$run_dir"; do
  [[ ! -L "$directory" && "$(realpath -m -- "$directory")" == "$directory" && (! -e "$directory" || -d "$directory") ]] ||
    fail 'controller inputs and results must use real Core directories'
done
[[ -f "$manifest" && ! -L "$manifest" ]] || fail 'input manifest must be a regular file'
schema="$(get SCHEMA_VERSION)"
case "$schema" in 1 | 2) ;; *) fail 'unsupported manifest schema' ;; esac
keys="$(awk -v schema="$schema" '$1 == "common" || $1 == schema { print $2 }' "$repo_root/lab/stub/configs/tier3-controller-fields.txt" | paste -sd ' ' -)"
awk -F= -v keys="$keys" '
  BEGIN { split(keys,list," "); for (i in list) allowed[list[i]]=1 }
  !/^[A-Z][A-Z0-9_]*=[^[:space:][:cntrl:]]+$/ || !($1 in allowed) || seen[$1]++ { exit 1 }
  END { for (key in allowed) if (seen[key] != 1) exit 1 }
' "$manifest" || fail 'invalid input manifest'
[[ "$(get RUN_ID)" == "$run_id" ]] || fail 'run ID does not bind input path'
channel="$(get CONTROLLER_CHANNEL)" stage="$(get CONTROLLER_RELEASE_STAGE)"
case "$channel:$stage" in stable:GA | release-candidate:RC) ;; *) fail 'invalid release channel' ;; esac
selector="$(get CONTROLLER_SELECTOR)"
case "$selector" in latest | cancellation-fixture) ;; *) fail 'controller selector must be latest or cancellation-fixture' ;; esac
[[ "$schema" == 1 || "$selector" == latest ]] || fail 'schema 2 requires the latest selector'
expected_version="$(get CONTROLLER_EXPECTED_VERSION)"
[[ "$expected_version" =~ ^[0-9]+([.][0-9]+){1,3}$ ]] || fail 'invalid expected version'
release_id="$(get CONTROLLER_RELEASE_ID)"
[[ "$release_id" =~ ^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$ ]] || fail 'invalid release ID'
[[ "$(get CONTROLLER_RELEASE_URL)" == "https://community.ui.com/releases/UniFi-Network-Application-${expected_version//./-}/$release_id" ]] || fail 'release URL does not bind version'
zip_url='' zip_sha='' runtime_selector='' api_origin='' auth_file=''
if [[ "$schema" == 1 ]]; then
  zip_url="$(get CONTROLLER_ZIP_URL)" zip_sha="$(get CONTROLLER_ZIP_SHA256)"
  [[ "$zip_url" =~ ^https://dl\.ui\.com/unifi/[A-Za-z0-9._-]+/UniFi\.unix\.zip$ && "$zip_sha" =~ ^[0-9a-f]{64}$ ]] || fail 'invalid official ZIP binding'
  zip_token="${zip_url#https://dl.ui.com/unifi/}"
  zip_token="${zip_token%/UniFi.unix.zip}"
  case "$zip_token" in "$expected_version" | "$expected_version"-*) ;; *) fail 'ZIP URL does not bind version' ;; esac
  runtime_selector="$(get CONTROLLER_RUNTIME_BASE_SELECTOR)"
  [[ "$runtime_selector" == lscr.io/linuxserver/unifi-network-application:latest ]] || fail 'unexpected runtime base'
else
  [[ ! -L "$(dirname "$manifest")" && "$(stat -c %a "$(dirname "$manifest")")" == 700 ]] || fail 'schema 2 input directory must be private and real (mode 0700)'
  api_origin="$(get CONTROLLER_API_ORIGIN)"
  [[ "$api_origin" =~ ^https://host\.docker\.internal:([1-9][0-9]{0,4})$ && "${BASH_REMATCH[1]}" -le 65535 ]] || fail 'schema 2 needs an isolated host.docker.internal HTTPS endpoint'
  inform_port="$(get CONTROLLER_INFORM_PORT)"
  [[ "$inform_port" =~ ^[1-9][0-9]{0,4}$ && "$inform_port" -le 65535 ]] || fail 'invalid companion inform port'
  auth_file="${manifest%.env}.auth.json"
  provenance_file="${manifest%.env}.controller.json"
  for sidecar in "$auth_file" "$provenance_file"; do
    [[ -f "$sidecar" && ! -L "$sidecar" && "$(stat -c %a "$sidecar")" == 600 ]] || fail 'sidecars must be private regular files'
  done
  [[ "$(wc -c <"$auth_file")" -le 65536 && "$(wc -c <"$provenance_file")" -le 4194304 ]] || fail 'oversized sidecar'
  for key in CONTROLLER_AUTH_SHA256 CONTROLLER_PROVENANCE_SHA256; do
    [[ "$(get "$key")" =~ ^[0-9a-f]{64}$ ]] || fail "invalid $key"
  done
  [[ "$(sha256sum "$auth_file" | awk '{print $1}')" == "$(get CONTROLLER_AUTH_SHA256)" ]] || fail 'authentication digest mismatch'
  [[ "$(sha256sum "$provenance_file" | awk '{print $1}')" == "$(get CONTROLLER_PROVENANCE_SHA256)" ]] || fail 'controller provenance digest mismatch'
  jq -se --arg run "$run_id" --arg channel "$channel" --arg version "$expected_version" '
    length == 1 and (.[0] | type == "object" and .schema_version == 1 and .run_id == $run and
    (.source | .owner == "unifi-stubd-labs" and .schema_version == 1 and
      .controller_channel == $channel and .controller_version == $version and .fresh_controller == true and
      all(.controller_instance, .machine_id; type == "string" and length > 0) and
      all(.testhost_manifest_sha256, .controller_image_sha256, .controller_package_sha256; type == "string" and test("^[0-9a-f]{64}$"))))
  ' "$provenance_file" >/dev/null || fail 'invalid per-run controller provenance'
fi
[[ "$(get CONTROLLER_RELEASE_QUERY_ENDPOINT)" == https://community.svc.ui.com/graphql ]] || fail 'unexpected release endpoint'
for key in CONTROLLER_RELEASE_QUERY_SHA256 CONTROLLER_RELEASE_RESPONSE_SHA256 CANDIDATE_SHA256; do
  [[ "$(get "$key")" =~ ^[0-9a-f]{64}$ ]] || fail "invalid $key"
done
mac="$(get CANDIDATE_MAC)" candidate_ip="$(get CANDIDATE_IP)" expected_speed="$(get EXPECTED_UPLINK_SPEED_MBPS)"
[[ "$mac" =~ ^02:([0-9a-f]{2}:){4}[0-9a-f]{2}$ ]] || fail 'candidate must use a local MAC'
[[ "$candidate_ip" =~ ^[0-9]{1,3}([.][0-9]{1,3}){3}$ && "$expected_speed" =~ ^[1-9][0-9]*$ ]] || fail 'invalid guest IP or observed uplink speed'
manifest_sha="$(sha256sum "$manifest" | awk '{print $1}')"
export COMPOSE_PROJECT_NAME="unifi-stubd-t3-$run_id"
controller_image="unifi-stubd-tier3-controller:$run_id"
helper_image="unifi-stubd-controller-adoption:t3-$run_id"
ui_image="unifi-stubd-controller-ui:t3-$run_id"
owned_images=("$helper_image" "$ui_image")
if [[ "$schema" == 1 ]]; then owned_images+=("$controller_image"); fi
export UNIFI_NETWORK_IMAGE="$controller_image" UNIFI_STUB_CONTROLLER_IMAGE="$controller_image"
export UNIFI_STUB_CONTROLLER_HELPER_IMAGE="$helper_image" UNIFI_STUB_CONTROLLER_UI_IMAGE="$ui_image"
export UNIFI_STUB_CONTROLLER_ZIP_URL="$zip_url" UNIFI_STUB_CONTROLLER_ZIP_SHA256="$zip_sha"
export UNIFI_STUB_CONTROLLER_RUNTIME_BASE_DIGEST="$runtime_selector"
export UNIFI_STUB_BRIDGE_MAC="$mac" UNIFI_STUB_BRIDGE_IP="$candidate_ip"
actual_version=not-run runtime_digest=not-run qualified=no adopted=false connected=false telemetry=false

# The orchestration container and the receiving daemon must agree on the native
# architecture; this gate has no implicit emulation fallback.
native_arch="$(uname -m)"
case "$native_arch" in aarch64 | arm64) native_arch=arm64 ;; x86_64 | amd64) native_arch=amd64 ;; *) fail 'unsupported native architecture' ;; esac
daemon_arch="$(docker info --format '{{.Architecture}}')"
case "$daemon_arch" in aarch64 | arm64) daemon_arch=arm64 ;; x86_64 | amd64) daemon_arch=amd64 ;; esac
[[ "$native_arch" == "$daemon_arch" ]] || fail 'Docker daemon architecture differs from this invocation'

if [[ "${1:-}" == --cleanup ]]; then
  [[ -f "$run_dir/owner.sha256" && ! -L "$run_dir/owner.sha256" && "$(cat "$run_dir/owner.sha256")" == "$manifest_sha" ]] || fail 'cleanup has no matching run ownership'
  remove_resources
  exit 0
fi
mkdir -p "$repo_root/dist/tier3-controller"
mkdir "$run_dir" || fail 'run ID already has state'
for alias in "${owned_images[@]}"; do
  image_ids="$(docker image ls --quiet --no-trunc "$alias")" || fail "cannot inspect image alias: $alias"
  [[ -z "$image_ids" ]] || fail "image alias already exists: $alias"
done
project_is_clean || fail 'Compose project already exists'
printf '%s\n' "$manifest_sha" >"$run_dir/owner.sha256"
trap finish EXIT
trap 'exit 129' HUP
trap 'exit 130' INT
trap 'exit 143' TERM

if [[ "$schema" == 1 ]]; then
  docker pull "$runtime_selector"
  [[ "$(docker image inspect --format '{{.Os}}/{{.Architecture}}' "$runtime_selector")" == "linux/$native_arch" ]] || fail 'controller runtime is not native Linux'
  runtime_digest="$(docker image inspect --format '{{range .RepoDigests}}{{println .}}{{end}}' "$runtime_selector" | awk '/^lscr.io\/linuxserver\/unifi-network-application@sha256:[0-9a-f]+$/ { count++; value=$0 } END { if (count != 1) exit 1; print value }')"
  export UNIFI_STUB_CONTROLLER_RUNTIME_BASE_DIGEST="$runtime_digest"
  compose build unifi controller-adoption controller-ui
  compose up -d unifi
else
  compose build controller-adoption controller-ui
fi
deadline=$(($(date +%s) + 300))
while ! controller_api version --expected-version "$expected_version" >"$run_dir/controller-version.log" 2>&1; do
  [[ "$(date +%s)" -lt "$deadline" ]] || fail 'controller did not become ready'
  sleep 3
done
actual_version="$expected_version"
if [[ "$schema" == 1 ]]; then controller_api bootstrap; fi
# The guest receives only 127.0.0.1:8080 through authenticated SSH forwarding.
controller_api loopback-inform
if [[ "$schema" == 1 ]]; then
  endpoint="$(compose port unifi 8080)"
  [[ "$endpoint" =~ ^127\.0\.0\.1:([0-9]{1,5})$ ]] || fail 'inform listener is not host loopback'
  inform_port="${BASH_REMATCH[1]}"
  [[ "$inform_port" -ge 1 && "$inform_port" -le 65535 ]] || fail 'invalid inform port'
fi
{
  write_binding
  printf 'INFORM_HOST=127.0.0.1\nINFORM_PORT=%s\nGUEST_INFORM_PORT=8080\n' "$inform_port"
} >"$run_dir/ready.env.incoming"
mv "$run_dir/ready.env.incoming" "$run_dir/ready.env"
controller_api wait-visible --mac "$mac" --timeout 300
[[ "$selector" != cancellation-fixture ]] || fail 'cancellation fixture cannot adopt or qualify a candidate'
controller_api adopt --mac "$mac"
adopted=true
controller_api wait-connected --mac "$mac" --timeout 300 >"$run_dir/controller-connected.env"
connected=true
ui_options=(-e "UNIFI_STUB_UI_DEVICE_IP=$candidate_ip" -e "UNIFI_STUB_UI_SPEED_MBPS=$expected_speed")
if [[ "$schema" == 2 ]]; then
  compose run -T --rm --no-deps "${ui_options[@]}" \
    -e UNIFI_STUB_CONTROLLER_AUTH_STDIN=1 -e UNIFI_STUB_CONTROLLER_MODE=unifi-os \
    -e "UNIFI_STUB_CONTROLLER_URL=$api_origin" controller-ui <"$auth_file" >"$run_dir/controller-ui.log" 2>&1
else
  compose run -T --rm --no-deps "${ui_options[@]}" controller-ui >"$run_dir/controller-ui.log" 2>&1
fi
telemetry=true
{
  write_binding
  printf 'ADOPTED=true\nCONNECTED=true\nTELEMETRY=true\n'
} >"$run_dir/accepted.env.incoming"
mv "$run_dir/accepted.env.incoming" "$run_dir/accepted.env"
# The Core qualifier proves the live guest exchange before the gate removes its
# Compose resources. A disconnected qualifier cannot leave this project running.
deadline=$(($(date +%s) + 60))
while [[ ! -f "$run_dir/guest-complete.sha256" ]]; do
  [[ "$(date +%s)" -lt "$deadline" ]] || fail 'qualifier did not acknowledge live guest evidence'
  sleep 1
done
[[ ! -L "$run_dir/guest-complete.sha256" && "$(cat "$run_dir/guest-complete.sha256")" == "$manifest_sha" ]] || fail 'qualifier completion has a different binding'
qualified=yes

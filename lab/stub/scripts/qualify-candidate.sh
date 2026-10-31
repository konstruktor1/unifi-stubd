#!/bin/bash

set -euo pipefail
umask 077

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)"
# shellcheck source=lab/stub/scripts/lib/manifest.sh
source "$script_dir/lib/manifest.sh"

die() {
  printf 'tier3-candidate: %s\n' "$*" >&2
  exit 1
}
need() { command -v "$1" >/dev/null 2>&1 || die "missing command: $1"; }
lower_ascii() { printf '%s' "$1" | LC_ALL=C tr '[:upper:]' '[:lower:]'; }

# shellcheck source=lab/stub/scripts/lib/qualification-context.sh
source "$script_dir/lib/qualification-context.sh"

scenario_manifest="$script_dir/../configs/qualification.env"
scenario_keys='SCHEMA_VERSION PROFILE GUEST_USER GUEST_FORWARD_HOST CORE_TARGET CORE_CLEANUP_TARGET CORE_INPUT_RELATIVE_DIR GUEST_ROOT_RELATIVE_PATH GUEST_STAGE_RELATIVE_PATH SERVICE_NAME FIXTURE_LINKS READY_TIMEOUT_SECONDS RESULT_TIMEOUT_SECONDS CORE_TERMINATE_TIMEOUT_SECONDS SSH_CONNECT_TIMEOUT_SECONDS PRODUCT_INTERRUPT_TIMEOUT_SECONDS PRODUCT_TERMINATE_TIMEOUT_SECONDS TRAFFIC_INGRESS_TARGET TRAFFIC_EGRESS_TARGET'
manifest_validate "$scenario_manifest" "$scenario_keys" || die "invalid scenario manifest: $scenario_manifest"
[[ "$(manifest_get "$scenario_manifest" SCHEMA_VERSION)" == 1 ]] || die 'unsupported scenario manifest schema'

profile="$(manifest_get "$scenario_manifest" PROFILE)"
guest_user="$(manifest_get "$scenario_manifest" GUEST_USER)"
guest_forward_host="$(manifest_get "$scenario_manifest" GUEST_FORWARD_HOST)"
core_target="$(manifest_get "$scenario_manifest" CORE_TARGET)"
core_cleanup_target="$(manifest_get "$scenario_manifest" CORE_CLEANUP_TARGET)"
core_input_relative_dir="$(manifest_get "$scenario_manifest" CORE_INPUT_RELATIVE_DIR)"
guest_root_relative="$(manifest_get "$scenario_manifest" GUEST_ROOT_RELATIVE_PATH)"
guest_stage_relative="$(manifest_get "$scenario_manifest" GUEST_STAGE_RELATIVE_PATH)"
service_name="$(manifest_get "$scenario_manifest" SERVICE_NAME)"
fixture_links="$(manifest_get "$scenario_manifest" FIXTURE_LINKS)"
ready_timeout="$(manifest_get "$scenario_manifest" READY_TIMEOUT_SECONDS)"
result_timeout="$(manifest_get "$scenario_manifest" RESULT_TIMEOUT_SECONDS)"
core_terminate_timeout="$(manifest_get "$scenario_manifest" CORE_TERMINATE_TIMEOUT_SECONDS)"
ssh_connect_timeout="$(manifest_get "$scenario_manifest" SSH_CONNECT_TIMEOUT_SECONDS)"
product_interrupt_timeout="$(manifest_get "$scenario_manifest" PRODUCT_INTERRUPT_TIMEOUT_SECONDS)"
product_terminate_timeout="$(manifest_get "$scenario_manifest" PRODUCT_TERMINATE_TIMEOUT_SECONDS)"
traffic_ingress_target="$(manifest_get "$scenario_manifest" TRAFFIC_INGRESS_TARGET)"
traffic_egress_target="$(manifest_get "$scenario_manifest" TRAFFIC_EGRESS_TARGET)"

env_value() {
  awk -F= -v key="$2" '$1 == key { count++; value = substr($0, length($1) + 2) } END { if (count != 1) exit 1; print value }' "$1"
}

validate_env() {
  local file="$1" keys="$2"
  [[ -s "$file" ]] || die "missing result file: $file"
  awk -F= -v keys="$keys" '
    BEGIN { split(keys, list, " "); for (i in list) allowed[list[i]] = 1 }
    !/^[A-Z][A-Z0-9_]*=[^[:space:][:cntrl:]]+$/ || !($1 in allowed) || seen[$1]++ { exit 1 }
    END { for (key in allowed) if (seen[key] != 1) exit 1 }
  ' "$file" || die "unexpected result contract: $file"
}

private_file() {
  local file="$1" limit="$2" mode
  [[ "$file" == /* && -f "$file" && ! -L "$file" ]] || die 'controller sidecar must be an absolute regular file'
  physical_directory "$(dirname "$file")"
  if ! mode="$(stat -c %a "$file" 2>/dev/null)"; then mode="$(stat -f %Lp "$file")"; fi
  [[ "$mode" == 600 && "$(wc -c <"$file")" -le "$limit" ]] || die 'controller sidecar must be mode 0600 and within its size limit'
}

# The caller supplies the remote command; its argument vector expands here.
# shellcheck disable=SC2029
ssh_guest() { ssh "${ssh_options[@]}" "$guest_user@$guest_forward_host" "$@"; }

wait_for_file() {
  local file="$1" timeout="$2" elapsed=0 child_status
  while [[ "$elapsed" -lt "$timeout" ]]; do
    [[ -s "$file" ]] && return
    if [[ -n "${tunnel_pid:-}" ]] && ! kill -0 "$tunnel_pid" 2>/dev/null; then
      die 'owned inform tunnel exited before controller acceptance'
    fi
    if [[ -n "$core_pid" ]] && ! kill -0 "$core_pid" 2>/dev/null; then
      set +e
      wait "$core_pid"
      child_status=$?
      set -e
      core_pid=
      die "controller gate exited before $(basename "$file") (status $child_status)"
    fi
    sleep 1
    elapsed=$((elapsed + 1))
  done
  die "timed out waiting for $(basename "$file")"
}

wait_pid() {
  local pid="$1" timeout="$2" elapsed=0
  while kill -0 "$pid" 2>/dev/null && [[ "$elapsed" -lt "$timeout" ]]; do
    sleep 1
    elapsed=$((elapsed + 1))
  done
  ! kill -0 "$pid" 2>/dev/null
}

cleanup_controller() {
  local failed=0 child_status
  if [[ -n "$core_pid" ]]; then
    if kill -0 "$core_pid" 2>/dev/null; then
      kill -TERM "$core_pid" || failed=1
      wait_pid "$core_pid" "$core_terminate_timeout" || failed=1
    fi
    if ! kill -0 "$core_pid" 2>/dev/null; then
      set +e
      wait "$core_pid"
      child_status=$?
      set -e
      case "$child_status" in 0 | 143) ;; *) failed=1 ;; esac
      core_pid=
    fi
  fi
  if [[ "$core_started" == yes && "$core_completed" != yes && -z "$core_pid" ]]; then
    (
      cd "$repo_root"
      env UNIFI_STUB_TIER3_GATE_MANIFEST="$core_manifest_relative" \
        make "$core_cleanup_target"
    ) >>"$controller_log" 2>&1 || failed=1
  fi
  return "$failed"
}

cleanup_guest() {
  # Pass local run paths and cleanup limits as remote positional arguments.
  # shellcheck disable=SC2029
  ssh "${ssh_options[@]}" "$guest_user@$guest_forward_host" \
    "sudo sh -s -- '$guest_root' '$guest_stage' '$candidate_sha256' '$service_exists' '$service_was_running' '$guest_root_prefix' '$guest_stage_prefix' '$fixture_links' '$traffic_ingress_target' '$traffic_egress_target' '$service_name' '$product_interrupt_timeout' '$product_terminate_timeout'" <<'GUEST_CLEANUP'
set -eu
root=$1; stage=$2; expected=$3; service_exists=$4; service_was_running=$5
root_prefix=$6; stage_prefix=$7; fixture_links=$8; ingress_target=$9
shift 9
egress_target=$1; service_name=$2; interrupt_timeout=$3; terminate_timeout=$4
case "$root" in "$root_prefix"/*) ;; *) exit 41;; esac
case "$stage" in "$stage_prefix"-*) ;; *) exit 42;; esac

for item in "traffic-ingress:$ingress_target" "traffic-egress:$egress_target"; do
  name=${item%%:*}; target=${item#*:}; pid_file="$root/$name.pid"
  [ -f "$pid_file" ] || continue
  pid=$(cat "$pid_file"); case "$pid" in ''|*[!0-9]*) exit 43;; esac
  if kill -0 "$pid" 2>/dev/null; then
    command=$(tr '\000' ' ' <"/proc/$pid/cmdline")
    case "$command" in *ping*"$target"*) kill -TERM "$pid";; *) exit 44;; esac
  fi
done

if [ -f "$root/product.pid" ]; then
  pid=$(cat "$root/product.pid"); case "$pid" in ''|*[!0-9]*) exit 45;; esac
  if kill -0 "$pid" 2>/dev/null; then
    [ "$(readlink -f "/proc/$pid/exe")" = "$root/unifi-stubd" ] || exit 46
    [ "$(sha256sum "/proc/$pid/exe" | awk '{print $1}')" = "$expected" ] || exit 47
    kill -INT "$pid"
    count=0; while kill -0 "$pid" 2>/dev/null && [ "$count" -lt "$interrupt_timeout" ]; do sleep 1; count=$((count + 1)); done
    if kill -0 "$pid" 2>/dev/null; then
      kill -TERM "$pid"
      count=0; while kill -0 "$pid" 2>/dev/null && [ "$count" -lt "$terminate_timeout" ]; do sleep 1; count=$((count + 1)); done
    fi
    ! kill -0 "$pid" 2>/dev/null || exit 48
  fi
fi

if [ -f "$root/fixture.created" ]; then
  old_ifs=$IFS
  IFS=,
  for link in $fixture_links; do
    if ip link show "$link" >/dev/null 2>&1; then
      ip link delete "$link"
    fi
  done
  IFS=$old_ifs
fi
[ ! -L "$root" ] && [ ! -L "$stage" ] || exit 49
rm -rf "$root" "$stage"

if [ "$service_exists" = yes ]; then
  if [ "$service_was_running" = yes ]; then
    rc-service "$service_name" start >/dev/null
    rc-service "$service_name" status >/dev/null
  else
    ! rc-service "$service_name" status >/dev/null 2>&1
  fi
fi
GUEST_CLEANUP
}

contract_hash() {
  local relative file digest mode
  local files=()
  while IFS= read -r relative; do
    [[ "$relative" =~ ^[.A-Za-z0-9][A-Za-z0-9._/-]*$ && "$relative" != *'..'* ]] || die 'unsafe qualification contract path'
    file="$repo_root/$relative"
    physical_directory "$(dirname "$file")"
    [[ -f "$file" && ! -L "$file" ]] || die "missing qualification contract file: $relative"
    files+=("$relative")
  done <"$script_dir/../configs/qualification-contract-files.txt"
  [[ "${#files[@]}" -gt 0 ]] || die 'empty qualification contract'
  {
    for relative in "${files[@]}"; do
      file="$repo_root/$relative"
      digest="$(shasum -a 256 "$file" | awk '{print $1}')"
      if ! mode="$(stat -c %a "$file" 2>/dev/null)"; then mode="$(stat -f %Lp "$file")"; fi
      printf '%s %s %s\n' "$digest" "$mode" "$relative"
    done
  } | shasum -a 256 | awk '{print $1}'
}

on_exit() {
  local status=$? cleanup=passed
  trap - EXIT HUP INT TERM
  set +e
  if [[ "$guest_owned" == yes ]]; then cleanup_guest || cleanup=failed; fi
  if [[ -n "$tunnel_pid" ]]; then
    kill -TERM "$tunnel_pid" 2>/dev/null || true
    wait_pid "$tunnel_pid" 10 || kill -KILL "$tunnel_pid" 2>/dev/null || true
    wait "$tunnel_pid" 2>/dev/null || true
    tunnel_pid=
  fi
  if [[ "$guest_start_attempted" == yes ]]; then
    if test_runner stop >"$evidence_dir/test-runner-stop.log" 2>&1; then
      test_runner destroy >"$evidence_dir/test-runner-destroy.log" 2>&1 || cleanup=failed
    else
      cleanup=failed
      printf 'tier3-candidate: Test Runner identity/cleanup unresolved; retain runtime for its owner\n' >&2
    fi
  fi
  cleanup_controller || cleanup=failed
  if [[ "$controller_auth_written" == yes ]]; then
    [[ "$core_auth_path" == "$repo_root/dist/tier3-controller-input/$run_id.auth.json" &&
      -f "$core_auth_path" && ! -L "$core_auth_path" &&
      "$(shasum -a 256 "$core_auth_path" 2>/dev/null | awk '{print $1}')" == "$TIER3_CONTROLLER_AUTH_SHA256" ]] || cleanup=failed
  fi
  if [[ "$controller_sidecars_written" == yes ]]; then
    [[ "$(shasum -a 256 "$core_provenance_path" 2>/dev/null | awk '{print $1}')" == "$TIER3_CONTROLLER_PROVENANCE_SHA256" ]] || cleanup=failed
  fi
  if [[ "$core_manifest_sha256" != not-written ]]; then
    [[ "$(shasum -a 256 "$core_manifest_path" 2>/dev/null | awk '{print $1}')" == "$core_manifest_sha256" ]] || cleanup=failed
  fi
  if [[ "$runner_manifest_sha256" != not-written ]]; then
    [[ "$(shasum -a 256 "$runner_manifest_path" 2>/dev/null | awk '{print $1}')" == "$runner_manifest_sha256" ]] || cleanup=failed
  fi
  [[ "$(shasum -a 256 "$runner_template_path" 2>/dev/null | awk '{print $1}')" == "$runner_template_sha256" ]] || cleanup=failed
  [[ "$(contract_hash 2>/dev/null)" == "$initial_contract_hash" ]] || cleanup=failed
  if [[ "$cleanup" == passed && "$controller_auth_written" == yes ]]; then
    unlink "$core_auth_path" || cleanup=failed
  fi
  if [[ "$qualified" != yes || "$cleanup" != passed ]]; then status=1; fi
  {
    printf 'RESULT=%s\nRUN_ID=%s\nMAC=%s\nSHA256=%s\nVERSION=%s\n' \
      "$([[ "$status" -eq 0 ]] && printf passed || printf failed)" "$run_id" "$candidate_mac" "$candidate_sha256" "$candidate_version"
    printf 'CANDIDATE_SOURCE_HEAD=%s\nCANDIDATE_SOURCE_BINDING_SHA256=%s\n' "$candidate_source_head" "$candidate_source_binding_sha256"
    printf 'CONTROLLER_CHANNEL=%s\nCONTROLLER_SELECTOR=latest\nCONTROLLER_EXPECTED_VERSION=%s\nCONTROLLER_ACTUAL_VERSION=%s\n' \
      "$controller_channel" "$controller_expected_version" "$controller_actual_version"
    printf 'CONTROLLER_SCHEMA_VERSION=%s\n' "$controller_schema"
    if [[ "$controller_schema" == 2 ]]; then printf 'CONTROLLER_CLEANUP=external-owner-required\n'; fi
    if [[ "$controller_sidecars_written" == yes ]]; then
      printf 'CONTROLLER_AUTH_SHA256=%s\nCONTROLLER_PROVENANCE_SHA256=%s\n' "$TIER3_CONTROLLER_AUTH_SHA256" "$TIER3_CONTROLLER_PROVENANCE_SHA256"
    fi
    printf 'TEST_RUNNER_HOST_SERVICE=%s\nTEST_RUNNER_MANIFEST_SHA256=%s\n' "$TIER3_TEST_RUNNER_HOST_SERVICE" "$runner_manifest_sha256"
    printf 'QUALIFIED=%s\nCLEANUP=%s\n' "$qualified" "$cleanup"
  } >"$evidence_dir/final.env"
  rmdir "$qualifier_lock" || status=1
  printf 'tier3-candidate: evidence=%s\n' "$evidence_dir"
  exit "$status"
}

test_runner() {
  local operation="$1"
  shift
  "$test_runner_entry" --project-name unifi-stubd "$operation" \
    --manifest "$runner_manifest_relative" --host-service "$TIER3_TEST_RUNNER_HOST_SERVICE" "$@"
}

monitor_testhosts() {
  "$monitor_entry" --project-name unifi-stubd list \
    --runner test-runner --tier 3 --backend qemu --project "$repo_root" --all --json --strict
}

[[ $# -eq 0 ]] || die 'product qualification takes no arguments; use the documented native inputs'
for name in LAB_DATA_ROOT TIER3_CANDIDATE_BINARY TIER3_CANDIDATE_SHA256 TIER3_CANDIDATE_VERSION TIER3_CANDIDATE_MAC TIER3_CANDIDATE_IP TIER3_RUN_ID TIER3_SSH_IDENTITY_FILE TIER3_SSH_KNOWN_HOSTS_FILE TIER3_TEST_RUNNER_MANIFEST TIER3_TEST_RUNNER_HOST_SERVICE TIER3_TEST_RUNNER_SSH_PORT; do
  [[ -n "${!name:-}" ]] || die "set $name"
done
guest_ip="$TIER3_CANDIDATE_IP"
load_qualification_binding
controller_schema="${TIER3_CONTROLLER_SCHEMA_VERSION:-1}"
case "$controller_schema" in 1 | 2) ;; *) die 'unsupported controller schema' ;; esac
field_file="$repo_root/lab/stub/configs/tier3-controller-fields.txt"
[[ -f "$field_file" && ! -L "$field_file" ]] || die 'Core controller field contract is missing'
binding_keys="$(awk -v schema="$controller_schema" '$1 == "common" || $1 == schema { print $2 }' "$field_file" | paste -sd ' ' -)"
release_keys="$(awk -v schema="$controller_schema" '($1 == "common" || $1 == schema) && $2 ~ /^CONTROLLER_/ { print $2 }' "$field_file" | paste -sd ' ' -)"
for key in $release_keys; do
  case "$key" in CONTROLLER_AUTH_SHA256 | CONTROLLER_PROVENANCE_SHA256) continue ;; esac
  name="TIER3_$key"
  [[ -n "${!name:-}" && "${!name}" =~ ^[^[:space:][:cntrl:]]+$ ]] || die "invalid or missing $name"
done
if [[ "$controller_schema" == 2 ]]; then
  : "${TIER3_CONTROLLER_AUTH_FILE:?set the private controller authentication file}"
  : "${TIER3_CONTROLLER_PROVENANCE_FILE:?set the private controller owner provenance file}"
  private_file "$TIER3_CONTROLLER_AUTH_FILE" 65536
  private_file "$TIER3_CONTROLLER_PROVENANCE_FILE" 4193280
  jq -se 'length == 1 and (.[0] | type == "object" and keys == ["password", "username"] and all(.username, .password; type == "string" and length > 0))' \
    "$TIER3_CONTROLLER_AUTH_FILE" >/dev/null || die 'invalid controller authentication JSON'
  jq -se --arg channel "$TIER3_CONTROLLER_CHANNEL" --arg version "$TIER3_CONTROLLER_EXPECTED_VERSION" '
    length == 1 and (.[0] | type == "object" and .schema_version == 1 and .owner == "unifi-stubd-labs" and .fresh_controller == true and
    .controller_channel == $channel and .controller_version == $version and
    all(.controller_instance, .machine_id; type == "string" and length > 0) and
    all(.testhost_manifest_sha256, .controller_image_sha256, .controller_package_sha256; type == "string" and test("^[0-9a-f]{64}$")))
  ' "$TIER3_CONTROLLER_PROVENANCE_FILE" >/dev/null || die 'missing fresh per-lane controller provenance and artifact bindings'
fi
[[ "$TIER3_CONTROLLER_SELECTOR" == latest ]] || die 'product qualification requires the latest controller selector'
for command in awk cp file grep jq make readlink scp sed shasum ssh ssh-keygen unlink; do need "$command"; done
[[ "$LAB_DATA_ROOT" == /* && -d "$LAB_DATA_ROOT" ]] || die 'LAB_DATA_ROOT must be an existing absolute directory'
physical_directory "$LAB_DATA_ROOT"
machine_root="${LAB_DATA_ROOT%/}/tier3/machines/$profile"
[[ -d "$machine_root" && "$(cd "$machine_root" && pwd -P)" == "$machine_root" ]] || die 'machine root must be canonical and present'
run_id="$TIER3_RUN_ID"
candidate_mac="$(lower_ascii "$TIER3_CANDIDATE_MAC")"
candidate_sha256="$(lower_ascii "$TIER3_CANDIDATE_SHA256")"
candidate_version="$TIER3_CANDIDATE_VERSION"
[[ "$run_id" =~ ^[a-z0-9][a-z0-9_-]{0,31}$ ]] || die 'unsafe run ID'
[[ "$candidate_mac" =~ ^02:([0-9a-f]{2}:){4}[0-9a-f]{2}$ ]] || die 'candidate MAC must use a local 02 address'
[[ "$candidate_sha256" =~ ^[0-9a-f]{64}$ && "$candidate_version" =~ ^[A-Za-z0-9][A-Za-z0-9._+-]{0,79}$ ]] || die 'invalid candidate binding'
[[ "$TIER3_CANDIDATE_BINARY" == "$repo_root"/dist/candidates/*/bin/unifi-stubd_linux_arm64 && -f "$TIER3_CANDIDATE_BINARY" && ! -L "$TIER3_CANDIDATE_BINARY" ]] || die 'use the retained Core Linux ARM64 candidate'
candidate_evidence_root="$(cd "$(dirname "$TIER3_CANDIDATE_BINARY")/.." && pwd -P)"
[[ "$candidate_evidence_root" == "$repo_root"/dist/candidates/* && "$TIER3_CANDIDATE_BINARY" == "$candidate_evidence_root/bin/unifi-stubd_linux_arm64" ]] || die 'candidate path must be canonical inside Core evidence'
candidate_sums="$candidate_evidence_root/SHA256SUMS"
[[ -f "$candidate_sums" && ! -L "$candidate_sums" ]] || die 'candidate evidence has no SHA256SUMS'
recorded_sha="$(awk '$2 == "bin/unifi-stubd_linux_arm64" { count++; value=$1 } END { if (count != 1) exit 1; print value }' "$candidate_sums")"
[[ "$recorded_sha" == "$candidate_sha256" && "$(shasum -a 256 "$TIER3_CANDIDATE_BINARY" | awk '{print $1}')" == "$candidate_sha256" ]] || die 'candidate digest differs from retained evidence'
file "$TIER3_CANDIDATE_BINARY" | grep -Eq 'ELF 64-bit.*(ARM aarch64|ARM64)' || die 'candidate is not Linux ARM64'
candidate_source_head="$(cat "$candidate_evidence_root/base-head.txt")"
candidate_source_binding_sha256="$(shasum -a 256 "$candidate_evidence_root/verified-source.tar" | awk '{print $1}')"
recorded_source_sha="$(awk '$2 == "verified-source.tar" { count++; value=$1 } END { if (count != 1) exit 1; print value }' "$candidate_sums")"
[[ "$recorded_source_sha" == "$candidate_source_binding_sha256" ]] || die 'candidate source archive differs from retained evidence'
[[ "$candidate_source_head" =~ ^[0-9a-f]{40}$ ]] || die 'candidate source base is missing'
[[ "$TIER3_TEST_RUNNER_MANIFEST" =~ ^build-runner/testhosts/[A-Za-z0-9][A-Za-z0-9._/-]*\.ya?ml$ && "$TIER3_TEST_RUNNER_MANIFEST" != *..* ]] || die 'use a Core-owned testhost manifest'
runner_template_path="$repo_root/$TIER3_TEST_RUNNER_MANIFEST"
[[ -f "$runner_template_path" && ! -L "$runner_template_path" ]] || die 'missing Test Runner manifest template'
[[ "$(grep -c '^  name: unifi-stubd-lab$' "$runner_template_path")" -eq 1 ]] || die 'Test Runner template must name the bound Linux machine'
runner_template_sha256="$(shasum -a 256 "$runner_template_path" | awk '{print $1}')"
runner_name="unifi-stubd-lab-$run_id"
runner_manifest_relative="$core_input_relative_dir/$run_id.testhost.yaml"
runner_manifest_path="$repo_root/$runner_manifest_relative"
[[ "$TIER3_TEST_RUNNER_HOST_SERVICE" =~ ^[A-Za-z0-9][A-Za-z0-9._-]*$ ]] || die 'invalid managed host service'
guest_forward_port="$TIER3_TEST_RUNNER_SSH_PORT"
[[ "$guest_forward_port" =~ ^[1-9][0-9]{0,4}$ && "$guest_forward_port" -le 65535 ]] || die 'invalid Test Runner SSH port'
for ssh_file in "$TIER3_SSH_IDENTITY_FILE" "$TIER3_SSH_KNOWN_HOSTS_FILE"; do
  [[ "$ssh_file" == /* && -f "$ssh_file" && ! -L "$ssh_file" ]] || die 'SSH inputs must be absolute regular files'
done
"$PYTHON" -B "$script_dir/../tools/qualification_binding.py" \
  --data-root "$LAB_DATA_ROOT" --host-service "$TIER3_TEST_RUNNER_HOST_SERVICE" \
  --candidate-ip "$guest_ip" \
  --host "$guest_forward_host" --port "$guest_forward_port" --user "$guest_user" \
  --identity-file "$TIER3_SSH_IDENTITY_FILE" --known-hosts-file "$TIER3_SSH_KNOWN_HOSTS_FILE"
ssh_trust_host="$guest_forward_host"
if [[ "$guest_forward_port" != 22 ]]; then ssh_trust_host="[$guest_forward_host]:$guest_forward_port"; fi
ssh-keygen -F "$ssh_trust_host" -f "$TIER3_SSH_KNOWN_HOSTS_FILE" >/dev/null || die 'known_hosts has no trusted native SSH endpoint'
controller_channel="$TIER3_CONTROLLER_CHANNEL"
controller_expected_version="$TIER3_CONTROLLER_EXPECTED_VERSION"
controller_actual_version=not-run

evidence_dir="$machine_root/evidence/candidate/$run_id"
core_run_dir="$repo_root/dist/tier3-controller/$run_id"
core_manifest_relative="$core_input_relative_dir/$run_id.env"
core_manifest_path="$repo_root/$core_manifest_relative"
core_auth_path="${core_manifest_path%.env}.auth.json"
core_provenance_path="${core_manifest_path%.env}.controller.json"
controller_log="$evidence_dir/controller.log"
ready_file="$core_run_dir/ready.env"
result_file="$core_run_dir/result.env"
accepted_file="$core_run_dir/accepted.env"
qualifier_lock="$machine_root/control/candidate.lock"
[[ ! -e "$evidence_dir" && ! -e "$core_run_dir" && ! -e "$core_manifest_path" && ! -e "$runner_manifest_path" ]] || die 'run ID already has state'
[[ ! -e "$core_auth_path" && ! -L "$core_auth_path" && ! -e "$core_provenance_path" && ! -L "$core_provenance_path" ]] || die 'run ID already has controller sidecars'
for directory in "$machine_root/control" "$qualifier_lock" "$evidence_dir" "$repo_root/dist" "$repo_root/$core_input_relative_dir" "$core_run_dir"; do
  physical_directory "$directory"
done
initial_contract_hash="$(contract_hash)"
mkdir -p "$machine_root/control" "$machine_root/evidence/candidate" "$repo_root/$core_input_relative_dir"
if [[ "$controller_schema" == 2 ]]; then chmod 700 "$repo_root/$core_input_relative_dir"; fi
mkdir "$evidence_dir"
mkdir "$qualifier_lock" || die 'another qualification owns the machine; inspect its cleanup before retrying'
core_pid=
tunnel_pid=
core_started=no core_completed=no guest_start_attempted=no guest_owned=no
service_exists=no service_was_running=no qualified=no
core_manifest_sha256=not-written
runner_manifest_sha256=not-written
controller_sidecars_written=no
controller_auth_written=no
TIER3_CONTROLLER_AUTH_SHA256=not-written
trap on_exit EXIT
trap 'exit 129' HUP
trap 'exit 130' INT
trap 'exit 143' TERM

if [[ "$controller_schema" == 2 ]]; then
  source_auth_sha256="$(shasum -a 256 "$TIER3_CONTROLLER_AUTH_FILE" | awk '{print $1}')"
  auth_copy_status=0
  cp "$TIER3_CONTROLLER_AUTH_FILE" "$core_auth_path" || auth_copy_status=$?
  controller_auth_written=yes
  TIER3_CONTROLLER_AUTH_SHA256="$(shasum -a 256 "$core_auth_path" | awk '{print $1}')"
  [[ "$auth_copy_status" -eq 0 ]] || die 'controller authentication copy failed'
  [[ "$TIER3_CONTROLLER_AUTH_SHA256" == "$source_auth_sha256" ]] || die 'copied controller authentication differs from its source'
  jq -cs --arg run "$run_id" 'if length == 1 then {schema_version:1,run_id:$run,source:.[0]} else error("expected one provenance object") end' "$TIER3_CONTROLLER_PROVENANCE_FILE" >"$core_provenance_path"
  chmod 600 "$core_auth_path" "$core_provenance_path"
  private_file "$core_provenance_path" 4194304
  TIER3_CONTROLLER_PROVENANCE_SHA256="$(shasum -a 256 "$core_provenance_path" | awk '{print $1}')"
  controller_sidecars_written=yes
fi

ssh_options=(
  -p "$guest_forward_port" -i "$TIER3_SSH_IDENTITY_FILE"
  -o BatchMode=yes -o IdentitiesOnly=yes -o PasswordAuthentication=no
  -o KbdInteractiveAuthentication=no -o StrictHostKeyChecking=yes -o UpdateHostKeys=no
  -o "UserKnownHostsFile=$TIER3_SSH_KNOWN_HOSTS_FILE"
  -o GlobalKnownHostsFile=/dev/null -o "ConnectTimeout=$ssh_connect_timeout"
)
scp_options=("${ssh_options[@]}")
scp_options[0]=-P

# Each run has a new public testhost identity. Idempotent start must never make
# this qualifier the owner of a previously running or unresolved testhost.
sed "s/^  name: unifi-stubd-lab$/  name: $runner_name/" "$runner_template_path" >"$runner_manifest_path"
runner_manifest_sha256="$(shasum -a 256 "$runner_manifest_path" | awk '{print $1}')"
monitor_testhosts >"$evidence_dir/test-runner-before.json" 2>"$evidence_dir/test-runner-before.stderr"
jq -e --arg name "$runner_name" --arg host "$TIER3_TEST_RUNNER_HOST_SERVICE" '
  .schema_version == 1 and (.items | type == "array")
  and all(.items[]; .instance != $name and (.details.host_service != $host or .state == "stopped"))
' "$evidence_dir/test-runner-before.json" >/dev/null || die 'selected host service has existing active, stale or unknown runtime; retain it for its owner'
"$test_runner_entry" --project-name unifi-stubd request validate --manifest "$runner_manifest_relative" >"$evidence_dir/test-runner-validate.log" 2>&1
"$test_runner_entry" --project-name unifi-stubd inspect --host-service "$TIER3_TEST_RUNNER_HOST_SERVICE" --json >"$evidence_dir/test-runner-inspect.json"
jq -e --arg host "$TIER3_TEST_RUNNER_HOST_SERVICE" '.name == $host and .profile == "linux-arm64" and .platform == "linux/arm64" and .provider == "qemu" and
  .source_activity.state == "idle" and (.requirements | type == "object" and length > 0) and
  (.paths | type == "object" and length > 0) and all(.requirements[]; . == true) and all(.paths[]; .exists == true)' \
  "$evidence_dir/test-runner-inspect.json" >/dev/null || die 'selected product source is not available and idle; retain its state for its owner'
# This public lifecycle owns a converted copy; the original machine is not reset.
guest_start_attempted=yes
test_runner start >"$evidence_dir/test-runner-start.log" 2>&1
test_runner wait --timeout 300 >"$evidence_dir/test-runner-wait.log" 2>&1
# The guest resolves its own architecture, tools and privilege.
# shellcheck disable=SC2016
ssh_guest 'set -eu; test "$(uname -s)" = Linux; test "$(uname -m)" = aarch64; for command in bridge ip nc nohup ping sha256sum sudo; do command -v "$command" >/dev/null; done; sudo -n true' >"$evidence_dir/guest-preflight.log" 2>&1
ssh_guest 'ip -4 addr show' | awk -v expected="$guest_ip" '$1 == "inet" { split($2, address, "/"); if (address[1] == expected) found=1 } END { exit !found }' || die 'candidate IP is not assigned on the selected guest'
# Resolve this guest's own home; no host/machine-specific home path is stored.
# shellcheck disable=SC2016
guest_home="$(ssh_guest 'printf "%s\n" "$HOME"')"
[[ "$guest_home" =~ ^/[A-Za-z0-9._/-]+$ && "$guest_home" != *..* ]] || die 'guest home is not a safe absolute path'
for relative in "$guest_root_relative" "$guest_stage_relative"; do
  [[ "$relative" =~ ^[.A-Za-z0-9][A-Za-z0-9._/-]*$ && "$relative" != *..* ]] || die 'guest fixture paths must stay relative to its home'
done
guest_root_prefix="$guest_home/$guest_root_relative"
guest_stage_prefix="$guest_home/$guest_stage_relative"
guest_root="$guest_root_prefix/$run_id"
guest_stage="$guest_stage_prefix-$run_id"
[[ "$(ssh_guest "test -e '$guest_root' -o -e '$guest_stage' && printf occupied || printf clear")" == clear ]] || die 'guest run paths already exist'
# Keep the loop variable on the guest.
# shellcheck disable=SC2016
ssh_guest 'set -eu; for link in stubbr0 uplink0 uplinkp0; do ! ip link show "$link" >/dev/null 2>&1; done'
service_exists="$(ssh_guest "test -x /etc/init.d/$service_name && printf yes || printf no")"
ssh_guest "mkdir -m 700 '$guest_stage'"
guest_owned=yes
if [[ "$service_exists" == yes ]] && ssh_guest "sudo rc-service '$service_name' status >/dev/null 2>&1"; then
  service_was_running=yes
  ssh_guest "sudo rc-service '$service_name' stop >/dev/null"
fi
printf 'SERVICE_EXISTS=%s\nSERVICE_WAS_RUNNING=%s\n' "$service_exists" "$service_was_running" >"$evidence_dir/guest-service-before.env"

# Stage and hash the actual binary before starting the controller or fixture.
scp "${scp_options[@]}" "$TIER3_CANDIDATE_BINARY" "$guest_user@$guest_forward_host:$guest_stage/unifi-stubd.incoming"
scp "${scp_options[@]}" "$repo_root/lab/stub/configs/hosts/stub-bridge-observe/config.yaml" "$guest_user@$guest_forward_host:$guest_stage/config.yaml"
scp "${scp_options[@]}" "$repo_root/lab/stub/scripts/setup-bridge-observe.sh" "$guest_user@$guest_forward_host:$guest_stage/setup.sh"
expected_speed="$(
  ssh_guest "sudo sh -s -- '$guest_stage' '$guest_root' '$candidate_sha256' '$candidate_version'" <<'GUEST_PREPARE'
set -eu
stage=$1; root=$2; expected=$3; version=$4
mkdir -p "$root"; chmod 700 "$root"
install -m 0755 "$stage/unifi-stubd.incoming" "$root/unifi-stubd"
install -m 0600 "$stage/config.yaml" "$root/config.yaml"
install -m 0755 "$stage/setup.sh" "$root/setup.sh"
[ "$(sha256sum "$root/unifi-stubd" | awk '{print $1}')" = "$expected" ]
[ "$("$root/unifi-stubd" -version)" = "$version" ]
: >"$root/fixture.created"
sh "$root/setup.sh"
cat /sys/class/net/uplink0/speed
GUEST_PREPARE
)"
[[ "$expected_speed" =~ ^[1-9][0-9]*$ ]] || die 'guest uplink has no observed positive speed'
{
  printf 'SCHEMA_VERSION=%s\nRUN_ID=%s\n' "$controller_schema" "$run_id"
  for key in $release_keys; do
    name="TIER3_$key"
    printf '%s=%s\n' "$key" "${!name}"
  done
  printf 'CANDIDATE_MAC=%s\nCANDIDATE_SHA256=%s\nCANDIDATE_VERSION=%s\nCANDIDATE_IP=%s\nEXPECTED_UPLINK_SPEED_MBPS=%s\n' \
    "$candidate_mac" "$candidate_sha256" "$candidate_version" "$guest_ip" "$expected_speed"
} >"$core_manifest_path"
core_manifest_sha256="$(shasum -a 256 "$core_manifest_path" | awk '{print $1}')"
cp "$core_manifest_path" "$evidence_dir/core-run-manifest.env"
cp "$candidate_sums" "$evidence_dir/candidate-SHA256SUMS"
printf 'CANDIDATE_SOURCE_HEAD=%s\nCANDIDATE_SOURCE_BINDING_SHA256=%s\nCONTRACT_SHA256=%s\n' \
  "$candidate_source_head" "$candidate_source_binding_sha256" "$initial_contract_hash" >"$evidence_dir/binding.env"
core_started=yes
(
  cd "$repo_root"
  exec env UNIFI_STUB_TIER3_GATE_MANIFEST="$core_manifest_relative" make "$core_target"
) >"$controller_log" 2>&1 &
core_pid=$!
wait_for_file "$ready_file" "$ready_timeout"
runtime_keys=
if [[ "$controller_schema" == 1 ]]; then runtime_keys=CONTROLLER_RUNTIME_BASE_DIGEST; fi
ready_keys="$binding_keys MANIFEST_SHA256 CONTROLLER_ACTUAL_VERSION $runtime_keys INFORM_HOST INFORM_PORT GUEST_INFORM_PORT"
validate_env "$ready_file" "$ready_keys"
while IFS='=' read -r key value; do
  [[ "$(env_value "$ready_file" "$key")" == "$value" ]] || die "controller readiness changed binding: $key"
done <"$core_manifest_path"
[[ "$(env_value "$ready_file" MANIFEST_SHA256)" == "$core_manifest_sha256" ]] || die 'controller manifest digest mismatch'
controller_actual_version="$(env_value "$ready_file" CONTROLLER_ACTUAL_VERSION)"
[[ "$controller_actual_version" == "$controller_expected_version" ]] || die 'controller actual version mismatch'
inform_host="$(env_value "$ready_file" INFORM_HOST)"
inform_port="$(env_value "$ready_file" INFORM_PORT)"
guest_inform_port="$(env_value "$ready_file" GUEST_INFORM_PORT)"
[[ "$inform_host" == 127.0.0.1 && "$inform_port" =~ ^[1-9][0-9]{0,4}$ && "$inform_port" -le 65535 && "$guest_inform_port" == 8080 ]] || die 'invalid loopback inform endpoint'
cp "$ready_file" "$evidence_dir/controller-ready.env"
forward_ready=no
for ((tunnel_attempt = 1; tunnel_attempt <= 3; tunnel_attempt++)); do
  tunnel_log="$evidence_dir/inform-tunnel.log"
  if [[ "$tunnel_attempt" -gt 1 ]]; then tunnel_log="$evidence_dir/inform-tunnel-attempt-$tunnel_attempt.log"; fi
  ssh "${ssh_options[@]}" -N -T -o ExitOnForwardFailure=yes -o ServerAliveInterval=10 -o ServerAliveCountMax=3 \
    -R "127.0.0.1:$guest_inform_port:$inform_host:$inform_port" "$guest_user@$guest_forward_host" >"$tunnel_log" 2>&1 &
  tunnel_pid=$!
  for ((attempt = 0; attempt < 10; attempt++)); do
    kill -0 "$tunnel_pid" 2>/dev/null || break
    if [[ "$(ssh_guest 'printf "GET / HTTP/1.0\r\n\r\n" | nc -w 3 127.0.0.1 8080 | head -c 4')" == HTTP ]]; then
      forward_ready=yes
      break
    fi
    sleep 1
  done
  [[ "$forward_ready" == yes ]] && break
  kill -0 "$tunnel_pid" 2>/dev/null && die 'guest loopback inform tunnel did not become ready'
  tunnel_status=0
  wait "$tunnel_pid" || tunnel_status=$?
  tunnel_pid=
  # Reap the owned child before retrying only its exact early transport failure.
  # Separate private logs preserve every attempt; other failures stay terminal.
  if [[ "$tunnel_attempt" -lt 3 && "$tunnel_status" -eq 255 ]] &&
    awk -v broken="banner exchange: Connection to $guest_forward_host port $guest_forward_port: Broken pipe" \
      -v reset="banner exchange: Connection to $guest_forward_host port $guest_forward_port: Connection reset by peer" \
      -v connect_reset="ssh: connect to host $guest_forward_host port $guest_forward_port: Connection reset by peer" '
      { sub(/\r$/, ""); allowed = ($0 == broken || $0 == reset || $0 == connect_reset) }
      END { exit !(NR == 1 && allowed) }
    ' "$tunnel_log"; then
    sleep 5
    continue
  fi
  die 'inform tunnel exited before readiness'
done
[[ "$forward_ready" == yes ]] || die 'guest loopback inform tunnel did not become ready'
controller_url="http://127.0.0.1:$guest_inform_port/inform"

# The candidate's executable and arguments follow the verified local binding.
# shellcheck disable=SC2029
ssh "${ssh_options[@]}" "$guest_user@$guest_forward_host" \
  "sudo sh -s -- '$guest_root' '$controller_url' '$candidate_mac' '$guest_ip' 'tier3-candidate-${candidate_sha256:0:8}'" <<'GUEST_START'
set -eu
root=$1; controller=$2; mac=$3; ip=$4; hostname=$5
nohup "$root/unifi-stubd" -config "$root/config.yaml" -controller "$controller" -no-discovery \
  -mac "$mac" -ip "$ip" -hostname "$hostname" \
  -ssh-state "$root/adoption.env" -status-path "$root/status.json" \
  -instance-guard-path "$root/instance.lock" -ssh-host-key "$root/ssh_host_key" \
  >"$root/product.log" 2>&1 </dev/null &
printf '%s\n' "$!" >"$root/product.pid"
ping -I 198.51.100.2 -c 1500 -i 0.2 -s 1200 -W 1 203.0.113.1 >"$root/traffic-ingress.log" 2>&1 &
printf '%s\n' "$!" >"$root/traffic-ingress.pid"
ping -I 198.51.100.1 -c 1500 -i 0.2 -s 800 -W 1 203.0.113.2 >"$root/traffic-egress.log" 2>&1 &
printf '%s\n' "$!" >"$root/traffic-egress.pid"
GUEST_START
sleep 2
product_pid="$(ssh_guest "sudo cat '$guest_root/product.pid'")"
[[ "$product_pid" =~ ^[1-9][0-9]*$ ]] || die 'invalid product PID'
product_exe="$(ssh_guest "sudo readlink -f '/proc/$product_pid/exe'")"
product_hash="$(ssh_guest "sudo sha256sum '/proc/$product_pid/exe' | awk '{print \$1}'")"
[[ "$product_exe" == "$guest_root/unifi-stubd" && "$product_hash" == "$candidate_sha256" ]] || die 'running process differs from candidate'
printf 'RUN_ID=%s\nPID=%s\nEXE=%s\nSHA256=%s\nVERSION=%s\nMAC=%s\n' \
  "$run_id" "$product_pid" "$product_exe" "$product_hash" "$candidate_version" "$candidate_mac" >"$evidence_dir/product-binding.env"
wait_for_file "$accepted_file" "$result_timeout"
accepted_keys="$binding_keys MANIFEST_SHA256 CONTROLLER_ACTUAL_VERSION $runtime_keys ADOPTED CONNECTED TELEMETRY"
validate_env "$accepted_file" "$accepted_keys"
for key in $binding_keys MANIFEST_SHA256 CONTROLLER_ACTUAL_VERSION $runtime_keys; do
  [[ "$(env_value "$accepted_file" "$key")" == "$(env_value "$ready_file" "$key")" ]] || die "controller acceptance changed binding: $key"
done
ssh_guest "set -eu; test \"\$(sudo readlink -f '/proc/$product_pid/exe')\" = '$guest_root/unifi-stubd'; test \"\$(sudo sha256sum '/proc/$product_pid/exe' | awk '{print \$1}')\" = '$candidate_sha256'; sudo grep -q '^STATE=connected$' '$guest_root/adoption.env'; sudo grep -Eq '^AUTHKEY=.+$' '$guest_root/adoption.env'; sudo grep -Eq '^CFGVERSION=.+$' '$guest_root/adoption.env'"
ssh_guest "sudo cat '$guest_root/status.json'" >"$evidence_dir/product-status.json"
jq -e --arg endpoint "$controller_url" '.last_inform | .controller_state == "connected" and .status_code == 200 and .key_source == "adopted" and .url == $endpoint' "$evidence_dir/product-status.json" >/dev/null ||
  die 'candidate status does not confirm the adopted loopback inform exchange'
printf '%s\n' "$core_manifest_sha256" >"$core_run_dir/guest-complete.sha256.incoming"
mv "$core_run_dir/guest-complete.sha256.incoming" "$core_run_dir/guest-complete.sha256"
wait_for_file "$result_file" 90
wait_pid "$core_pid" 60 || die 'controller gate did not exit after result'
wait "$core_pid" || die 'controller gate failed'
core_pid=
controller_cleanup_keys=
if [[ "$controller_schema" == 2 ]]; then controller_cleanup_keys=CONTROLLER_CLEANUP; fi
result_keys="$binding_keys MANIFEST_SHA256 CONTROLLER_ACTUAL_VERSION $runtime_keys RESULT ADOPTED CONNECTED TELEMETRY CLEANUP $controller_cleanup_keys"
validate_env "$result_file" "$result_keys"
if [[ "$controller_schema" == 2 ]]; then
  [[ "$(env_value "$result_file" CONTROLLER_CLEANUP)" == external-owner-required ]] || die 'unexpected external controller cleanup ownership'
fi
for key in $binding_keys MANIFEST_SHA256 CONTROLLER_ACTUAL_VERSION $runtime_keys; do
  [[ "$(env_value "$result_file" "$key")" == "$(env_value "$ready_file" "$key")" ]] || die "controller result changed binding: $key"
done
for expectation in 'RESULT passed' 'ADOPTED true' 'CONNECTED true' 'TELEMETRY true' 'CLEANUP passed'; do
  [[ "$(env_value "$result_file" "${expectation%% *}")" == "${expectation#* }" ]] || die "controller result mismatch: $expectation"
done
cp "$result_file" "$evidence_dir/controller-result.env"
cp "$core_run_dir/controller-ui.log" "$evidence_dir/controller-ui.log"
cp "$core_run_dir/controller-connected.env" "$evidence_dir/controller-connected.env"
core_completed=yes
qualified=yes
printf 'tier3-candidate: exact candidate adopted, connected, and reporting controller metrics\n'

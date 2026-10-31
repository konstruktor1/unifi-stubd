#!/bin/bash

set -euo pipefail
umask 077

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)"
qualifier="$script_dir/qualify-candidate.sh"
release_endpoint=https://community.svc.ui.com/graphql
release_group=e6712595-81bb-4829-8e42-9e2630fabcfe
runtime_base_default=lscr.io/linuxserver/unifi-network-application:latest

die() {
  printf 'tier3-matrix: %s\n' "$*" >&2
  exit 1
}
need() { command -v "$1" >/dev/null 2>&1 || die "missing command: $1"; }

# shellcheck source=lab/stub/scripts/lib/qualification-context.sh
source "$script_dir/lib/qualification-context.sh"

env_value() {
  awk -F= -v key="$2" '$1 == key { count++; value = substr($0, length($1) + 2) } END { if (count != 1) exit 1; print value }' "$1"
}

finish() {
  local command_status=$? result=failed final_status
  trap - EXIT HUP INT TERM
  if [[ -n "$lane_pid" ]]; then
    kill -TERM "$lane_pid" 2>/dev/null || true
    set +e
    wait "$lane_pid"
    set -e
    lane_pid=
  fi
  if [[ "$matrix_complete" == yes && "$stable_result" == passed && "$stable_cleanup" == passed ]] &&
    [[ "$scope" == stable-only || ("$rc_result" == passed && "$rc_cleanup" == passed) ]]; then
    result=passed
  fi
  final_status=$command_status
  [[ "$result" == passed ]] && final_status=0
  [[ "$result" == failed && "$final_status" -eq 0 ]] && final_status=1
  {
    printf 'RESULT=%s\nMATRIX_RUN_ID=%s\nSELECTOR=latest\n' "$result" "$matrix_id"
    printf 'SCOPE=%s\n' "$scope"
    if [[ "$controller_schema" == 2 ]]; then printf 'CONTROLLER_CLEANUP=external-owner-required\n'; fi
    printf 'CANDIDATE_SHA256=%s\nCANDIDATE_VERSION=%s\n' "$TIER3_CANDIDATE_SHA256" "$TIER3_CANDIDATE_VERSION"
    printf 'CONTROLLER_RELEASE_QUERY_SHA256=%s\nCONTROLLER_RELEASE_RESPONSE_SHA256=%s\n' "$release_query_sha" "$release_response_sha"
    printf 'STABLE_RUN_ID=%s\nSTABLE_RESULT=%s\nSTABLE_CLEANUP=%s\nSTABLE_VERSION=%s\n' "$stable_run_id" "$stable_result" "$stable_cleanup" "$stable_version"
    printf 'RELEASE_CANDIDATE_RUN_ID=%s\nRELEASE_CANDIDATE_RESULT=%s\nRELEASE_CANDIDATE_CLEANUP=%s\nRELEASE_CANDIDATE_VERSION=%s\n' "$rc_run_id" "$rc_result" "$rc_cleanup" "$rc_version"
  } >"$matrix_evidence/final.env"
  printf 'tier3-matrix: result=%s evidence=%s\n' "$result" "$matrix_evidence"
  exit "$final_status"
}

forward_signal() {
  local signal="$1" status="$2"
  trap - HUP INT TERM
  if [[ -n "$lane_pid" ]]; then
    kill -s "$signal" "$lane_pid" 2>/dev/null || true
    set +e
    wait "$lane_pid"
    set -e
    lane_pid=
  fi
  exit "$status"
}

resolve_lane() {
  local key="$1" channel="$2" stage="$3" item_file="$4" env_file="$5"
  local id slug version created zip_url zip_sha release_url
  jq -e --arg key "$key" '.data[$key].items[0]' "$release_response" >"$item_file"
  id="$(jq -er '.id' "$item_file")"
  slug="$(jq -er '.slug' "$item_file")"
  version="$(jq -er '.version' "$item_file")"
  created="$(jq -er '.createdAt' "$item_file")"
  [[ "$id" =~ ^[0-9a-f-]{36}$ && "$slug" =~ ^[A-Za-z0-9-]+$ && "$version" =~ ^[0-9]+([.][0-9]+){1,3}$ ]] || die "unsafe $channel release metadata"
  if [[ "$controller_schema" == 1 ]]; then
    zip_url="$(jq -er '[.links[]?.url | select(endswith("/UniFi.unix.zip"))] | if length == 1 then .[0] else error("expected one Unix ZIP") end' "$item_file")"
    [[ "$zip_url" =~ ^https://dl\.ui\.com/[A-Za-z0-9._/-]+/UniFi\.unix\.zip$ ]] || die "unsafe $channel ZIP URL"
    zip_sha="$(jq -er '[.instructions[]? | select(.type == "TEXT") | .content | match("([0-9a-f]{64}) \\| UniFi\\.unix\\.zip"; "g") | .captures[0].string] | if length == 1 then .[0] else error("expected one Unix ZIP SHA-256") end' "$item_file")"
    [[ "$zip_sha" =~ ^[0-9a-f]{64}$ ]] || die "invalid $channel ZIP SHA-256"
    curl -fsSIL --max-time 30 --proto '=https' --proto-redir '=https' "$zip_url" >"$matrix_evidence/$channel-artifact-head.txt"
  fi
  release_url="https://community.ui.com/releases/$slug/$id"
  {
    printf 'CONTROLLER_CHANNEL=%s\nCONTROLLER_SELECTOR=latest\n' "$channel"
    printf 'CONTROLLER_EXPECTED_VERSION=%s\nCONTROLLER_RELEASE_STAGE=%s\n' "$version" "$stage"
    printf 'CONTROLLER_RELEASE_ID=%s\nCONTROLLER_RELEASE_URL=%s\n' "$id" "$release_url"
    printf 'CONTROLLER_RELEASE_CREATED_AT=%s\nCONTROLLER_RESOLVED_AT=%s\n' "$created" "$resolved_at"
    if [[ "$controller_schema" == 1 ]]; then printf 'CONTROLLER_ZIP_URL=%s\nCONTROLLER_ZIP_SHA256=%s\n' "$zip_url" "$zip_sha"; fi
    printf 'CONTROLLER_RELEASE_QUERY_ENDPOINT=%s\nCONTROLLER_RELEASE_QUERY_SHA256=%s\n' "$release_endpoint" "$release_query_sha"
    printf 'CONTROLLER_RELEASE_RESPONSE_SHA256=%s\n' "$release_response_sha"
  } >"$env_file"
}

run_lane() {
  local env_file="$1" run_id="$2" channel="$3" status final_file result cleanup key prefix name
  local lane_env=("TIER3_RUN_ID=$run_id" "TIER3_CONTROLLER_SCHEMA_VERSION=$controller_schema")
  while IFS= read -r key; do
    lane_env+=("TIER3_$key=$(env_value "$env_file" "$key")")
  done < <(awk -v schema="$controller_schema" '($1 == "common" || $1 == schema) && $2 ~ /^CONTROLLER_/ && $2 !~ /^(CONTROLLER_RUNTIME_BASE_SELECTOR|CONTROLLER_API_ORIGIN|CONTROLLER_INFORM_PORT|CONTROLLER_AUTH_SHA256|CONTROLLER_PROVENANCE_SHA256)$/ { print $2 }' "$field_file")
  if [[ "$controller_schema" == 1 ]]; then
    lane_env+=("TIER3_CONTROLLER_RUNTIME_BASE_SELECTOR=$runtime_base")
  else
    case "$channel" in stable) prefix=TIER3_STABLE_CONTROLLER ;; release-candidate) prefix=TIER3_RC_CONTROLLER ;; esac
    for key in API_ORIGIN INFORM_PORT AUTH_FILE PROVENANCE_FILE; do
      name="${prefix}_$key"
      lane_env+=("TIER3_CONTROLLER_$key=${!name}")
    done
  fi
  set +e
  env "${lane_env[@]}" "$qualifier" &
  lane_pid=$!
  wait "$lane_pid"
  status=$?
  lane_pid=
  set -e
  final_file="$machine_root/evidence/candidate/$run_id/final.env"
  [[ -s "$final_file" ]] || die "$channel qualifier produced no final evidence"
  result="$(env_value "$final_file" RESULT)"
  cleanup="$(env_value "$final_file" CLEANUP)"
  [[ "$result" == passed || "$result" == failed ]] || die "$channel qualifier returned an invalid result"
  [[ "$cleanup" == passed || "$cleanup" == failed ]] || die "$channel qualifier returned an invalid cleanup result"
  cp "$final_file" "$matrix_evidence/$channel-result.env"
  [[ "$status" -eq 0 && "$result" == passed ]] || return 1
  return 0
}

[[ $# -eq 0 ]] || die 'matrix qualification takes no arguments; use TIER3_MATRIX_SCOPE'
scope="${TIER3_MATRIX_SCOPE:-stable-and-release-candidate}"
case "$scope" in stable-and-release-candidate | stable-only) ;; *) die 'TIER3_MATRIX_SCOPE must be stable-and-release-candidate or stable-only' ;; esac
for name in LAB_DATA_ROOT TIER3_CANDIDATE_BINARY TIER3_CANDIDATE_SHA256 TIER3_CANDIDATE_VERSION TIER3_CANDIDATE_MAC TIER3_CANDIDATE_IP TIER3_MATRIX_RUN_ID TIER3_SSH_IDENTITY_FILE TIER3_SSH_KNOWN_HOSTS_FILE TIER3_TEST_RUNNER_MANIFEST TIER3_TEST_RUNNER_HOST_SERVICE TIER3_TEST_RUNNER_SSH_PORT; do
  [[ -n "${!name:-}" ]] || die "set $name"
done
load_qualification_binding
for command in awk cp curl env jq shasum; do need "$command"; done
[[ -x "$qualifier" ]] || die "qualifier is not executable: $qualifier"
matrix_id="$TIER3_MATRIX_RUN_ID"
[[ "$matrix_id" =~ ^[a-z0-9][a-z0-9_-]{0,28}$ ]] || die 'unsafe TIER3_MATRIX_RUN_ID (maximum 29 characters)'
runtime_base="${TIER3_CONTROLLER_RUNTIME_BASE_SELECTOR:-$runtime_base_default}"
[[ "$runtime_base" =~ ^[A-Za-z0-9][A-Za-z0-9./:@_-]{0,255}$ ]] || die 'unsafe controller runtime-base selector'
controller_schema="${TIER3_CONTROLLER_SCHEMA_VERSION:-1}"
case "$controller_schema" in 1 | 2) ;; *) die 'unsupported controller schema' ;; esac
field_file="$repo_root/lab/stub/configs/tier3-controller-fields.txt"
[[ -f "$field_file" && ! -L "$field_file" ]] || die 'Core controller field contract is missing'
if [[ "$controller_schema" == 2 ]]; then
  prefixes=(TIER3_STABLE_CONTROLLER)
  if [[ "$scope" == stable-and-release-candidate ]]; then prefixes+=(TIER3_RC_CONTROLLER); fi
  for prefix in "${prefixes[@]}"; do
    for key in API_ORIGIN INFORM_PORT AUTH_FILE PROVENANCE_FILE; do
      name="${prefix}_$key"
      [[ -n "${!name:-}" ]] || die "schema 2 requires a fresh external controller per lane: set $name"
    done
    name="${prefix}_PROVENANCE_FILE"
    [[ -f "${!name}" && ! -L "${!name}" && "$(wc -c <"${!name}")" -le 4194304 ]] || die 'controller provenance must be a bounded regular file'
    physical_directory "$(dirname "${!name}")"
  done
  if [[ "$scope" == stable-and-release-candidate ]]; then
    stable_instance="$(jq -ser 'if length == 1 then .[0].controller_instance | select(type == "string" and length > 0) else error("expected one controller provenance object") end' "$TIER3_STABLE_CONTROLLER_PROVENANCE_FILE")"
    rc_instance="$(jq -ser 'if length == 1 then .[0].controller_instance | select(type == "string" and length > 0) else error("expected one controller provenance object") end' "$TIER3_RC_CONTROLLER_PROVENANCE_FILE")"
    [[ "$stable_instance" != "$rc_instance" ]] || die 'controller instances must differ between Stable and RC lanes'
  fi
fi

machine_root="${LAB_DATA_ROOT%/}/tier3/machines/unifi-stubd-lab"
[[ "$LAB_DATA_ROOT" == /* && -d "$machine_root" ]] || die 'use the existing native project data binding and assigned product machine'
physical_directory "$LAB_DATA_ROOT"
matrix_evidence="$machine_root/evidence/candidate-matrix/$matrix_id"
stable_run_id="$matrix_id-ga"
rc_run_id="$matrix_id-rc"
stable_result=not-run
stable_cleanup=not-run
stable_version=not-resolved
rc_result=not-run
rc_cleanup=not-run
rc_version=not-resolved
release_query_sha=not-run
release_response_sha=not-run
matrix_complete=no
lane_pid=
[[ ! -e "$matrix_evidence" ]] || die 'matrix run ID already has evidence'
physical_directory "$matrix_evidence"
mkdir -p "$matrix_evidence"
chmod 700 "$matrix_evidence"
trap finish EXIT
trap 'forward_signal HUP 129' HUP
trap 'forward_signal INT 130' INT
trap 'forward_signal TERM 143' TERM

resolved_at="$(date -u '+%Y-%m-%dT%H:%M:%SZ')"
release_request="$matrix_evidence/controller-release-request.json"
release_response="$matrix_evidence/controller-release-response.json"
# GraphQL variable names are literal request data, not shell expansions.
# shellcheck disable=SC2016
query='query LatestControllerChannels($groupId: ID!) { stable: releases(groupId: $groupId, stage: GA, statuses: [PUBLISHED], sortBy: LATEST, limit: 1) { totalCount items { id slug title version groupId stage status createdAt updatedAt links { url title checksums { md5 sha256 } } instructions { type ... on TextContent { content } } } } releaseCandidate: releases(groupId: $groupId, stage: RC, statuses: [PUBLISHED], sortBy: LATEST, limit: 1) { totalCount items { id slug title version groupId stage status createdAt updatedAt links { url title checksums { md5 sha256 } } instructions { type ... on TextContent { content } } } } }'
jq -n --arg query "$query" --arg groupId "$release_group" '{query: $query, variables: {groupId: $groupId}}' >"$release_request"
release_query_sha="$(shasum -a 256 "$release_request" | awk '{print $1}')"
curl -fsS --max-time 30 --max-filesize 4194304 --proto '=https' -H 'content-type: application/json' --data-binary "@$release_request" "$release_endpoint" >"$release_response"
release_response_sha="$(shasum -a 256 "$release_response" | awk '{print $1}')"
jq -e --arg group "$release_group" --arg scope "$scope" '
  (.errors? == null)
  and (.data.stable.items | length == 1)
  and (.data.stable.items[0] | .title == "UniFi Network Application" and .groupId == $group and .stage == "GA" and .status == "PUBLISHED")
  and ($scope == "stable-only" or (
    (.data.releaseCandidate.items | length == 1)
    and (.data.releaseCandidate.items[0] | .title == "UniFi Network Application" and .groupId == $group and .stage == "RC" and .status == "PUBLISHED")
  ))
' "$release_response" >/dev/null || die 'Ubiquiti query did not return the requested current channel metadata'
{
  printf 'CONTROLLER_RELEASE_QUERY_ENDPOINT=%s\n' "$release_endpoint"
  printf 'CONTROLLER_RELEASE_QUERY_SHA256=%s\n' "$release_query_sha"
  printf 'CONTROLLER_RELEASE_RESPONSE_SHA256=%s\n' "$release_response_sha"
  printf 'CONTROLLER_RESOLVED_AT=%s\n' "$resolved_at"
} >"$matrix_evidence/controller-release-binding.env"

stable_env="$matrix_evidence/stable-release.env"
rc_env="$matrix_evidence/release-candidate-release.env"
resolve_lane stable stable GA "$matrix_evidence/stable-release.json" "$stable_env"
stable_version="$(env_value "$stable_env" CONTROLLER_EXPECTED_VERSION)"
if [[ "$scope" == stable-and-release-candidate ]]; then
  resolve_lane releaseCandidate release-candidate RC "$matrix_evidence/release-candidate-release.json" "$rc_env"
  rc_version="$(env_value "$rc_env" CONTROLLER_EXPECTED_VERSION)"
fi

if run_lane "$stable_env" "$stable_run_id" stable; then stable_result=passed; else stable_result=failed; fi
stable_cleanup="$(env_value "$machine_root/evidence/candidate/$stable_run_id/final.env" CLEANUP)"
[[ "$stable_cleanup" == passed ]] || die 'Stable cleanup failed; RC lane is unsafe to start'
if [[ "$scope" == stable-only ]]; then
  matrix_complete=yes
  [[ "$stable_result" == passed ]] || die 'candidate did not pass Stable'
  exit 0
fi
if run_lane "$rc_env" "$rc_run_id" release-candidate; then rc_result=passed; else rc_result=failed; fi
rc_cleanup="$(env_value "$machine_root/evidence/candidate/$rc_run_id/final.env" CLEANUP)"
matrix_complete=yes
[[ "$rc_cleanup" == passed ]] || die 'Release Candidate cleanup failed'
[[ "$stable_result" == passed && "$rc_result" == passed ]] || die 'candidate did not pass both controller channels'

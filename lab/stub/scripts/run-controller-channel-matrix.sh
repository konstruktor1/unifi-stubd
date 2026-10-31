#!/bin/bash
set -euo pipefail
umask 077

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
if [[ -n "${UNIFI_STUB_TIER3_GATE_MANIFEST:-}" ]]; then
  exec bash "$script_dir/run-tier3-controller-gate.sh"
fi
repo_root="$(cd -- "$script_dir/../../.." && pwd)"
release_endpoint='https://community.svc.ui.com/graphql'
release_group='e6712595-81bb-4829-8e42-9e2630fabcfe'
runtime_base_selector='lscr.io/linuxserver/unifi-network-application:latest'

fail() {
  echo "controller matrix: $*" >&2
  exit 2
}

need() {
  command -v "$1" >/dev/null 2>&1 || fail "missing command: $1"
}

image_alias_absent() {
  local image_ids
  image_ids="$(docker image ls --quiet --no-trunc "$1")" || return 1
  [[ -z "$image_ids" ]]
}

project_is_clean() {
  local project="$1" label containers networks volumes
  label="label=com.docker.compose.project=$project"
  containers="$(docker ps -aq --filter "$label")" || return 1
  networks="$(docker network ls -q --filter "$label")" || return 1
  volumes="$(docker volume ls -q --filter "$label")" || return 1
  [[ -z "$containers" && -z "$networks" && -z "$volumes" ]]
}

compose_channel() {
  docker compose \
    -f "$repo_root/lab/stub/compose.yaml" \
    -f "$repo_root/lab/stub/compose.tests.yaml" \
    -f "$repo_root/lab/stub/compose.controller-channel.yaml" \
    "$@"
}

resolve_lane() {
  local prefix="$1" key="$2" channel="$3"
  local id slug version version_slug expected_slug created zip_url zip_release_token zip_sha
  id="$(jq -er --arg key "$key" '.data[$key].items[0].id' "$release_response")"
  slug="$(jq -er --arg key "$key" '.data[$key].items[0].slug' "$release_response")"
  version="$(jq -er --arg key "$key" '.data[$key].items[0].version' "$release_response")"
  created="$(jq -er --arg key "$key" '.data[$key].items[0].createdAt' "$release_response")"
  [[ "$id" =~ ^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$ &&
    "$version" =~ ^[0-9]+([.][0-9]+){1,3}$ &&
    "$created" =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}([.][0-9]+)?Z$ ]] ||
    fail "unsafe $channel release metadata"
  version_slug="${version//./-}"
  expected_slug="UniFi-Network-Application-$version_slug"
  [[ "$slug" == "$expected_slug" ]] ||
    fail "$channel release slug does not bind the expected version"
  zip_url="$(jq -er --arg key "$key" '[.data[$key].items[0].links[]?.url | select(endswith("/UniFi.unix.zip"))] | if length == 1 then .[0] else error("expected one Unix ZIP") end' "$release_response")" ||
    fail "$channel $version does not publish one official Unix ZIP; current-channel execution is blocked"
  [[ "$zip_url" =~ ^https://dl\.ui\.com/[A-Za-z0-9._/-]+/UniFi\.unix\.zip$ ]] ||
    fail "unsafe $channel ZIP URL"
  zip_release_token="${zip_url#https://dl.ui.com/unifi/}"
  zip_release_token="${zip_release_token%/UniFi.unix.zip}"
  case "$zip_release_token" in
  "$version" | "$version"-*) ;;
  *) fail "$channel ZIP URL does not bind the expected version" ;;
  esac
  zip_sha="$(jq -er --arg key "$key" '[.data[$key].items[0].instructions[]? | select(.type == "TEXT") | .content | match("([0-9a-f]{64}) \\| UniFi\\.unix\\.zip"; "g") | .captures[0].string] | if length == 1 then .[0] else error("expected one Unix ZIP SHA-256") end' "$release_response")" ||
    fail "$channel $version does not publish one official Unix ZIP SHA-256"
  [[ "$zip_sha" =~ ^[0-9a-f]{64}$ ]] || fail "invalid $channel ZIP SHA-256"
  printf -v "${prefix}_version" '%s' "$version"
  printf -v "${prefix}_zip_url" '%s' "$zip_url"
  printf -v "${prefix}_zip_sha" '%s' "$zip_sha"
}

remove_owned_image_alias() {
  local image_alias="$1"
  image_alias_absent "$image_alias" || docker image rm "$image_alias" >/dev/null
  image_alias_absent "$image_alias"
}

cleanup() {
  local command_status=$? cleanup_result=passed final_status
  trap - EXIT HUP INT TERM
  set +e
  if [[ -n "${current_project:-}" ]] && ! project_is_clean "$current_project"; then
    compose_channel down --volumes --remove-orphans >/dev/null 2>&1 || cleanup_result=failed
    project_is_clean "$current_project" || cleanup_result=failed
  fi
  for image_alias in "${stable_controller_image:-}" "${rc_controller_image:-}" "${product_image:-}" "${helper_image:-}" "${ui_image:-}"; do
    if [[ -n "$image_alias" ]] && ! remove_owned_image_alias "$image_alias"; then
      cleanup_result=failed
    fi
  done
  final_status=$command_status
  [[ "$cleanup_result" == passed ]] || final_status=1
  echo "controller matrix: result=$([[ "$final_status" -eq 0 ]] && echo passed || echo failed) cleanup=$cleanup_result artifacts=$evidence_dir"
  exit "$final_status"
}

[[ "$repo_root" == /workspace ]] || fail 'run through Build Runner (make integration-nonunifi-metrics)'
[[ $# -eq 0 ]] || fail 'this command accepts no arguments'
for command in awk curl docker jq sed tee; do need "$command"; done
docker compose version >/dev/null 2>&1 || fail 'docker compose is required'
run_id="${UNIFI_STUB_CONTROLLER_MATRIX_RUN_ID:-matrix-$(date -u '+%Y%m%d%H%M%S')-$$}"
[[ "$run_id" =~ ^[a-z0-9][a-z0-9-]{0,39}$ ]] || fail 'unsafe matrix run ID'
evidence_dir="$repo_root/dist/controller-channel-matrix/$run_id"
mkdir -p "$(dirname -- "$evidence_dir")"
mkdir "$evidence_dir" || fail "evidence already exists: $evidence_dir"

product_image="unifi-stubd:product-$run_id"
helper_image="unifi-stubd-controller-adoption:gate-$run_id"
ui_image="unifi-stubd-controller-ui:gate-$run_id"
stable_controller_image="unifi-stubd-controller:$run_id-stable"
rc_controller_image="unifi-stubd-controller:$run_id-rc"
for image_alias in "$product_image" "$helper_image" "$ui_image" "$stable_controller_image" "$rc_controller_image"; do
  image_alias_absent "$image_alias" || fail "owned image alias already exists: $image_alias"
done

current_project=
trap cleanup EXIT
trap 'exit 129' HUP
trap 'exit 130' INT
trap 'exit 143' TERM

resolved_at="$(date -u '+%Y-%m-%dT%H:%M:%SZ')"
release_response="$evidence_dir/controller-release-response.json"
# shellcheck disable=SC2016
query='query LatestControllerChannels($groupId: ID!) { stable: releases(groupId: $groupId, stage: GA, statuses: [PUBLISHED], sortBy: LATEST, limit: 1) { items { id slug title version groupId stage status createdAt links { url } instructions { type ... on TextContent { content } } } } releaseCandidate: releases(groupId: $groupId, stage: RC, statuses: [PUBLISHED], sortBy: LATEST, limit: 1) { items { id slug title version groupId stage status createdAt links { url } instructions { type ... on TextContent { content } } } } }'
jq -n --arg query "$query" --arg groupId "$release_group" '{query: $query, variables: {groupId: $groupId}}' |
  curl -fsS --max-time 30 --proto '=https' -H 'content-type: application/json' \
    --data-binary @- "$release_endpoint" >"$release_response"
jq -e --arg group "$release_group" '
    (.errors? == null)
    and (.data.stable.items | length == 1)
    and (.data.releaseCandidate.items | length == 1)
    and (.data.stable.items[0] | .title == "UniFi Network Application" and .groupId == $group and .stage == "GA" and .status == "PUBLISHED")
    and (.data.releaseCandidate.items[0] | .title == "UniFi Network Application" and .groupId == $group and .stage == "RC" and .status == "PUBLISHED")
' "$release_response" >/dev/null || fail 'official channel query did not return exactly one current Stable and RC'

stable_version='' stable_zip_url='' stable_zip_sha=''
rc_version='' rc_zip_url='' rc_zip_sha=''
resolve_lane stable stable stable
resolve_lane rc releaseCandidate release-candidate
echo "controller matrix: resolved stable=$stable_version release-candidate=$rc_version at=$resolved_at"

docker pull "$runtime_base_selector"
runtime_repository="${runtime_base_selector%:*}"
runtime_base_digest="$(docker image inspect --format '{{range .RepoDigests}}{{println .}}{{end}}' "$runtime_base_selector" | awk -v prefix="$runtime_repository@sha256:" 'index($0, prefix) == 1 { count++; value=$0 } END { if (count != 1) exit 1; print value }')" ||
  fail 'latest controller runtime base did not resolve to one immutable digest'

candidate_version="$(sed -n '1p' "$repo_root/VERSION")"
[[ "$candidate_version" =~ ^[0-9][0-9A-Za-z.+~-]{0,63}$ ]] || fail 'invalid product version'
export UNIFI_STUB_IMAGE="$product_image"
export UNIFI_STUB_CONTROLLER_HELPER_IMAGE="$helper_image"
export UNIFI_STUB_CONTROLLER_UI_IMAGE="$ui_image"
export UNIFI_STUBD_VERSION="$candidate_version"
export UNIFI_STUB_BRIDGE_MAC=02:15:6d:00:08:21
export UNIFI_STUB_BRIDGE_IP=172.31.242.25
UNIFI_NETWORK_IMAGE="$runtime_base_digest" \
  COMPOSE_PROJECT_NAME="unifi-stubd-build-$run_id" docker compose \
  -f "$repo_root/lab/stub/compose.yaml" \
  -f "$repo_root/lab/stub/compose.tests.yaml" \
  build stub-bridge-observe controller-adoption controller-ui
run_lane() {
  local channel="$1" expected_version="$2" zip_url="$3" zip_sha="$4" controller_image="$5" lane_suffix="$6"
  local lane_status
  current_project="unifi-stubd-$run_id-$lane_suffix"
  export COMPOSE_PROJECT_NAME="$current_project"
  export UNIFI_STUB_CONTROLLER_RUNTIME_BASE_DIGEST="$runtime_base_digest"
  export UNIFI_STUB_CONTROLLER_ZIP_URL="$zip_url"
  export UNIFI_STUB_CONTROLLER_ZIP_SHA256="$zip_sha"
  export UNIFI_STUB_CONTROLLER_IMAGE="$controller_image"
  export UNIFI_NETWORK_IMAGE="$controller_image"
  export UNIFI_STUB_LAB_EXPECTED_NETWORK_VERSION="$expected_version"
  compose_channel build unifi || fail "$channel controller image build failed"
  set +e
  sh "$script_dir/run-docker-integration.sh" 2>&1 | tee "$evidence_dir/$channel.log"
  lane_status=$?
  set -e
  [[ "$lane_status" -eq 0 ]] || return "$lane_status"
  project_is_clean "$current_project" || fail "$channel Compose project was not completely removed"
  current_project=
}

run_lane stable "$stable_version" "$stable_zip_url" "$stable_zip_sha" "$stable_controller_image" ga ||
  fail 'Stable product lane failed'
run_lane release-candidate "$rc_version" "$rc_zip_url" "$rc_zip_sha" "$rc_controller_image" rc ||
  fail 'Release Candidate product lane failed'

echo "controller matrix: same product image passed the controller-facing gate on Stable $stable_version and RC $rc_version"

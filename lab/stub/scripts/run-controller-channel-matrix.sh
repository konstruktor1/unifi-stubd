#!/bin/bash
set -euo pipefail
umask 077

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
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

write_snapshot_manifest() {
  local output="$1"
  local source_file
  shift
  (
    cd "$repo_root"
    printf '%s\0' "$@" | LC_ALL=C sort -zu |
      while IFS= read -r -d '' source_file; do
        [[ -f "$source_file" ]] || fail "missing manifest input: $source_file"
        sha256sum -- "$source_file"
      done
  ) >"$output"
}

write_product_manifest() {
  local output="$1" source_file
  local inputs=(.dockerignore Dockerfile VERSION go.mod go.sum go.work go.work.sum)
  while IFS= read -r -d '' source_file; do
    inputs+=("$source_file")
  done < <(find cmd internal -type f \( -name '*.go' -o -path 'internal/device/profiles/*/profile.yaml' \) -print0)
  write_snapshot_manifest "$output" "${inputs[@]}"
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
  zip_url="$(jq -er --arg key "$key" '[.data[$key].items[0].links[]?.url | select(endswith("/UniFi.unix.zip"))] | if length == 1 then .[0] else error("expected one Unix ZIP") end' "$release_response")"
  [[ "$zip_url" =~ ^https://dl\.ui\.com/[A-Za-z0-9._/-]+/UniFi\.unix\.zip$ ]] ||
    fail "unsafe $channel ZIP URL"
  zip_release_token="${zip_url#https://dl.ui.com/unifi/}"
  zip_release_token="${zip_release_token%/UniFi.unix.zip}"
  case "$zip_release_token" in
  "$version" | "$version"-*) ;;
  *) fail "$channel ZIP URL does not bind the expected version" ;;
  esac
  zip_sha="$(jq -er --arg key "$key" '[.data[$key].items[0].instructions[]? | select(.type == "TEXT") | .content | match("([0-9a-f]{64}) \\| UniFi\\.unix\\.zip"; "g") | .captures[0].string] | if length == 1 then .[0] else error("expected one Unix ZIP SHA-256") end' "$release_response")"
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
  local finished_at result final_pending
  trap - EXIT HUP INT TERM
  set +e
  if [[ -n "${current_project:-}" ]] && ! project_is_clean "$current_project"; then
    compose_channel down --volumes --remove-orphans >/dev/null 2>&1 || cleanup_result=failed
    project_is_clean "$current_project" || cleanup_result=failed
  fi
  for image_alias in "${stable_controller_image:-}" "${rc_controller_image:-}" "${product_image:-}" "${helper_image:-}"; do
    if [[ -n "$image_alias" ]] && ! remove_owned_image_alias "$image_alias"; then
      cleanup_result=failed
    fi
  done
  if [[ "$command_status" -eq 0 && "$matrix_complete" == true && "$cleanup_result" == passed ]]; then
    mv "$accepted_binary_pending" "$accepted_binary" || {
      cleanup_result=failed
      rm -f "$accepted_binary_pending"
    }
  else
    rm -f "${accepted_binary_pending:-}"
  fi
  final_status=$command_status
  [[ "$matrix_complete" == true && "$cleanup_result" == passed ]] || final_status=1
  result=failed
  [[ "$final_status" -ne 0 ]] || result=passed
  finished_at="$(date -u '+%Y-%m-%dT%H:%M:%SZ')"
  final_pending="$evidence_dir/final.env.pending"
  if ! {
    printf 'RESULT=%s\n' "$result"
    printf 'RUN_ID=%s\n' "$run_id"
    printf 'STARTED_AT=%s\n' "$started_at"
    printf 'FINISHED_AT=%s\n' "$finished_at"
    printf 'COMMAND_EXIT_CODE=%s\n' "$command_status"
    printf 'EXIT_CODE=%s\n' "$final_status"
    printf 'CONTROLLER_RESOLVED_AT=%s\n' "${resolved_at:-not-resolved}"
    printf 'RUNNER_KIND=Build Runner\n'
    printf 'RUNNER_PROVIDER_VERSION=%s\n' "$runner_provider_version"
    printf 'RUNNER_PROVIDER_COMMIT=%s\n' "$runner_provider_commit"
    printf 'RUNNER_SESSION_ID=%s\n' "$runner_session_id"
    printf 'RUNNER_PROJECT_NAME=%s\n' "$runner_project_name"
    printf 'RUNNER_GO_VERSION=%s\n' "$runner_go_version"
    printf 'RUNNER_OS_ARCH=%s\n' "$runner_os_arch"
    printf 'GATE_COMMAND=%s\n' "$gate_command"
    printf 'PRODUCT_SOURCE_BINDING_SHA256=%s\n' "${product_source_binding:-unknown}"
    printf 'GATE_BINDING_SHA256=%s\n' "${gate_binding:-unknown}"
    printf 'PRODUCT_IMAGE_ID=%s\n' "${product_image_id:-unknown}"
    printf 'PRODUCT_VERSION=%s\n' "${candidate_version:-unknown}"
    printf 'PRODUCT_OS=%s\n' "${product_os:-unknown}"
    printf 'PRODUCT_ARCH=%s\n' "${product_arch:-unknown}"
    printf 'PRODUCT_GO_COMPILER_VERSION=%s\n' "${product_go_compiler_version:-unknown}"
    printf 'PRODUCT_BINARY_SHA256=%s\n' "${product_binary_sha:-unknown}"
    printf 'STABLE_VERSION=%s\n' "${stable_version:-not-resolved}"
    printf 'CONTROLLER_RUNTIME_BASE_DIGEST=%s\n' "${runtime_base_digest:-unknown}"
    printf 'STABLE_CONTROLLER_IMAGE_DIGEST=%s\n' "${stable_controller_image_digest:-unknown}"
    printf 'STABLE_RESULT=%s\n' "$stable_result"
    printf 'RELEASE_CANDIDATE_VERSION=%s\n' "${rc_version:-not-resolved}"
    printf 'RELEASE_CANDIDATE_CONTROLLER_IMAGE_DIGEST=%s\n' "${rc_controller_image_digest:-unknown}"
    printf 'RELEASE_CANDIDATE_RESULT=%s\n' "$rc_result"
    printf 'CONTROLLER_UI_RESULT=not-tested\n'
    printf 'CLEANUP=%s\n' "$cleanup_result"
  } >"$final_pending" || ! mv "$final_pending" "$evidence_dir/final.env"; then
    rm -f "$final_pending" "${accepted_binary:-}"
    echo "controller matrix: failed to write durable final evidence" >&2
    exit 1
  fi
  echo "controller matrix: result=$([[ "$final_status" -eq 0 ]] && echo passed || echo failed) cleanup=$cleanup_result evidence=$evidence_dir"
  exit "$final_status"
}

[[ "$repo_root" == /workspace ]] || fail 'run through Build Runner (make integration-nonunifi-metrics)'
[[ $# -eq 0 ]] || fail 'this command accepts no arguments'
: "${GOCACHE:?Build Runner must supply GOCACHE}"
for command in awk cmp curl docker find go jq sed sha256sum sort tee uname; do need "$command"; done
docker compose version >/dev/null 2>&1 || fail 'docker compose is required'
started_at="$(date -u '+%Y-%m-%dT%H:%M:%SZ')"
run_id="${UNIFI_STUB_CONTROLLER_MATRIX_RUN_ID:-matrix-$(date -u '+%Y%m%d%H%M%S')-$$}"
[[ "$run_id" =~ ^[a-z0-9][a-z0-9-]{0,39}$ ]] || fail 'unsafe matrix run ID'
evidence_dir="$repo_root/dist/controller-channel-matrix/$run_id"
mkdir -p "$(dirname -- "$evidence_dir")"
mkdir "$evidence_dir" || fail "evidence already exists: $evidence_dir"

product_image="unifi-stubd:product-$run_id"
helper_image="unifi-stubd-controller-adoption:gate-$run_id"
stable_controller_image="unifi-stubd-controller:$run_id-stable"
rc_controller_image="unifi-stubd-controller:$run_id-rc"
for image_alias in "$product_image" "$helper_image" "$stable_controller_image" "$rc_controller_image"; do
  image_alias_absent "$image_alias" || fail "owned image alias already exists: $image_alias"
done

matrix_complete=false
current_project=
stable_result=not-run
rc_result=not-run
stable_controller_image_digest=unknown
rc_controller_image_digest=unknown
runner_provider_version="${UNIFI_STUB_BUILD_RUNNER_VERSION:-unknown}"
runner_provider_commit="${UNIFI_STUB_BUILD_RUNNER_COMMIT:-unknown}"
runner_session_id="${BUILD_RUNNER_SESSION_ID:-unknown}"
runner_project_name="${BUILD_RUNNER_PROJECT_NAME:-unknown}"
runner_go_version="$(go version)"
runner_os_arch="$(uname -srm)"
gate_command='make integration-nonunifi-metrics'
trap cleanup EXIT
trap 'exit 129' HUP
trap 'exit 130' INT
trap 'exit 143' TERM

product_source_manifest="$evidence_dir/product-source.sha256"
gate_manifest="$evidence_dir/gate-source.sha256"
gate_inputs=(
  .dockerignore Makefile go.mod
  build-runner/instances/controller-gate.yaml
  build-runner/toolsets/controller-gate/image-setup.sh
  lab/stub/Dockerfile.controller-adoption
  lab/stub/Dockerfile.controller-channel
  lab/stub/compose.controller-channel.yaml
  lab/stub/compose.tests.yaml
  lab/stub/compose.yaml
  lab/stub/configs/hosts/stub-bridge-observe/config.yaml
  lab/stub/mongo-init-unifi.sh
  lab/stub/scripts/run-bridge-observe.sh
  lab/stub/scripts/run-controller-channel-matrix.sh
  lab/stub/scripts/run-docker-integration.sh
  lab/stub/scripts/setup-bridge-observe.sh
  lab/stub/tools/controller-adoption/main.go
)
write_product_manifest "$product_source_manifest"
write_snapshot_manifest "$gate_manifest" "${gate_inputs[@]}"
product_source_binding="$(sha256sum -- "$product_source_manifest" | awk '{print $1}')"
gate_binding="$(sha256sum -- "$gate_manifest" | awk '{print $1}')"

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
export UNIFI_STUBD_VERSION="$candidate_version"
export UNIFI_STUB_BRIDGE_MAC=02:15:6d:00:08:21
export UNIFI_STUB_BRIDGE_IP=172.31.242.25
UNIFI_NETWORK_IMAGE="$runtime_base_digest" \
  COMPOSE_PROJECT_NAME="unifi-stubd-build-$run_id" docker compose \
  -f "$repo_root/lab/stub/compose.yaml" \
  -f "$repo_root/lab/stub/compose.tests.yaml" \
  build stub-bridge-observe controller-adoption
product_image_id="$(docker image inspect --format '{{.Id}}' "$product_image")"
[[ "$product_image_id" =~ ^sha256:[0-9a-f]{64}$ ]] || fail 'invalid product image digest'
product_os="$(docker image inspect --format '{{.Os}}' "$product_image_id")"
product_arch="$(docker image inspect --format '{{.Architecture}}' "$product_image_id")"
[[ "$product_os" == linux && "$product_arch" =~ ^(amd64|arm64)$ ]] ||
  fail "unsupported product platform: $product_os/$product_arch"
run_lane() {
  local channel="$1" expected_version="$2" zip_url="$3" zip_sha="$4" controller_image="$5" lane_suffix="$6"
  local lane_status controller_image_digest
  current_project="unifi-stubd-$run_id-$lane_suffix"
  export COMPOSE_PROJECT_NAME="$current_project"
  export UNIFI_STUB_CONTROLLER_RUNTIME_BASE_DIGEST="$runtime_base_digest"
  export UNIFI_STUB_CONTROLLER_ZIP_URL="$zip_url"
  export UNIFI_STUB_CONTROLLER_ZIP_SHA256="$zip_sha"
  export UNIFI_STUB_CONTROLLER_IMAGE="$controller_image"
  export UNIFI_NETWORK_IMAGE="$controller_image"
  export UNIFI_STUB_LAB_EXPECTED_NETWORK_VERSION="$expected_version"
  export UNIFI_STUB_DOCKER_EXPECTED_PRODUCT_IMAGE_ID="$product_image_id"
  compose_channel build unifi || fail "$channel controller image build failed"
  controller_image_digest="$(docker image inspect --format '{{.Id}}' "$controller_image")"
  [[ "$controller_image_digest" =~ ^sha256:[0-9a-f]{64}$ ]] ||
    fail "$channel controller image has no immutable digest"
  if [[ "$channel" == stable ]]; then
    stable_controller_image_digest="$controller_image_digest"
  else
    rc_controller_image_digest="$controller_image_digest"
  fi
  set +e
  sh "$script_dir/run-docker-integration.sh" 2>&1 | tee "$evidence_dir/$channel.log"
  lane_status=${PIPESTATUS[0]}
  set -e
  [[ "$lane_status" -eq 0 ]] || return "$lane_status"
  project_is_clean "$current_project" || fail "$channel Compose project was not completely removed"
  [[ "$(docker image inspect --format '{{.Id}}' "$product_image")" == "$product_image_id" ]] ||
    fail 'the product image tag changed between controller lanes'
  [[ "$(docker image inspect --format '{{.Id}}' "$controller_image")" == "$controller_image_digest" ]] ||
    fail "$channel controller image tag changed during its lane"
  remove_owned_image_alias "$controller_image" ||
    fail "$channel owned controller image alias was not removed"
  current_project=
}

run_lane stable "$stable_version" "$stable_zip_url" "$stable_zip_sha" "$stable_controller_image" ga ||
  fail 'Stable product lane failed'
stable_result=passed
run_lane release-candidate "$rc_version" "$rc_zip_url" "$rc_zip_sha" "$rc_controller_image" rc ||
  fail 'Release Candidate product lane failed'
rc_result=passed

write_product_manifest "$evidence_dir/product-source.after.sha256"
write_snapshot_manifest "$evidence_dir/gate-source.after.sha256" "${gate_inputs[@]}"
cmp -s "$product_source_manifest" "$evidence_dir/product-source.after.sha256" ||
  fail 'product source changed during the controller gate'
cmp -s "$gate_manifest" "$evidence_dir/gate-source.after.sha256" ||
  fail 'gate source changed during the controller gate'
rm -f "$evidence_dir/product-source.after.sha256" "$evidence_dir/gate-source.after.sha256"

accepted_binary="$evidence_dir/unifi-stubd-$product_os-$product_arch"
accepted_binary_pending="$accepted_binary.pending"
docker run --rm --entrypoint cat "$product_image_id" /usr/local/bin/unifi-stubd >"$accepted_binary_pending"
chmod 0755 "$accepted_binary_pending"
product_go_compiler_version="$(go version -m "$accepted_binary_pending" | awk 'NR == 1 { print $2 }')"
[[ "$product_go_compiler_version" =~ ^go[0-9]+([.][0-9]+){1,2}([a-z]+[0-9]+)?$ ]] ||
  fail 'accepted product binary has no valid Go compiler version'
product_binary_sha="$(sha256sum -- "$accepted_binary_pending" | awk '{print $1}')"
[[ "$product_binary_sha" =~ ^[0-9a-f]{64}$ ]] || fail 'invalid accepted product binary digest'
matrix_complete=true
echo "controller matrix: same product image passed the product gate on Stable $stable_version and RC $rc_version"

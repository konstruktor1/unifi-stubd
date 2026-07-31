#!/bin/sh
set -eu

script_dir="$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)"
repo_root="$(CDPATH='' cd -- "$script_dir/../../.." && pwd)"
if [ "$repo_root" != /workspace ]; then
  echo "run the Docker lab through Build Runner (make integration-nonunifi-metrics)" >&2
  exit 2
fi

resources_cleaned=0
COMPOSE_PROJECT_NAME="${COMPOSE_PROJECT_NAME:-unifi-stubd-controller-gate}"
export COMPOSE_PROJECT_NAME

: "${UNIFI_NETWORK_IMAGE:?the channel matrix must supply a resolved controller image}"
: "${UNIFI_STUB_LAB_EXPECTED_NETWORK_VERSION:?the channel matrix must supply the resolved controller version}"
: "${UNIFI_STUB_DOCKER_EXPECTED_PRODUCT_IMAGE_ID:?the channel matrix must supply the product image ID}"
export UNIFI_NETWORK_IMAGE UNIFI_STUB_LAB_EXPECTED_NETWORK_VERSION
compose() {
  docker compose \
    -f "$repo_root/lab/stub/compose.yaml" \
    -f "$repo_root/lab/stub/compose.tests.yaml" \
    -f "$repo_root/lab/stub/compose.controller-channel.yaml" \
    "$@"
}

verify_product_image() {
  container_id="$(compose ps -q stub-bridge-observe)"
  [ -n "$container_id" ] || {
    echo "docker integration: product container is missing" >&2
    return 1
  }
  actual="$(docker container inspect --format '{{.Image}}' "$container_id")"
  if [ "$actual" != "$UNIFI_STUB_DOCKER_EXPECTED_PRODUCT_IMAGE_ID" ]; then
    echo "docker integration: product image mismatch" >&2
    return 1
  fi
  echo "docker integration: product image verified id=$actual"
}

controller_api() {
  compose run --rm --no-deps controller-adoption "$@"
}

cleanup_runtime_resources() {
  if [ "${resources_cleaned:-0}" = "1" ]; then
    return
  fi
  if ! compose down --volumes --remove-orphans; then
    echo "docker integration: failed to remove disposable Compose project=$COMPOSE_PROJECT_NAME" >&2
    return 2
  fi
  resources_cleaned=1
}

cleanup() {
  status=$?
  set +e
  [ "$status" -eq 0 ] || compose logs --no-color --tail 200 unifi-db unifi stub-bridge-observe
  cleanup_runtime_resources
  cleanup_status=$?
  if [ "$status" -ne 0 ]; then
    return "$status"
  fi
  return "$cleanup_status"
}

wait_for_controller_version() {
  deadline=$(($(date +%s) + ${UNIFI_STUB_DOCKER_CONTROLLER_TIMEOUT:-240}))
  last_error=""
  while [ "$(date +%s)" -le "$deadline" ]; do
    if output="$(controller_api version --expected-version "$UNIFI_STUB_LAB_EXPECTED_NETWORK_VERSION" 2>&1)"; then
      printf '%s\n' "$output"
      return 0
    fi
    last_error="$output"
    sleep 3
  done
  [ -z "$last_error" ] || printf '%s\n' "$last_error" >&2
  return 1
}

run_switch_traffic_gate() {
  echo "docker integration: generating bounded bidirectional switch uplink traffic"
  expected_speed="$(compose exec -T stub-bridge-observe cat /sys/class/net/uplink0/speed | tr -d '\r')"
  case "$expected_speed" in '' | *[!0-9]*)
    echo "uplink0 did not expose a positive observed speed" >&2
    return 1
    ;;
  esac
  [ "$expected_speed" -gt 0 ] || {
    echo "uplink0 did not expose a positive observed speed" >&2
    return 1
  }
  # shellcheck disable=SC2016
  compose exec -T stub-bridge-observe sh -ec '
        ping -I "$1" -c 360 -i 0.2 -s 1200 -W 1 "$3" >/dev/null 2>&1 &
        ingress_pid=$!
        ping -I "$2" -c 360 -i 0.2 -s 800 -W 1 "$4" >/dev/null 2>&1 &
        egress_pid=$!
        wait "$ingress_pid" || true
        wait "$egress_pid" || true
    ' sh 198.51.100.2 198.51.100.1 203.0.113.1 203.0.113.2 &
  traffic_pid=$!

  if ! controller_api wait-switch-telemetry \
    --mac "$UNIFI_STUB_BRIDGE_MAC" \
    --expected-uplink-speed "$expected_speed" \
    --timeout 65; then
    wait "$traffic_pid" || true
    return 1
  fi
  wait "$traffic_pid"
}

run_adoption_gate() {
  echo "docker integration: controller US8 adoption gate"
  compose up -d stub-bridge-observe
  verify_product_image
  controller_api wait-visible --mac "$UNIFI_STUB_BRIDGE_MAC" --timeout 90
  controller_api adopt --mac "$UNIFI_STUB_BRIDGE_MAC"
}

trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

echo "docker integration: test identity US8=$UNIFI_STUB_BRIDGE_MAC/$UNIFI_STUB_BRIDGE_IP"
echo "docker integration: resetting disposable Compose project=$COMPOSE_PROJECT_NAME"
compose down --volumes --remove-orphans

echo "docker integration: starting controller image=$UNIFI_NETWORK_IMAGE"
compose up -d unifi
wait_for_controller_version
controller_api bootstrap

run_adoption_gate
run_switch_traffic_gate

echo "docker integration: cleaning product fixture"
cleanup_runtime_resources
echo "docker integration: ok"

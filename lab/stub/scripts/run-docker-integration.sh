#!/bin/sh
set -eu

script_dir="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
repo_root="$(CDPATH= cd -- "$script_dir/../../.." && pwd)"
captures_dir="$repo_root/lab/stub/captures"
payload_dir="${TMPDIR:-/tmp}/unifi-stubd-docker-integration.$$"
resources_cleaned=0
default_unifi_network_image="lscr.io/linuxserver/unifi-network-application:10.3.58-ls129"

if [ -z "${UNIFI_NETWORK_IMAGE:-}" ]; then
    export UNIFI_NETWORK_IMAGE="$default_unifi_network_image"
    export UNIFI_STUB_LAB_EXPECTED_NETWORK_VERSION="${UNIFI_STUB_LAB_EXPECTED_NETWORK_VERSION:-10.3.58}"
else
    export UNIFI_NETWORK_IMAGE
fi

compose() {
    docker compose \
        -f "$repo_root/lab/stub/compose.yaml" \
        -f "$repo_root/lab/stub/compose.tests.yaml" \
        "$@"
}

format_lab_mac() {
    value="$1"
    octet4=$((value / 65536 % 256))
    octet5=$((value / 256 % 256))
    octet6=$((value % 256))
    printf '02:15:6d:%02x:%02x:%02x' "$octet4" "$octet5" "$octet6"
}

test_volume_name() {
    suffix="$1"
    docker volume ls -q | grep "_${suffix}$" | head -n 1 || true
}

reset_test_service() {
    service="$1"
    volume_suffix="$2"
    compose stop "$service" >/dev/null 2>&1 || true
    compose rm -fsv "$service" >/dev/null 2>&1 || true
    volume="$(test_volume_name "$volume_suffix")"
    if [ -n "$volume" ]; then
        docker volume rm "$volume" >/dev/null 2>&1 || true
    fi
}

cleanup_runtime_resources() {
    if [ "${resources_cleaned:-0}" = "1" ]; then
        return
    fi
    resources_cleaned=1
    if [ "${UNIFI_STUB_DOCKER_KEEP_RESOURCES:-0}" = "1" ]; then
        return
    fi
    reset_test_service stub-bridge-observe stub-bridge-observe-state
    reset_test_service stub-port-map stub-port-map-state
    reset_test_service stub-gateway-smoke stub-gateway-smoke-state
    if [ -n "${UNIFI_STUB_BRIDGE_MAC:-}" ]; then
        controller_api forget --mac "$UNIFI_STUB_BRIDGE_MAC" >/dev/null 2>&1 || true
        controller_api wait-clean --mac "$UNIFI_STUB_BRIDGE_MAC" --timeout "${UNIFI_STUB_DOCKER_CLEANUP_TIMEOUT:-30}" >/dev/null 2>&1 || true
    fi
    if [ -n "${UNIFI_STUB_PORTMAP_MAC:-}" ]; then
        controller_api forget --mac "$UNIFI_STUB_PORTMAP_MAC" >/dev/null 2>&1 || true
        controller_api wait-clean --mac "$UNIFI_STUB_PORTMAP_MAC" --timeout "${UNIFI_STUB_DOCKER_CLEANUP_TIMEOUT:-30}" >/dev/null 2>&1 || true
    fi
    if [ -n "${UNIFI_STUB_GATEWAY_MAC:-}" ]; then
        controller_api forget --mac "$UNIFI_STUB_GATEWAY_MAC" >/dev/null 2>&1 || true
        controller_api wait-clean --mac "$UNIFI_STUB_GATEWAY_MAC" --timeout "${UNIFI_STUB_DOCKER_CLEANUP_TIMEOUT:-30}" >/dev/null 2>&1 || true
    fi
}

cleanup() {
    status=$?
    set +e
    cleanup_runtime_resources
    rm -rf "$payload_dir"
    return "$status"
}

event_line_count() {
    events_path="$captures_dir/events.jsonl"
    if [ ! -f "$events_path" ]; then
        printf '0\n'
        return
    fi
    wc -l < "$events_path" | tr -d ' '
}

wait_for_events() {
    start_line="$1"
    shift
    deadline=$(($(date +%s) + ${UNIFI_STUB_DOCKER_EVENT_TIMEOUT:-60}))
    while [ "$(date +%s)" -le "$deadline" ]; do
        if python3 "$script_dir/assert-events.py" "$captures_dir/events.jsonl" "$start_line" "$@"; then
            return 0
        fi
        sleep 2
    done
    python3 "$script_dir/assert-events.py" "$captures_dir/events.jsonl" "$start_line" "$@"
}

wait_for_event_count() {
    start_line="$1"
    min_count="$2"
    shift 2
    deadline=$(($(date +%s) + ${UNIFI_STUB_DOCKER_EVENT_TIMEOUT:-60}))
    while [ "$(date +%s)" -le "$deadline" ]; do
        if python3 "$script_dir/assert-events.py" "$captures_dir/events.jsonl" "$start_line" --min-count "$min_count" "$@"; then
            return 0
        fi
        sleep 2
    done
    python3 "$script_dir/assert-events.py" "$captures_dir/events.jsonl" "$start_line" --min-count "$min_count" "$@"
}

wait_for_mitm() {
    deadline=$(($(date +%s) + ${UNIFI_STUB_DOCKER_MITM_TIMEOUT:-90}))
    while [ "$(date +%s)" -le "$deadline" ]; do
        if compose run --rm --no-deps --entrypoint /bin/sh stub-port-map -ec 'nc -z -w 1 unifi 8080' >/dev/null 2>&1; then
            return 0
        fi
        sleep 2
    done
    echo "inform MITM did not become reachable on unifi:8080" >&2
    return 1
}

wait_for_controller_version() {
    deadline=$(($(date +%s) + ${UNIFI_STUB_DOCKER_CONTROLLER_TIMEOUT:-180}))
    last_error=""
    while [ "$(date +%s)" -le "$deadline" ]; do
        if [ -n "${UNIFI_STUB_LAB_EXPECTED_NETWORK_VERSION:-}" ]; then
            if output="$(controller_api version --expected-version "$UNIFI_STUB_LAB_EXPECTED_NETWORK_VERSION" 2>&1)"; then
                printf '%s\n' "$output"
                return 0
            fi
            last_error="$output"
        else
            if output="$(controller_api version 2>&1)"; then
                printf '%s\n' "$output"
                return 0
            fi
            last_error="$output"
        fi
        sleep 3
    done
    if [ -n "$last_error" ]; then
        printf '%s\n' "$last_error" >&2
    fi
    if [ -n "${UNIFI_STUB_LAB_EXPECTED_NETWORK_VERSION:-}" ]; then
        controller_api version --expected-version "$UNIFI_STUB_LAB_EXPECTED_NETWORK_VERSION"
    else
        controller_api version
    fi
}

send_one_shot_until_event() {
    service="$1"
    mac="$2"
    start_line="$3"
    attempts="${UNIFI_STUB_DOCKER_INFORM_ATTEMPTS:-5}"
    attempt=1
    while [ "$attempt" -le "$attempts" ]; do
        compose run --rm "$service" -once || true
        if python3 "$script_dir/assert-events.py" "$captures_dir/events.jsonl" "$start_line" "$mac" >/dev/null 2>&1; then
            python3 "$script_dir/assert-events.py" "$captures_dir/events.jsonl" "$start_line" "$mac"
            return 0
        fi
        attempt=$((attempt + 1))
        sleep 2
    done
    python3 "$script_dir/assert-events.py" "$captures_dir/events.jsonl" "$start_line" "$mac"
}

controller_api() {
    python3 "$script_dir/controller-adoption.py" "$@"
}

wait_for_stub_adoption_state() {
    service="$1"
    deadline=$(($(date +%s) + ${UNIFI_STUB_DOCKER_ADOPTION_TIMEOUT:-90}))
    while [ "$(date +%s)" -le "$deadline" ]; do
        if compose exec -T "$service" sh -lc 'test -f /var/lib/unifi-stubd/adoption.env && grep -q "^STATE=connected$" /var/lib/unifi-stubd/adoption.env && grep -q "^AUTHKEY=." /var/lib/unifi-stubd/adoption.env' >/dev/null 2>&1; then
            echo "stub adoption state persisted: service=$service state=connected authkey_set=true"
            return 0
        fi
        sleep 2
    done
    echo "stub adoption state did not become connected: service=$service" >&2
    return 1
}

if ! command -v docker >/dev/null 2>&1; then
    echo "docker is required" >&2
    exit 2
fi
if ! docker compose version >/dev/null 2>&1; then
    echo "docker compose is required" >&2
    exit 2
fi
if ! command -v python3 >/dev/null 2>&1; then
    echo "python3 is required" >&2
    exit 2
fi

mkdir -p "$captures_dir" "$payload_dir"
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

run_seed="$(date +%s)"
export UNIFI_STUB_BRIDGE_MAC="${UNIFI_STUB_BRIDGE_MAC:-$(format_lab_mac "$run_seed")}"
export UNIFI_STUB_PORTMAP_MAC="${UNIFI_STUB_PORTMAP_MAC:-$(format_lab_mac "$((run_seed + 1))")}"
export UNIFI_STUB_GATEWAY_MAC="${UNIFI_STUB_GATEWAY_MAC:-$(format_lab_mac "$((run_seed + 2))")}"
export UNIFI_STUB_BRIDGE_IP="${UNIFI_STUB_BRIDGE_IP:-172.31.242.$((80 + run_seed % 60))}"
export UNIFI_STUB_PORTMAP_IP="${UNIFI_STUB_PORTMAP_IP:-172.31.242.$((150 + run_seed % 60))}"
export UNIFI_STUB_GATEWAY_IP="${UNIFI_STUB_GATEWAY_IP:-172.31.242.$((30 + run_seed % 40))}"
echo "docker integration: test identities bridge=$UNIFI_STUB_BRIDGE_MAC/$UNIFI_STUB_BRIDGE_IP port-map=$UNIFI_STUB_PORTMAP_MAC/$UNIFI_STUB_PORTMAP_IP gateway=$UNIFI_STUB_GATEWAY_MAC/$UNIFI_STUB_GATEWAY_IP"

echo "docker integration: resetting temporary test services"
reset_test_service stub-bridge-observe stub-bridge-observe-state
reset_test_service stub-port-map stub-port-map-state
reset_test_service stub-gateway-smoke stub-gateway-smoke-state

echo "docker integration: validating compose configuration"
compose config >/dev/null

echo "docker integration: building test image"
compose build stub-bridge-observe stub-port-map stub-gateway-smoke

echo "docker integration: dry-run bridge-observe payload"
compose run --rm --no-deps stub-bridge-observe -dry-run > "$payload_dir/bridge-observe.txt"
python3 "$script_dir/assert-payload.py" bridge-observe "$payload_dir/bridge-observe.txt"

echo "docker integration: dry-run management-lan preexisting-interface payload"
compose run --rm --no-deps stub-bridge-observe \
    -dry-run \
    -management-lan-enabled \
    -management-lan-vlan 42 \
    -management-lan-mode preexisting-interface \
    -management-lan-interface eth0 \
    -management-lan-ip "$UNIFI_STUB_BRIDGE_IP" \
    > "$payload_dir/management-lan.txt"
python3 "$script_dir/assert-payload.py" management-lan "$payload_dir/management-lan.txt"

echo "docker integration: dry-run port-map payload"
compose run --rm --no-deps stub-port-map -dry-run > "$payload_dir/port-map.txt"
python3 "$script_dir/assert-payload.py" port-map "$payload_dir/port-map.txt"

echo "docker integration: dry-run gateway payload"
compose run --rm --no-deps stub-gateway-smoke -dry-run > "$payload_dir/gateway-smoke.txt"
python3 "$script_dir/assert-payload.py" gateway-smoke "$payload_dir/gateway-smoke.txt"

echo "docker integration: starting existing controller/MITM lab services"
compose up -d inform-mitm
wait_for_mitm
echo "docker integration: controller image=$UNIFI_NETWORK_IMAGE"
wait_for_controller_version
controller_api forget --mac "$UNIFI_STUB_BRIDGE_MAC" >/dev/null 2>&1 || true
controller_api forget --mac "$UNIFI_STUB_PORTMAP_MAC" >/dev/null 2>&1 || true
controller_api forget --mac "$UNIFI_STUB_GATEWAY_MAC" >/dev/null 2>&1 || true
controller_api wait-clean --mac "$UNIFI_STUB_BRIDGE_MAC" --timeout "${UNIFI_STUB_DOCKER_ADOPTION_TIMEOUT:-90}"
controller_api wait-clean --mac "$UNIFI_STUB_PORTMAP_MAC" --timeout "${UNIFI_STUB_DOCKER_ADOPTION_TIMEOUT:-90}"
controller_api wait-clean --mac "$UNIFI_STUB_GATEWAY_MAC" --timeout "${UNIFI_STUB_DOCKER_ADOPTION_TIMEOUT:-90}"

start_line="$(event_line_count)"

echo "docker integration: one-shot bridge-observe inform"
send_one_shot_until_event \
    stub-bridge-observe \
    "$UNIFI_STUB_BRIDGE_MAC" \
    "$start_line"

echo "docker integration: one-shot port-map inform"
send_one_shot_until_event \
    stub-port-map \
    "$UNIFI_STUB_PORTMAP_MAC" \
    "$start_line"

echo "docker integration: one-shot gateway inform"
send_one_shot_until_event \
    stub-gateway-smoke \
    "$UNIFI_STUB_GATEWAY_MAC" \
    "$start_line"

echo "docker integration: checking MITM inform events"
wait_for_events "$start_line" \
    "$UNIFI_STUB_BRIDGE_MAC" \
    "$UNIFI_STUB_PORTMAP_MAC" \
    "$UNIFI_STUB_GATEWAY_MAC"

echo "docker integration: controller switch adoption smoke"
reset_test_service stub-bridge-observe stub-bridge-observe-state
controller_api forget --mac "$UNIFI_STUB_BRIDGE_MAC" >/dev/null 2>&1 || true
controller_api wait-clean --mac "$UNIFI_STUB_BRIDGE_MAC" --timeout "${UNIFI_STUB_DOCKER_ADOPTION_TIMEOUT:-90}"
compose up -d stub-bridge-observe
controller_api wait-pending --mac "$UNIFI_STUB_BRIDGE_MAC" --timeout "${UNIFI_STUB_DOCKER_ADOPTION_TIMEOUT:-90}"
controller_api adopt --mac "$UNIFI_STUB_BRIDGE_MAC"
controller_api wait-adopted --mac "$UNIFI_STUB_BRIDGE_MAC" --timeout "${UNIFI_STUB_DOCKER_ADOPTION_TIMEOUT:-90}"
switch_post_adopt_line="$(event_line_count)"
wait_for_stub_adoption_state stub-bridge-observe
wait_for_event_count "$switch_post_adopt_line" 1 "$UNIFI_STUB_BRIDGE_MAC"

echo "docker integration: controller gateway adoption smoke"
reset_test_service stub-gateway-smoke stub-gateway-smoke-state
controller_api forget --mac "$UNIFI_STUB_GATEWAY_MAC" >/dev/null 2>&1 || true
controller_api wait-clean --mac "$UNIFI_STUB_GATEWAY_MAC" --timeout "${UNIFI_STUB_DOCKER_ADOPTION_TIMEOUT:-90}"
compose up -d stub-gateway-smoke
controller_api wait-pending --mac "$UNIFI_STUB_GATEWAY_MAC" --timeout "${UNIFI_STUB_DOCKER_ADOPTION_TIMEOUT:-90}"
controller_api adopt --mac "$UNIFI_STUB_GATEWAY_MAC"
controller_api wait-adopted --mac "$UNIFI_STUB_GATEWAY_MAC" --timeout "${UNIFI_STUB_DOCKER_ADOPTION_TIMEOUT:-90}"
gateway_post_adopt_line="$(event_line_count)"
wait_for_stub_adoption_state stub-gateway-smoke
wait_for_event_count "$gateway_post_adopt_line" 1 "$UNIFI_STUB_GATEWAY_MAC"

echo "docker integration: cleaning temporary test resources"
cleanup_runtime_resources

echo "docker integration: ok"

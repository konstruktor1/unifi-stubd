#!/bin/sh
set -eu

script_dir="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
repo_root="$(CDPATH= cd -- "$script_dir/../../.." && pwd)"
captures_dir="$repo_root/lab/stub/captures"
payload_dir="${TMPDIR:-/tmp}/unifi-stubd-docker-integration.$$"
resources_cleaned=0
go_cmd="${GO:-go}"
default_unifi_network_image="lscr.io/linuxserver/unifi-network-application:10.4.57-ls135"
default_gateway_command_expectation="absent"

if [ -z "${UNIFI_NETWORK_IMAGE:-}" ]; then
    export UNIFI_NETWORK_IMAGE="$default_unifi_network_image"
    export UNIFI_STUB_LAB_EXPECTED_NETWORK_VERSION="${UNIFI_STUB_LAB_EXPECTED_NETWORK_VERSION:-10.4.57}"
else
    export UNIFI_NETWORK_IMAGE
fi
export UNIFI_STUB_INFORM_PROXY_TARGET="${UNIFI_STUB_INFORM_PROXY_TARGET:-http://unifi-controller:8080}"
export UNIFI_STUB_DISCOVERY_PROXY_TARGET="${UNIFI_STUB_DISCOVERY_PROXY_TARGET:-unifi-controller:10001}"
gateway_command_expectation="${UNIFI_STUB_DOCKER_GATEWAY_COMMAND_EXPECTATION:-$default_gateway_command_expectation}"

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
    for mac in "${UNIFI_STUB_BRIDGE_MAC:-}" "${UNIFI_STUB_PORTMAP_MAC:-}" "${UNIFI_STUB_GATEWAY_MAC:-}"; do
        if [ -n "$mac" ]; then
            controller_api forget --mac "$mac" >/dev/null 2>&1 || true
            controller_api wait-clean --mac "$mac" --timeout "${UNIFI_STUB_DOCKER_CLEANUP_TIMEOUT:-30}" >/dev/null 2>&1 || true
        fi
    done
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
        if assert_events "$captures_dir/events.jsonl" "$start_line" "$@" >/dev/null 2>&1; then
            assert_events "$captures_dir/events.jsonl" "$start_line" "$@"
            return 0
        fi
        sleep 2
    done
    assert_events "$captures_dir/events.jsonl" "$start_line" "$@"
}

wait_for_event_kind() {
    start_line="$1"
    kind="$2"
    shift 2
    deadline=$(($(date +%s) + ${UNIFI_STUB_DOCKER_EVENT_TIMEOUT:-90}))
    while [ "$(date +%s)" -le "$deadline" ]; do
        if assert_events "$captures_dir/events.jsonl" "$start_line" --kind "$kind" "$@" >/dev/null 2>&1; then
            assert_events "$captures_dir/events.jsonl" "$start_line" --kind "$kind" "$@"
            return 0
        fi
        sleep 2
    done
    assert_events "$captures_dir/events.jsonl" "$start_line" --kind "$kind" "$@"
}

wait_for_event_count() {
    start_line="$1"
    kind="$2"
    min_count="$3"
    shift 3
    deadline=$(($(date +%s) + ${UNIFI_STUB_DOCKER_EVENT_TIMEOUT:-90}))
    while [ "$(date +%s)" -le "$deadline" ]; do
        if assert_events "$captures_dir/events.jsonl" "$start_line" --kind "$kind" --min-count "$min_count" "$@" >/dev/null 2>&1; then
            assert_events "$captures_dir/events.jsonl" "$start_line" --kind "$kind" --min-count "$min_count" "$@"
            return 0
        fi
        sleep 2
    done
    assert_events "$captures_dir/events.jsonl" "$start_line" --kind "$kind" --min-count "$min_count" "$@"
}

wait_for_inform_proxy() {
    deadline=$(($(date +%s) + ${UNIFI_STUB_DOCKER_MITM_TIMEOUT:-90}))
    while [ "$(date +%s)" -le "$deadline" ]; do
        if compose run --rm --no-deps --entrypoint /bin/sh stub-port-map -ec 'nc -z -w 1 unifi 8080' >/dev/null 2>&1; then
            return 0
        fi
        sleep 2
    done
    echo "inform proxy did not become reachable on unifi:8080" >&2
    return 1
}

wait_for_controller_version() {
    deadline=$(($(date +%s) + ${UNIFI_STUB_DOCKER_CONTROLLER_TIMEOUT:-240}))
    last_error=""
    while [ "$(date +%s)" -le "$deadline" ]; do
        if [ -n "${UNIFI_STUB_LAB_EXPECTED_NETWORK_VERSION:-}" ]; then
            if output="$(controller_api version --expected-version "$UNIFI_STUB_LAB_EXPECTED_NETWORK_VERSION" 2>&1)"; then
                printf '%s\n' "$output"
                return 0
            fi
        else
            if output="$(controller_api version 2>&1)"; then
                printf '%s\n' "$output"
                return 0
            fi
        fi
        last_error="$output"
        sleep 3
    done
    if [ -n "$last_error" ]; then
        printf '%s\n' "$last_error" >&2
    fi
    return 1
}

send_one_shot_until_event() {
    service="$1"
    mac="$2"
    start_line="$3"
    attempts="${UNIFI_STUB_DOCKER_INFORM_ATTEMPTS:-5}"
    attempt=1
    while [ "$attempt" -le "$attempts" ]; do
        compose run --rm --no-deps "$service" -once || true
        if assert_events "$captures_dir/events.jsonl" "$start_line" "$mac" >/dev/null 2>&1; then
            assert_events "$captures_dir/events.jsonl" "$start_line" "$mac"
            return 0
        fi
        attempt=$((attempt + 1))
        sleep 2
    done
    assert_events "$captures_dir/events.jsonl" "$start_line" "$mac"
}

go_tool() {
    tool="$1"
    shift
    (cd "$repo_root" && "$go_cmd" run "./lab/stub/tools/$tool" "$@")
}

assert_events() {
    go_tool assert-events "$@"
}

assert_payload() {
    go_tool assert-payload "$@"
}

assert_status() {
    go_tool assert-status "$@"
}

controller_api() {
    go_tool controller-adoption "$@"
}

reset_controller_fixtures() {
    timeout="${UNIFI_STUB_DOCKER_ADOPTION_TIMEOUT:-90}"
    if output="$(controller_api reset-fixtures --timeout "$timeout" 2>&1)"; then
        printf '%s\n' "$output"
        case "$output" in
            *adopted_removed=0)
                ;;
            *)
                echo "docker integration: restarting disposable controller to clear removed gateway state"
                compose restart unifi
                wait_for_controller_version
                ;;
        esac
        return 0
    fi
    case "$output" in
        *api.err.DeviceBusy*)
            echo "docker integration: restarting disposable controller to release busy test fixtures"
            compose restart unifi
            wait_for_controller_version
            controller_api reset-fixtures --timeout "$timeout"
            ;;
        *)
            printf '%s\n' "$output" >&2
            return 1
            ;;
    esac
}

wait_for_stub_adoption_state() {
    service="$1"
    deadline=$(($(date +%s) + ${UNIFI_STUB_DOCKER_ADOPTION_TIMEOUT:-90}))
    while [ "$(date +%s)" -le "$deadline" ]; do
        if compose exec -T "$service" sh -ec '
            state=/var/lib/unifi-stubd/adoption.env
            status=/var/lib/unifi-stubd/status.json
            test -f "$state"
            test -f "$status"
            grep -q "^STATE=connected$" "$state"
            grep -Eq "^AUTHKEY=.+$" "$state"
            grep -Eq "^CFGVERSION=.+$" "$state"
            grep -q "^USE_AES_GCM=true$" "$state"
            grep -q "\"status_code\": 200" "$status"
            grep -q "\"controller_state\": \"connected\"" "$status"
            grep -q "\"used_aes_gcm\": true" "$status"
        ' >/dev/null 2>&1; then
            echo "stub adoption evidence: service=$service state=connected authkey_set=true cfgversion_set=true use_aes_gcm=true"
            return 0
        fi
        sleep 2
    done
    echo "stub adoption state did not become connected with AES-GCM evidence: service=$service" >&2
    compose exec -T "$service" sh -ec 'sed -E "s/^(AUTHKEY=).*/\\1<redacted>/" /var/lib/unifi-stubd/adoption.env 2>/dev/null || true; cat /var/lib/unifi-stubd/status.json 2>/dev/null || true' >&2 || true
    return 1
}

gateway_status_json() {
    compose exec -T stub-gateway-smoke \
        /usr/local/bin/unifi-stubd \
        -status-json \
        -config /usr/local/share/unifi-stubd-lab/configs/hosts/stub-gateway-smoke/config.yaml
}

wait_for_gateway_command_lifecycle() {
    case "$gateway_command_expectation" in
        required)
            assertion="wg-genkeypair-complete"
            deadline=$(($(date +%s) + ${UNIFI_STUB_DOCKER_COMMAND_TIMEOUT:-90}))
            while [ "$(date +%s)" -le "$deadline" ]; do
                if gateway_status_json | assert_status "$assertion" >/dev/null 2>&1; then
                    break
                fi
                sleep 2
            done
            ;;
        absent)
            assertion="no-command-stable"
            ;;
    esac
    gateway_status_json | assert_status "$assertion"
    sleep "${UNIFI_STUB_DOCKER_COMMAND_STABILITY_SECONDS:-75}"
    gateway_status_json | assert_status "$assertion"
}

run_controller_adoption_gate() {
    label="$1"
    service="$2"
    volume_suffix="$3"
    mac="$4"

    echo "docker integration: controller $label adoption gate"
    reset_test_service "$service" "$volume_suffix"
    controller_api forget --mac "$mac" >/dev/null 2>&1 || true
    controller_api wait-clean --mac "$mac" --timeout "${UNIFI_STUB_DOCKER_ADOPTION_TIMEOUT:-90}"
    adoption_start_line="$(event_line_count)"
    compose up -d "$service"
    wait_for_event_kind "$adoption_start_line" discovery "$mac"
    controller_api wait-pending --mac "$mac" --timeout "${UNIFI_STUB_DOCKER_ADOPTION_TIMEOUT:-90}"
    controller_api adopt --mac "$mac"
    wait_for_event_kind "$adoption_start_line" setparam "$mac"
    wait_for_stub_adoption_state "$service"
    wait_for_event_kind "$adoption_start_line" gcm "$mac"
    wait_for_event_count "$adoption_start_line" gcm 2 "$mac"
    controller_api wait-connected \
        --mac "$mac" \
        --timeout "${UNIFI_STUB_DOCKER_ADOPTION_TIMEOUT:-90}" \
        --stable-samples "${UNIFI_STUB_DOCKER_STABLE_SAMPLES:-3}" \
        --poll-interval "${UNIFI_STUB_DOCKER_STABLE_POLL_SECONDS:-3}"
    if [ "$label" = "gateway" ]; then
        wait_for_gateway_command_lifecycle
    fi
}

if ! command -v docker >/dev/null 2>&1; then
    echo "docker is required" >&2
    exit 2
fi
if ! docker compose version >/dev/null 2>&1; then
    echo "docker compose is required" >&2
    exit 2
fi
if ! command -v "$go_cmd" >/dev/null 2>&1; then
    echo "go is required" >&2
    exit 2
fi
case "$gateway_command_expectation" in
    required|absent)
        ;;
    *)
        echo "UNIFI_STUB_DOCKER_GATEWAY_COMMAND_EXPECTATION must be required or absent" >&2
        exit 2
        ;;
esac

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
echo "docker integration: gateway command expectation=$gateway_command_expectation"

echo "docker integration: resetting temporary test services"
reset_test_service stub-bridge-observe stub-bridge-observe-state
reset_test_service stub-port-map stub-port-map-state
reset_test_service stub-gateway-smoke stub-gateway-smoke-state

echo "docker integration: validating compose configuration"
compose config >/dev/null

if [ "${UNIFI_STUB_DOCKER_SKIP_BUILD:-0}" = "1" ]; then
    echo "docker integration: using prebuilt test image"
else
    echo "docker integration: building test image"
    compose build stub-bridge-observe stub-port-map stub-gateway-smoke
fi

echo "docker integration: dry-run bridge-observe payload"
compose run --rm --no-deps stub-bridge-observe -dry-run > "$payload_dir/bridge-observe.txt"
assert_payload bridge-observe "$payload_dir/bridge-observe.txt"

echo "docker integration: dry-run management-lan preexisting-interface payload"
compose run --rm --no-deps stub-bridge-observe \
    -dry-run \
    -management-lan-enabled \
    -management-lan-vlan 42 \
    -management-lan-mode preexisting-interface \
    -management-lan-interface eth0 \
    -management-lan-ip "$UNIFI_STUB_BRIDGE_IP" \
    > "$payload_dir/management-lan.txt"
assert_payload management-lan "$payload_dir/management-lan.txt"

echo "docker integration: dry-run port-map payload"
compose run --rm --no-deps stub-port-map -dry-run > "$payload_dir/port-map.txt"
assert_payload port-map "$payload_dir/port-map.txt"

echo "docker integration: dry-run gateway payload"
compose run --rm --no-deps stub-gateway-smoke -dry-run > "$payload_dir/gateway-smoke.txt"
assert_payload gateway-smoke "$payload_dir/gateway-smoke.txt"

echo "docker integration: starting pinned controller image=$UNIFI_NETWORK_IMAGE"
compose up -d unifi inform-mitm
wait_for_controller_version
controller_api bootstrap
reset_controller_fixtures
echo "docker integration: starting inform/discovery proxy target=$UNIFI_STUB_INFORM_PROXY_TARGET discovery_target=$UNIFI_STUB_DISCOVERY_PROXY_TARGET"
wait_for_inform_proxy

for mac in "$UNIFI_STUB_BRIDGE_MAC" "$UNIFI_STUB_PORTMAP_MAC" "$UNIFI_STUB_GATEWAY_MAC"; do
    controller_api forget --mac "$mac" >/dev/null 2>&1 || true
    controller_api wait-clean --mac "$mac" --timeout "${UNIFI_STUB_DOCKER_ADOPTION_TIMEOUT:-90}"
done

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

echo "docker integration: checking inform proxy events"
wait_for_events "$start_line" \
    "$UNIFI_STUB_BRIDGE_MAC" \
    "$UNIFI_STUB_PORTMAP_MAC" \
    "$UNIFI_STUB_GATEWAY_MAC"

echo "docker integration: checking discovery proxy events"
wait_for_event_kind "$start_line" discovery \
    "$UNIFI_STUB_BRIDGE_MAC" \
    "$UNIFI_STUB_PORTMAP_MAC" \
    "$UNIFI_STUB_GATEWAY_MAC"

run_controller_adoption_gate \
    switch \
    stub-bridge-observe \
    stub-bridge-observe-state \
    "$UNIFI_STUB_BRIDGE_MAC"

run_controller_adoption_gate \
    gateway \
    stub-gateway-smoke \
    stub-gateway-smoke-state \
    "$UNIFI_STUB_GATEWAY_MAC"

echo "docker integration: cleaning temporary test resources"
cleanup_runtime_resources

echo "docker integration: ok"

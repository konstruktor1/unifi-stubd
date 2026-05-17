#!/bin/bash
# Start the selected UXG-Lite firmware processes inside the lab container.
set -euo pipefail

preload="${UNIFI_FW_SIM_PRELOAD:-/mock/libubnthal_redirect.so}"
log_dir="${UNIFI_FW_SIM_LOG_DIR:-/tmp}"
model="${UNIFI_FW_SIM_MODEL:-UXG}"
mac="${UNIFI_FW_SIM_MAC:-02:15:6d:00:a6:77}"
udapi_socket="${UNIFI_FW_SIM_UDAPI_SOCKET:-/var/run/ubnt-udapi-server.sock}"
event_socket="${UNIFI_FW_SIM_EVENT_SOCKET:-/var/run/ubnt-udapi-bridge-event.sock}"
ready_path="${UNIFI_FW_SIM_READY_PATH:-/run/ubios-udapi-server-bridge-event-notifier.sock}"
allow_partial="${UNIFI_FW_SIM_ALLOW_PARTIAL:-1}"

if [[ ! -r "$preload" ]]; then
    echo "missing LD_PRELOAD shim: $preload" >&2
    exit 1
fi

mkdir -p /data/udapi-config/ubios-udapi-server "$log_dir"

: > "$log_dir/ubios-udapi-server.run.log"
: > "$log_dir/ubios-udapi-server.run.err"
: > "$log_dir/udapi-bridge.run.log"
: > "$log_dir/udapi-bridge.run.err"
: > "$log_dir/mcad.run.out"
: > "$log_dir/mcad.run.err"

env LD_PRELOAD="$preload" \
    /usr/bin/ubios-udapi-server \
        -c /data/udapi-config/ubios-udapi-server/ubios-udapi-server.state \
        -s "$udapi_socket" \
        -e "$event_socket" \
        -x -t \
    >"$log_dir/ubios-udapi-server.run.log" \
    2>"$log_dir/ubios-udapi-server.run.err" &
udapi_pid=$!

for _ in {1..60}; do
    if [[ -S "$udapi_socket" || -S "$ready_path" ]]; then
        break
    fi
    if ! kill -0 "$udapi_pid" 2>/dev/null; then
        echo "ubios-udapi-server exited before creating a readiness socket" >&2
        tail -80 "$log_dir/ubios-udapi-server.run.err" >&2 || true
        exit 1
    fi
    sleep 1
done

if [[ ! -S "$udapi_socket" ]]; then
    echo "warning: $udapi_socket is not present; continuing with partial UXG-Lite simulation" >&2
    if [[ "$allow_partial" != "1" ]]; then
        exit 1
    fi
fi

env LD_PRELOAD="$preload" \
    /usr/bin/udapi-bridge \
    -m "$model" \
    -M "$mac" \
    --rest-api-port 1080 \
    --rest-api-secure-port 0 \
    --rest-api-interface lo \
    -l - -x - \
    >"$log_dir/udapi-bridge.run.log" \
    2>"$log_dir/udapi-bridge.run.err" &
bridge_pid=$!

env LD_PRELOAD="$preload" \
    /usr/bin/mcad -n 0 -s /tmp/.mcad -v \
    >"$log_dir/mcad.run.out" \
    2>"$log_dir/mcad.run.err" &
mcad_pid=$!

stop_processes() {
    kill "$mcad_pid" "$bridge_pid" "$udapi_pid" 2>/dev/null || true
    wait "$mcad_pid" "$bridge_pid" "$udapi_pid" 2>/dev/null || true
}

trap stop_processes INT TERM

wait -n "$mcad_pid" "$bridge_pid" "$udapi_pid"
status=$?
stop_processes
exit "$status"

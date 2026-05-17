#!/bin/bash
set -euo pipefail

preload="${UXGPRO_SIM_PRELOAD:-/mock/libubnthal_redirect.so}"
log_dir="${UXGPRO_SIM_LOG_DIR:-/tmp}"

if [[ ! -r "$preload" ]]; then
    echo "missing LD_PRELOAD shim: $preload" >&2
    echo "build it first, for example: docker compose --profile build-shim run --rm shim-builder" >&2
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
        -x -t \
    >"$log_dir/ubios-udapi-server.run.log" \
    2>"$log_dir/ubios-udapi-server.run.err" &
udapi_pid=$!

for _ in {1..60}; do
    if [[ -S /var/run/ubnt-udapi-server.sock ]]; then
        break
    fi
    if ! kill -0 "$udapi_pid" 2>/dev/null; then
        echo "ubios-udapi-server exited before creating its socket" >&2
        tail -80 "$log_dir/ubios-udapi-server.run.err" >&2 || true
        exit 1
    fi
    sleep 1
done

if [[ ! -S /var/run/ubnt-udapi-server.sock ]]; then
    echo "timed out waiting for /var/run/ubnt-udapi-server.sock" >&2
    tail -120 "$log_dir/ubios-udapi-server.run.err" >&2 || true
    exit 1
fi

if [[ -n "${UXGPRO_SIM_STATIC_ADDRESS:-}" ]]; then
    static_interface="${UXGPRO_SIM_STATIC_INTERFACE:-eth0}"
    ip link set "$static_interface" up || true
    ip addr flush dev "$static_interface" scope global || true
    ip addr add "$UXGPRO_SIM_STATIC_ADDRESS" dev "$static_interface"
fi

/usr/bin/udapi-bridge \
    -m UXGPRO \
    -M 00:15:6d:de:ad:00 \
    --rest-api-port 1080 \
    --rest-api-secure-port 0 \
    --rest-api-interface lo \
    -l - -x - \
    >"$log_dir/udapi-bridge.run.log" \
    2>"$log_dir/udapi-bridge.run.err" &
bridge_pid=$!

env LD_PRELOAD="$preload" \
    /usr/bin/mcad -n -s -v \
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

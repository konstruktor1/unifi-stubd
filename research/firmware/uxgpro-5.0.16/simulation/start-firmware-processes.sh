#!/bin/bash
# Start the selected UXG-Pro firmware processes inside the lab container.
set -euo pipefail

preload="${UNIFI_FW_SIM_PRELOAD:-${UXGPRO_SIM_PRELOAD:-/mock/libubnthal_redirect.so}}"
log_dir="${UNIFI_FW_SIM_LOG_DIR:-${UXGPRO_SIM_LOG_DIR:-/tmp}}"
model="${UNIFI_FW_SIM_MODEL:-${UXGPRO_SIM_MODEL:-UXGPRO}}"
mac="${UNIFI_FW_SIM_MAC:-${UXGPRO_SIM_MAC:-00:15:6d:de:ad:00}}"

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
: > "$log_dir/dropbear.run.log"
: > "$log_dir/dropbear.run.err"

dropbear_pid=""

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

static_address="${UNIFI_FW_SIM_STATIC_ADDRESS:-${UXGPRO_SIM_STATIC_ADDRESS:-}}"
if [[ -n "$static_address" ]]; then
    static_interface="${UNIFI_FW_SIM_STATIC_INTERFACE:-${UXGPRO_SIM_STATIC_INTERFACE:-eth0}}"
    ip link set "$static_interface" up || true
    ip addr flush dev "$static_interface" scope global || true
    ip addr add "$static_address" dev "$static_interface"
fi

dummy_interfaces_value="${UNIFI_FW_SIM_DUMMY_INTERFACES:-${UXGPRO_SIM_DUMMY_INTERFACES:-}}"
if [[ -n "$dummy_interfaces_value" ]]; then
    IFS=';' read -r -a dummy_interfaces <<< "$dummy_interfaces_value"
    for dummy_interface in "${dummy_interfaces[@]}"; do
        [[ -z "$dummy_interface" ]] && continue
        IFS=',' read -r iface mac address <<< "$dummy_interface"
        [[ -z "${iface:-}" ]] && continue

        ip link show "$iface" >/dev/null 2>&1 || ip link add "$iface" type dummy
        if [[ -n "${mac:-}" ]]; then
            ip link set "$iface" address "$mac"
        fi
        ip link set "$iface" up
        if [[ -n "${address:-}" ]]; then
            ip addr flush dev "$iface" scope global || true
            ip addr add "$address" dev "$iface"
        fi
    done
fi

start_dropbear="${UNIFI_FW_SIM_START_DROPBEAR:-${UXGPRO_SIM_START_DROPBEAR:-0}}"
if [[ "$start_dropbear" == "1" ]]; then
    mkdir -p /etc/dropbear
    /usr/sbin/dropbear -F -E -R -p 0.0.0.0:22 \
        >"$log_dir/dropbear.run.log" \
        2>"$log_dir/dropbear.run.err" &
    dropbear_pid=$!
fi

bridge_env=(
    "LD_PRELOAD=$preload"
)
bridge_redirect_debug="${UNIFI_FW_SIM_BRIDGE_REDIRECT_DEBUG:-${UXGPRO_SIM_BRIDGE_REDIRECT_DEBUG:-}}"
if [[ -n "$bridge_redirect_debug" ]]; then
    bridge_env+=("UBNTHAL_REDIRECT_DEBUG=$bridge_redirect_debug")
fi

env "${bridge_env[@]}" \
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
    /usr/bin/mcad -n -s -v \
    >"$log_dir/mcad.run.out" \
    2>"$log_dir/mcad.run.err" &
mcad_pid=$!

stop_processes() {
    pids=("$mcad_pid" "$bridge_pid" "$udapi_pid")
    if [[ -n "$dropbear_pid" ]]; then
        pids+=("$dropbear_pid")
    fi
    kill "${pids[@]}" 2>/dev/null || true
    wait "${pids[@]}" 2>/dev/null || true
}

trap stop_processes INT TERM

wait_pids=("$mcad_pid" "$bridge_pid" "$udapi_pid")
if [[ -n "$dropbear_pid" ]]; then
    wait_pids+=("$dropbear_pid")
fi

wait -n "${wait_pids[@]}"
status=$?
stop_processes
exit "$status"

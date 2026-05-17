#!/bin/bash
# Start the selected UDM Pro SE firmware processes inside the lab container.
set -euo pipefail

preload="${UNIFI_FW_SIM_PRELOAD:-/mock/libubnthal_redirect.so}"
log_dir="${UNIFI_FW_SIM_LOG_DIR:-/tmp}"
model="${UNIFI_FW_SIM_MODEL:-UDMPROSE}"
mac="${UNIFI_FW_SIM_MAC:-02:15:6d:00:ea:2c}"
udapi_socket="${UNIFI_FW_SIM_UDAPI_SOCKET:-/var/run/ubnt-udapi-server.sock}"
event_socket="${UNIFI_FW_SIM_EVENT_SOCKET:-/var/run/ubnt-udapi-bridge-event.sock}"
ready_path="${UNIFI_FW_SIM_READY_PATH:-/run/ubios-udapi-server-bridge-event-notifier.sock}"
allow_partial="${UNIFI_FW_SIM_ALLOW_PARTIAL:-1}"
trace="${UNIFI_FW_SIM_TRACE:-0}"
trace_dir="${UNIFI_FW_SIM_TRACE_DIR:-$log_dir/trace}"
ready_wait_seconds="${UNIFI_FW_SIM_READY_WAIT_SECONDS:-60}"
udapi_wait_seconds="${UNIFI_FW_SIM_UDAPI_WAIT_SECONDS:-180}"

if [[ ! -r "$preload" ]]; then
    echo "missing LD_PRELOAD shim: $preload" >&2
    exit 1
fi

mkdir -p /data/udapi-config/ubios-udapi-server "$log_dir"
if [[ "$trace" = "1" ]]; then
    if ! command -v strace >/dev/null 2>&1; then
        echo "trace requested but strace is not installed in this image" >&2
        exit 1
    fi
    mkdir -p "$trace_dir"
fi

: > "$log_dir/ubios-udapi-server.run.log"
: > "$log_dir/ubios-udapi-server.run.err"
: > "$log_dir/udapi-bridge.run.log"
: > "$log_dir/udapi-bridge.run.err"
: > "$log_dir/mcad.run.out"
: > "$log_dir/mcad.run.err"

run_firmware_process() {
    local name="$1"
    shift

    if [[ "$trace" = "1" ]]; then
        env LD_PRELOAD="$preload" \
            strace -ff -tt -s 256 \
                -o "$trace_dir/$name" \
                -e trace=open,openat,close,read,write,stat,lstat,newfstatat,access,faccessat,ioctl,socket,connect,bind,listen,getsockopt,setsockopt,sendto,recvfrom,execve \
                "$@"
    else
        env LD_PRELOAD="$preload" "$@"
    fi
}

run_firmware_process ubios-udapi-server \
    /usr/bin/ubios-udapi-server \
        -c /data/udapi-config/ubios-udapi-server/ubios-udapi-server.state \
        -s "$udapi_socket" \
        -e "$event_socket" \
        -x -t \
    >"$log_dir/ubios-udapi-server.run.log" \
    2>"$log_dir/ubios-udapi-server.run.err" &
udapi_pid=$!

for ((elapsed = 0; elapsed < ready_wait_seconds; elapsed++)); do
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

for ((elapsed = 0; elapsed < udapi_wait_seconds; elapsed++)); do
    if [[ -S "$udapi_socket" ]]; then
        break
    fi
    if ! kill -0 "$udapi_pid" 2>/dev/null; then
        echo "ubios-udapi-server exited before creating $udapi_socket" >&2
        tail -80 "$log_dir/ubios-udapi-server.run.err" >&2 || true
        exit 1
    fi
    sleep 1
done

if [[ ! -S "$udapi_socket" ]]; then
    echo "warning: $udapi_socket is not present; continuing with partial UDM Pro SE simulation" >&2
    if [[ "$allow_partial" != "1" ]]; then
        exit 1
    fi
fi

run_firmware_process udapi-bridge \
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

run_firmware_process mcad \
    /usr/bin/mcad -n 0 -s /tmp/.mcad -v \
    >"$log_dir/mcad.run.out" \
    2>"$log_dir/mcad.run.err" &
mcad_pid=$!

stop_processes() {
    kill "$mcad_pid" "$bridge_pid" "$udapi_pid" 2>/dev/null || true
    wait "$mcad_pid" "$bridge_pid" "$udapi_pid" 2>/dev/null || true
}

trap stop_processes INT TERM

set +e
wait -n "$mcad_pid" "$bridge_pid" "$udapi_pid"
status=$?
set -e

if [[ "$allow_partial" = "1" ]]; then
    echo "warning: a firmware process exited with status $status; keeping partial UDM Pro SE simulation container alive" >&2
    echo "warning: inspect /tmp/*.run.* logs and extend deterministic mocks before controller attachment" >&2
    stop_processes
    sleep infinity &
    wait $!
    exit 0
fi

stop_processes
exit "$status"

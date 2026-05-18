#!/bin/bash
# Process helpers for the reduced UDM Pro SE firmware chain.

run_firmware_process() {
    local name="$1"
    shift

    # The LD_PRELOAD shim supplies the missing hardware/filesystem surface. The
    # optional trace mode records exactly which additional mocks are still needed.
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

start_udapi_server() {
    run_firmware_process ubios-udapi-server \
        /usr/bin/ubios-udapi-server \
            -c /data/udapi-config/ubios-udapi-server/ubios-udapi-server.state \
            -s "$udapi_socket" \
            -e "$event_socket" \
            -x -t \
        >"$log_dir/ubios-udapi-server.run.log" \
        2>"$log_dir/ubios-udapi-server.run.err" &
    udapi_pid=$!
}

wait_for_udapi_readiness() {
    local elapsed

    # udapi-server exposes an early bridge-event socket before the primary REST
    # socket. Waiting for both gives useful logs without failing too early.
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
}

start_udapi_bridge() {
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
}

start_mcad() {
    run_firmware_process mcad \
        /usr/bin/mcad -n 0 -s /tmp/.mcad -v \
        >"$log_dir/mcad.run.out" \
        2>"$log_dir/mcad.run.err" &
    mcad_pid=$!
}

stop_firmware_processes() {
    kill "${mcad_pid:-}" "${bridge_pid:-}" "${udapi_pid:-}" 2>/dev/null || true
    wait "${mcad_pid:-}" "${bridge_pid:-}" "${udapi_pid:-}" 2>/dev/null || true
}

wait_for_firmware_exit() {
    local status

    set +e
    wait -n "$mcad_pid" "$bridge_pid" "$udapi_pid"
    status=$?
    set -e

    # Partial mode is deliberate: it keeps the container around so the next
    # missing firmware assumption can be read from the logs and modeled.
    if [[ "$allow_partial" = "1" ]]; then
        echo "warning: a firmware process exited with status $status; keeping partial UDM Pro SE simulation container alive" >&2
        echo "warning: inspect /tmp/*.run.* logs and extend deterministic mocks before controller attachment" >&2
        stop_firmware_processes
        sleep infinity &
        wait $!
        exit 0
    fi

    stop_firmware_processes
    exit "$status"
}

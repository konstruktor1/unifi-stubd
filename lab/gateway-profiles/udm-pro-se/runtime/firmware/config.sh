#!/bin/bash
# Runtime configuration for the reduced firmware process chain.

load_firmware_config() {
    preload="${UNIFI_FW_SIM_PRELOAD:-/mock/libubnthal_redirect.so}"
    log_dir="${UNIFI_FW_SIM_LOG_DIR:-/tmp}"
    model="${UNIFI_FW_SIM_MODEL:-UDMPROSE}"
    mac="${UNIFI_FW_SIM_MAC:-02:15:6d:00:ea:2c}"
    udapi_socket="${UNIFI_FW_SIM_UDAPI_SOCKET:-/var/run/ubnt-udapi-server.sock}"
    event_socket="${UNIFI_FW_SIM_EVENT_SOCKET:-/var/run/ubnt-udapi-bridge-event.sock}"
    ready_path="${UNIFI_FW_SIM_READY_PATH:-/run/ubios-udapi-server-bridge-event-notifier.sock}"
    kernel_dir="${UNIFI_FW_SIM_KERNEL_DIR:-/opt/unifi-fw-sim/kernel}"
    allow_partial="${UNIFI_FW_SIM_ALLOW_PARTIAL:-1}"
    trace="${UNIFI_FW_SIM_TRACE:-0}"
    trace_dir="${UNIFI_FW_SIM_TRACE_DIR:-$log_dir/trace}"
    ready_wait_seconds="${UNIFI_FW_SIM_READY_WAIT_SECONDS:-60}"
    udapi_wait_seconds="${UNIFI_FW_SIM_UDAPI_WAIT_SECONDS:-180}"
}

validate_firmware_config() {
    if [[ ! -r "$preload" ]]; then
        echo "missing LD_PRELOAD shim: $preload" >&2
        exit 1
    fi

    if [[ "$trace" = "1" ]] && ! command -v strace >/dev/null 2>&1; then
        echo "trace requested but strace is not installed in this image" >&2
        exit 1
    fi
}

prepare_firmware_runtime() {
    mkdir -p /data/udapi-config/ubios-udapi-server "$log_dir"

    if [[ "$trace" = "1" ]]; then
        mkdir -p "$trace_dir"
    fi

    : > "$log_dir/ubios-udapi-server.run.log"
    : > "$log_dir/ubios-udapi-server.run.err"
    : > "$log_dir/udapi-bridge.run.log"
    : > "$log_dir/udapi-bridge.run.err"
    : > "$log_dir/mcad.run.out"
    : > "$log_dir/mcad.run.err"
}

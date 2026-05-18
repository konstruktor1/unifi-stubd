#!/bin/bash
# Runtime configuration for the Docker webportal profile.

load_webportal_config() {
    log_dir="${UNIFI_FW_SIM_WEB_LOG_DIR:-/tmp/udm-pro-se-webportal}"
    postgres_password="${UNIFI_CORE_POSTGRES_PASSWORD:-unifi-core-lab-pass}"
    kernel_dir="${UNIFI_FW_SIM_KERNEL_DIR:-/opt/unifi-fw-sim/kernel}"
    systemd_dbus_stub="${UNIFI_FW_SIM_SYSTEMD_DBUS_STUB:-/usr/local/lib/udm-pro-se-systemd-dbus/index.cjs}"
    network_app_stub="${UNIFI_FW_SIM_NETWORK_APP_STUB:-/usr/local/lib/udm-pro-se-network-app/index.cjs}"

    webportal_runtime_dir="${UNIFI_FW_SIM_WEB_RUNTIME_DIR:-/usr/local/lib/udm-pro-se-runtime/webportal}"
    wrapper_dir="$webportal_runtime_dir/wrappers"
    template_dir="$webportal_runtime_dir/templates"
    http_dir="$webportal_runtime_dir/http"
}

prepare_webportal_runtime() {
    mkdir -p "$log_dir"
}

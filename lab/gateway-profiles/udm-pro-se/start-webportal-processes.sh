#!/bin/bash
# Start the UDM Pro SE firmware simulation plus the minimal UniFi OS webportal.
set -euo pipefail

runtime_dir="${UNIFI_FW_SIM_RUNTIME_DIR:-/usr/local/lib/udm-pro-se-runtime}"

# Webportal startup is intentionally split into sourced modules and runtime
# assets. The large wrapper scripts, nginx snippets, and AWK filters live under
# runtime/webportal/ instead of being generated inline from this entry point.
# shellcheck disable=SC1091
. "$runtime_dir/common/kernel-artifacts.sh"
# shellcheck disable=SC1091
. "$runtime_dir/webportal/config.sh"
# shellcheck disable=SC1091
. "$runtime_dir/webportal/install-wrappers.sh"
# shellcheck disable=SC1091
. "$runtime_dir/webportal/http.sh"
# shellcheck disable=SC1091
. "$runtime_dir/webportal/services.sh"

load_webportal_config
prepare_webportal_runtime
record_kernel_artifacts "$kernel_dir" "$log_dir"

install_webportal_wrappers
ensure_lab_lan_bridge
start_postgres
write_unifi_core_override
prepare_support_bundle_paths
start_dbus
start_systemd_stub
start_network_app_stub
start_ulp_go
start_unifi_core

exec /usr/local/bin/udm-pro-se-sim-start

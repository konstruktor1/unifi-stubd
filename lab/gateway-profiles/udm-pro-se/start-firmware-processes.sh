#!/bin/bash
# Start the selected UDM Pro SE firmware processes inside the lab container.
set -euo pipefail

runtime_dir="${UNIFI_FW_SIM_RUNTIME_DIR:-/usr/local/lib/udm-pro-se-runtime}"

# The startup logic is split under runtime/ so Docker and documentation can
# point at the same project-owned implementation instead of hidden heredocs.
# shellcheck disable=SC1091
. "$runtime_dir/common/kernel-artifacts.sh"
# shellcheck disable=SC1091
. "$runtime_dir/firmware/config.sh"
# shellcheck disable=SC1091
. "$runtime_dir/firmware/processes.sh"

load_firmware_config
validate_firmware_config
prepare_firmware_runtime
record_kernel_artifacts "$kernel_dir" "$log_dir"

start_udapi_server
wait_for_udapi_readiness
start_udapi_bridge
start_mcad

trap stop_firmware_processes INT TERM
wait_for_firmware_exit

#!/bin/sh
# Configure a registered UTM clone for the UDM Pro SE QEMU lab.
set -eu

profile_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
artifacts="${UDM_PRO_SE_VM_ARTIFACTS:-$profile_dir/artifacts}"
kernel_deploy_dir="${UDM_PRO_SE_KERNEL_DEPLOY_DIR:-$artifacts/deploy/kernel}"

# Defaults are sourced before resolving variables so callers can override any
# value from the environment without editing the committed profile files.
utm_defaults="$profile_dir/utm/defaults.env"
if [ -f "$utm_defaults" ]; then
    # shellcheck disable=SC1090
    . "$utm_defaults"
fi

# The installer modules share the variables resolved below. Keep the source
# order explicit: generic helpers first, then PlistBuddy, then the UTM sections
# that call those helpers.
# shellcheck disable=SC1091
. "$profile_dir/utm/install/common.sh"
# shellcheck disable=SC1091
. "$profile_dir/utm/install/plist.sh"
# shellcheck disable=SC1091
. "$profile_dir/utm/install/boot.sh"
# shellcheck disable=SC1091
. "$profile_dir/utm/install/drives.sh"
# shellcheck disable=SC1091
. "$profile_dir/utm/install/network.sh"
# shellcheck disable=SC1091
. "$profile_dir/utm/install/system.sh"

utm_documents="${UDM_PRO_SE_UTM_DOCUMENTS:-$HOME/Library/Containers/com.utmapp.UTM/Data/Documents}"
utm_name="${UDM_PRO_SE_UTM_NAME:-UDM-Pro-SE-QEMU}"
utm_ifname="${UDM_PRO_SE_UTM_IFNAME:-en0}"
utm_network_mode="${UDM_PRO_SE_UTM_NETWORK_MODE:-Host}"
utm_network_backend="${UDM_PRO_SE_UTM_NETWORK_BACKEND:-utm}"
utm_network_hardware="${UDM_PRO_SE_UTM_NETWORK_HARDWARE:-virtio-net-pci}"
utm_network_pci_addr="${UDM_PRO_SE_UTM_NETWORK_PCI_ADDR:-0x2}"
utm_sfp_wan_mode="${UDM_PRO_SE_UTM_SFP_WAN_MODE:-Shared}"
utm_sfp_wan_mac="${UDM_PRO_SE_UTM_SFP_WAN_MAC:-02:15:6d:00:ea:35}"
utm_lan_mode="${UDM_PRO_SE_UTM_LAN_MODE:-Host}"
utm_lan_mac="${UDM_PRO_SE_UTM_LAN_MAC:-02:15:6d:00:ea:34}"
utm_https_host_port="${UDM_PRO_SE_UTM_HTTPS_HOST_PORT:-10443}"
utm_memory_size="${UDM_PRO_SE_UTM_MEMORY:-4096}"
utm_serial_port="${UDM_PRO_SE_UTM_SERIAL_PORT:-15555}"
utm_serial_wait="${UDM_PRO_SE_UTM_SERIAL_WAIT:-false}"
utm_bootargs_file="${UDM_PRO_SE_UTM_BOOTARGS_FILE:-$profile_dir/utm/bootargs.txt}"
utm_bundle="$utm_documents/$utm_name.utm"
config="$utm_bundle/config.plist"
plistbuddy=/usr/libexec/PlistBuddy

# The sequence below mirrors the generated UTM bundle shape:
# resolve local inputs, copy boot payloads into Data/, replace the writable
# disk, then rewrite the config.plist sections for system, network, and QEMU.
utm_resolve_artifact_paths
utm_resolve_cmdline
utm_require_inputs
utm_require_tools
utm_deploy_boot_inputs
utm_configure_drives
utm_configure_system
utm_configure_network
utm_configure_qemu

plutil -lint "$config"
utm_print_summary

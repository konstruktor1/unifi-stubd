# shellcheck shell=sh
# UTM network devices and QEMU command-line arguments.

utm_configure_network() {
    pb_reset_array ":Network"
    if [ "$utm_network_backend" = "utm" ]; then
        # Network 0 intentionally represents the first SFP+ WAN role. Shared
        # mode gives the guest outbound internet. The plist PortForward below
        # records the intended localhost HTTPS mapping; verify at runtime that
        # the current UTM build actually binds the host port.
        "$plistbuddy" -c "Add :Network:0 dict" "$config"
        if [ "$utm_sfp_wan_mode" = "Bridged" ]; then
            pb_set_string ":Network:0:BridgeInterface" "$utm_ifname"
        else
            pb_delete ":Network:0:BridgeInterface"
        fi
        pb_set_string ":Network:0:Hardware" "$utm_network_hardware"
        pb_set_string ":Network:0:MacAddress" "$utm_sfp_wan_mac"
        pb_set_string ":Network:0:Mode" "$utm_sfp_wan_mode"
        pb_set_bool ":Network:0:IsolateFromHost" "false"
        pb_reset_array ":Network:0:PortForward"
        if [ -n "$utm_https_host_port" ] && [ "$utm_sfp_wan_mode" = "Shared" ]; then
            "$plistbuddy" -c "Add :Network:0:PortForward:0 dict" "$config"
            pb_set_string ":Network:0:PortForward:0:GuestAddress" ""
            pb_set_int ":Network:0:PortForward:0:GuestPort" "443"
            pb_set_string ":Network:0:PortForward:0:HostAddress" "127.0.0.1"
            pb_set_int ":Network:0:PortForward:0:HostPort" "$utm_https_host_port"
            pb_set_string ":Network:0:PortForward:0:Protocol" "TCP"
        fi

        # Network 1 is the host-only 2.5G RJ45 LAN role. The guest initramfs
        # keeps this interface attached to br0 so the web surface remains on
        # the LAN side instead of being opened as WAN ingress.
        "$plistbuddy" -c "Add :Network:1 dict" "$config"
        pb_delete ":Network:1:BridgeInterface"
        pb_set_string ":Network:1:Hardware" "$utm_network_hardware"
        pb_set_string ":Network:1:MacAddress" "$utm_lan_mac"
        pb_set_string ":Network:1:Mode" "$utm_lan_mode"
        pb_set_bool ":Network:1:IsolateFromHost" "false"
        pb_reset_array ":Network:1:PortForward"
    fi
}

utm_configure_qemu() {
    pb_set_bool ":QEMU:UEFIBoot" "false"
    pb_set_bool ":QEMU:DebugLog" "true"
    pb_set_bool ":QEMU:Hypervisor" "true"
    pb_set_bool ":QEMU:RNGDevice" "true"
    pb_set_bool ":QEMU:TPMDevice" "false"
    pb_set_bool ":QEMU:TSO" "false"
    pb_set_bool ":QEMU:PS2Controller" "false"
    pb_set_bool ":QEMU:BalloonDevice" "false"
    pb_set_string ":QEMU:MachinePropertyOverride" "gic-version=3,highmem=off"
    pb_reset_array ":QEMU:AdditionalArguments"

    if [ "$utm_network_backend" = "manual" ]; then
        # Manual backend is a fallback for UTM versions where Network[] does not
        # expose the needed vmnet mode. It pins one NIC through raw QEMU args.
        case "$utm_network_mode" in
        Host)
            pb_add_arg "-netdev"
            pb_add_arg "vmnet-host,id=udm_lan"
            ;;
        Bridged)
            pb_add_arg "-netdev"
            pb_add_arg "vmnet-bridged,id=udm_lan,ifname=$utm_ifname"
            ;;
        *)
            echo "unsupported pinned UTM network mode: $utm_network_mode" >&2
            echo "supported values: Host, Bridged" >&2
            exit 1
            ;;
        esac
        pb_add_arg "-device"
        pb_add_arg "$utm_network_hardware,netdev=udm_lan,mac=02:15:6d:00:ea:36,addr=$utm_network_pci_addr"
    fi

    pb_add_arg "-no-reboot"
    pb_add_arg "-append"
    pb_add_quoted_arg "$cmdline"
}

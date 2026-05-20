# shellcheck shell=sh
# UTM system, serial-only runtime, sharing, and summary output.

utm_configure_system() {
    pb_set_string ":Information:Name" "$utm_name"
    pb_delete ":Information:ConsoleOnly"

    # Keep the machine shape close to the direct QEMU runner: ARM64 virt, four
    # vCPUs, and 4 GiB RAM by default. The real UDM SE has more board-specific
    # hardware, but QEMU virt is the bootable boundary for this reference.
    pb_set_int ":System:CPUCount" "4"
    pb_set_int ":System:MemorySize" "$utm_memory_size"
    pb_set_string ":System:Architecture" "aarch64"
    pb_set_string ":System:Target" "virt"
    pb_set_string ":System:CPU" "default"

    # The VM is serial-only on purpose. Removing Display/Sound avoids UTM SPICE
    # paths that previously blocked or confused the boot tests.
    pb_reset_array ":Display"
    pb_reset_array ":Sound"
    pb_reset_array ":Serial"
    "$plistbuddy" -c "Add :Serial:0 dict" "$config"
    pb_set_string ":Serial:0:Mode" "TcpServer"
    pb_set_string ":Serial:0:Target" "Auto"
    pb_set_int ":Serial:0:TcpPort" "$utm_serial_port"
    pb_set_bool ":Serial:0:WaitForConnection" "$utm_serial_wait"
    pb_set_bool ":Serial:0:RemoteConnectionAllowed" "false"

    # Disable convenience integrations. They are not needed for firmware boot
    # and keeping them off makes the VM closer to a network appliance boundary.
    pb_delete ":Sharing"
    "$plistbuddy" -c "Add :Sharing dict" "$config"
    pb_set_bool ":Sharing:ClipboardSharing" "false"
    pb_set_string ":Sharing:DirectoryShareMode" "None"
    pb_set_bool ":Sharing:DirectoryShareReadOnly" "true"
    pb_set_bool ":Input:UsbSharing" "false"
    pb_set_string ":Input:UsbBusSupport" "Disabled"
    pb_set_int ":Input:MaximumUsbShare" "0"
}

utm_print_summary() {
    echo "configured registered UTM profile: $utm_bundle"
    echo "network backend: $utm_network_backend"
    echo "network hardware: $utm_network_hardware"
    if [ "$utm_network_backend" = "utm" ]; then
        echo "network 0: $utm_sfp_wan_mode SFP+ WAN eth9, mac $utm_sfp_wan_mac"
        echo "network 1: $utm_lan_mode 2.5G LAN eth8, mac $utm_lan_mac"
        if [ "$utm_sfp_wan_mode" = "Bridged" ]; then
            echo "bridged interface: $utm_ifname"
        fi
        if [ -n "$utm_https_host_port" ] && [ "$utm_sfp_wan_mode" = "Shared" ]; then
            echo "https forward: https://127.0.0.1:$utm_https_host_port/ -> guest 443"
        fi
    else
        echo "manual network mode: $utm_network_mode"
    fi
    echo "memory: ${utm_memory_size} MiB"
    if [ "$utm_network_backend" = "manual" ]; then
        echo "network pci addr: $utm_network_pci_addr"
        if [ "$utm_network_mode" = "Bridged" ]; then
            echo "bridged interface: $utm_ifname"
        fi
    fi
    echo "writable disk: $drive_path"
    echo "kernel source: $kernel"
    echo "initrd source: $initrd"
    echo "kernel/initrd copied into: $utm_bundle/Data"
    echo "serial console: 127.0.0.1:$utm_serial_port"
    echo "serial wait before boot: $utm_serial_wait"
    echo "read serial with:"
    echo "  (sleep 600) | nc 127.0.0.1 '$utm_serial_port'"
    echo "start without attach:"
    echo "  utmctl start --hide '$utm_name'"
}

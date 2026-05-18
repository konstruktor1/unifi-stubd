#!/bin/bash
# Install lab command wrappers from versioned runtime files.

install_runtime_file() {
    local source="$1"
    local target="$2"
    local mode="${3:-0755}"

    install -m "$mode" "$source" "$target"
}

move_real_binary_once() {
    local target="$1"
    local real="$2"

    if [[ -x "$target" && ! -e "$real" ]]; then
        mv "$target" "$real"
    fi
}

write_ubnt_tools_wrapper() {
    move_real_binary_once /sbin/ubnt-tools /sbin/ubnt-tools.real
    install_runtime_file "$wrapper_dir/ubnt-tools" /sbin/ubnt-tools
}

write_ubnt_systool_wrapper() {
    move_real_binary_once /sbin/ubnt-systool /sbin/ubnt-systool.real
    install_runtime_file "$wrapper_dir/ubnt-systool" /sbin/ubnt-systool
}

write_systemd_run_wrapper() {
    move_real_binary_once /usr/bin/systemd-run /usr/bin/systemd-run.real
    install_runtime_file "$wrapper_dir/systemd-run" /usr/bin/systemd-run
    rm -f /usr/local/bin/systemd-run
}

write_systemctl_wrapper() {
    move_real_binary_once /usr/bin/systemctl /usr/bin/systemctl.real
    move_real_binary_once /bin/systemctl /bin/systemctl.real
    install_runtime_file "$wrapper_dir/systemctl" /usr/bin/systemctl
    cp /usr/bin/systemctl /bin/systemctl
    chmod 0755 /bin/systemctl
}

write_lab_sudoers() {
    install_runtime_file "$template_dir/unifi-core-lab.sudoers" /etc/sudoers.d/unifi-core-lab 0440
}

write_timedatectl_wrapper() {
    move_real_binary_once /usr/bin/timedatectl /usr/bin/timedatectl.real
    install_runtime_file "$wrapper_dir/timedatectl" /usr/bin/timedatectl
}

write_tar_wrapper() {
    install_runtime_file "$wrapper_dir/tar" /usr/local/bin/tar
}

write_udapi_lab_wrappers() {
    local tool

    for tool in mca-ctrl mca-dump ubios-udapi-client; do
        if [[ -x "/usr/bin/$tool" && ! -e "/usr/bin/$tool.real" ]]; then
            if ! grep -q "udm-pro-se-udapi-lab-shim" "/usr/bin/$tool" 2>/dev/null; then
                mv "/usr/bin/$tool" "/usr/bin/$tool.real"
            fi
        fi

        install_runtime_file "$wrapper_dir/udapi-tool" "/usr/bin/$tool"
    done
}

install_webportal_wrappers() {
    write_ubnt_tools_wrapper
    write_ubnt_systool_wrapper
    write_systemd_run_wrapper
    write_systemctl_wrapper
    write_lab_sudoers
    write_timedatectl_wrapper
    write_tar_wrapper
    write_udapi_lab_wrappers
}

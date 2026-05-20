#!/bin/bash
# Shared kernel-payload logging for Docker firmware and webportal paths.

record_kernel_artifacts() {
    local kernel_dir="$1"
    local log_dir="$2"
    local out="$log_dir/kernel-artifacts.txt"

    : > "$out"
    if [[ ! -d "$kernel_dir" ]]; then
        echo "kernel_dir_missing=$kernel_dir" > "$out"
        return
    fi

    {
        echo "kernel_dir=$kernel_dir"
        if [[ -f "$kernel_dir/MANIFEST.txt" ]]; then
            sed -n '1,120p' "$kernel_dir/MANIFEST.txt"
        else
            find "$kernel_dir" -maxdepth 3 -type f | sort
        fi
    } > "$out"
}

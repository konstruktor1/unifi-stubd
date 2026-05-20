# shellcheck shell=sh
# UTM drive image and Linux kernel/initrd drive entries.

find_writable_drive_index() {
    # Cloned UTM profiles often contain extra CDROM or kernel pseudo-drives.
    # Keep the first writable Disk entry and replace only that image.
    index=0
    while "$plistbuddy" -c "Print :Drive:$index" "$config" >/dev/null 2>&1; do
        readonly_value=$("$plistbuddy" -c "Print :Drive:$index:ReadOnly" "$config" 2>/dev/null || echo "true")
        image_type=$("$plistbuddy" -c "Print :Drive:$index:ImageType" "$config" 2>/dev/null || echo "")
        if [ "$readonly_value" = "false" ] && [ "$image_type" = "Disk" ]; then
            echo "$index"
            return 0
        fi
        index=$((index + 1))
    done
    return 1
}

utm_configure_drives() {
    drive_index=$(find_writable_drive_index || true)
    if [ -z "$drive_index" ]; then
        echo "no writable disk found in cloned UTM profile: $config" >&2
        exit 1
    fi

    drive_image=$("$plistbuddy" -c "Print :Drive:$drive_index:ImageName" "$config")
    drive_id=$("$plistbuddy" -c "Print :Drive:$drive_index:Identifier" "$config" 2>/dev/null || uuidgen)
    drive_path="$utm_bundle/Data/$drive_image"
    tmp_drive="$drive_path.tmp"

    # UTM stores the writable disk as qcow2 in the bundle even though the VM
    # builder creates a raw GPT image. Convert atomically to avoid corrupting an
    # existing clone if qemu-img fails.
    qemu-img convert -O qcow2 "$disk" "$tmp_drive"
    mv "$tmp_drive" "$drive_path"

    # Rebuild Drive[] from scratch so stale display, initrd, or disk entries
    # from the cloned template cannot change the boot order.
    pb_reset_array ":Drive"
    "$plistbuddy" -c "Add :Drive:0 dict" "$config"
    pb_set_string ":Drive:0:Identifier" "$drive_id"
    pb_set_string ":Drive:0:ImageName" "$drive_image"
    pb_set_string ":Drive:0:ImageType" "Disk"
    pb_set_string ":Drive:0:Interface" "VirtIO"
    pb_set_int ":Drive:0:InterfaceVersion" "1"
    pb_set_bool ":Drive:0:ReadOnly" "false"

    # UTM models Linux kernel and initrd as read-only Drive entries. The actual
    # bytes were copied into Data/ by boot.sh; these entries only wire them into
    # UTM's QEMU invocation.
    "$plistbuddy" -c "Add :Drive:1 dict" "$config"
    pb_set_string ":Drive:1:Identifier" "$(uuidgen)"
    pb_set_string ":Drive:1:ImageName" "$(basename "$utm_kernel")"
    pb_set_string ":Drive:1:ImageType" "LinuxKernel"
    pb_set_string ":Drive:1:Interface" "None"
    pb_set_int ":Drive:1:InterfaceVersion" "1"
    pb_set_bool ":Drive:1:ReadOnly" "true"

    "$plistbuddy" -c "Add :Drive:2 dict" "$config"
    pb_set_string ":Drive:2:Identifier" "$(uuidgen)"
    pb_set_string ":Drive:2:ImageName" "$(basename "$utm_initrd")"
    pb_set_string ":Drive:2:ImageType" "LinuxInitrd"
    pb_set_string ":Drive:2:Interface" "None"
    pb_set_int ":Drive:2:InterfaceVersion" "1"
    pb_set_bool ":Drive:2:ReadOnly" "true"
}

# shellcheck shell=sh
# Thin PlistBuddy helpers used by the UTM profile installer.

pb_set_string() {
    path=$1
    value=$2
    "$plistbuddy" -c "Set $path $value" "$config" 2>/dev/null ||
        "$plistbuddy" -c "Add $path string $value" "$config"
}

pb_set_bool() {
    path=$1
    value=$2
    "$plistbuddy" -c "Set $path $value" "$config" 2>/dev/null ||
        "$plistbuddy" -c "Add $path bool $value" "$config"
}

pb_set_int() {
    path=$1
    value=$2
    "$plistbuddy" -c "Set $path $value" "$config" 2>/dev/null ||
        "$plistbuddy" -c "Add $path integer $value" "$config"
}

pb_reset_array() {
    path=$1
    "$plistbuddy" -c "Delete $path" "$config" 2>/dev/null || true
    "$plistbuddy" -c "Add $path array" "$config"
}

pb_delete() {
    path=$1
    "$plistbuddy" -c "Delete $path" "$config" 2>/dev/null || true
}

pb_add_arg() {
    "$plistbuddy" -c "Add :QEMU:AdditionalArguments: string $1" "$config"
}

pb_add_quoted_arg() {
    value=$(printf '%s' "$1" | sed 's/\\/\\\\/g; s/"/\\"/g')
    "$plistbuddy" -c "Add :QEMU:AdditionalArguments: string \\\"$value\\\"" "$config"
}

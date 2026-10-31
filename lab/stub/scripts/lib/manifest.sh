#!/bin/bash

# Strict, non-executing KEY=value manifest reader for Tier 3 helpers.

manifest_validate() {
  local file="$1" allowed_keys="$2"
  [[ -f "$file" && ! -L "$file" ]] || return 1
  awk -F= -v allowed_keys="$allowed_keys" '
    BEGIN {
      split(allowed_keys, keys, " ")
      for (key_index in keys) allowed[keys[key_index]] = 1
    }
    /^[[:space:]]*($|#)/ { next }
    !/^[A-Z][A-Z0-9_]*=[^[:space:][:cntrl:]]+$/ { exit 1 }
    !($1 in allowed) || seen[$1]++ { exit 1 }
    END {
      for (key in allowed) if (seen[key] != 1) exit 1
    }
  ' "$file"
}

manifest_get() {
  local file="$1" key="$2"
  awk -F= -v key="$key" '
    $1 == key {
      count++
      value = substr($0, length($1) + 2)
    }
    END {
      if (count != 1) exit 1
      print value
    }
  ' "$file"
}

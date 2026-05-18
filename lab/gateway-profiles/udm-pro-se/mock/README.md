# UDM Pro SE Mocks

This directory contains project-owned mock sources for the UDM Pro SE firmware
lab profile.

- `ldpreload/`: C `LD_PRELOAD` shim modules. Filesystem interposition is split
  into `fs_paths.c`, `fs_open.c`, `fs_io.c`, `process_control.c`, and
  `socket_trace.c`; `swconfig.c` provides the RTL8370-style switch ABI,
  `auth.c` handles the narrow lab root-user check, `response_patch.c` keeps
  setup/readiness responses deterministic, `common.c` holds feature flags, and
  `ubnthal_redirect.h` shares declarations.
- `files/`: deterministic mock filesystem inputs copied into `/mock` for Docker
  and QEMU/UTM lab runs.
- `files/ubnthal/`: synthetic `board` and `system.info` identity files. These
  are lab values and must not be replaced with real serials, MACs, QR IDs, or
  manufacturing data.

Runtime mock data is staged into ignored artifact directories by the Docker and
QEMU/UTM lab scripts. Do not put extracted vendor firmware content here.

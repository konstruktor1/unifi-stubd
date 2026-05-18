# LD_PRELOAD Mock Source

This C shim is the shared userspace hardware boundary for the UDM Pro SE Docker
and QEMU/UTM lab paths. It lets selected firmware processes see deterministic
lab hardware state without granting them direct access to host `/proc`, `/sys`,
MTD, process-control, or switch-driver behavior.

The split is by responsibility:

- Path and file interception map selected firmware reads/writes into `/mock`.
- `swconfig.c` provides the RTL8370-style `libsw.so`/OpenWrt `swconfig` ABI
  surface that `ubios-udapi-server` expects.
- `response_patch.c` applies narrow byte-preserving response patches for VM
  setup/readiness cases.
- `auth.c` handles only the local root-user compatibility needed by the setup
  path.
- `process_control.c` keeps unsafe process actions as logged lab no-ops.
- `socket_trace.c` is optional diagnostics.

This is not an ASIC emulator and it cannot fix early kernel boot. It applies
only after userspace starts, which is why Docker can use it for UI/API
inspection while the UTM VM remains the stronger boot reference.

# Common Runtime Modules

Common modules are sourced by both Docker entry points: the reduced firmware
path and the webportal path. Put code here only when both paths need the same
behavior and the helper has no assumptions about UniFi Core, PostgreSQL, or the
web setup stack.

`kernel-artifacts.sh` records the mounted QEMU/UTM kernel deployment manifest
when it exists. Docker still runs on the host kernel; this manifest is for
comparison and traceability, not a claim that Docker booted the UDM kernel.

Keep common helpers side-effect-light. Anything that starts services or patches
web configuration belongs in the narrower runtime subdirectory.

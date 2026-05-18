# UDM Pro SE VM Initramfs Payload

This directory contains project-owned files that are copied into the lab
initramfs or into the UDM root filesystem during the QEMU/UTM boot path.

- `init-top/`: early initramfs hooks that run before the UDM mount scripts.
- `init-bottom/`: initramfs hooks that run after the UDM root is mounted but
  before `/sbin/init` switches into the firmware root.
- `module-lists/`: kernel module paths shared by the builder and initramfs
  hooks.
- `etc/`: static files copied directly into the initramfs.
- `etc/unifi-stubd-vm/`: lab note and copied module list visible inside the
  initramfs for diagnostics.
- `rootfs-payload/`: files installed into the mounted UDM root by the
  `init-bottom` installer.

Important payload areas:

- `rootfs-payload/sbin/`: narrow `ubnt-tools` and `ubnt-systool` wrappers for
  deterministic board/system metadata inside the VM.
- `rootfs-payload/etc/systemd/system/`: lab systemd units and drop-ins.
- `rootfs-payload/etc/systemd/system/network-init.service.d/`: orders the VM
  netdev preparation before firmware network initialization.
- `rootfs-payload/etc/systemd/system/udapi-server.service.d/`: replaces the
  UDAPI startup command with the lab-compatible invocation.
- `rootfs-payload/etc/systemd/system/unifi-core.service.d/`: adapts UniFi Core
  service type for the VM path.
- `rootfs-payload/etc/systemd/system/unifi.service.d/`: sets Network service
  environment used by the VM lab.
- `rootfs-payload/usr/local/lib/unifi-stubd-vm/`: guest helper scripts for
  netdev preparation, nginx setup config, support-state dumps, and link upkeep.
- `rootfs-payload/usr/local/share/unifi-stubd-vm/systemd/`: shared systemd
  drop-in that applies the mock library to selected firmware services.
- `rootfs-payload/usr/local/share/unifi-stubd-vm/http/`: templates and AWK
  filters used to keep the setup web path reachable while blocking destructive
  reset/reboot routes.
- `rootfs-payload/usr/local/share/unifi-stubd-vm/ssh/`: debug SSH drop-in used
  only when the local lab explicitly supplies a debug public key.

The VM initramfs does not replace UniFi OS. It keeps the vendor initramfs,
rootfs, overlay layout, and final `/sbin/init` path, then adds lab-only files
needed because QEMU `virt` is not an Annapurna Labs AL324 board.

`build-lab-initramfs.sh` removes `README.md` files from `rootfs-payload/` after
copying that tree, so documentation files in this source tree do not become
guest files.

Keep vendor firmware content out of this tree. These files are local lab
compatibility shims only.

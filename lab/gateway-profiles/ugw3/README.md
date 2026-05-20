# UGW3 QEMU Simulation

This folder contains project-owned helper files for running selected UGW3
firmware userspace components through QEMU-MIPS.

Unlike UXG-Pro and UXG-Lite, UGW3 is a legacy EdgeOS/USG firmware. The helper
container is therefore a Debian runner that mounts an extracted vendor rootfs
from a Docker volume and chroots into it with `qemu-mips-static`.

Current state:

- `mcad` starts.
- `/tmp/.mcad` is created inside the chroot.
- `mca-ctrl -t dump` works.
- Hardware identity is still placeholder data until a legacy board mock is
  added.

Use `docker-howto.md` for the reproducible setup.

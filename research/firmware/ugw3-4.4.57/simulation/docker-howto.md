# UGW3 Docker/QEMU Simulation How-To

Run commands from the repository root.

## Requirements

- Docker.
- The official UGW3 firmware tar downloaded separately into ignored
  `artifacts/`.
- No host MIPS binfmt registration is required for this runner. It executes
  `/usr/bin/qemu-mips-static` inside the chroot explicitly.

## Paths

```sh
RESEARCH=research/firmware/ugw3-4.4.57
ARTIFACTS="$RESEARCH/artifacts"
FW="$ARTIFACTS/7920-UGW3-4.4.57-803dc5671c6745dbb68c8dfa10145a8f.tar"
mkdir -p "$ARTIFACTS"
```

Verify the image:

```sh
shasum -a 256 "$FW"
```

Expected hash:

```text
08a35a626e9733018b2e49af92aa3474255136bb0178e0697f44c6d8042cdd74
```

## Extract Firmware Tar

```sh
rm -rf "$ARTIFACTS/extracted"
mkdir -p "$ARTIFACTS/extracted"
tar -C "$ARTIFACTS/extracted" -xf "$FW"
```

Expected files:

```text
vmlinux.tmp
vmlinux.tmp.md5
squashfs.tmp
squashfs.tmp.md5
version.tmp
compat
```

## Extract Rootfs To Docker Volume

Use a Docker volume so Linux metadata is preserved and macOS filesystem
collisions do not affect extraction.

```sh
docker volume create unifi-ugw3-rootfs

docker run --rm \
  -v "$PWD/$ARTIFACTS/extracted:/firmware:ro" \
  -v unifi-ugw3-rootfs:/rootfs \
  debian:bookworm-slim \
  sh -lc 'apt-get update &&
    apt-get install -y --no-install-recommends squashfs-tools &&
    rm -rf /rootfs/* &&
    unsquashfs -no-xattrs -f -d /rootfs /firmware/squashfs.tmp'
```

## Start Runner

```sh
docker compose -f "$RESEARCH/simulation/compose.yaml" up -d --build
```

Inspect logs:

```sh
docker compose -f "$RESEARCH/simulation/compose.yaml" logs --tail 120 ugw3-qemu
```

Run the local management CLI inside the same chroot:

```sh
docker compose -f "$RESEARCH/simulation/compose.yaml" exec ugw3-qemu \
  sh -lc 'chroot /firmware-rootfs /usr/bin/qemu-mips-static /usr/bin/mca-ctrl -t dump'
```

Stop:

```sh
docker compose -f "$RESEARCH/simulation/compose.yaml" down
```

## Current Limitation

The runner starts the legacy management agent, but it does not yet emulate USG
hardware identity. Expect placeholder values for model, serial, MAC, board
revision, and interface tables until a legacy board mock is added.

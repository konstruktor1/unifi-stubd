# UXG-Lite Docker Simulation How-To

Run commands from the repository root.

## Requirements

- Docker with Linux/ARM64 support.
- `squashfs-tools`.
- The official UXG firmware image downloaded separately into ignored
  `artifacts/`.

## Paths

```sh
PROFILE=lab/gateway-profiles/uxg-lite
ARTIFACTS="$PROFILE/artifacts"
FW="$ARTIFACTS/dad0-UXG-5.0.16-996c83e4-42a4-4dc7-bfa3-26894dc59cd7.bin"
SIM=/tmp/unifi-fw-sim-uxg-lite
mkdir -p "$ARTIFACTS"
```

Verify the image:

```sh
shasum -a 256 "$FW"
```

Expected hash:

```text
e2e361fc9b4296628f1b4fa10280449695c9df8fb8da95a440a7462044a4765c
```

## Extract Rootfs

The first SquashFS header starts at byte offset `15644578`.

```sh
SQUASHFS_OFFSET=15644578
tail -c +$((SQUASHFS_OFFSET + 1)) "$FW" > "$ARTIFACTS/rootfs.squashfs"
```

Import through a Linux filesystem or Docker volume to avoid macOS
case-insensitive path collisions:

```sh
docker volume create unifi-uxg-lite-rootfs

docker run --rm \
  -v "$PWD/$ARTIFACTS:/firmware:ro" \
  -v unifi-uxg-lite-rootfs:/rootfs \
  debian:bookworm-slim \
  sh -lc 'apt-get update &&
    apt-get install -y --no-install-recommends squashfs-tools &&
    rm -rf /rootfs/* &&
    unsquashfs -no-xattrs -f -d /rootfs /firmware/rootfs.squashfs'

docker run --rm \
  -v unifi-uxg-lite-rootfs:/rootfs \
  debian:bookworm-slim \
  tar -C /rootfs --numeric-owner -cpf - . \
  | docker import --platform linux/arm64 - uxg-lite-fw:5.0.16
```

## Prepare Mock Hardware Files

```sh
rm -rf "$SIM"
mkdir -p \
  "$SIM/ubnthal/status" \
  "$SIM/proc/sys/crypto" \
  "$SIM/proc/sys/kernel" \
  "$SIM/proc/sys/net/core" \
  "$SIM/proc/sys/net/ipv4" \
  "$SIM/proc/sys/net/ipv6/conf/all" \
  "$SIM/proc/sys/net/netfilter"

cp lab/gateway-profiles/uxg-lite/ubnthal_redirect.c \
  "$SIM/ubnthal_redirect.c"
```

Create `"$SIM/ubnthal/board"`:

```text
format=0002
version=0002
boardid=a677
vendorid=0777
bomrev=00000001
model_name=UXG
model_short=UXG Lite
model_number=UXG-Lite
model_description=Gateway Lite
model_url=http://ui.com
serial=02156D00A677
hwaddrbbase=02:15:6d:00:a6:77
hwaddrbase=02:15:6d:00:a6:77
```

Create `"$SIM/ubnthal/system.info"`:

```text
cpu=IPQ5018
cpuid=00000000
flashSize=16777216
ramsize=1073741824
vendorid=0777
systemid=a677
shortname=UXG
boardrevision=1
serialno=02156D00A677
manufid=003d
mfgweek=202607
qrid=SIMULATED
cpu_rev_id=00010000
macaddr=02:15:6d:00:a6:77
eth0.macaddr=02:15:6d:00:a6:77
eth1.macaddr=02:15:6d:00:a6:78
firmware=5.0.16
```

Create the simple mock values:

```sh
printf 'false\n' > "$SIM/ubnthal/status/IsLocated"
printf '0\n' > "$SIM/proc/sys/crypto/fips_enabled"
printf 'UXG-Lite\n' > "$SIM/proc/sys/kernel/hostname"
printf '(none)\n' > "$SIM/proc/sys/kernel/domainname"
printf '212992\n' > "$SIM/proc/sys/net/core/rmem_max"
printf '212992\n' > "$SIM/proc/sys/net/core/wmem_max"
printf '4096\n' > "$SIM/proc/sys/net/core/somaxconn"
printf '0\n' > "$SIM/proc/sys/net/ipv4/ip_forward"
printf '0\n' > "$SIM/proc/sys/net/ipv6/conf/all/forwarding"
printf '0\n' > "$SIM/proc/sys/net/netfilter/nf_conntrack_helper"
```

## Build Shim

```sh
SIM_DIR="$SIM" docker compose \
  -f "$PROFILE/compose.yaml" \
  --profile build-shim \
  run --rm shim-builder
```

## Start Partial Simulation

```sh
SIM_DIR="$SIM" docker compose \
  -f "$PROFILE/compose.yaml" \
  up -d --build firmware
```

Inspect:

```sh
docker compose -f "$PROFILE/compose.yaml" ps
docker compose -f "$PROFILE/compose.yaml" logs --tail 120 firmware
```

Inside the container, this profile currently reaches the UbiOS bridge-event
socket but not the normal UDAPI server socket:

```sh
docker compose -f "$PROFILE/compose.yaml" exec firmware \
  find /run /tmp -maxdepth 2 -type s -o -type f
```

Stop:

```sh
SIM_DIR="$SIM" docker compose \
  -f "$PROFILE/compose.yaml" \
  down
```

## Known Blocker

`ubios-udapi-server` does not create `/var/run/ubnt-udapi-server.sock` in the
current wrapper. Because of that, `mcad` starts but `mca-ctrl -t dump` cannot
connect to `/tmp/.mcad`.

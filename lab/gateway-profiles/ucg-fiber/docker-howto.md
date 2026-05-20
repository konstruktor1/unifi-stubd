# UCG-Fiber Docker Simulation How-To

Run commands from the repository root.

## Requirements

- Docker with Linux/ARM64 support.
- `squashfs-tools`.
- The official UCG-Fiber firmware image downloaded separately into ignored
  `research/firmware/ucg-fiber-5.0.16/artifacts/`.

## Paths

```sh
PROFILE=lab/gateway-profiles/ucg-fiber
RESEARCH=research/firmware/ucg-fiber-5.0.16
ARTIFACTS="$RESEARCH/artifacts"
FW="$ARTIFACTS/ca3a-UCGF-5.0.16-109206e5-e71e-4be4-b9c8-a4d4ae2ac799.bin"
SIM=/tmp/unifi-fw-sim-ucg-fiber
mkdir -p "$ARTIFACTS"
```

Verify the image:

```sh
shasum -a 256 "$FW"
```

Expected hash:

```text
7c8635974513413f19b4542c85b188d93a2fef38707ad53d237ce6a657e88ce5
```

## Extract Rootfs

The first SquashFS header starts at byte offset `12914826`.

```sh
SQUASHFS_OFFSET=12914826
tail -c +$((SQUASHFS_OFFSET + 1)) "$FW" > "$ARTIFACTS/rootfs.squashfs"
```

Import through a Linux filesystem or Docker volume to avoid macOS
case-insensitive path collisions:

```sh
docker volume create unifi-ucg-fiber-rootfs

docker run --rm \
  -v "$PWD/$ARTIFACTS:/firmware:ro" \
  -v unifi-ucg-fiber-rootfs:/rootfs \
  debian:bookworm-slim \
  sh -lc 'apt-get update &&
    apt-get install -y --no-install-recommends squashfs-tools &&
    rm -rf /rootfs/* &&
    unsquashfs -no-xattrs -f -d /rootfs /firmware/rootfs.squashfs'

docker run --rm \
  -v unifi-ucg-fiber-rootfs:/rootfs \
  debian:bookworm-slim \
  tar -C /rootfs --numeric-owner -cpf - . \
  | docker import --platform linux/arm64 - ucg-fiber-fw:5.0.16
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

cp "$PROFILE/ubnthal_redirect.c" "$SIM/ubnthal_redirect.c"
```

Create `"$SIM/ubnthal/board"`:

```text
format=0002
version=0002
boardid=a6a8
vendorid=0777
bomrev=00000001
model_name=UCGF
model_short=UCG Fiber
model_number=UCG-Fiber
model_description=UniFi Cloud Gateway Fiber
model_url=http://ui.com
serial=02156D00A6A8
hwaddrbbase=02:15:6d:00:a6:a8
hwaddrbase=02:15:6d:00:a6:a8
```

Create `"$SIM/ubnthal/system.info"`:

```text
cpu=IPQ9574
cpuid=00000000
flashSize=16777216
ramsize=4294967296
vendorid=0777
systemid=a6a8
shortname=UCGF
boardrevision=1
serialno=02156D00A6A8
manufid=003d
mfgweek=202607
qrid=SIMULATED
cpu_rev_id=00010000
macaddr=02:15:6d:00:a6:a8
eth0.macaddr=02:15:6d:00:a6:a8
eth1.macaddr=02:15:6d:00:a6:a9
eth2.macaddr=02:15:6d:00:a6:aa
eth3.macaddr=02:15:6d:00:a6:ab
eth4.macaddr=02:15:6d:00:a6:ac
eth5.macaddr=02:15:6d:00:a6:ad
eth6.macaddr=02:15:6d:00:a6:ae
firmware=5.0.16
```

Create the simple mock values:

```sh
printf 'false\n' > "$SIM/ubnthal/status/IsLocated"
printf '0\n' > "$SIM/proc/sys/crypto/fips_enabled"
printf 'UCG-Fiber\n' > "$SIM/proc/sys/kernel/hostname"
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

Inside the container, inspect sockets and logs before attaching this profile to
any controller lab:

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

The wrapper starts `ubios-udapi-server`, `udapi-bridge`, and `mcad`, and
`udapi-bridge` opens `lo:1080`. In the current mock environment,
`ubios-udapi-server` does not create `/var/run/ubnt-udapi-server.sock`, so
`mcad` does not expose `/tmp/.mcad` and `mca-ctrl -t dump` cannot complete.

Treat this as a startup-analysis profile until the missing runtime dependency
is identified and mocked with deterministic lab values.

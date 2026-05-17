# UDM Pro SE Docker Simulation How-To

Run commands from the repository root.

## Requirements

- Docker with Linux/ARM64 support.
- `squashfs-tools`.
- The official UDM Pro SE firmware image downloaded separately into ignored
  `research/firmware/udm-pro-se-5.0.16/artifacts/`.

## Paths

```sh
PROFILE=lab/gateway-profiles/udm-pro-se
RESEARCH=research/firmware/udm-pro-se-5.0.16
ARTIFACTS="$RESEARCH/artifacts"
FW="$ARTIFACTS/473c-UDMPROSE-5.0.16-511eddc1-cb19-476d-a02d-fcaf3dbddc29.bin"
SIM=/tmp/unifi-fw-sim-udm-pro-se
mkdir -p "$ARTIFACTS"
```

Verify the image:

```sh
shasum -a 256 "$FW"
```

Expected hash:

```text
7cf58f4563522220716f5025a7b2954b070df6be9364d7f60af0bc644512bce4
```

## Extract Rootfs

The first SquashFS header starts at byte offset `16141142`.

```sh
SQUASHFS_OFFSET=16141142
tail -c +$((SQUASHFS_OFFSET + 1)) "$FW" > "$ARTIFACTS/rootfs.squashfs"
```

Import through a Linux filesystem or Docker volume to avoid macOS
case-insensitive path collisions:

```sh
docker volume create unifi-udm-pro-se-rootfs

docker run --rm \
  -v "$PWD/$ARTIFACTS:/firmware:ro" \
  -v unifi-udm-pro-se-rootfs:/rootfs \
  debian:bookworm-slim \
  sh -lc 'apt-get update &&
    apt-get install -y --no-install-recommends squashfs-tools &&
    rm -rf /rootfs/* &&
    unsquashfs -no-xattrs -f -d /rootfs /firmware/rootfs.squashfs'

docker run --rm \
  -v unifi-udm-pro-se-rootfs:/rootfs \
  debian:bookworm-slim \
  tar -C /rootfs --numeric-owner -cpf - . \
  | docker import --platform linux/arm64 - udm-pro-se-fw:5.0.16
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

The UDM Pro SE startup path writes sysctl values during initialization, so this
profile mounts the mock directory read-write.

Create `"$SIM/ubnthal/board"`:

```text
format=0002
version=0002
boardid=ea2c
vendorid=0777
bomrev=00000001
model_name=UDMPROSE
model_short=UDM-SE
model_number=UDM-SE
model_description=UniFi Dream Machine SE
model_url=http://ui.com
serial=02156D00EA2C
hwaddrbbase=02:15:6d:00:ea:2c
hwaddrbase=02:15:6d:00:ea:2c
```

Create `"$SIM/ubnthal/system.info"`:

```text
cpu=AL324
cpuid=00000000
flashSize=16777216
ramsize=4294967296
vendorid=0777
systemid=ea2c
shortname=UDM-SE
boardrevision=1
serialno=02156D00EA2C
manufid=003d
mfgweek=202607
qrid=SIMULATED
cpu_rev_id=00010000
macaddr=02:15:6d:00:ea:2c
eth0.macaddr=02:15:6d:00:ea:2c
eth1.macaddr=02:15:6d:00:ea:2d
eth2.macaddr=02:15:6d:00:ea:2e
eth3.macaddr=02:15:6d:00:ea:2f
eth4.macaddr=02:15:6d:00:ea:30
eth5.macaddr=02:15:6d:00:ea:31
eth6.macaddr=02:15:6d:00:ea:32
eth7.macaddr=02:15:6d:00:ea:33
eth8.macaddr=02:15:6d:00:ea:34
eth9.macaddr=02:15:6d:00:ea:35
eth10.macaddr=02:15:6d:00:ea:36
firmware=5.0.16
```

Create the simple mock values:

```sh
printf 'false\n' > "$SIM/ubnthal/status/IsLocated"
printf '0\n' > "$SIM/proc/sys/crypto/fips_enabled"
printf 'UDM-Pro-SE\n' > "$SIM/proc/sys/kernel/hostname"
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

This wrapper starts far enough to prove that the firmware reads the mocked
board identity and writes redirected sysctl values through the mock tree. It
currently stops at `Failed to connect to switch chip`; with
`UNIFI_FW_SIM_ALLOW_PARTIAL=1` the container stays alive for log inspection.
In this partial state, `mca-ctrl -t dump` cannot complete because `/tmp/.mcad`
is not available.

Treat it as a startup-analysis profile until `mca-ctrl -t dump` works with
deterministic lab values and `mcad` uses the mocked identity instead of fallback
runtime data.

# Docker Simulation How-To

This is the reproducible setup used to run selected UXG-Pro firmware userspace
components in an isolated Docker container.

The goal is protocol and process analysis. Keep the container networkless unless
you intentionally connect it to a disposable lab controller.

## Requirements

- Docker Desktop or Docker Engine with Linux/ARM64 support.
- `squashfs-tools` for `unsquashfs`.
- The official `UXGPROV2-5.0.16.bin` firmware image downloaded separately.

On macOS with Homebrew:

```sh
brew install squashfs
```

On Debian/Ubuntu:

```sh
sudo apt-get update
sudo apt-get install -y squashfs-tools
```

## Paths

Run commands from the repository root.

```sh
RESEARCH=research/firmware/uxgpro-5.0.16
ARTIFACTS="$RESEARCH/artifacts"
FW="$ARTIFACTS/UXGPROV2-5.0.16.bin"
SIM=/tmp/unifi-fw-sim
mkdir -p "$ARTIFACTS"
```

Put the firmware image at:

```text
research/firmware/uxgpro-5.0.16/artifacts/UXGPROV2-5.0.16.bin
```

Verify the image:

```sh
shasum -a 256 "$FW"
```

Expected hash:

```text
18a7f198f71edc0161365114356239b0b370b4b90f664bb90f253b33f8b5658c
```

## Extract Rootfs

For this image, the SquashFS filesystem begins at byte offset `15807210` and
ends before the updater section at byte offset `346989298`.

```sh
SQUASHFS_OFFSET=15807210
UPDATER_OFFSET=346989298
ROOTFS_LEN=$((UPDATER_OFFSET - SQUASHFS_OFFSET))

dd if="$FW" \
  of="$ARTIFACTS/rootfs.squashfs" \
  bs=1 \
  skip="$SQUASHFS_OFFSET" \
  count="$ROOTFS_LEN" \
  status=progress

rm -rf "$ARTIFACTS/rootfs"
unsquashfs -d "$ARTIFACTS/rootfs" "$ARTIFACTS/rootfs.squashfs"
```

The `dd bs=1` form is slow but exact and portable. On a machine with GNU
coreutils you can replace it with a faster equivalent.

## Import Docker Image

```sh
tar -C "$ARTIFACTS/rootfs" --numeric-owner -cpf - . \
  | docker import --platform linux/arm64 - uxgpro-fw:5.0.16

docker image inspect uxgpro-fw:5.0.16 \
  --format 'id={{.Id}} os={{.Os}} arch={{.Architecture}} size={{.Size}}'
```

## Prepare Mock Hardware Files

The firmware expects Ubiquiti hardware metadata below `/proc/ubnthal` and a few
sysctls. The LD_PRELOAD shim redirects selected reads into `/mock`.

```sh
rm -rf "$SIM"
mkdir -p \
  "$SIM/ubnthal/status" \
  "$SIM/proc/sys/net/core" \
  "$SIM/proc/sys/net/ipv4" \
  "$SIM/proc/sys/net/ipv6/conf/all" \
  "$SIM/proc/sys/net/netfilter" \
  "$SIM/proc/sys/kernel"

cp "$RESEARCH/simulation/ubnthal_redirect.c" "$SIM/ubnthal_redirect.c"

cat > "$SIM/ubnthal/board" <<'EOF'
format=0002
version=0002
boardid=ea19
vendorid=0777
bomrev=0002d30b
model_name=UXGPRO
model_short=UXG Pro
model_number=UXG-PRO
model_description=Gateway Pro
model_url=http://ui.com
serial=000000000000
hwaddrbbase=00:15:6d:de:ad:00
hwaddrbase=00:15:6d:de:ad:00
EOF

cat > "$SIM/ubnthal/system.info" <<'EOF'
cpu=AL324V2
cpuid=411ed073
flashSize=16777216
ramsize=4294967296
vendorid=0777
systemid=ea19
shortname=UXGPRO
boardrevision=1
serialno=000000000000
manufid=003d
mfgweek=202607
qrid=SIMULATED
cpu_rev_id=00010000
macaddr=02:00:00:00:00:01
eth0.macaddr=00:15:6d:de:ad:00
firmware=5.0.16
EOF

printf 'false\n' > "$SIM/ubnthal/status/IsLocated"
printf '212992\n' > "$SIM/proc/sys/net/core/rmem_max"
printf '212992\n' > "$SIM/proc/sys/net/core/wmem_max"
printf '4096\n' > "$SIM/proc/sys/net/core/somaxconn"
printf '0\n' > "$SIM/proc/sys/net/ipv4/ip_forward"
printf '0\n' > "$SIM/proc/sys/net/ipv6/conf/all/forwarding"
printf '0\n' > "$SIM/proc/sys/net/netfilter/nf_conntrack_helper"
printf 'UXG-Pro\n' > "$SIM/proc/sys/kernel/hostname"
printf '(none)\n' > "$SIM/proc/sys/kernel/domainname"
printf '5\n' > "$SIM/proc/sys/kernel/printk_ratelimit"
```

## Build LD_PRELOAD Shim

Build the shim as ARM64 code in a disposable Debian container:

```sh
docker run --rm --platform linux/arm64 \
  -v "$SIM:/mock" \
  debian:bullseye \
  sh -lc 'apt-get update &&
    apt-get install -y --no-install-recommends gcc libc6-dev &&
    gcc -shared -fPIC -Wall -Wextra -O2 -ldl \
      -o /mock/libubnthal_redirect.so \
      /mock/ubnthal_redirect.c'
```

## Start Networkless Firmware Container

### Compose Option

The simulation folder contains a Dockerfile and Compose file for repeatable
startup. The Compose service still requires the imported base image
`uxgpro-fw:5.0.16` and the mock directory prepared above.

Build the shim through Compose:

```sh
SIM_DIR="$SIM" docker compose \
  -f "$RESEARCH/simulation/compose.yaml" \
  --profile build-shim \
  run --rm shim-builder
```

Start the firmware wrapper:

```sh
SIM_DIR="$SIM" docker compose \
  -f "$RESEARCH/simulation/compose.yaml" \
  up -d --build firmware
```

Compose starts `ubios-udapi-server`, `udapi-bridge`, and `mcad` automatically
through `/usr/local/bin/uxgpro-sim-start`.

Stop it:

```sh
SIM_DIR="$SIM" docker compose \
  -f "$RESEARCH/simulation/compose.yaml" \
  down
```

To run this firmware container against a local UniFi Network Application
controller, use `controller-lab.compose.yaml` as described in
`controller-lab.md`.

### Manual Docker Option

```sh
docker rm -f uxgpro-fw-fullsim 2>/dev/null || true

docker run -d \
  --name uxgpro-fw-fullsim \
  --platform linux/arm64 \
  --network none \
  --cap-drop ALL \
  --cap-add DAC_OVERRIDE \
  --cap-add FOWNER \
  --cap-add NET_ADMIN \
  --cap-add NET_RAW \
  --security-opt no-new-privileges \
  --mount "type=bind,source=$SIM,target=/mock" \
  --tmpfs /tmp:exec,nosuid,nodev \
  --tmpfs /run:exec,nosuid,nodev \
  --tmpfs /var/run:exec,nosuid,nodev \
  --tmpfs /var/log:nosuid,nodev \
  --tmpfs /data:exec,nosuid,nodev \
  uxgpro-fw:5.0.16 \
  /bin/sleep infinity
```

Check the container:

```sh
docker ps --filter name=uxgpro-fw-fullsim
```

## Start Firmware Processes

Start UDAPI first:

```sh
docker exec uxgpro-fw-fullsim /bin/bash --noprofile --norc -lc '
  mkdir -p /data/udapi-config/ubios-udapi-server
  : > /tmp/ubios-udapi-server.run.log
  : > /tmp/ubios-udapi-server.run.err
  nohup env LD_PRELOAD=/mock/libubnthal_redirect.so \
    /usr/bin/ubios-udapi-server \
      -c /data/udapi-config/ubios-udapi-server/ubios-udapi-server.state \
      -x -t \
    >/tmp/ubios-udapi-server.run.log \
    2>/tmp/ubios-udapi-server.run.err &
'
```

Wait for the socket:

```sh
docker exec uxgpro-fw-fullsim sh -lc \
  'until [ -S /var/run/ubnt-udapi-server.sock ]; do sleep 1; done; ls -l /var/run/ubnt-udapi-server.sock'
```

Start the REST bridge:

```sh
docker exec uxgpro-fw-fullsim /bin/bash --noprofile --norc -lc '
  : > /tmp/udapi-bridge.run.log
  : > /tmp/udapi-bridge.run.err
  nohup /usr/bin/udapi-bridge \
    -m UXGPRO \
    -M 00:15:6d:de:ad:00 \
    --rest-api-port 1080 \
    --rest-api-secure-port 0 \
    --rest-api-interface lo \
    -l - -x - \
    >/tmp/udapi-bridge.run.log \
    2>/tmp/udapi-bridge.run.err &
'
```

Start `mcad`:

```sh
docker exec uxgpro-fw-fullsim /bin/bash --noprofile --norc -lc '
  : > /tmp/mcad.run.out
  : > /tmp/mcad.run.err
  nohup env LD_PRELOAD=/mock/libubnthal_redirect.so \
    /usr/bin/mcad -n -s -v \
    >/tmp/mcad.run.out \
    2>/tmp/mcad.run.err &
'
```

## Inspect Runtime State

```sh
docker exec uxgpro-fw-fullsim pgrep -af 'ubios-udapi-server|udapi-bridge|mcad'

docker exec uxgpro-fw-fullsim /usr/bin/mca-ctrl -t dump
docker exec uxgpro-fw-fullsim /usr/bin/ubios-udapi-client -r GET /device
docker exec uxgpro-fw-fullsim /usr/bin/ubios-udapi-client -r GET /interfaces
docker exec uxgpro-fw-fullsim /usr/bin/ubios-udapi-client -r GET /system/configuration
```

Useful logs:

```sh
docker exec uxgpro-fw-fullsim tail -120 /tmp/ubios-udapi-server.run.err
docker exec uxgpro-fw-fullsim tail -120 /tmp/udapi-bridge.run.err
docker exec uxgpro-fw-fullsim tail -120 /tmp/mcad.run.err
```

## Isolated Adoption Probe

This only changes the simulated container state. The address is from
`192.0.2.0/24`, reserved for documentation.

```sh
docker exec uxgpro-fw-fullsim /usr/bin/mca-ctrl \
  -t connect \
  -s http://192.0.2.10:8080/inform \
  -k 00112233445566778899aabbccddeeff

docker exec uxgpro-fw-fullsim sh -lc '
  for f in \
    /run/system.inform \
    /run/system.controller \
    /run/system.controller.state \
    /run/system.controller.connect.error \
    /run/system.managed \
    /run/system.state
  do
    [ -e "$f" ] && printf "%s=" "$f" && cat "$f" && printf "\n"
  done
'
```

Expected result in a networkless container: `mcad` accepts the inform URL and
then reports the controller as unreachable.

## Cleanup

```sh
docker rm -f uxgpro-fw-fullsim
docker image rm uxgpro-fw:5.0.16
rm -rf "$SIM"
```

Remove `research/firmware/uxgpro-5.0.16/artifacts/` if you no longer need the
local firmware extraction.

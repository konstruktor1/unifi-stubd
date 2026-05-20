# Run All Lab Containers

This runbook starts every committed lab container stack in this repository. It
is for local research only; keep firmware images, extracted root filesystems,
captures, controller data, keys, and tokens out of Git.

## Controller Labs

There are two UniFi controller labs on purpose:

- `lab/stub/compose.yaml` is the generic `unifi-stubd` controller lab for the
  synthetic switch stub. It publishes the UniFi UI on
  `https://127.0.0.1:8443`.
- `lab/gateway-profiles/uxgpro/controller-lab.compose.yaml` is the separate
  UXG-Pro real-firmware adoption and MITM lab. When both labs run at the same
  time, publish this second UniFi UI on `https://127.0.0.1:9443`.

Use separate controller data because firmware adoption tests should not share a
site database with generic stub tests.

## Runtime Inputs

The firmware stacks require local ignored inputs:

- `research/firmware/ugw3-4.4.57/artifacts/extracted/squashfs.tmp`
- `research/firmware/uxg-lite-5.0.16/artifacts/rootfs.squashfs`
- `research/firmware/uxgpro-5.0.16/artifacts/rootfs.squashfs`
- `research/firmware/ucg-fiber-5.0.16/artifacts/rootfs.squashfs`
- `research/firmware/udm-pro-se-5.0.16/artifacts/rootfs.squashfs`
- mock hardware directories under `/tmp/unifi-fw-sim*`
- UDM Pro SE shared kernel payload under
  `lab/gateway-profiles/udm-pro-se-vm/artifacts/deploy/kernel/` when the UDM
  Pro SE Docker profile should log the same kernel inputs as the VM reference.

The UXG-Pro rootfs slice can be recreated from the official firmware image
using the offsets in `lab/gateway-profiles/uxgpro/docker-howto.md`.

## Recreate Docker Rootfs Inputs

Run from the repository root.

```sh
docker volume create unifi-ugw3-rootfs
docker run --rm \
  -v "$PWD/research/firmware/ugw3-4.4.57/artifacts/extracted:/firmware:ro" \
  -v unifi-ugw3-rootfs:/rootfs \
  debian:bookworm-slim \
  sh -lc 'apt-get update &&
    apt-get install -y --no-install-recommends squashfs-tools &&
    rm -rf /rootfs/* &&
    unsquashfs -quiet -no-xattrs -f -d /rootfs /firmware/squashfs.tmp'
```

```sh
docker volume create unifi-uxg-lite-rootfs
docker run --rm \
  -v "$PWD/research/firmware/uxg-lite-5.0.16/artifacts:/firmware:ro" \
  -v unifi-uxg-lite-rootfs:/rootfs \
  debian:bookworm-slim \
  sh -lc 'apt-get update &&
    apt-get install -y --no-install-recommends squashfs-tools &&
    rm -rf /rootfs/* &&
    unsquashfs -quiet -no-xattrs -f -d /rootfs /firmware/rootfs.squashfs'

docker run --rm \
  -v unifi-uxg-lite-rootfs:/rootfs \
  debian:bookworm-slim \
  tar -C /rootfs --numeric-owner -cpf - . \
  | docker import --platform linux/arm64 - uxg-lite-fw:5.0.16
```

```sh
docker volume create unifi-uxgpro-rootfs
docker run --rm \
  -v "$PWD/research/firmware/uxgpro-5.0.16/artifacts:/firmware:ro" \
  -v unifi-uxgpro-rootfs:/rootfs \
  debian:bookworm-slim \
  sh -lc 'apt-get update &&
    apt-get install -y --no-install-recommends squashfs-tools &&
    rm -rf /rootfs/* &&
    unsquashfs -quiet -no-xattrs -f -d /rootfs /firmware/rootfs.squashfs'

docker run --rm \
  -v unifi-uxgpro-rootfs:/rootfs \
  debian:bookworm-slim \
  tar -C /rootfs --numeric-owner -cpf - . \
  | docker import --platform linux/arm64 - uxgpro-fw:5.0.16
```

```sh
docker volume create unifi-ucg-fiber-rootfs
docker run --rm \
  -v "$PWD/research/firmware/ucg-fiber-5.0.16/artifacts:/firmware:ro" \
  -v unifi-ucg-fiber-rootfs:/rootfs \
  debian:bookworm-slim \
  sh -lc 'apt-get update &&
    apt-get install -y --no-install-recommends squashfs-tools &&
    rm -rf /rootfs/* &&
    unsquashfs -quiet -no-xattrs -f -d /rootfs /firmware/rootfs.squashfs'

docker run --rm \
  -v unifi-ucg-fiber-rootfs:/rootfs \
  debian:bookworm-slim \
  tar -C /rootfs --numeric-owner -cpf - . \
  | docker import --platform linux/arm64 - ucg-fiber-fw:5.0.16
```

```sh
docker volume create unifi-udm-pro-se-rootfs
docker run --rm \
  -v "$PWD/research/firmware/udm-pro-se-5.0.16/artifacts:/firmware:ro" \
  -v unifi-udm-pro-se-rootfs:/rootfs \
  debian:bookworm-slim \
  sh -lc 'apt-get update &&
    apt-get install -y --no-install-recommends squashfs-tools &&
    rm -rf /rootfs/* &&
    unsquashfs -quiet -no-xattrs -f -d /rootfs /firmware/rootfs.squashfs'

docker run --rm \
  -v unifi-udm-pro-se-rootfs:/rootfs \
  debian:bookworm-slim \
  tar -C /rootfs --numeric-owner -cpf - . \
  | docker import --platform linux/arm64 - udm-pro-se-fw:5.0.16
```

For the UDM Pro SE profile, also stage the shared mock and kernel inputs:

```sh
lab/gateway-profiles/udm-pro-se-vm/scripts/prepare-vm.sh
lab/gateway-profiles/udm-pro-se-vm/scripts/fetch-foreign-kernel.sh
lab/gateway-profiles/udm-pro-se-vm/scripts/prepare-mocks.sh
lab/gateway-profiles/udm-pro-se-vm/scripts/build-lab-initramfs.sh
lab/gateway-profiles/udm-pro-se-vm/scripts/deploy-kernel-artifacts.sh
```

## Build Firmware Wrapper Images

```sh
docker build --pull=false -t uxg-lite-fw-sim:5.0.16 \
  lab/gateway-profiles/uxg-lite

docker build --pull=false -t uxgpro-fw-sim:5.0.16 \
  lab/gateway-profiles/uxgpro

docker build --pull=false -t ucg-fiber-fw-sim:5.0.16 \
  lab/gateway-profiles/ucg-fiber

docker build --pull=false -t udm-pro-se-fw-sim:5.0.16 \
  lab/gateway-profiles/udm-pro-se

docker build --pull=false -t ugw3-fw-qemu:4.4.57 \
  lab/gateway-profiles/ugw3
```

## Start All Firmware Stacks

```sh
docker compose -f lab/gateway-profiles/ugw3/compose.yaml \
  up -d --no-build ugw3-qemu

SIM_DIR=/tmp/unifi-fw-sim-uxg-lite \
docker compose -f lab/gateway-profiles/uxg-lite/compose.yaml \
  up -d --no-build firmware

SIM_DIR=/tmp/unifi-fw-sim \
docker compose -f lab/gateway-profiles/uxgpro/compose.yaml \
  up -d --no-build firmware

SIM_DIR=/tmp/unifi-fw-sim-ucg-fiber \
docker compose -f lab/gateway-profiles/ucg-fiber/compose.yaml \
  up -d --no-build firmware

SIM_DIR=/tmp/unifi-fw-sim-udm-pro-se \
docker compose -f lab/gateway-profiles/udm-pro-se/compose.yaml \
  up -d --no-build firmware
```

Start the UDM Pro SE webportal override only when you want the partial UniFi OS
setup UI instead of the networkless firmware process command:

```sh
SIM_DIR=/tmp/unifi-fw-sim-udm-pro-se \
docker compose \
  -f lab/gateway-profiles/udm-pro-se/compose.yaml \
  -f lab/gateway-profiles/udm-pro-se/webportal.compose.yaml \
  up -d --no-build firmware
```

## Start Both Controller Labs

```sh
mkdir -p lab/stub/captures lab/gateway-profiles/uxgpro/captures

docker compose -f lab/stub/compose.yaml up -d --build
```

Create a temporary UXG-Pro controller-lab compose file that moves the second
UniFi UI from host port `8443` to `9443`:

```sh
TMP=/tmp/uxgpro-controller-lab-9443.yaml
docker compose -f lab/gateway-profiles/uxgpro/controller-lab.compose.yaml \
  config > "$TMP"
perl -0pi -e 's/published: "8443"/published: "9443"/g' "$TMP"

SIM_DIR=/tmp/unifi-fw-sim \
docker compose -f "$TMP" up -d --no-build
```

## Verify

```sh
docker ps -a --format 'table {{.Names}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}'
```

Expected long-running containers after startup:

- `stub-lab-unifi-db`
- `stub-lab-unifi-network`
- `stub-lab-inform-mitm`
- `stub`
- `uxgpro-unifi-db`
- `uxgpro-unifi-network`
- `uxgpro-inform-mitm`
- `uxgpro-fw-lab`
- `ugw3-fw-simulation-ugw3-qemu-1`
- `uxg-lite-fw-simulation-firmware-1`
- `uxgpro-fw-fullsim`
- `ucg-fiber-fw-simulation-firmware-1`
- `udm-pro-se-fw-simulation-firmware-1`

Partial ARM64 firmware profiles may start their process chain without exposing
the final `mcad` control socket. Use the per-profile `firmware.md` files for
current blockers.

## Stop

```sh
docker compose -f lab/stub/compose.yaml down
docker compose -f /tmp/uxgpro-controller-lab-9443.yaml down
docker compose -f lab/gateway-profiles/ugw3/compose.yaml down
docker compose -f lab/gateway-profiles/uxg-lite/compose.yaml down
docker compose -f lab/gateway-profiles/uxgpro/compose.yaml down
docker compose -f lab/gateway-profiles/ucg-fiber/compose.yaml down
docker compose -f lab/gateway-profiles/udm-pro-se/compose.yaml down
```

# Alle Lab-Container starten

Dieses Runbook startet alle im Repository abgelegten Lab-Container-Stacks. Es
ist nur fuer lokale Research-Arbeit gedacht. Firmware-Images, extrahierte
Rootfs-Baeume, Captures, Controller-Daten, Keys und Tokens bleiben aus Git
draussen.

## Zwei Controller-Labs

Es gibt absichtlich zwei UniFi-Controller-Labs:

- `lab/stub/compose.yaml` ist das generische `unifi-stubd`-Controller-Lab fuer
  den synthetischen Switch-Stub. Die UI liegt auf `https://127.0.0.1:8443`.
- `lab/gateway-profiles/uxgpro/controller-lab.compose.yaml` ist das getrennte
  UXG-Pro-Real-Firmware-Adoption- und MITM-Lab. Wenn beide Labs gleichzeitig
  laufen, wird diese zweite UI auf `https://127.0.0.1:9443` gelegt.

Die Controller-Daten bleiben getrennt, weil Firmware-Adoption-Tests nicht die
gleiche Site-Datenbank wie generische Stub-Tests nutzen sollen.

## Runtime-Eingaben

Die Firmware-Stacks brauchen lokale, ignorierte Eingaben:

- `research/firmware/ugw3-4.4.57/artifacts/extracted/squashfs.tmp`
- `research/firmware/uxg-lite-5.0.16/artifacts/rootfs.squashfs`
- `research/firmware/uxgpro-5.0.16/artifacts/rootfs.squashfs`
- `research/firmware/ucg-fiber-5.0.16/artifacts/rootfs.squashfs`
- `research/firmware/udm-pro-se-5.0.16/artifacts/rootfs.squashfs`
- Mock-Hardware-Verzeichnisse unter `/tmp/unifi-fw-sim*`
- Gemeinsame UDM-Pro-SE-Kernel-Ablage unter
  `lab/gateway-profiles/udm-pro-se-vm/artifacts/deploy/kernel/`, wenn das
  Docker-Profil dieselben Kernel-Eingaben wie die VM-Referenz protokollieren
  soll.

Der UXG-Pro-Rootfs-Slice kann aus dem offiziellen Firmware-Image mit den
Offsets in `lab/gateway-profiles/uxgpro/docker-howto.md` neu erstellt werden.

## Docker-Rootfs-Eingaben neu erzeugen

Aus dem Repository-Root ausfuehren.

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

Die ARM64-Profile werden jeweils gleich importiert:

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
docker run --rm -v unifi-uxg-lite-rootfs:/rootfs debian:bookworm-slim \
  tar -C /rootfs --numeric-owner -cpf - . \
  | docker import --platform linux/arm64 - uxg-lite-fw:5.0.16
```

Den gleichen Ablauf fuer `unifi-uxgpro-rootfs` mit
`research/firmware/uxgpro-5.0.16/artifacts` und fuer
`unifi-ucg-fiber-rootfs` mit
`research/firmware/ucg-fiber-5.0.16/artifacts` ausfuehren. Die Import-Ziele
sind `uxgpro-fw:5.0.16` und `ucg-fiber-fw:5.0.16`.
Den gleichen Ablauf fuer `unifi-udm-pro-se-rootfs` mit
`research/firmware/udm-pro-se-5.0.16/artifacts` ausfuehren. Das Import-Ziel
ist `udm-pro-se-fw:5.0.16`.

Fuer das UDM-Pro-SE-Profil auch die gemeinsamen Mock- und Kernel-Eingaben
erzeugen:

```sh
lab/gateway-profiles/udm-pro-se-vm/scripts/prepare-vm.sh
lab/gateway-profiles/udm-pro-se-vm/scripts/fetch-foreign-kernel.sh
lab/gateway-profiles/udm-pro-se-vm/scripts/prepare-mocks.sh
lab/gateway-profiles/udm-pro-se-vm/scripts/build-lab-initramfs.sh
lab/gateway-profiles/udm-pro-se-vm/scripts/deploy-kernel-artifacts.sh
```

## Wrapper-Images bauen

```sh
docker build --pull=false -t ugw3-fw-qemu:4.4.57 \
  lab/gateway-profiles/ugw3
docker build --pull=false -t uxg-lite-fw-sim:5.0.16 \
  lab/gateway-profiles/uxg-lite
docker build --pull=false -t uxgpro-fw-sim:5.0.16 \
  lab/gateway-profiles/uxgpro
docker build --pull=false -t ucg-fiber-fw-sim:5.0.16 \
  lab/gateway-profiles/ucg-fiber
docker build --pull=false -t udm-pro-se-fw-sim:5.0.16 \
  lab/gateway-profiles/udm-pro-se
```

## Firmware-Stacks starten

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

Das UDM-Pro-SE-Webportal-Override nur starten, wenn statt des networkless
Firmware-Prozesskommandos die teilweise UniFi-OS-Setup-UI gebraucht wird:

```sh
SIM_DIR=/tmp/unifi-fw-sim-udm-pro-se \
docker compose \
  -f lab/gateway-profiles/udm-pro-se/compose.yaml \
  -f lab/gateway-profiles/udm-pro-se/webportal.compose.yaml \
  up -d --no-build firmware
```

## Beide Controller-Labs starten

```sh
mkdir -p lab/stub/captures lab/gateway-profiles/uxgpro/captures
docker compose -f lab/stub/compose.yaml up -d --build
```

Fuer das UXG-Pro-Controller-Lab eine temporaere Compose-Datei erzeugen, damit
die zweite UI Hostport `9443` statt `8443` nutzt:

```sh
TMP=/tmp/uxgpro-controller-lab-9443.yaml
docker compose -f lab/gateway-profiles/uxgpro/controller-lab.compose.yaml \
  config > "$TMP"
perl -0pi -e 's/published: "8443"/published: "9443"/g' "$TMP"

SIM_DIR=/tmp/unifi-fw-sim \
docker compose -f "$TMP" up -d --no-build
```

## Pruefen

```sh
docker ps -a --format 'table {{.Names}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}'
```

Die ARM64-Firmware-Profile koennen teilweise starten, ohne schon den finalen
`mcad`-Control-Socket bereitzustellen. Der aktuelle Blocker steht jeweils in
`firmware.md` im Profilverzeichnis.

## Stoppen

```sh
docker compose -f lab/stub/compose.yaml down
docker compose -f /tmp/uxgpro-controller-lab-9443.yaml down
docker compose -f lab/gateway-profiles/ugw3/compose.yaml down
docker compose -f lab/gateway-profiles/uxg-lite/compose.yaml down
docker compose -f lab/gateway-profiles/uxgpro/compose.yaml down
docker compose -f lab/gateway-profiles/ucg-fiber/compose.yaml down
docker compose -f lab/gateway-profiles/udm-pro-se/compose.yaml down
```

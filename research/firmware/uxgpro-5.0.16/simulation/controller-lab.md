# Controller Lab Compose

`controller-lab.compose.yaml` runs the UXG-Pro firmware wrapper together with a
local UniFi Network Application and MongoDB.

This lab is intentionally separate from `compose.yaml`. The default simulation
uses `network_mode: none`; the controller lab gives the firmware container a
private Docker network so `mcad` can reach `http://unifi:8080/inform`.

## Components

- `firmware`: UXG-Pro firmware wrapper built from the local
  `uxgpro-fw:5.0.16` image.
- `inform-mitm`: mitmproxy reverse proxy for `http://unifi:8080/inform`.
- `unifi`: LinuxServer.io UniFi Network Application, reachable from the MITM
  as `http://unifi-controller:8080`.
- `unifi-db`: MongoDB `7.0`, pinned because MongoDB does not support automatic
  major-version upgrades.

The firmware and database share a Docker network marked `internal: true`. The
firmware container cannot reach the host network or the internet. The
controller also joins a normal Docker network so its UI can be published to
localhost on HTTPS port `8443`.

The lab starts Dropbear inside the firmware container because `mcad` waits for
an SSH daemon before it sends normal inform traffic. No SSH port is published
to the host; it is only reachable inside the private Docker lab network.

The lab uses static internal addresses because `ubios-udapi-server` rewrites
`/etc/resolv.conf` inside the firmware container and also takes control of
`eth0`. The firmware service therefore gets an `/etc/hosts` entry for `unifi`,
and the simulation start script restores a static lab address on `eth0` before
starting `mcad`:

| Service | Default address |
| --- | --- |
| `unifi` / MITM | `172.31.240.12` |
| `unifi-controller` | `172.31.240.10` |
| `unifi-db` | `172.31.240.11` |
| `firmware` | `172.31.240.20` |

Override `UNIFI_LAB_SUBNET`, `UNIFI_LAB_CONTROLLER_IP`,
`UNIFI_LAB_MITM_IP`, `UNIFI_LAB_MONGO_IP`, and `UNIFI_LAB_FIRMWARE_IP` if that
subnet conflicts with a local Docker setup.

References:

- https://docs.mitmproxy.org/stable/concepts/modes/
- https://docs.linuxserver.io/images/docker-unifi-network-application/
- https://help.ui.com/hc/en-us/articles/218506997-UniFi-Network-Required-Ports-Reference

## Start

Prepare the firmware rootfs image and mock directory first by following
`docker-howto.md` through `Build LD_PRELOAD Shim`.

From the repository root:

```sh
RESEARCH=research/firmware/uxgpro-5.0.16
SIM=/tmp/unifi-fw-sim

mkdir -p "$RESEARCH/simulation/captures"

SIM_DIR="$SIM" docker compose \
  -f "$RESEARCH/simulation/controller-lab.compose.yaml" \
  up -d --build
```

First startup can take several minutes while MongoDB initializes and the UniFi
Network Application creates its config.

Open the controller UI:

```text
https://localhost:8443
```

The certificate is self-signed.

## Inform Host

In the UniFi Network Application setup, keep device communication on port
`8080`. For this Docker-only lab, set the Inform Host override to:

```text
unifi
```

The firmware container resolves that name through `/etc/hosts` and uses:

```text
http://unifi:8080/inform
```

This mirrors the default firmware inform URL and avoids exposing port `8080` on
the host. The hostname points at the MITM container, which forwards to the real
controller service.

## MITM Capture

The `inform-mitm` service uses mitmproxy reverse mode:

```text
firmware -> http://unifi:8080/inform -> inform-mitm -> http://unifi-controller:8080/inform
```

Capture output is ignored by Git and written below:

```text
research/firmware/uxgpro-5.0.16/simulation/captures/
```

Files written there:

- `inform.flows`: mitmproxy flow archive.
- `events.jsonl`: one JSON line per request and response.
- `*-request.bin`: raw HTTP request body.
- `*-response.bin`: raw HTTP response body.

The Inform body is usually a binary `TNBU` packet and may contain adoption
keys or controller state once decrypted by a device. Treat the capture folder
as sensitive lab data.

Sanitized findings from one local run are documented in
`inform-mitm-analysis.md`.

Watch live MITM logs:

```sh
SIM_DIR="$SIM" docker compose \
  -f "$RESEARCH/simulation/controller-lab.compose.yaml" \
  logs -f inform-mitm
```

## Inspect

```sh
SIM_DIR="$SIM" docker compose \
  -f "$RESEARCH/simulation/controller-lab.compose.yaml" \
  ps

SIM_DIR="$SIM" docker compose \
  -f "$RESEARCH/simulation/controller-lab.compose.yaml" \
  exec firmware /usr/bin/mca-ctrl -t dump

SIM_DIR="$SIM" docker compose \
  -f "$RESEARCH/simulation/controller-lab.compose.yaml" \
  logs -f unifi

SIM_DIR="$SIM" docker compose \
  -f "$RESEARCH/simulation/controller-lab.compose.yaml" \
  logs -f inform-mitm
```

Force the simulated firmware to retry the lab controller:

```sh
SIM_DIR="$SIM" docker compose \
  -f "$RESEARCH/simulation/controller-lab.compose.yaml" \
  exec firmware /usr/bin/mca-ctrl \
    -t connect \
    -s http://unifi:8080/inform
```

## Reset

Remove containers and lab data volumes:

```sh
SIM_DIR="$SIM" docker compose \
  -f "$RESEARCH/simulation/controller-lab.compose.yaml" \
  down -v
```

This deletes the local UniFi Network Application and MongoDB state for the lab.

## Notes

- The controller image requires an external MongoDB instance.
- MongoDB user creation only runs on first initialization of the database
  volume. If credentials or database names change, reset with `down -v`.
- Do not use these default lab passwords outside an isolated throwaway setup.
- To expose additional UniFi ports for real external devices, add port mappings
  deliberately and keep port `8080` unchanged on both sides.

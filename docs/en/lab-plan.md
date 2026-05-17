# Lab Plan

## Goal

The MVP is reached when the controller sees a fake switch as adoptable and keeps it connected after adoption.

## Recommended Setup

```text
UniFi Network Controller
  192.0.2.10

Linux host / Proxmox lab
  192.0.2.50
  unifi-stubd

Optional:
  real UniFi switch for comparison PCAPs
```

## Controller Ports

Relevant ports for this project:

- UDP `10001`: device discovery.
- TCP `8080`: device inform.
- UDP `3478`: STUN, relevant later.
- TCP `8443`: controller UI/API.
- TCP `5671`: traffic flow logging on UXG, relevant later.
- UDP `10101`: client fingerprinting, relevant later.

Source: [UniFi Required Ports Reference](https://help.ui.com/hc/en-us/articles/218506997-UniFi-Network-Required-Ports-Reference)

## Capture

On the controller or mirror port:

```sh
sudo tcpdump -i any -nn -s0 -w unifi-inform.pcap 'udp port 10001 or tcp port 8080'
```

On the stub host:

```sh
sudo tcpdump -i any -nn -s0 'udp port 10001 or tcp port 8080'
```

## Debug Logs

Useful UniFi Controller log areas:

- `discover`
- `inform`
- `devmgr`
- `ssh`

Typical errors to look for:

- `invalid inform_ip`
- `inform decrypt error`
- `Inform Invalid`
- `ADOPTING -> UNKNOWN`
- `INFORM_ERROR`

## Lab Sequence

1. `make lint`
2. `make test`
3. `go run ./cmd/unifi-stubd -dry-run`
4. `go run ./cmd/unifi-stubd -once`
5. Check whether the controller sees a device.
6. Open the PCAP and inspect TLVs.
7. Implement default-key inform POST.
8. Decode controller responses.
9. Trigger adoption and persist `setparam`.

## Controller Lab Command

Lab state:

- UniFi Controller: `192.0.2.10`
- Host IP for the controller path: `192.0.2.50`
- Fake MAC: `02:11:22:33:44:55`
- Fake model: `US16P150`
- Fake ports: `16`

The `192.0.2.0/24` addresses are documentation examples. Replace them with
addresses from an isolated lab network.

Send discovery plus one minimal inform heartbeat:

```sh
go run ./cmd/unifi-stubd \
  -profile us16p150 \
  -mac 02:11:22:33:44:55 \
  -ip 192.0.2.50 \
  -hostname proxmox-vmbr0 \
  -controller http://192.0.2.10:8080/inform \
  -once
```

Send a direct L3 inform heartbeat without UDP discovery:

```sh
go run ./cmd/unifi-stubd \
  -profile us16p150 \
  -mac 02:11:22:33:44:55 \
  -ip 192.0.2.50 \
  -hostname proxmox-vmbr0 \
  -controller http://192.0.2.10:8080/inform \
  -no-discovery \
  -once
```

Built-in SSH for advanced adoption:

```sh
sudo install -m 0755 unifi-stubd /usr/local/bin/unifi-stubd
sudo install -d -m 0755 /etc/unifi-stubd /var/lib/unifi-stubd
sudo install -m 0600 packaging/linux/etc/unifi-stubd/config.yaml /etc/unifi-stubd/config.yaml
sudo /usr/local/bin/unifi-stubd
```

The controller can then use `ubnt` / `ubnt` against port `22` for advanced adoption. Management SSH on the lab system should be moved to another port in this setup.

## Docker Controller Lab

For local gateway-profile simulation without vendor firmware, use the Docker
Compose lab in `lab/controller-gateway-stubs.compose.yaml`. It starts a UniFi
Network Application, MongoDB, an inform MITM, and one selected `unifi-stubd`
gateway profile.

Each gateway profile has its own Dockerfile under `lab/gateway-profiles/`, so
the image entrypoint carries the selected profile. Compose only supplies
lab-specific runtime values such as MAC address, IP address, hostname, and
controller URL.

Start the Gateway Lite profile:

```sh
mkdir -p lab/captures
docker compose -f lab/controller-gateway-stubs.compose.yaml \
  --profile uxg-lite \
  up -d --build
```

Available stub profiles are `ugw3`, `uxg-lite`, `uxgpro`, and `ucg-fiber`. The
`gateways` Compose profile starts all four for packet-shape comparison, but adoption
testing should normally use one gateway per clean controller site.

Open the UI at `https://localhost:8443`. In the UniFi setup, keep TCP `8080`
for device communication and set the Inform Host override to `unifi`. The
stub's `http://unifi:8080/inform` traffic goes through the MITM container and
raw local captures are written to ignored `lab/captures/`.

The same repository also tracks real firmware simulation profiles in
`research/firmware/profiles.yaml`. Those entries are not just `unifi-stubd`
profiles: each real firmware profile needs a local vendor firmware image,
extracted rootfs, architecture notes, and a matched process wrapper. Currently
only UXG-Pro `5.0.16` has a working adopted controller lab.

The Compose lab uses LinuxServer.io's UniFi Network Application image with an
external MongoDB container. Ubiquiti's current self-hosting direction is UniFi
OS Server, but Ubiquiti documents that it is not provided as a standalone
Docker/Podman container.

OpenRC service:

```sh
sudo install -m 0755 packaging/linux/etc/init.d/unifi-stubd /etc/init.d/unifi-stubd
sudo rc-update add unifi-stubd default
sudo rc-service unifi-stubd restart
```

Systemd service:

```sh
sudo install -m 0644 packaging/linux/usr/lib/systemd/system/unifi-stubd.service /etc/systemd/system/unifi-stubd.service
sudo systemctl daemon-reload
sudo systemctl enable --now unifi-stubd.service
```

## Packaging

Build all supported package formats:

```sh
make package
```

Build one format:

```sh
make package-deb
make package-rpm
make package-arch
make package-tgz
make package-freebsd-tgz
```

Override version, release, or target architecture:

```sh
PKG_VERSION=0.1.0 PKG_RELEASE=1 PKG_GOARCH=amd64 \
  PKG_MAINTAINER='Name <email@example.com>' make package
```

Output files are written to `dist/packages/`. Native Debian, RPM, and Arch Linux packages are built with nFPM from `packaging/nfpm.yaml`; the Linux and FreeBSD `.tar.gz` packages are built from their OS-specific staging trees. FreeBSD/OPNsense is currently stub-only.

Layout:

- Code: `/usr/local/bin/unifi-stubd`
- Config: `/etc/unifi-stubd/config.yaml`
- Adoption SSH key: `/etc/unifi-stubd/ssh_host_rsa_key`
- Runtime state: `/var/lib/unifi-stubd/adoption.env`
- Logs: `/var/log/unifi-stubd.log`, `/var/log/unifi-stubd.err`

Use `lab/` for lab switch identities, or inspect
`packaging/installed-files.md` for the packaged Linux and FreeBSD file trees.

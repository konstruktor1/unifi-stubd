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

For the plain `unifi-stubd` switch stub, use the dedicated Docker Compose lab
in `lab/stub/compose.yaml`. The directory, Compose service, default container
name, hostname, and persistent volume are declared as `stub`:

```text
lab/stub/compose.yaml
lab/stub/configs/hosts/stub/config.yaml
services.stub
container_name: stub
hostname: stub
volume: stub_state
```

Start the generic stub service and its controller/MITM dependencies:

```sh
mkdir -p lab/stub/captures
docker compose -f lab/stub/compose.yaml up -d --build stub
```

The `stub` service builds the root `Dockerfile` and passes
`${UNIFI_STUB_PROFILE:-us8}` at runtime. The default emulated UniFi profile is
`us8`; the Docker path and container identity remain `stub`. Shared lab
defaults for the generic and temporary test hosts live under
`lab/stub/configs/hosts/<hostname>/config.yaml`.

For gateway firmware simulation, use the per-profile labs under
`lab/gateway-profiles/`. Those directories are real firmware wrappers, not
`internal/device` stub profile copies.

For the current repository status and the UDM Pro SE VM reference summary, see
`project-status.md`.

Current gateway firmware labs:

- `lab/gateway-profiles/ugw3/`: QEMU-MIPS runner for an extracted UGW3 rootfs.
- `lab/gateway-profiles/uxg-lite/`: ARM64 UbiOS userspace wrapper; partial
  simulation.
- `lab/gateway-profiles/uxgpro/`: ARM64 UbiOS userspace wrapper plus
  controller/MITM lab.
- `lab/gateway-profiles/ucg-fiber/`: ARM64 UbiOS userspace wrapper; partial
  simulation.
- `lab/gateway-profiles/udm-pro-se/`: ARM64 UbiOS userspace wrapper; reaches
  the UDAPI socket and `mca-ctrl -t dump` with a deterministic RTL8370-style
  switch mock. The optional Docker webportal override exposes a partial UniFi
  OS setup surface through modular `network-app/` and `systemd-dbus/`
  CommonJS facades.
- `lab/gateway-profiles/udm-pro-se-vm/`: real `qemu-system-aarch64` VM boot
  profile using copied local UDM Pro SE firmware artifacts. The direct vendor
  kernel hangs before serial output on QEMU `virt`; the foreign-kernel
  `udm-systemd` path reaches UDM firmware `systemd`, applies userspace hardware
  mocks, completes `network-init.service`, starts `ubios-udapi-server`,
  exercises `/firewall/nat`, and reaches a serial login prompt.

Run a firmware simulation:

```sh
docker compose -f lab/gateway-profiles/ugw3/compose.yaml up -d --build
docker compose -f lab/gateway-profiles/uxg-lite/compose.yaml up -d --build
docker compose -f lab/gateway-profiles/uxgpro/compose.yaml up -d --build
docker compose -f lab/gateway-profiles/ucg-fiber/compose.yaml up -d --build
docker compose -f lab/gateway-profiles/udm-pro-se/compose.yaml up -d --build
```

For the UDM Pro SE Docker webportal inspection path, prepare the shared kernel
payload and start the override:

```sh
lab/gateway-profiles/udm-pro-se-vm/scripts/prepare-vm.sh
lab/gateway-profiles/udm-pro-se-vm/scripts/fetch-foreign-kernel.sh
lab/gateway-profiles/udm-pro-se-vm/scripts/prepare-mocks.sh
lab/gateway-profiles/udm-pro-se-vm/scripts/build-lab-initramfs.sh
lab/gateway-profiles/udm-pro-se-vm/scripts/deploy-kernel-artifacts.sh
SIM_DIR=/tmp/unifi-fw-sim-udm-pro-se \
  docker compose \
    -f lab/gateway-profiles/udm-pro-se/compose.yaml \
    -f lab/gateway-profiles/udm-pro-se/webportal.compose.yaml \
    up -d --build firmware
```

That path is documented in `lab/gateway-profiles/udm-pro-se/docker-howto.md`.
It is a setup/API inspection wrapper; native firmware boot behavior belongs to
the QEMU/UTM VM profile.

Run the UDM Pro SE VM profile:

```sh
lab/gateway-profiles/udm-pro-se-vm/scripts/prepare-vm.sh
lab/gateway-profiles/udm-pro-se-vm/scripts/run-direct-kernel.sh
```

Run the UDM Pro SE VM path that reaches firmware `systemd`:

```sh
lab/gateway-profiles/udm-pro-se-vm/scripts/prepare-vm.sh
lab/gateway-profiles/udm-pro-se-vm/scripts/fetch-foreign-kernel.sh
lab/gateway-profiles/udm-pro-se-vm/scripts/prepare-mocks.sh
lab/gateway-profiles/udm-pro-se-vm/scripts/build-lab-initramfs.sh
lab/gateway-profiles/udm-pro-se-vm/scripts/deploy-kernel-artifacts.sh
UDM_PRO_SE_FOREIGN_MODE=udm-systemd \
  lab/gateway-profiles/udm-pro-se-vm/scripts/run-foreign-kernel.sh
```

The direct QEMU runner can use one transparent `vmnet-bridged` LAN NIC. The UTM
profile uses two NICs that map closer to the UDM Pro SE front panel: UTM
`Shared` / NAT becomes guest `eth9` for the first SFP+ WAN role, and UTM `Host`
becomes guest `eth8` for the 2.5G RJ45 LAN role attached to `br0`. The guest
keeps `br0` on `192.168.1.1/24` and adds `192.168.128.2/24` for host-only
access. The versioned UTM inputs live in
`lab/gateway-profiles/udm-pro-se-vm/utm/`; the generated UTM bundle remains
local. The shared kernel deployment payload is staged under
`lab/gateway-profiles/udm-pro-se-vm/artifacts/deploy/kernel/` and is mounted by
the Docker profile read-only for comparison. After nginx starts in the guest,
check the web path from an isolated lab segment with:

```sh
curl -k https://192.168.1.1/
```

In UTM Shared/NAT mode, first verify the guest address assigned by UTM and test
that address directly:

```sh
curl -k https://<utm-shared-guest-ip>/
curl -k https://<utm-shared-guest-ip>/api/system
```

The profile also writes a UTM `Network:0:PortForward` entry that is intended to
publish guest `443` on the Mac as:

```text
https://127.0.0.1:10443/
```

The latest observed UTM CLI run did not bind `127.0.0.1:10443` natively even
though the plist entry was present. If that URL works, check whether it is
UTM's own forward or an explicit local TCP helper before treating it as a UTM
feature.

The lab initramfs uses the vendor setup nginx template for this QEMU path and
keeps WAN ingress explicit. The older QEMU user-mode forwarding path remains
available explicitly with `UDM_PRO_SE_VM_NET=user-lan`, but it is not the
preferred VM reference path.

Run the UXG-Pro controller/MITM lab:

```sh
mkdir -p lab/gateway-profiles/uxgpro/captures
docker compose -f lab/gateway-profiles/uxgpro/controller-lab.compose.yaml up -d --build
```

Firmware images, extracted rootfs trees, raw captures, adoption keys,
controller tokens, certificates, and private controller data stay out of Git.

The same repository also tracks safe firmware research summaries in
`research/firmware/profiles.yaml`. Currently only UXG-Pro `5.0.16` has a
working adopted controller lab.

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
- Adoption SSH key: `/var/lib/unifi-stubd/ssh_host_rsa_key`
- Runtime state: `/var/lib/unifi-stubd/adoption.env`
- Logs: `/var/log/unifi-stubd.log`, `/var/log/unifi-stubd.err`

Use `lab/` for lab switch identities, or inspect
`packaging/installed-files.md` for the packaged Linux and FreeBSD file trees.

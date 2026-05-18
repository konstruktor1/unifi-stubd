# UTM Profile Files

This directory contains the project-owned UTM profile inputs for the UDM Pro SE
VM lab. It intentionally does not contain a copied UTM `config.plist`: real UTM
bundles contain local UUIDs, drive image names, paths, and generated state. Use
`scripts/install-utm-profile.sh` to apply these reproducible inputs to a cloned
UTM VM.

## Files

- `defaults.env`: default UTM values used by `install-utm-profile.sh`.
- `bootargs.txt`: Linux kernel command line injected into the generated UTM
  device tree and passed through QEMU `-append`.
- `../kernel/README.md`: describes the shared kernel deployment payload used
  by both UTM and Docker.
- `install/common.sh`: shared input, path, and tool checks.
- `install/plist.sh`: PlistBuddy helper functions.
- `install/boot.sh`: kernel, initramfs, and generated DTB deployment.
- `install/drives.sh`: writable disk conversion and UTM drive entries.
- `install/network.sh`: UTM NICs, port-forward intent, and QEMU network
  arguments.
- `install/system.sh`: serial-only VM shape, disabled sharing/input, and
  summary output.

Generated files stay outside Git:

- `artifacts/deploy/kernel/`
- `artifacts/utm/virt-udm-bootargs.dts`
- `artifacts/utm/virt-udm-bootargs.dtb`
- UTM bundle `Data/udm-foreign-kernel`
- UTM bundle `Data/udm-lab-initramfs.cpio.gz`
- UTM bundle `Data/virt-udm-bootargs.dtb`
- The converted writable UTM disk image inside the cloned bundle.

The installer is intentionally split so profile mutations are auditable:

- `install/common.sh` validates paths and external tools before mutating a UTM
  clone.
- `install/plist.sh` centralizes `PlistBuddy` operations.
- `install/boot.sh` deploys boot inputs from `artifacts/deploy/kernel/` or
  compatible fallback artifacts.
- `install/drives.sh` prepares writable disk images and drive entries.
- `install/network.sh` owns the two-NIC UDM port mapping and writes the
  intended HTTPS forward.
- `install/system.sh` enforces the serial-only, 4 GiB, non-graphical VM shape.

## Expected UTM Shape

The installed profile is a serial-only `aarch64` / `virt` VM with 4 GiB RAM:

```text
System:
  Architecture: aarch64
  Target: virt
  CPU count: 4
  Memory: 4096 MiB
  UEFI: disabled
  Display: none
  Serial: TCP server on 127.0.0.1:15555
```

Network mapping:

```text
UTM Network 0:
  Mode: Shared
  Hardware: virtio-net-pci
  MAC: 02:15:6d:00:ea:35
  Guest role: eth9, first 10G SFP+ WAN
  Port forward intent: host 127.0.0.1:10443 -> guest 443

UTM Network 1:
  Mode: Host
  Hardware: virtio-net-pci
  MAC: 02:15:6d:00:ea:34
  Guest role: eth8, 2.5G RJ45 LAN on br0
```

Inside the guest, the lab initramfs keeps `eth8` attached to `br0` with
`192.168.1.1/24` and adds `192.168.128.2/24` for macOS host-only access. The
web setup surface is therefore expected on guest port `443`.

The installer writes a UTM plist port-forward entry for:

```text
https://127.0.0.1:10443/
```

The latest observed UTM CLI run used `vmnet-shared` and `vmnet-host` network
backends but did not bind `127.0.0.1:10443` natively, even though the plist
entry was present. Treat that URL as a configured intent until it is verified
with `lsof` or `curl`. If it works because a local TCP helper is running, that
helper is separate from UTM's own port-forward implementation.

Latest observed UTM result:

- `install-utm-profile.sh` reapplied the profile with serial-only display and
  4 GiB RAM.
- `utmctl start --hide UDM-Pro-SE-QEMU` started the VM. `utmctl` can print
  macOS OSStatus event errors while the VM still reaches `started`.
- The serial socket on `127.0.0.1:15555` showed `ubios-udapi-server` and the
  lab UDAPI payloads.
- Direct HTTPS to the UTM Shared/NAT guest address worked, and `/api/system`
  reported `hasInternet=true`, `eth9` as the SFP+ WAN role, and `br0` with
  `192.168.1.1/24` plus `192.168.128.2/24`.
- Login and browser refresh stayed on the Network dashboard when the same host
  name was used consistently. Mixing `localhost` and `127.0.0.1` creates
  separate cookie scopes.
- `/api/setup/support/generate` timed out in the VM path and is not yet a
  reliable UTM check.

## Apply To A UTM Clone

Prepare the firmware, foreign kernel, mocks, and lab initramfs first:

```sh
lab/gateway-profiles/udm-pro-se-vm/scripts/prepare-vm.sh
lab/gateway-profiles/udm-pro-se-vm/scripts/fetch-foreign-kernel.sh
lab/gateway-profiles/udm-pro-se-vm/scripts/prepare-mocks.sh
lab/gateway-profiles/udm-pro-se-vm/scripts/build-lab-initramfs.sh
lab/gateway-profiles/udm-pro-se-vm/scripts/deploy-kernel-artifacts.sh
```

Create or clone a UTM VM named `UDM-Pro-SE-QEMU`, then apply the profile:

```sh
lab/gateway-profiles/udm-pro-se-vm/scripts/install-utm-profile.sh
```

Local overrides can be set in the environment:

```sh
UDM_PRO_SE_UTM_NAME=UDM-Pro-SE-QEMU \
UDM_PRO_SE_UTM_IFNAME=en0 \
UDM_PRO_SE_UTM_HTTPS_HOST_PORT=10443 \
  lab/gateway-profiles/udm-pro-se-vm/scripts/install-utm-profile.sh
```

Read the serial console:

```sh
(sleep 600) | nc 127.0.0.1 15555
```

Start without opening the UTM UI:

```sh
utmctl start --hide UDM-Pro-SE-QEMU
```

## Verify Web Access

Use the direct UTM Shared/NAT guest address first. The exact address is assigned
by UTM and can change between runs:

```sh
curl -k https://<utm-shared-guest-ip>/
curl -k https://<utm-shared-guest-ip>/api/system
```

Then check whether the configured localhost port is actually bound:

```sh
lsof -nP -iTCP:10443 -sTCP:LISTEN
curl -k https://127.0.0.1:10443/
```

If no listener exists on `127.0.0.1:10443`, UTM did not activate the native
forward in that run. Use the direct guest address or start an explicit local
TCP forwarder for browser convenience, and document that helper separately.

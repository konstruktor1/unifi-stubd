# Roadmap

## Phase 0: Pin the Lab

- [ ] Choose and document a UniFi Network Controller version.
- [ ] Run the controller in an isolated lab.
- [ ] Capture a real UniFi switch inform sequence if hardware is available.
- [ ] Enable debug logs for `inform`, `discover`, `devmgr`.

## Phase 1: Discovery

- [x] Discovery TLV builder.
- [x] Broadcast/multicast sender.
- [ ] Controller validation: does the fake device appear as Pending Adoption?
- [ ] Document TLV diff against a real switch.

## Phase 2: Inform Without Adoption

- [x] `TNBU` header encoder/decoder.
- [x] AES-CBC + zlib foundation.
- [x] AES-GCM foundation.
- [x] Minimal inform client with default key.
- [ ] Validate default key against controller lab.
- [x] Decode and log controller responses.

## Phase 3: Adoption

- [ ] Parse `_type: setparam`.
- [ ] Split `mgmt_cfg` into key/value data.
- [ ] Persist `authkey`, `cfgversion`, `use_aes_gcm`, `inform_url`.
- [ ] Send two quick inform requests after adoption.
- [ ] Reach connected state.

## Phase 4: Fake Switch Payload

- [x] Minimal switch payload.
- [ ] Extend `port_table` with stable defaults.
- [ ] Fill `mac_table` from Linux bridge FDB.
- [ ] Read port counters from sysfs.
- [ ] Model several virtual ports for `vmbr0`, `tap*`, `veth*`.

## Phase 5: Operations

- [x] OpenRC service.
- [x] systemd unit.
- [x] YAML configuration wired into the daemon.
- [x] Package builders for Debian, RPM, Arch Linux, and tgz.
- [ ] Rotating debug log.
- [ ] Healthcheck/status command.
- [ ] README with Proxmox lab example.

## Phase 6: Later Research

- [x] Built-in SSH adoption for `syswrapper.sh set-adopt` and `mca-cli-op set-inform`.
- [ ] Gateway profile `UGW3`/`UXG`.
- [ ] Synthesize DPI fields from NetFlow/OPNsense/ntopng.
- [ ] Compatibility matrix per UniFi Network version.

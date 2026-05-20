# Roadmap

## Phase 0: Pin the Lab

- [x] Validate a controller-known 10G profile: `USAGGPRO` online/adopted.
- [x] Document operation modes and the current live lab state.
- [x] Start a UniFi Network compatibility matrix.
- [x] Run the controller in an isolated lab.
- [ ] Capture a real UniFi switch inform sequence if hardware is available.
- [ ] Enable debug logs for `inform`, `discover`, `devmgr`.

## Phase 1: Discovery

- [x] Discovery TLV builder.
- [x] Broadcast/multicast sender.
- [x] Controller validation: Docker lab confirms a fake device appears as Pending Adoption.
- [ ] Document TLV diff against a real switch.

## Phase 2: Inform Without Adoption

- [x] `TNBU` header encoder/decoder.
- [x] AES-CBC + zlib foundation.
- [x] AES-GCM foundation.
- [x] Minimal inform client with default key.
- [x] Validate default-key inform path against the Docker controller lab.
- [x] Decode and log controller responses.

## Phase 3: Adoption

- [x] Parse `_type: setparam`.
- [x] Split `mgmt_cfg` into key/value data.
- [x] Persist `authkey`, `cfgversion`, `use_aes_gcm`, `inform_url`.
- [x] Verify post-adoption connected state in the Docker controller lab.
- [x] Reach connected state with `USAGGPRO`.

## Phase 4: Fake Switch Payload

- [x] Minimal switch payload.
- [x] Extend `port_table` with stable defaults.
- [x] Add read-only `bridge-observe` mode for Linux bridge/sysfs data.
- [x] Keep `observe` as a migration alias for `bridge-observe`.
- [x] Add read-only `port-map` mode for explicit interface/disabled/unmapped port sources.
- [x] Fill `mac_table` from Linux bridge FDB when configured.
- [x] Read port counters from sysfs when configured.
- [x] Optional `/proc/net/dev` counter source through `proc_source: procfs`.
- [x] Model several virtual ports for `vmbr0`, `tap*`, `veth*`.
- [x] Classify bridge members as bridge/uplink/access before mapping.
- [x] Filter remote MACs learned on the physical uplink out of local access
  port MAC tables.
- [x] Report unused bridge-observe ports as disconnected instead of synthetic-up.
- [x] Support explicit SFP/SFP+ uplink placement through `uplink_port` for
  bridge-observe profiles whose physical link is not the GE fallback.
- [ ] Add explicit `uplink_neighbor.remote_port` metadata and feed it into the
  switch payload when known.
- [ ] Add automatic uplink-neighbor derivation from passive LLDP, with manual
  `uplink_neighbor` as a deterministic override.

## Phase 5: Operations

- [x] OpenRC service.
- [x] systemd unit.
- [x] YAML configuration wired into the daemon.
- [x] Package builders for Debian, RPM, Arch Linux, and tgz.
- [x] Stub-only FreeBSD/OPNsense tgz and rc.d artifact.
- [ ] Rotating debug log.
- [x] Healthcheck/status command.
- [x] Platform capability status for LLDP, logs, procfs, D-Bus, and traffic sources.
- [x] README and operation-mode docs for bridge-observe/port-map lab use.
- [x] Docker controller integration smoke test.

## Phase 6: Later Research

- [x] Built-in SSH adoption for `syswrapper.sh set-adopt` and `mca-cli-op set-inform`.
- [ ] Active macvlan/ipvlan lifecycle after the dry-run plan is proven.
- [x] Passive LLDP import from `lldpd`.
- [ ] LLDP VLAN/MED details, CDP/FDP, and event subscriptions.
- [ ] Document and test topology direction when a stub uses a physical host MAC
  that is also visible to a real upstream UniFi switch.
- [ ] Add a supported deployment pattern for synthetic stub MACs on Proxmox
  bridge representations.
- [x] Experimental gateway identity profiles `UGW3`, `UXG`, and `UXGPRO`.
- [ ] Full gateway status payload for `UGW3`/`UXG`.
- [ ] Synthesize DPI fields from NetFlow/OPNsense/ntopng.
- [ ] Move large firmware research labs to a companion repo or clearly separate research package.
- [x] Compatibility matrix per UniFi Network version.
- [x] JSON Schema for YAML configuration.

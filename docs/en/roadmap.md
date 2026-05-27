# Roadmap

## Phase 0: Pin the Lab

- [x] Validate a controller-known 10G profile: `USAGGPRO` online/adopted.
- [x] Document operation modes and the current live lab state.
- [x] Start a UniFi Network compatibility matrix.
- [x] Run the controller in an isolated lab.
- [ ] Capture a real UniFi switch inform sequence if hardware is available.
- [ ] Capture real UXG/Cloud Gateway inform baselines when hardware access is
  available.
- [ ] Enable debug logs for `inform`, `discover`, `devmgr`.
- [ ] Add an anonymized capture review checklist for payload diffs before any
  fixture is committed.

## Phase 1: Discovery

- [x] Discovery TLV builder.
- [x] Broadcast/multicast sender.
- [x] Controller validation: Docker lab confirms a fake device appears as Pending Adoption.
- [x] Explicit `discovery_targets` for routed or FreeBSD/OPNsense labs.
- [x] Optional `discovery_interface` source binding.
- [ ] Document TLV diff against a real switch.
- [ ] Validate discovery through routed management networks without all-ones
  broadcast.
- [ ] Decide whether STUN metadata is needed for any supported stub profile.

## Phase 2: Inform Without Adoption

- [x] `TNBU` header encoder/decoder.
- [x] AES-CBC + zlib foundation.
- [x] AES-GCM foundation.
- [x] Minimal inform client with default key.
- [x] Validate default-key inform path against the Docker controller lab.
- [x] Decode and log controller responses.
- [x] Record safe metadata for ignored provisioning responses.
- [ ] Add fixture coverage for `include_blocks` responses such as `gw_caps`,
  `dns_shield_servers`, active leases, and WAN status blocks.
- [ ] Add retry/backoff and timeout policy tests for controller inform failures.
- [ ] Add bounded status history for recent inform response types without
  leaking authkeys or controller secrets.

## Phase 3: Adoption

- [x] Parse `_type: setparam`.
- [x] Split `mgmt_cfg` into key/value data.
- [x] Persist `authkey`, `cfgversion`, `use_aes_gcm`, `inform_url`.
- [x] Verify post-adoption connected state in the Docker controller lab.
- [x] Reach connected state with `USAGGPRO`.
- [x] Treat forget/delete/remove/restore-default responses as local stub reset.
- [x] Built-in SSH adoption compatibility for constrained `set-adopt` and
  `set-inform` command shapes.
- [ ] Add adoption failure diagnostics that distinguish bad key, stale model,
  stale MAC, unreachable controller, and controller-side rejection.
- [ ] Add migration tests for adoption state when moving from `/tmp` runs to
  package-managed services.

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
- [x] Enrich learned MAC entries with local ARP IPv4 metadata when available.
- [x] Support configured `port_neighbors` with deterministic hostname and IP
  metadata.
- [x] Preserve disconnected unused ports instead of controller-visible fake
  traffic on empty ports.
- [ ] Add explicit `uplink_neighbor.remote_port` metadata and feed it into the
  switch payload when known.
- [ ] Add automatic uplink-neighbor derivation from passive LLDP, with manual
  `uplink_neighbor` as a deterministic override.
- [ ] Add LLDP/CDP/FDP neighbor normalization for local interface, chassis,
  remote port, VLAN, and system name fields.
- [ ] Add topology regression tests for physical-MAC collision versus synthetic
  locally administered stub MACs.
- [ ] Add profile coverage for additional common UniFi switch models used in
  homelabs.

## Phase 5: Operations

- [x] OpenRC service.
- [x] systemd unit.
- [x] YAML configuration wired into the daemon.
- [x] Package builders for Debian, RPM, Arch Linux, and tgz.
- [x] Stub-only FreeBSD/OPNsense tgz and rc.d artifact.
- [x] GitHub pre-release with neutral package artifacts and checksums.
- [x] GitHub Pages alpha repositories for APT, RPM, Arch Linux, and
  FreeBSD/OPNsense tarballs.
- [x] External private host-configuration layout for real lab configs and
  install protocols.
- [ ] Rotating debug log.
- [x] Healthcheck/status command.
- [x] Platform capability status for LLDP, logs, procfs, D-Bus, and traffic sources.
- [x] README and operation-mode docs for bridge-observe/port-map lab use.
- [x] Docker controller integration smoke test.
- [ ] Project release GPG key and signed APT/RPM/Arch package metadata.
- [ ] FreeBSD `pkg` package and repository once real FreeBSD package artifacts
  replace the current tarball-only path.
- [ ] Separate `alpha`, `testing`, and `stable` package channels.
- [ ] Package-manager install smoke tests in CI for APT, RPM, and Pacman repos.
- [ ] Pages repository freshness check in CI for APT `Packages.gz`, RPM
  `repomd.xml`, Arch `unifi-stubd.db`, FreeBSD tarballs, and checksums.
- [ ] Publish release SBOM and provenance artifacts alongside checksums.
- [ ] Add package upgrade, downgrade, uninstall, and purge tests with explicit
  adoption-state retention behavior.
- [ ] Add package rollback runbook that maps package-managed services back to a
  captured temporary command.
- [ ] GitHub Actions update for the Node 24 runner transition.
- [ ] Encrypted backup or rotation policy for the external private host-config
  store.
- [ ] Live controller post-install gate: no duplicate devices and no
  `ADOPT_FAILED` loop after package rollout.
- [ ] Operator runbook for collecting safe `status-json`, service logs, and
  package versions without leaking private controller data.

## Phase 6: Later Research

- [x] Built-in SSH adoption for `syswrapper.sh set-adopt` and `mca-cli-op set-inform`.
- [ ] Active macvlan/ipvlan lifecycle after the dry-run plan is proven.
- [ ] Review a narrow local-only adapter for `planned-host-vlan`; keep it
  separate from controller provisioning.
- [x] Passive LLDP import from `lldpd`.
- [ ] LLDP VLAN/MED details, CDP/FDP, and event subscriptions.
- [ ] Document and test topology direction when a stub uses a physical host MAC
  that is also visible to a real upstream UniFi switch.
- [ ] Add a supported deployment pattern for synthetic stub MACs on Proxmox
  bridge representations.
- [x] Experimental gateway identity profiles `UGW3`, `UXG-Lite`, `UXGPRO`, and
  `UCGF`.
- [x] Gateway WAN/LAN port reporting for UXG-shaped lab stubs.
- [x] Config-gated traffic-rate reporting from read-only interface counters.
- [ ] Full gateway status payload for `UGW3`/`UXG`.
- [ ] Gateway UI validation for Settings > Internet, Devices, and Ports views
  across pinned UniFi Network versions.
- [x] Config-driven WAN uptime, latency, downtime, and connected-state hints.
- [ ] WAN peak utilization and ISP label where the controller accepts read-only
  values.
- [ ] Active lease and host-table shaping for gateway client visibility without
  inventing per-client traffic counters.
- [ ] UniFi UI verification for switch and gateway traffic activity after
  package-based deployment.
- [ ] Broader WAN activity and link-state compatibility tests across UniFi
  Network versions.
- [ ] Implement `traffic_source` adapters only after the counter-based path is
  stable; candidates are NetFlow/IPFIX, OPNsense telemetry, and ntopng.
- [ ] Synthesize DPI fields from NetFlow/OPNsense/ntopng.
- [ ] FreeBSD `SIOCGIFMEDIA` interface media and speed reader.
- [ ] FreeBSD interface counter parity for `port-map` and future
  `bridge-observe`.
- [ ] Full FreeBSD `bridge-observe` parity once native counter/media readers are
  available.
- [ ] Move large firmware research labs to a companion repo or clearly separate research package.
- [ ] UXG-Lite and UCG-Fiber ARM64 firmware wrapper debugging with `strace` or
  a broader LD_PRELOAD shim.
- [ ] UDM Pro SE deterministic netdev and switch-driver mock surface for the
  firmware reference lab.
- [ ] UGW3 legacy board identity and EEPROM mock layer for the firmware runner.
- [ ] GPL/source-bundle handling process before any external firmware-derived
  source or structured data is copied into project-owned code.
- [x] Compatibility matrix per UniFi Network version.
- [x] JSON Schema for YAML configuration.
- [ ] Generate config reference docs from the JSON Schema to keep README,
  packaged examples, and schema aligned.

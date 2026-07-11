# Compatibility Matrix

This matrix tracks controller-facing lab validation. It is intentionally about
UniFi Network behavior, not firmware source inventory.

| UniFi Network | Lab target | Result | Notes |
| --- | --- | --- | --- |
| Docker lab `lscr.io/linuxserver/unifi-network-application:10.4.57-ls135`, validated 2026-07-11 | `US8` bridge-observe and `UXG`/`uxg-lite` gateway-smoke | Adopted and stably connected | `make integration-docker` enforces `server_version=10.4.57` and verifies discovery, adoption, persisted auth, AES-GCM, and stable connected samples. Unlike 10.3.58, this controller issued no gateway command; the gate proved zero handled, delivered, duplicate, or pending commands before and after the 75-second stability window while retaining a healthy controller-driven inform schedule. |
| Docker lab `lscr.io/linuxserver/unifi-network-application:10.3.58-ls129`, validated 2026-07-10 | `US8` bridge-observe and `UXG`/`uxg-lite` gateway-smoke | Adopted and stably connected | `make integration-docker` bootstraps empty controller volumes, enforces `server_version=10.3.58`, captures forwarded discovery, verifies Pending Adoption and controller-triggered adoption, decodes sanitized default-key `setparam` evidence, checks persisted connected/auth/cfgversion/GCM state, records at least two AES-GCM informs, and requires three consecutive connected controller samples. The gateway gate also proves one handled `wg-genkeypair` with a WireGuard-clamped X25519 key, one delivered result, no duplicate or pending result, and no task recurrence after 75 seconds. It verifies that the controller's positive interval hint becomes the locally bounded effective and next delay with no active backoff. |
| Private real UniFi OS Server controller, captured 2026-05-20 | Linux/Proxmox bridge-observe with 48-port switch profile | Adopted in lab | Validated real controller adoption, AES-GCM post-adoption heartbeats, access-port clients from bridge FDB, disconnected unused ports, SFP+ uplink placement with `uplink_port`, and `uplink_neighbor` topology metadata. Controller URL, site data, device MACs, and token are intentionally not recorded in Git. Physical host MAC tests showed UniFi Network can reverse topology direction when a real upstream switch already reports that MAC. |
| LinuxServer.io `unifi-network-application:latest`, captured 2026-05-17 | UXG-Pro firmware `5.0.16.30689` and host-side `unifi-stubd` diagnostics | Adopted in lab | Exact application version was not pinned in the Compose file; pin it before treating this as release-grade compatibility data. |
| Future pinned version | Switch profiles `US8`, `US16P150`, `US16XG`, `USAGGPRO`, `USW-Pro-XG-48` | Not yet recorded | Add one row per tested UniFi Network version. |

Compatibility entries must include the controller/application version, profile
or firmware identity, adoption result, cipher mode, and any controller response
types that were ignored by safety policy.

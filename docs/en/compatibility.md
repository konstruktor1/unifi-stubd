# Compatibility Matrix

This matrix tracks controller-facing lab validation. It is intentionally about
UniFi Network behavior, not firmware source inventory.

| UniFi Network | Lab target | Result | Notes |
| --- | --- | --- | --- |
| Docker lab `lscr.io/linuxserver/unifi-network-application:10.3.58-ls129`, captured 2026-05-20 | `US8` bridge-observe and `UXG`/`uxg-lite` gateway-smoke | Adopted in lab | `make integration-docker` validates dry-run gateway tables, MITM inform events, controller pending state, controller-triggered adoption, persisted local `STATE=connected`, and controller `/status` reporting `server_version=10.3.58`. |
| Private real UniFi OS Server controller, captured 2026-05-20 | Linux/Proxmox bridge-observe with 48-port switch profile | Adopted in lab | Validated real controller adoption, AES-GCM post-adoption heartbeats, access-port clients from bridge FDB, disconnected unused ports, SFP+ uplink placement with `uplink_port`, and `uplink_neighbor` topology metadata. Controller URL, site data, device MACs, and token are intentionally not recorded in Git. Physical host MAC tests showed UniFi Network can reverse topology direction when a real upstream switch already reports that MAC. |
| LinuxServer.io `unifi-network-application:latest`, captured 2026-05-17 | UXG-Pro firmware `5.0.16.30689` and host-side `unifi-stubd` diagnostics | Adopted in lab | Exact application version was not pinned in the Compose file; pin it before treating this as release-grade compatibility data. |
| Future pinned version | Switch profiles `US8`, `US16P150`, `US16XG`, `USAGGPRO`, `USW-Pro-XG-48` | Not yet recorded | Add one row per tested UniFi Network version. |

Compatibility entries must include the controller/application version, profile
or firmware identity, adoption result, cipher mode, and any controller response
types that were ignored by safety policy.

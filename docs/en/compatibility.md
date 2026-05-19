# Compatibility Matrix

This matrix tracks controller-facing lab validation. It is intentionally about
UniFi Network behavior, not firmware source inventory.

| UniFi Network | Lab target | Result | Notes |
| --- | --- | --- | --- |
| LinuxServer.io `unifi-network-application:latest`, captured 2026-05-17 | UXG-Pro firmware `5.0.16.30689` and host-side `unifi-stubd` diagnostics | Adopted in lab | Exact application version was not pinned in the Compose file; pin it before treating this as release-grade compatibility data. |
| Future pinned version | Switch profiles `US8`, `US16P150`, `US16XG`, `USAGGPRO`, `USW-Pro-XG-48` | Not yet recorded | Add one row per tested UniFi Network version. |

Compatibility entries must include the controller/application version, profile
or firmware identity, adoption result, cipher mode, and any controller response
types that were ignored by safety policy.

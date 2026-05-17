# Simulation Fixtures

This folder contains committed, sanitized protocol fixtures from local firmware
simulation runs.

Raw MITM captures stay under `../captures/` and are ignored by Git. They can
contain adoption material, controller state, keys, private controller URLs, or
device data once the lab reaches adoption. Fixtures in this folder intentionally
omit raw HTTP bodies and body hashes.

- `inform-telegrams.jsonl`: sanitized HTTP/TNBU envelope summaries.
- `decoded-gateway-inform-sample.json`: sanitized decoded firmware inform
  payload showing the gateway `if_table` and `network_table` fields used for
  WAN/LAN interface reporting in the current lab.
- `adoption-mitm-timeline.json`: sanitized adoption sequence from portal
  `Adopt` through adopted steady state. It records packet sizes, decoded
  message types, state changes, and config versions, but omits raw bodies,
  auth keys, tokens, certificates, password hashes, and `system_cfg` details.
- `adopted-system-config-summary.json`: sanitized structural summary of the
  large adopted `system_cfg` response. It records UDAPI categories, service
  names, interface IDs, rule counts, and safe implementation implications, but
  omits the raw `system_cfg` values.

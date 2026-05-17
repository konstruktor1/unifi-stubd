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

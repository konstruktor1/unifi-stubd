# Simulation Fixtures

This folder contains committed, sanitized protocol fixtures from local firmware
simulation runs.

Raw MITM captures stay under `../captures/` and are ignored by Git. They can
contain adoption material, controller state, keys, private controller URLs, or
device data once the lab reaches adoption. Fixtures in this folder intentionally
omit raw HTTP bodies and body hashes.


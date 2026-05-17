# Firmware Simulation Helpers

This folder contains project-owned helper source for running parts of the
UXG-Pro firmware in an isolated, networkless container.

See `docker-howto.md` for the full container setup.
See `controller-lab.md` for a local UniFi Network Application lab that can run
against the firmware container.
See `inform-mitm-analysis.md` for sanitized findings from a local MITM run.

`ubnthal_redirect.c` is an LD_PRELOAD shim. It lets selected firmware binaries
read mock hardware data from `/mock` instead of host kernel paths and no-ops a
small set of system-management calls that are not meaningful in the container.

Build example inside an ARM64 Debian/Bullseye environment:

```sh
gcc -shared -fPIC -Wall -Wextra -O2 -ldl \
  -o /mock/libubnthal_redirect.so \
  /mock/ubnthal_redirect.c
```

Runtime example:

```sh
env LD_PRELOAD=/mock/libubnthal_redirect.so \
  /usr/bin/mcad -n -s -v
```

Keep the container networkless for analysis unless a lab controller endpoint is
explicitly configured.

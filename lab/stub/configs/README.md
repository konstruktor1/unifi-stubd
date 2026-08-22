# Stub Lab Configuration

The disposable controller gate uses one runtime configuration:
`hosts/stub-bridge-observe/config.yaml`. The gate supplies a fresh MAC and IP
for every run and mounts this directory read-only.

`hosts/opnsense-api-source.example.yaml` is an independent example for the
OPNsense generator; it is not part of the controller gate.

Controller credentials, adoption keys, captures, private URLs, real addresses,
MAC tables, client data, real source files, and generated host configurations
must not be stored anywhere in this repository, including ignored paths. Keep
them in private deployment storage outside the checkout; fixtures here must use
documentation addresses and synthetic identities only.

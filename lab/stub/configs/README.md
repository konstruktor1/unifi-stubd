# Stub Lab Configuration

The disposable product gate uses one runtime configuration:
`hosts/stub-bridge-observe/config.yaml`. The gate supplies a fresh MAC and IP
for every run and mounts this directory read-only.

`hosts/opnsense-api-source.example.yaml` is an independent example for the
OPNsense generator; it is not part of the controller gate.

Controller credentials, adoption keys, captures, private URLs, real addresses,
MAC tables, and client data must not be stored here. Real-network snapshots
belong only in the ignored `hosts/<hostname>/real/` or `temp/` paths.

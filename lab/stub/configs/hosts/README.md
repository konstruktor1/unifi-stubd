# Host Configuration Layout

The tracked controller-gate input is
`stub-bridge-observe/config.yaml`. `opnsense-api-source.example.yaml` is a
separate read-only generator example.

Real or temporary host configurations belong in ignored
`<hostname>/real/config.yaml` or `<hostname>/temp/config.yaml` paths. Do not
commit private controller URLs, credentials, adoption state, real addresses,
interface names, MAC tables, or client data.

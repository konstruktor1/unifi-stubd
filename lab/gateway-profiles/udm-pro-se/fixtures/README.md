# UDM Pro SE Fixtures

Fixtures here are reduced reference outputs from the UDM Pro SE lab path. They
exist so future changes can compare behavior without committing raw firmware
logs or private controller data.

`mca-dump-summary.json` is a sanitized summary of `mca-ctrl -t dump` from the
mocked Docker firmware path. It should answer "what shape did the local
management agent expose?" without preserving volatile process output,
controller state, or full device dumps.

Add fixtures only when they are deterministic, small, and useful for regression
review. Do not store support bundles, raw captures, real serials, tokens,
private controller URLs, or host-specific addresses in this directory.

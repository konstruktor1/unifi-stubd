# Product Roadmap

Last updated: 2026-07-29.

The roadmap describes the Stub product. Labs, runners, packages, and tests
support delivery; they are not product phases.

## 1. US8 with visible real metrics — current

Deliver a non-UniFi Linux bridge that an operator can add to UniFi Network as
an `US8`.

Product outcome:

- it is discovered and can be adopted;
- it remains connected;
- it reports real link state, speed, counters, and rates from the named local
  source;
- controller input cannot provision the host.

The current [compatibility evidence](compatibility.md) verifies discovery,
adoption, connection, and the real metric data path on dynamically resolved
Stable and Release Candidate. Focused tests cover the host safety boundary, but
a retained result has not yet bound them to the exact qualified source. Direct
evidence that an operator sees those values on both controller device pages is
also missing, so this product phase is not accepted yet.

## 2. Usable Linux release — next

Turn this behavior into a package and operator path that work on a
supported Linux host.

Exit:

- the package contains the data-path-qualified product binary;
- the supported package configuration selects US8 bridge observation and a
  real local metric source;
- installation and packaged service startup pass on a supported Linux host;
- the quick start leads an operator from local metric source to adoption and
  controller readback;
- limitations state that the supported slice is US8 on a Linux bridge.

The binary-bound package set exists. Its starter config now explicitly selects
`US8`, but it still uses non-observing `stub` mode without a named real metric
source and is therefore not the supported acceptance configuration.
Host-specific bridge configuration, installed-service, and operator-path
acceptance remain open.

## 3. Further representations — only on demand

Add another profile or platform only when it has a concrete operator outcome,
a named local read-only source, an explicit safety decision, and its own
Stable/RC acceptance. Evidence from a neighboring model does not expand the
product automatically.

## Immediate sequence

1. Complete the supported US8 bridge configuration with a named real local
   metric source; the explicit packaged US8 identity alone is not acceptance
   evidence.
2. Install the run-bound package on a supported Linux host and start it with
   that configuration.
3. On current Stable and Release Candidate, adopt it and directly confirm that
   an operator sees the real link, speed, and traffic values.
4. Retain the focused host-safety result for that exact candidate source.
5. Make that exact installation, adoption, metric-source, limitation, and
   recovery path the operator quick start before publishing the package.

New gateway telemetry, WAN-health semantics, and additional profiles remain
later options; they do not precede the usable release path.

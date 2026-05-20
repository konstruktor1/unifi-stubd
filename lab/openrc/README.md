# OpenRC Lab Service Files

This directory holds the dedicated OpenRC variant of the observe-bridge lab
fixture. Use it when the bridge should be managed as its own service, with
normal `rc-service` start/stop behavior and configuration in `/etc/conf.d`.

The service calls `lab/observe-bridge.sh` rather than duplicating bridge logic.
That keeps the manual, `local.d`, and OpenRC paths aligned: if the bridge
topology changes, update the helper first and keep this service as orchestration
only.

These files are separate from the installable `unifi-stubd` service units under
`packaging/linux/`. They are lab fixtures for observation and packet-flow
experiments, not the daemon packaging path.

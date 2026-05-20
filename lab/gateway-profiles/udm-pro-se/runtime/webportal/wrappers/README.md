# Webportal Wrappers

These executables are installed ahead of selected firmware commands in the
Docker webportal path. They let UniFi Core complete known setup and support
flows without handing the container arbitrary host control.

The wrappers are intentionally command-specific:

- service wrappers report or start only known lab services;
- support wrappers package deterministic logs and metadata;
- identity wrappers return stable synthetic device information;
- UDAPI wrappers translate Docker `eth0` into the lab's UDM-style WAN view.

When adding behavior, keep it allowlisted and logged. A wrapper may accept a
known UniFi Core command shape, but it must not execute arbitrary controller
shell, reboot/update the host, or mutate host networking.

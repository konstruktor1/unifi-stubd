# Webportal Templates

Templates here are rendered during Docker webportal startup before UniFi Core
and its helper commands are launched. They define lab-local Core settings and
the narrow sudo surface needed for support-bundle generation.

`unifi-core-default.yaml.in` is configuration, not a captured runtime file. It
sets database and feature overrides that make the setup UI inspectable in the
container. `unifi-core-lab.sudoers` allows only the approved transient commands
the lab wrappers can safely handle.

Do not broaden these templates into arbitrary shell, reboot, update, or host
network privileges.

# Webportal HTTP Assets

These files adapt generated UniFi Core/nginx configuration for a Docker lab
where the browser talks to localhost and destructive setup actions must stay
blocked.

`shared-runnable-lab.conf` adds the small lab-only sidecar routes needed by
Core. The AWK filters preserve the generated configuration while changing only
the pieces the lab must control: local preview reachability and setup API
guards for reset/reboot-style actions.

Treat these as runtime patches, not captured vendor configuration. Keep each
filter narrow enough that a future diff can show exactly what behavior changed.

# Local.d Observe Bridge Hooks

This directory is for Alpine/OpenRC hosts where the observe bridge should be
created as part of the generic `local` service instead of through a dedicated
init script. The hooks are intentionally tiny: they delegate all real work to
`lab/observe-bridge.sh`, so bridge behavior has one implementation.

The start hook creates the lab-only `stubbr0` bridge and its veth members. The
stop hook tears that bridge back down. This is useful for repeatable
observe-mode tests where a controller or packet capture needs to see a stable
bridge shape across reboots.

Install these only on disposable lab hosts. They alter local network device
state and should not be treated as package-managed production service files.

# Host Configuration Layout

Each host has its own directory below `hosts/`.

Tracked Docker lab defaults use:

```text
hosts/<hostname>/config.yaml
```

Real-network or temporary host snapshots must stay local and ignored by Git:

```text
hosts/<real-hostname>/real/config.yaml
hosts/<real-hostname>/temp/config.yaml
```

Use `real/` for the current real-network service config and `temp/` for short
lived experiments copied from a host. These files may contain private
controller URLs, real lab addresses, MACs, interface names, or adoption paths,
so they are ignored by `hosts/.gitignore`.

If a real host config should become shareable, copy only a sanitized version to
a tracked file such as `hosts/<hostname>/config.example.yaml`, replace real
addresses with documentation ranges like `192.0.2.0/24`, and remove controller
tokens, adoption keys, private URLs, and real client data.

Tracked `config.example.yaml` files may describe real lab topology patterns, but
must use example addresses and locally administered MAC addresses. The current
SFP+ examples capture two reusable patterns:

- `server-lan1-sfp-lab`: Proxmox/Linux bridge switch with OPNsense on SFP+ port
  49, aggregation on SFP+ port 50, and a TAP member ignored because that virtual
  side is already represented by the physical SFP+ uplink.
- `opnsense-uxg-sfp-lab`: UXG Pro-shaped gateway with WAN on SFP+ port 3 and
  LAN to `server-lan1` on SFP+ port 4.
- `opnsense-api-source.example.yaml`: separate read-only source file for the
  `unifi-stubd-opnsense` companion generator. It contains API endpoint and port
  mapping examples, but no real API credentials. See
  `docs/en/opnsense-generator.md` and
  `docs/en/opnsense-generator-reference.md` for the standalone generator docs.

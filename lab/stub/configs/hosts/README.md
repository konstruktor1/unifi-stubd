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

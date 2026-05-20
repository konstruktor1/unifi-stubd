# Research

Status: 2026-05-16

## Core Thesis

UniFi does not show real UniFi devices in the `Devices` tab because of LLDP alone. Devices appear because they speak Ubiquiti discovery and inform protocols. To display a Proxmox bridge, OPNsense/pfSense VM, or another non-UniFi system, `unifi-stubd` therefore needs a minimal fake UniFi device lifecycle:

1. Send UDP discovery.
2. Talk `/inform` to the controller.
3. Accept adoption/authkey.
4. Send periodic status payloads.
5. Ignore provisioning commands or acknowledge them as successful without changing host configuration.

## Key References

The detailed attribution matrix is maintained in
[CREDITS.md](../../CREDITS.md). The entries below are the sources that shaped
the protocol and product direction.

- [Unofficial UniFi Guide: Discovery](https://jrjparks.github.io/unofficial-unifi-guide/protocols/discovery.html)
- [Unofficial UniFi Guide: Inform](https://jrjparks.github.io/unofficial-unifi-guide/protocols/inform.html)
- [Unofficial UniFi Guide: Adoption](https://jrjparks.github.io/unofficial-unifi-guide/adoption.html)
- [jeffreykog/unifi-inform-protocol](https://github.com/jeffreykog/unifi-inform-protocol)
- [fxkr/unifi-protocol-reverse-engineering](https://github.com/fxkr/unifi-protocol-reverse-engineering)
- [Tamarack: Reverse Engineering the UniFi Inform Protocol](https://tamarack.cloud/blog/reverse-engineering-unifi-inform-protocol)
- [Ubiquiti: UniFi Required Ports Reference](https://help.ui.com/hc/en-us/articles/218506997-UniFi-Network-Required-Ports-Reference)
- [Ubiquiti: Remote Adoption / Layer 3](https://help.ui.com/hc/en-us/articles/204909754-Remote-Adoption-Layer-3)
- [Ubiquiti: UniFi Security Gateway Datasheet](https://dl.ui.com/datasheets/unifi/UniFi_Security_Gateway_DS.pdf)
- [Ubiquiti: UniFi Security Gateway Quick Start Guide](https://dl.ui.com/qsg/USG/USG_EN.html)
- [Ubiquiti: UXG-Pro Tech Specs](https://techspecs.ui.com/unifi/advanced-hosting/uxg-pro?s=me)

## Attribution Boundary

Research repositories are credited as sources of protocol facts, historical
context, and design ideas. The implementation in `unifi-stubd` is independent
Go code. Do not copy code from research repositories unless its license has
been reviewed and the attribution files are updated.

## Older Projects

### wvengen/unifi-controllable-switch

[wvengen/unifi-controllable-switch](https://github.com/wvengen/unifi-controllable-switch) is the closest ancestor for this project. It patched a TOUGHswitch so it appeared in the UniFi Controller as a switch and could be adopted. Particularly relevant:

- `devel/unifi_announce.py`: discovery TLVs.
- `devel/unifi_inform.py`: old inform packet with `TNBU`, AES-CBC, and payload fields.
- `src/syswrapper.sh`: adoption via `set-adopt <inform_url> <authkey>`.
- `src/unifi-inform-status`: examples for `if_table`, `port_table`, `sys_stats`.

The project failed long-term because of controller version drift and firmware patching complexity, not because of the core idea.

### stephanlascar/unifi-gateway

[stephanlascar/unifi-gateway](https://github.com/stephanlascar/unifi-gateway) was a pfSense/USG emulator PoC. It is not production-ready, but contains useful gateway payload ideas:

- `dpi-clients`
- `dpi-stats`
- `dpi-stats-table`
- `config_port_table`
- WAN/LAN configuration

Gateway emulation is a later research goal, not the MVP.

### jda/pixiedust

[jda/pixiedust](https://github.com/jda/pixiedust) analyzes inform traffic and uses PCAPs. Useful for:

- Authkey extraction.
- Observing `setparam`.
- Comparing fresh UniFi switch payloads.

### ZAP-Quebec/unifi-inform

[ZAP-Quebec/unifi-inform](https://github.com/ZAP-Quebec/unifi-inform) is an older Go implementation of the inform protocol. Useful for header structure, flags, and message model.

## Official Boundaries

UniFi documents third-party devices only in a limited way. Third-party gateways can exist in VLAN/routing scenarios, but Traffic Identification/DPI is strongly tied to UniFi gateways. For this project that means:

- Switch-like visibility is realistic.
- Port traffic and MAC tables are realistic.
- Full DPI without a UniFi gateway is probably not realistic.
- Simulated DPI values would be a separate reverse-engineering project.

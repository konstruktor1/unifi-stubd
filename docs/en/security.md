# Security Notes

`unifi-stubd` is intended for an isolated lab.

See the repository-level [security policy](../../SECURITY.md) before reporting
or publishing security-sensitive details.
Private reports can be sent to `info@spinas.org`.

## Authkeys

The UniFi `authkey` is a symmetric key for inform payloads. It must not land in logs, screenshots, or Git history.

Do not share `adoption.env`, SSH host keys, or controller API tokens.

## No Host Provisioning

The controller must not blindly mutate host configuration. For the MVP:

- Persist `setparam`.
- Acknowledge `noop`.
- Record only sanitized metadata for gateway `system_cfg`, such as byte length
  and top-level keys.
- Mark restart/upgrade/provisioning commands as ignored by policy.
- Do not execute shell commands from the controller.

## Network Boundaries

Discovery and inform belong only in the lab or management network. The project should not run on production VLANs with unrelated controllers.

Packaged configs keep the adoption SSH shim closed by default
(`ssh_listen: ""`). The normal adoption path is inform-based through
`controller_url`. Enable `ssh_listen` only in an isolated lab when the
controller must use advanced adoption or `set-inform` over SSH; the factory
credentials are then exposed on the configured listen address.

## Personal and Client Data

MAC tables, DHCP information, DPI data, and NetFlow can contain personal metadata. Example PCAPs belong in `.gitignore` and should be anonymized before sharing.

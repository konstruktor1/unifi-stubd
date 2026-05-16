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
- Only log restart/upgrade/provisioning commands.
- Do not execute shell commands from the controller.

## Network Boundaries

Discovery and inform belong only in the lab or management network. The project should not run on production VLANs with unrelated controllers.

## Personal and Client Data

MAC tables, DHCP information, DPI data, and NetFlow can contain personal metadata. Example PCAPs belong in `.gitignore` and should be anonymized before sharing.

# UTM Installer Modules

`install-utm-profile.sh` sources these files to mutate a cloned UTM bundle in
small, reviewable steps. The split matters because UTM profile changes are easy
to get wrong: boot inputs, disks, network devices, serial settings, and QEMU
arguments each have different failure modes.

Responsibilities:

- `common.sh`: resolves paths, loads defaults, validates tools, and refuses to
  proceed when required local artifacts are missing.
- `plist.sh`: wraps PlistBuddy so the other modules do not open-code plist
  mutation.
- `boot.sh`: places the foreign kernel, lab initramfs, and generated DTB where
  UTM expects boot inputs.
- `drives.sh`: converts the VM disk and attaches it to the cloned bundle.
- `network.sh`: writes the two-NIC UDM mapping, including SFP+ WAN on Shared/NAT
  and 2.5G LAN on Host networking. It records the localhost HTTPS forward as an
  intent; runtime tests still verify whether UTM actually binds the port.
- `system.sh`: enforces the serial-only 4 GiB VM shape and removes display,
  sharing, sound, and USB extras that are not useful for this lab.

Generated UTM bundle files and local VM state belong in ignored artifacts or in
the UTM bundle, never in this source directory.

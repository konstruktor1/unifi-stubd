# Mock Files

This directory is the committed, deterministic part of the `/mock` filesystem.
Docker copies it into the simulation directory, and the VM preparation scripts
stage it into the mock root used by the lab initramfs.

Only stable identity inputs belong here. Volatile state such as MTD bytes,
sysctl values, hwmon temperatures, and generated persistence files is created
by runtime scripts under ignored local directories.

The current committed subtree is `ubnthal/`, which mimics the small part of
`/proc/ubnthal` that UDM firmware reads during setup and board identification.
Keep the values synthetic and predictable so Docker and QEMU/UTM tests describe
the same lab device.

# Research

This directory stores reproducible research notes and project-owned helper
source used while studying UniFi device behavior.

Do not commit vendor firmware images, extracted root filesystems, private lab
data, PCAPs, adoption keys, controller URLs, SSH host keys, MAC tables, or
client data here. Store large or proprietary inputs under an ignored
`artifacts/` directory and document their checksums instead.

`firmware/profiles.yaml` tracks the real-firmware simulation catalog. The
catalog distinguishes synthetic `unifi-stubd` profiles from real firmware
profiles that need a local vendor image, extracted rootfs, architecture notes,
and a model-specific process wrapper.

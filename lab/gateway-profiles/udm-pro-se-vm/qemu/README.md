# QEMU Network Presets

These opt-in environment snippets support direct `run-foreign-kernel.sh`
experiments. They are intentionally separate from the UTM profile because
direct QEMU is used for narrow kernel/initramfs iteration, while UTM is the
persisted full-VM reference used for browser and two-NIC validation.

`transparent-lan.env` keeps the web path on a LAN-facing guest interface, close
to the current VM model. `user-lan-forward.env` preserves the older QEMU
user-mode forwarding shape for quick localhost checks when vmnet networking is
not available.

Source one of these files only for a single local run. Do not move UTM defaults
here; UTM-owned network shape lives under `../utm/`.

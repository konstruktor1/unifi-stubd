# UDM Pro SE VM Scripts

These scripts define the reproducible UDM Pro SE VM research pipeline. They
turn a locally supplied firmware image into ignored boot artifacts, build the
lab initramfs, prepare userspace mocks, and run either direct QEMU or UTM-backed
VM tests.

The important boundary is that committed scripts are project-owned, while
firmware images, extracted root filesystems, generated disks, fetched kernels,
toolchains, and logs remain under ignored `../artifacts/`.

Normal VM preparation is ordered as:

```sh
./lab/gateway-profiles/udm-pro-se-vm/scripts/prepare-vm.sh
./lab/gateway-profiles/udm-pro-se-vm/scripts/fetch-foreign-kernel.sh
./lab/gateway-profiles/udm-pro-se-vm/scripts/prepare-mocks.sh
./lab/gateway-profiles/udm-pro-se-vm/scripts/build-lab-initramfs.sh
./lab/gateway-profiles/udm-pro-se-vm/scripts/deploy-kernel-artifacts.sh
```

Direct QEMU scripts are for focused boot experiments. `run-direct-kernel.sh`
and `run-vendor-uboot.sh` document the native vendor boundary: vendor kernel
and vendor U-Boot do not produce useful serial output on QEMU `virt`.
`run-foreign-kernel.sh` is the working mixed path that keeps the UDM
initramfs/rootfs handoff while using a QEMU-virt-capable kernel.

`install-utm-profile.sh` is the bridge from generated artifacts into a cloned
UTM VM. Keep UTM plist mutation in `../utm/install/` modules so it stays
auditable.

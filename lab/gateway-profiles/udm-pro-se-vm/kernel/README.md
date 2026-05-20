# Kernel Deployment Artifacts

The project does not commit firmware kernels, initramfs images, extracted
module trees, or generated UTM device trees. Those files are local research
inputs and generated outputs.

Use `scripts/deploy-kernel-artifacts.sh` to stage the current local kernel
payload into the ignored directory `artifacts/deploy/kernel/`.

The staged layout is shared by the UTM and Docker lab paths:

```text
artifacts/deploy/kernel/
  MANIFEST.txt
  vendor/
    kernel.Image
    kernel.fit
    udm-pro-se.dtb
    initramfs.cpio.gz
  foreign/
    debian-arm64-linux
    debian-arm64-initrd.gz
    modules/
  lab/
    lab-initramfs.cpio.gz
```

UTM uses the foreign kernel and lab initramfs as boot inputs. Docker mounts the
same directory read-only at `/opt/unifi-fw-sim/kernel` so the container can
inspect the exact kernel payload used by the VM reference. Docker still runs on
the host kernel; the mounted payload is a lab input, not a container boot
kernel.

`MANIFEST.txt` records the staged file list and SHA-256 values. The Docker
startup scripts copy the first part of that manifest into their log directory,
and the UTM installer copies the staged foreign kernel, lab initramfs, and
generated DTB into the cloned UTM bundle.

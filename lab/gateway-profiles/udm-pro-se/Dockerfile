# syntax=docker/dockerfile:1.6
# UDM Pro SE firmware simulation wrapper around an imported ARM64 rootfs image.

ARG BASE_IMAGE=udm-pro-se-fw:5.0.16
FROM ${BASE_IMAGE}

LABEL org.opencontainers.image.title="UDM Pro SE firmware simulation wrapper"
LABEL org.opencontainers.image.description="Local research wrapper around an imported UDM Pro SE firmware rootfs"

# Keep image mutations small and explicit. The imported firmware rootfs remains
# the base; this layer only removes dpkg cache hooks that break local installs
# and adds strace for optional lab diagnostics.
RUN mkdir -p /tmp \
    && mkdir -p /opt/unifi-fw-sim/kernel \
    && chmod 1777 /tmp \
    && rm -f /etc/dpkg/dpkg.cfg.d/015-ubnt-dpkg-status /etc/dpkg/dpkg.cfg.d/020-ubnt-dpkg-cache \
    && apt-get update \
    && apt-get install -y --no-install-recommends strace \
    && rm -rf /var/lib/apt/lists/*

# Project-owned helpers are copied as source trees so Docker and the QEMU/UTM
# reference can reuse the same mock logic without vendoring firmware content.
COPY mock/ldpreload/ /opt/unifi-fw-sim/ldpreload/
COPY runtime/ /usr/local/lib/udm-pro-se-runtime/
COPY --chmod=0755 start-firmware-processes.sh /usr/local/bin/udm-pro-se-sim-start
COPY --chmod=0755 start-webportal-processes.sh /usr/local/bin/udm-pro-se-webportal-start
COPY systemd-dbus/ /usr/local/lib/udm-pro-se-systemd-dbus/
COPY network-app/ /usr/local/lib/udm-pro-se-network-app/
COPY udapi-lab-shim.cjs /usr/local/lib/udm-pro-se-udapi-lab-shim.cjs

CMD ["/bin/sleep", "infinity"]

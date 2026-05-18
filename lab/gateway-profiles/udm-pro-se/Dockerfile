# syntax=docker/dockerfile:1.6
# UDM Pro SE firmware simulation wrapper around an imported ARM64 rootfs image.

ARG BASE_IMAGE=udm-pro-se-fw:5.0.16
FROM ${BASE_IMAGE}

LABEL org.opencontainers.image.title="UDM Pro SE firmware simulation wrapper"
LABEL org.opencontainers.image.description="Local research wrapper around an imported UDM Pro SE firmware rootfs"

RUN mkdir -p /tmp \
    && chmod 1777 /tmp \
    && rm -f /etc/dpkg/dpkg.cfg.d/015-ubnt-dpkg-status /etc/dpkg/dpkg.cfg.d/020-ubnt-dpkg-cache \
    && apt-get update \
    && apt-get install -y --no-install-recommends strace \
    && rm -rf /var/lib/apt/lists/*

COPY ubnthal_redirect.c /opt/unifi-fw-sim/ubnthal_redirect.c
COPY --chmod=0755 start-firmware-processes.sh /usr/local/bin/udm-pro-se-sim-start
COPY --chmod=0755 start-webportal-processes.sh /usr/local/bin/udm-pro-se-webportal-start
COPY systemd-dbus-stub.cjs /usr/local/lib/udm-pro-se-systemd-dbus-stub.cjs
COPY network-app-stub.cjs /usr/local/lib/udm-pro-se-network-app-stub.cjs
COPY udapi-lab-shim.cjs /usr/local/lib/udm-pro-se-udapi-lab-shim.cjs

CMD ["/bin/sleep", "infinity"]

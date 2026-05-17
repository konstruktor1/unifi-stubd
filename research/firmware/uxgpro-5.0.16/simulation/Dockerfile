# syntax=docker/dockerfile:1.6

ARG BASE_IMAGE=uxgpro-fw:5.0.16
FROM ${BASE_IMAGE}

LABEL org.opencontainers.image.title="UXG-Pro firmware simulation wrapper"
LABEL org.opencontainers.image.description="Local research wrapper around an imported UXG-Pro firmware rootfs"

COPY ubnthal_redirect.c /opt/unifi-fw-sim/ubnthal_redirect.c
COPY --chmod=0755 start-firmware-processes.sh /usr/local/bin/uxgpro-sim-start

CMD ["/bin/sleep", "infinity"]

# syntax=docker/dockerfile:1.6

ARG BASE_IMAGE=uxg-lite-fw:5.0.16
FROM ${BASE_IMAGE}

LABEL org.opencontainers.image.title="UXG-Lite firmware simulation wrapper"
LABEL org.opencontainers.image.description="Local research wrapper around an imported UXG-Lite firmware rootfs"

COPY research/firmware/uxgpro-5.0.16/simulation/ubnthal_redirect.c /opt/unifi-fw-sim/ubnthal_redirect.c
COPY --chmod=0755 research/firmware/uxg-lite-5.0.16/simulation/start-firmware-processes.sh /usr/local/bin/uxg-lite-sim-start

CMD ["/bin/sleep", "infinity"]

# syntax=docker/dockerfile:1.6
# UGW3 firmware simulation image for running the extracted MIPS rootfs under QEMU.

FROM debian:bookworm-slim

LABEL org.opencontainers.image.title="UGW3 firmware QEMU simulation runner"
LABEL org.opencontainers.image.description="Local research runner for an extracted UGW3 MIPS rootfs"

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
        qemu-user-static \
    && rm -rf /var/lib/apt/lists/*

COPY --chmod=0755 start-ugw3-qemu.sh /usr/local/bin/ugw3-qemu-start

CMD ["/usr/local/bin/ugw3-qemu-start"]

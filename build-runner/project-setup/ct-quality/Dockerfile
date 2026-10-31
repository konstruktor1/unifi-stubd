# syntax=docker/dockerfile:1

FROM docker.io/library/golang:1.27.3-bookworm@sha256:d9b6a1beb9474adc2a1ce842573639b44b2a0dd49b6669f51d3821baf9ec37c9 AS go-tools
RUN CGO_ENABLED=0 GOBIN=/tool-bin go install mvdan.cc/sh/v3/cmd/shfmt@v3.14.1

FROM docker.io/library/node:24.22.2-bookworm-slim@sha256:c86ecead9e55a5210ac1727de38a4ca840bf03dcaf34dc9749addbf383b1920c AS node-tools
FROM docker.io/koalaman/shellcheck:v0.11.0@sha256:61862eba1fcf09a484ebcc6feea46f1782532571a34ed51fedf90dd25f925a8d AS shellcheck

FROM docker.io/library/python:3.15.2-slim-bookworm@sha256:4e03260adeeebafccd084f3584e6b84823aba5651cfb53ef8a078ed441ddf9cd

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        bash binutils-gold ca-certificates gcc git libc6-dev libatomic1 procps util-linux \
    && rm -rf /var/lib/apt/lists/*

COPY --from=go-tools /usr/local/go /usr/local/go
COPY --from=go-tools /tool-bin/shfmt /usr/local/bin/shfmt
COPY --from=node-tools /usr/local/bin/node /usr/local/bin/node
COPY --from=node-tools /usr/local/lib/node_modules/npm /usr/local/lib/node_modules/npm
COPY --from=shellcheck /bin/shellcheck /usr/local/bin/shellcheck

RUN ln -s /usr/local/go/bin/go /usr/local/bin/go \
    && ln -s /usr/local/go/bin/gofmt /usr/local/bin/gofmt \
    && ln -s ../lib/node_modules/npm/bin/npm-cli.js /usr/local/bin/npm \
    && ln -s ../lib/node_modules/npm/bin/npx-cli.js /usr/local/bin/npx \
    && python3 -m pip install --no-cache-dir \
        attrs==26.1.0 jsonschema==4.26.0 jsonschema-specifications==2025.9.1 \
        referencing==0.37.0 rpds-py==2026.9.1 ruff==0.17.3 \
    && python3 -m pip check

# These are BR's declared container namespace and contained CT runtime roles.
ENV PATH=/usr/local/go/bin:${PATH} \
    PYTHONDONTWRITEBYTECODE=1 \
    GOTOOLCHAIN=local \
    HOME=/workspace/.provider-provider/runtime/unifi-stubd/ct-quality/home \
    GOPATH=/workspace/.provider-provider/runtime/unifi-stubd/ct-quality/go \
    GOMODCACHE=/workspace/.provider-provider/runtime/unifi-stubd/ct-quality/go/pkg/mod \
    GOCACHE=/workspace/.provider-provider/runtime/unifi-stubd/ct-quality/go-build \
    NPM_CONFIG_CACHE=/workspace/.provider-provider/runtime/unifi-stubd/ct-quality/npm \
    PIP_CACHE_DIR=/workspace/.provider-provider/runtime/unifi-stubd/ct-quality/pip \
    XDG_CACHE_HOME=/workspace/.provider-provider/runtime/unifi-stubd/ct-quality/cache \
    TMPDIR=/workspace/.provider-provider/runtime/unifi-stubd/ct-quality/tmp \
    TEMP=/workspace/.provider-provider/runtime/unifi-stubd/ct-quality/tmp \
    TMP=/workspace/.provider-provider/runtime/unifi-stubd/ct-quality/tmp

WORKDIR /workspace
CMD ["bash"]

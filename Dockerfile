# syntax=docker/dockerfile:1.7

FROM --platform=$BUILDPLATFORM golang:1.25-bookworm AS build

ARG TARGETOS=linux
ARG TARGETARCH
ARG VERSION=dev

WORKDIR /src

COPY go.mod go.sum go.work go.work.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w -X main.version=${VERSION}" \
    -o /out/unifi-stubd ./cmd/unifi-stubd

FROM alpine:3.22

RUN apk add --no-cache ca-certificates iproute2 tzdata \
    && mkdir -p /etc/unifi-stubd /var/lib/unifi-stubd

COPY --from=build /out/unifi-stubd /usr/local/bin/unifi-stubd

VOLUME ["/var/lib/unifi-stubd"]

ENTRYPOINT ["/usr/local/bin/unifi-stubd"]

# syntax=docker/dockerfile:1

ARG GO_VERSION="1.22"
ARG ALPINE_VERSION="3.18"

# --------------------------------------------------------
# Builder
# --------------------------------------------------------

FROM golang:${GO_VERSION}-bookworm AS builder

# Always set by buildkit
ARG TARGETPLATFORM
ARG TARGETARCH
ARG TARGETOS

# needed in makefile
ARG COMMIT
ARG VERSION
ARG CGO_ENABLED=1
ARG LINK_STATICALLY="false"

# Consume Args to env
ENV \
  COMMIT=${COMMIT} \
  VERSION=${VERSION} \
  GOOS=${TARGETOS} \
  GOARCH=${TARGETARCH} \
  CGO_ENABLED=${CGO_ENABLED} \
  LINK_STATICALLY=${LINK_STATICALLY}

# Install dependencies
RUN set -eux; \
  apt-get update && \
  apt-get install -y --no-install-recommends \
  build-essential \
  ca-certificates \
  binutils-gold \
  git && \
  rm -rf /var/lib/apt/lists/*

# Set the workdir
WORKDIR /go/src/github.com/burnt-labs/xion

# Download go dependencies
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/root/.cache/go-build \
  --mount=type=cache,target=/root/pkg/mod \
  set -eux; \
  go install cosmossdk.io/tools/cosmovisor/cmd/cosmovisor@v1.5.0; \
  go mod download

# Copy local files
COPY . .

# Build xiond binary
RUN --mount=type=cache,target=/root/.cache/go-build \
  --mount=type=cache,target=/root/pkg/mod \
  set -eux; \
  make test-version; \
  make install;

RUN set -eux; \
  mkdir -p /go/lib; \
  cp -L /lib*/ld-linux-*.so.* /go/lib; \
  ldd /go/bin/xiond | \
  awk '{print $1}' | \
  xargs -I {} find / -name {} -not -path "/go/lib/*" 2>/dev/null | \
  xargs -I {} cp -L {} /go/lib/;

# --------------------------------------------------------
# Heighliner
# --------------------------------------------------------

# Build final image from scratch
FROM scratch AS heighliner

WORKDIR /bin
ENV PATH=/bin

# Install busybox
COPY --from=busybox:1.36-musl /bin/busybox /bin/busybox

# users and group
COPY --from=busybox:1.36-musl /etc/passwd /etc/group /etc/

# Install trusted CA certificates
COPY --from=alpine:3 /etc/ssl/cert.pem /etc/ssl/cert.pem

# Install xiond
COPY --from=builder /go/lib/* /lib/
COPY --from=builder /go/bin/xiond /bin/xiond

# Install jq
COPY --from=ghcr.io/strangelove-ventures/infra-toolkit:v0.1.4 /usr/local/bin/jq /bin/jq

# link shell
RUN ["busybox", "ln", "/bin/busybox", "sh"]

# Add hard links for read-only utils
# Will then only have one copy of the busybox minimal binary file with all utils pointing to the same underlying inode
RUN set -eux; \
  for bin in \
  cat \
  date \
  df \
  du \
  env \
  grep \
  head \
  less \
  ls \
  md5sum \
  pwd \
  sha1sum \
  sha256sum \
  sha3sum \
  sha512sum \
  sleep \
  stty \
  tail \
  tar \
  tee \
  tr \
  watch \
  which \
  ; do busybox ln /bin/busybox $bin; \
  done;

RUN set -eux; \
  busybox ln -s /lib /lib64; \
  busybox mkdir -p /tmp /home/heighliner; \
  busybox addgroup --gid 1025 -S heighliner; \
  busybox adduser --uid 1025 -h /home/heighliner -S heighliner -G heighliner; \
  busybox chown 1025:1025 /tmp /home/heighliner; \
  busybox unlink busybox;

WORKDIR /home/heighliner
USER heighliner

# --------------------------------------------------------
# Runner
# --------------------------------------------------------

FROM alpine:${ALPINE_VERSION} AS release
COPY --from=builder /go/lib/* /lib/
COPY --from=builder /go/bin/xiond /usr/bin/xiond
COPY --from=builder /go/bin/cosmovisor /usr/bin/cosmovisor

# api
EXPOSE 1317
# grpc
EXPOSE 9090
# p2p
EXPOSE 26656
# rpc
EXPOSE 26657
# prometheus
EXPOSE 26660

RUN set -euxo pipefail; \
  ln -s /lib /lib64; \
  apk add --no-cache bash openssl curl htop jq lz4 tini; \
  addgroup --gid 1000 -S xiond; \
  adduser --uid 1000 -S xiond \
  --disabled-password \
  --gecos xiond \
  --ingroup xiond; \
  mkdir -p /home/xiond; \
  chown -R xiond:xiond /home/xiond

USER xiond:xiond
WORKDIR /home/xiond/.xiond
CMD ["/usr/bin/xiond"]

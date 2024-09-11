# syntax=docker/dockerfile:1

ARG GO_VERSION="1.22.6"
ARG ALPINE_VERSION="3.20"

# --------------------------------------------------------
# Builder
# --------------------------------------------------------

FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder

# Always set by buildkit
ARG TARGETPLATFORM
ARG TARGETARCH
ARG TARGETOS

# needed in makefile
ARG COMMIT
ARG VERSION
ARG CGO_ENABLED=1
ARG BUILD_TAGS=muslc
ARG LEDGER_ENABLED="true"
ARG LINK_STATICALLY="true"

# Consume Args to env
ENV COMMIT=${COMMIT} \
	VERSION=${VERSION} \
	GOOS=${TARGETOS} \
	GOARCH=${TARGETARCH} \
	BUILD_TAGS=${BUILD_TAGS} \
	CGO_ENABLED=${CGO_ENABLED} \
	LEDGER_ENABLED=${LEDGER_ENABLED} \
	LINK_STATICALLY=${LINK_STATICALLY}

# Install dependencies
RUN set -eux; \
	apk add --no-cache \
	build-base \
	ca-certificates \
	linux-headers \
	binutils-gold \
	gcc \
	git

# Install rust
ENV RUSTUP_HOME=/usr/local/rustup \
	CARGO_HOME=/usr/local/cargo \
	PATH=/usr/local/cargo/bin:$PATH \
	RUST_VERSION=1.81.0

RUN set -eux; \
	apkArch="$(apk --print-arch)"; \
	case "$apkArch" in \
	x86_64) rustArch='x86_64-unknown-linux-musl'; rustupSha256='1455d1df3825c5f24ba06d9dd1c7052908272a2cae9aa749ea49d67acbe22b47' ;; \
	aarch64) rustArch='aarch64-unknown-linux-musl'; rustupSha256='7087ada906cd27a00c8e0323401a46804a03a742bd07811da6dead016617cc64' ;; \
	*) echo >&2 "unsupported architecture: $apkArch"; exit 1 ;; \
	esac; \
	url="https://static.rust-lang.org/rustup/archive/1.27.1/${rustArch}/rustup-init"; \
	wget "$url"; \
	echo "${rustupSha256} *rustup-init" | sha256sum -c -; \
	chmod +x rustup-init; \
	./rustup-init -y --no-modify-path --profile minimal --default-toolchain $RUST_VERSION --default-host ${rustArch}; \
	rm rustup-init; \
	chmod -R a+w $RUSTUP_HOME $CARGO_HOME; \
	rustup --version; \
	cargo --version; \
	rustc --version;

# Set the workdir (xiond)
WORKDIR /go/src/github.com/burnt-labs/xion

COPY go.mod go.sum ./

# Set the workdir (wasmvm)
RUN set -eux; \
	WASMVM_REPO="github.com/CosmWasm/wasmvm"; \
	WASMVM_MOD_VERSION="$(grep ${WASMVM_REPO} /go/src/github.com/burnt-labs/xion/go.mod | cut -d ' ' -f 1)"; \
	WASMVM_VERSION="$(go list -m ${WASMVM_MOD_VERSION} | cut -d ' ' -f 2)"; \
	[ ${TARGETPLATFORM} = "linux/amd64" ] && ARCH="x86_64"; \
	[ ${TARGETPLATFORM} = "linux/arm64" ] && ARCH="aarch64"; \
	[ -z "$ARCH" ] && echo "Arch ${TARGETARCH} not recognized" && exit 1; \
	mkdir -p /go/src/github.com/CosmWasm/wasmvm; \
	cd /go/src/github.com/CosmWasm/wasmvm; \
	git clone --depth 1 --branch ${WASMVM_VERSION} https://${WASMVM_REPO}.git . ; \
	cd libwasmvm; \
	cargo build --release --target-dir=target --target ${ARCH}-unknown-linux-musl --example wasmvmstatic; \
	cp "target/${ARCH}-unknown-linux-musl/release/examples/libwasmvmstatic.a" /lib/libwasmvm_muslc.${ARCH}.a

# Set the workdir (xiond)
WORKDIR /go/src/github.com/burnt-labs/xion

# Download go dependencies
RUN --mount=type=cache,target=/root/.cache/go-build \
	--mount=type=cache,target=/root/pkg/mod \
	set -eux; \
	go mod download

# Copy local files
COPY . .

# Build xiond binary
RUN --mount=type=cache,target=/root/.cache/go-build \
	--mount=type=cache,target=/root/pkg/mod \
	set -eux; \
	make test-version; \
	make install;

# Install cosmovisor
RUN --mount=type=cache,target=/root/.cache/go-build \
	--mount=type=cache,target=/root/pkg/mod \
	set -eux; \
	go install cosmossdk.io/tools/cosmovisor/cmd/cosmovisor@v1.5.0;

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
COPY --from=builder /etc/ssl/cert.pem /etc/ssl/cert.pem

# Install xiond
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
	apk add --no-cache bash openssl curl htop jq lz4 tini; \
	addgroup --gid 1000 -S xiond; \
	adduser --uid 1000 -S xiond \
	--gecos xiond \
	--ingroup xiond \
	--disabled-password; \
	mkdir -p /home/xiond; \
	chown -R xiond:xiond /home/xiond

USER xiond:xiond
WORKDIR /home/xiond/.xiond
CMD ["/usr/bin/xiond"]

FROM golang:1.19.7-buster as builder
MAINTAINER BitQuery

RUN apt-get update && apt-get install -y ca-certificates build-essential clang ocl-icd-opencl-dev ocl-icd-libopencl1 jq libhwloc-dev

ENV RUSTUP_HOME=/usr/local/rustup \
    CARGO_HOME=/usr/local/cargo \
    PATH=/usr/local/cargo/bin:$PATH \
    RUST_VERSION=1.63.0

RUN set -eux; \
    dpkgArch="$(dpkg --print-architecture)"; \
    case "${dpkgArch##*-}" in \
        amd64) rustArch='x86_64-unknown-linux-gnu'; rustupSha256='5cc9ffd1026e82e7fb2eec2121ad71f4b0f044e88bca39207b3f6b769aaa799c' ;; \
        arm64) rustArch='aarch64-unknown-linux-gnu'; rustupSha256='e189948e396d47254103a49c987e7fb0e5dd8e34b200aa4481ecc4b8e41fb929' ;; \
        *) echo >&2 "unsupported architecture: ${dpkgArch}"; exit 1 ;; \
    esac; \
    url="https://static.rust-lang.org/rustup/archive/1.25.1/${rustArch}/rustup-init"; \
    wget "$url"; \
    echo "${rustupSha256} *rustup-init" | sha256sum -c -; \
    chmod +x rustup-init; \
    ./rustup-init -y --no-modify-path --profile minimal --default-toolchain $RUST_VERSION --default-host ${rustArch}; \
    rm rustup-init; \
    chmod -R a+w $RUSTUP_HOME $CARGO_HOME; \
    rustup --version; \
    cargo --version; \
    rustc --version;

COPY ./ /opt/filecoin
WORKDIR /opt/filecoin

### make configurable filecoin-ffi build
ARG FFI_BUILD_FROM_SOURCE=0
ENV FFI_BUILD_FROM_SOURCE=${FFI_BUILD_FROM_SOURCE}

RUN make clean deps

ARG RUSTFLAGS=""
ARG GOFLAGS=""

RUN make buildall



FROM ubuntu:20.04 AS base
MAINTAINER BitQuery

COPY --from=builder /etc/ssl/certs            /etc/ssl/certs
COPY --from=builder /lib/*/libdl.so.2 \
   /lib/*/librt.so.1 \
   /lib/*/libgcc_s.so.1 \
   /lib/*/libutil.so.1 \
   /usr/lib/*/libltdl.so.7 \
   /usr/lib/*/libnuma.so.1 \
   /usr/lib/*/libhwloc.so.5 \
   /usr/lib/*/libOpenCL.so.1 \
   /lib/

RUN useradd -r -u 532 -U fc \
 && mkdir -p /etc/OpenCL/vendors \
 && echo "libnvidia-opencl.so.1" > /etc/OpenCL/vendors/nvidia.icd



FROM base AS lotus
MAINTAINER BitQuery

COPY --from=builder /opt/filecoin/lotus /usr/local/bin/
COPY --from=builder /opt/filecoin/lotus-shed /usr/local/bin/
COPY scripts/docker-lotus-entrypoint.sh /

ARG DOCKER_LOTUS_IMPORT_SNAPSHOT https://snapshots.mainnet.filops.net/minimal/latest
ENV DOCKER_LOTUS_IMPORT_SNAPSHOT ${DOCKER_LOTUS_IMPORT_SNAPSHOT}
ENV FILECOIN_PARAMETER_CACHE /var/tmp/filecoin-proof-parameters
ENV LOTUS_PATH /var/lib/lotus
ENV DOCKER_LOTUS_IMPORT_WALLET ""

RUN mkdir /var/lib/lotus /var/tmp/filecoin-proof-parameters
RUN chown fc: /var/lib/lotus /var/tmp/filecoin-proof-parameters

VOLUME /var/lib/lotus
VOLUME /var/tmp/filecoin-proof-parameters

USER fc

EXPOSE 1234

ENTRYPOINT ["/docker-lotus-entrypoint.sh"]

CMD ["-help"]


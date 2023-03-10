FROM golang:1.18.8-buster as builder

MAINTAINER BitQuery

ARG RELEASE_VERSION
ENV RELEASE_VERSION=${RELEASE_VERSION:-v1.20.3rpc}

COPY . .

RUN apt-get update \
    && apt-get install -y ca-certificates build-essential clang ocl-icd-opencl-dev ocl-icd-libopencl1 jq libhwloc-dev \
    && make clean \
    && make all \
    && make install


FROM debian:bullseye-slim as runner

COPY ./run.sh /root/
COPY --from=builder /usr/local/bin/lotus /usr/local/bin/lotus
COPY --from=builder /etc/ssl/certs            /etc/ssl/certs
COPY --from=builder /lib/*/libdl.so.2         /lib/
COPY --from=builder /lib/*/librt.so.1         /lib/
COPY --from=builder /lib/*/libgcc_s.so.1      /lib/
COPY --from=builder /lib/*/libutil.so.1       /lib/
COPY --from=builder /usr/lib/*/libltdl.so.7   /lib/
COPY --from=builder /usr/lib/*/libnuma.so.1   /lib/
COPY --from=builder /usr/lib/*/libhwloc.so.5  /lib/
COPY --from=builder /usr/lib/*/libOpenCL.so.1 /lib/

RUN apt update \
    && apt -y install iputils-ping curl \
    && rm -rf /var/lib/apt/lists/* \
    && chmod +x /root/run.sh

EXPOSE 1234 25416

ENTRYPOINT ["/root/run.sh"]

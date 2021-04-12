#
# Dockerfile for the netspot server
#
# The first two parts are the builder which compiles the source
# into the 'netspot' binary. The last part is merely a 
# lightweight image with netspot as entrypoint.
#

# ========================================================================== #
# SYSTEM PREPARE FOR BUILD
# ========================================================================== #
FROM golang:1.16.3-alpine AS base

# utilities
ENV PACKAGES "build-base nano bash linux-headers git flex bison wget make bluez-dev bluez"

# SYSTEM
RUN apk update; apk add $PACKAGES


# ========================================================================== #
# LIBPCAP COMPILATION
# ========================================================================== #
FROM base as libpcap

# env
ARG LIBPCAP_VERSION=1.10.0
ENV LIBPCAP_VERSION ${LIBPCAP_VERSION}
ENV LIBPCAP_DIR /libpcap

# download
RUN mkdir -p ${LIBPCAP_DIR}
ADD https://www.tcpdump.org/release/libpcap-${LIBPCAP_VERSION}.tar.gz ${LIBPCAP_DIR}
RUN cd ${LIBPCAP_DIR}; tar -xvf libpcap-${LIBPCAP_VERSION}.tar.gz

# build
RUN cd ${LIBPCAP_DIR}/libpcap-${LIBPCAP_VERSION}; ./configure; make



# ========================================================================== #
# NETSPOT COMPILATION
# ========================================================================== #
FROM libpcap as builder

# git information
ARG GIT_COMMIT
ENV GIT_COMMIT ${GIT_COMMIT}

# Go stuff
ENV GOPATH /go
ENV CGO_ENABLED 1
ENV CGO_LDFLAGS "-L${LIBPCAP_DIR}/libpcap-${LIBPCAP_VERSION}"
ENV CGO_CFLAGS "-O2 -I${LIBPCAP_DIR}/libpcap-${LIBPCAP_VERSION}"

# prepare go path
RUN mkdir -p ${GOPATH}/src/netspot

# copy source code
COPY analyzer   ${GOPATH}/src/netspot/analyzer
COPY api/       ${GOPATH}/src/netspot/api
COPY cmd/       ${GOPATH}/src/netspot/cmd
COPY config/    ${GOPATH}/src/netspot/config
COPY exporter/  ${GOPATH}/src/netspot/exporter
COPY miner/     ${GOPATH}/src/netspot/miner
COPY stats/     ${GOPATH}/src/netspot/stats
COPY Makefile go.mod go.sum netspot.go ${GOPATH}/src/netspot/

# build
RUN cd $GOPATH/src/netspot; make build_netspot_static

# ========================================================================== #
# NETSPOT IMAGE
# ========================================================================== #
FROM alpine:latest

ENV NETSPOT /usr/bin/netspot
COPY --from=builder /go/src/netspot/bin/* ${NETSPOT}
RUN apk add libcap
RUN adduser -HD -s /dev/null netspot; chown netspot:netspot ${NETSPOT}
RUN setcap cap_net_admin,cap_net_raw=eip ${NETSPOT}
USER netspot
CMD [${NETSPOT}, "serve", "-e", "tcp://127.0.0.1:11000"]

#
# DockerFile for the NetSpot server
#
# The first part is the builder which compiles the source
# into the 'netspot' binary. The second part is merely a 
# lightweight image with netspot as entrypoint.
#
FROM golang:1.16.3-alpine AS base

# utilities
ENV PACKAGES "build-base nano bash linux-headers git flex bison wget make bluez-dev bluez"

# SYSTEM
RUN apk update; apk add $PACKAGES


# 
# 
# 
FROM base as libpcap

# env
ENV LIBPCAP_VERSION 1.10.0
ENV LIBPCAP_DIR /libpcap

# download
RUN mkdir -p ${LIBPCAP_DIR}
ADD https://www.tcpdump.org/release/libpcap-${LIBPCAP_VERSION}.tar.gz ${LIBPCAP_DIR}
RUN cd ${LIBPCAP_DIR}; tar -xvf libpcap-${LIBPCAP_VERSION}.tar.gz

# build
RUN cd ${LIBPCAP_DIR}/libpcap-${LIBPCAP_VERSION}; ./configure; make

# 
# 
# 
FROM libpcap as builder

# SYSTEM
ENV GOPATH /go
ENV CGO_ENABLED 1
ENV CGO_LDFLAGS "-L${LIBPCAP_DIR}/libpcap-${LIBPCAP_VERSION}"
ENV CGO_CFLAGS "-O2 -I${LIBPCAP_DIR}/libpcap-${LIBPCAP_VERSION}"

RUN mkdir -p "${GOPATH}/src/netspot"
COPY . $GOPATH/src/netspot
RUN cd $GOPATH/src/netspot; make build_netspot_static

# 
# 
# 
FROM alpine:latest

ENV NETSPOT /usr/bin/netspot
COPY --from=builder /go/src/netspot/bin/* ${NETSPOT}
RUN apk add libcap
RUN adduser -HD -s /dev/null netspot; chown netspot:netspot ${NETSPOT}
RUN setcap cap_net_admin,cap_net_raw=eip ${NETSPOT}
CMD [${NETSPOT}, "serve", "-e", "tcp://127.0.0.1:11000"]

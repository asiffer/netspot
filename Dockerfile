#
# DockerFile for the NetSpot server
#
# The first part is the builder which compiles the source
# into the 'netspot' binary. The second part is merely a 
# lightweight image with netspot as entrypoint.
#
FROM golang:1.12.7-alpine3.10 AS builder
# set the destination architecture and OS
ARG GOARCH
ARG GOOS
ENV ARCH=$GOARCH
ENV OS=$GOOS
# the working directory of the
# docker container
WORKDIR /go/src/netspot
# move source code
COPY analyzer/*.go ./analyzer/
COPY api/*.go ./api/
COPY influxdb/*.go ./influxdb/
COPY miner/*.go miner/
COPY miner/counters/*.go miner/counters/
COPY stats/*.go stats/
COPY netspot.go ./
# move makefile
COPY Makefile ./
# some extra library are needed to build
# build-base: make
# libpcap-dev: gopacket
# git: go update
RUN apk add build-base libpcap-dev git
# get netspot dependencies
RUN make netspot_deps
# build according to OS and ARCH
RUN make build_netspot ARCH=$ARCH OS=$OS

FROM alpine:latest
RUN apk add libpcap
WORKDIR /usr/bin/
COPY --from=builder /go/src/netspot/bin/* .
RUN mkdir -p /etc/netspot
COPY extra/*.toml /etc/netspot/
CMD ["/usr/bin/netspot"]

# server port for HTTP REST API (and golang RPC)
EXPOSE 11000 11001
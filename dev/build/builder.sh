#!/bin/bash
#
# This script must be called in a
# 'docker exec' command (see run.sh)

# set some build environment variables
# arguments:
# $1: target-arch (arm-linux, x86_64-linux, aarch64-linux)
prepare_env() {
    case $1 in
    arm-linux)
        prefix=arm-linux-musleabihf
        goarch=arm
        ;;
    x86_64-linux)
        prefix=x86_64-linux-musl
        goarch=amd64
        ;;
    aarch64-linux)
        prefix=aarch64-linux-musl
        goarch=arm64
        ;;
    *)
        echo -n "Unknown arch $1"
        ;;
    esac

    export CC=/bin/$prefix-gcc
    export AR=/bin/$prefix-ar
    export LD=/bin/$prefix-ld
    export GOARCH=$goarch
}

# function to build libpcap
# arguments:
# $1: target-arch (arm-linux, x86_64-linux, aarch64-linux)
build_libpcap() {
    cd $LIBPCAP_DIR/libpcap-$LIBPCAP_VERSION
    # remove previous builds
    if [ -f "Makefile" ]; then
        make clean
    fi
    ./configure --host=x86_64-linux --target=$1
    make -B CC=$CC AR=$AR LD=$LD CFLAGS='-I/usr/include'
}

# build netspot
# arguments
# $1 target-arch (arm-linux, x86_64-linux, aarch64-linux)
build_netspot() {
    cd $GOPATH/src/netspot
    export LD_LIBRARY_PATH=$LIBPCAP_DIR/libpcap-$LIBPCAP_VERSION
    CC=$CC AR=$AR LD=$LD go build -o bin/netspot-$1 -a -ldflags '-X "main.Version=v2.0" -s -w -extldflags "-static"' netspot.go
}

#
# Run
# $1 is the targeted arch
prepare_env $1
build_libpcap $1
build_netspot $1

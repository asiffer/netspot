#!/bin/bash
# This script install musl-based
# cross-compilation toolchains
#
# Prerequisite
# ============
#
# Packages
# --------
# bash
# wget
# tar
#
# Environment variables
# ---------------------
# GCC_VERSION
# MUSL_CROSS_MAKE_VERSION
# TARGET_ARCH
#

BASE_URL=https://github.com/just-containers/musl-cross-make/releases/download/$MUSL_CROSS_MAKE_VERSION

cd /tmp
for ARCH in $TARGET_ARCH; do
    FULL_ARCH=$ARCH-musl
    if [ "$ARCH" == "arm-linux" ]; then
        FULL_ARCH=$ARCH-musleabihf
    fi

    # get the toolchain
    wget "${BASE_URL}/gcc-$GCC_VERSION-$FULL_ARCH.tar.xz"
    # decompress and install it in the root
    tar -C / -xvf "gcc-$GCC_VERSION-$FULL_ARCH-musl.tar.xz"
done

FROM golang:1.15.3-alpine

# LIBPCAP
ENV LIBPCAP_VERSION 1.9.1
ENV LIBPCAP_DIR /libpcap

# SYSTEM
ENV PACKAGES "nano bash linux-headers git flex bison wget make bluez-dev bluez"
ENV CGO_ENABLED 1
ENV CGO_LDFLAGS "-L${LIBPCAP_DIR}/libpcap-${LIBPCAP_VERSION}"
ENV CGO_CFLAGS "-O2 -I${LIBPCAP_DIR}/libpcap-${LIBPCAP_VERSION}"

# CROSS COMPILATION OPTIONS
# see https://github.com/just-containers/musl-cross-make/releases/
ENV GCC_VERSION 9.2.0
ENV MUSL_CROSS_MAKE_VERSION v15
ENV TARGET_ARCH "x86_64-linux-musl arm-linux-musleabihf aarch64-linux-musl"

# SYSTEM
RUN apk update; apk add $PACKAGES

# CROSS COMPILE TOOLCHAIN
COPY install_toolchain.sh /install_toolchain.sh
RUN /install_toolchain.sh

# LIBPCAP
COPY get_libpcap_sources.sh /get_libpcap_sources.sh
RUN /get_libpcap_sources.sh

CMD ["/bin/bash"]
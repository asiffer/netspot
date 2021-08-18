---
title: Developer toolchain
weight: -9
summary: How to build netspot
---

`netspot` is distributed as statically-compiled binaries. The final executables
notably embed the `libpcap` and `musl` libraries.
Using a musl-based system (like [alpine](https://alpinelinux.org/))
is then far easier to compile `netspot`.
Obviously, you can dynamically link `netspot` to `libpcap` and the
more common GNU libc, but the installation will require these
dependencies on the target system. Here, we only detail the static
build for different architectures.

The [dev/](https://github.com/asiffer/netspot/tree/master/dev) includes
some utilities to build `netspot` statically.

### Docker image

First of all, you have to prepare the docker image to build netspot.
A Dockerfile is provided in the [dev/image/](https://github.com/asiffer/netspot/tree/master/dev/image)
subfolder.

```docker
FROM golang:1.15.6-alpine

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
ENV TARGET_ARCH "x86_64-linux arm-linux aarch64-linux"

# SYSTEM
RUN apk update; apk add $PACKAGES

# CROSS COMPILE TOOLCHAIN
COPY install_toolchain.sh /install_toolchain.sh
RUN /install_toolchain.sh

# LIBPCAP
COPY get_libpcap_sources.sh /get_libpcap_sources.sh
RUN /get_libpcap_sources.sh

CMD ["/bin/bash"]
```

So you basically need to build this image (from the `dev/image/` folder)

```bash
docker build -t alpine-crossbuild-libpcap:latest .
```

By default, this image will produce binaries for three different architectures:
`amd64`, `arm` and `arm64`, but you can only select those you want
by setting the `TARGET_ARCH` environment variable.

```bash
docker build --build-arg TARGET_ARCH="x86_64-linux" -t alpine-crossbuild-libpcap:latest .
```

<!-- prettier-ignore -->
!!! warning
    Some environment variables must be set before building and not
    when a container starts (see the bash scripts below).

The Dockerfile includes two bash scripts:

- `install_toolchain.sh`, to download the cross-compilers
- `get_libpcap_sources.sh`, to download the sources of `libpcap`

You can inspect these files to check what environment variables
they require.

### Compilation

Now your image is ready, you just have to compile `netspot`.
The [dev/build/](https://github.com/asiffer/netspot/tree/master/dev/build)
folder gathers scripts for this purpose.

The `builder.sh` script is the main file you have to run to compile
`netspot` for a specific architecture

```bash
builder.sh <ARCH>
```

<!-- prettier-ignore -->
!!! warning
    The available `ARCH` depend on the previous image.
    By default you can choose between

    - x86_64-linux
    - arm-linux
    - aarch64-linux

This script must be executed within an instance of the previous image
(container). So first, you have to run a container (here we suppose that `netspot` code
lives in `GOPATH/src/netspot` but it is likely to be different accodint to your
dev workflow):

```bash
docker run --detach -it -v "${GOPATH}/src/netspot:/go/src/netspot" --name netspot-build alpine-crossbuild-libpcap:latest
```

Then you can run the compilation for a specific architecture:

```bash
docker exec netspot-build /go/src/netspot/dev/build/builder.sh <ARCH>
```

<!-- prettier-ignore -->
!!! info
    The final binaries are located in the `bin/` folder.

The second script (`run.sh`) is an example that combine both steps.
You can adapt it to your workflow.

# Development

This folder gtaher some scripts to cross-compile `netspot`.

First the folder `image/` contains a Dockerfile to build
a development environment with
- `libpcap` sources
- `musl`-based cross-compiling toolchains (x86_64, arm, arm64)

Second, the `build/` folder stores some scripts to
build both `libpcap` and
`netspot` (statically linked with `musl`).

## Usage

To create the docker image you just need to build it:
```bash
cd image/
# now you can run 'build.sh' which does the following
sudo docker build -t alpine-crossbuild-libpcap:latest .
```

To perform `libpcap`/`netspot` build for all the architectures
you can invoke the `run.sh` script.
```bash
cd build/
# build for the three architectures: x86_64, arm and aarch64
./run.sh
```
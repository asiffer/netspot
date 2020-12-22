#!/bin/bash
#
#
# Usage
# Build for all the architectures:
#   $ run.sh
# Build for a single architecture:
#   $ run.sh x86_64-linux
#
# Available architectures:
#   - x86_64-linux
#   - arm-linux
#   - aarch64-linux
#

CONTAINER=netspot-build
IMAGE=alpine-crossbuild-libpcap:latest
ARCH="x86_64-linux arm-linux aarch64-linux"

function is_arch_available() {
    for arch in ${ARCH}; do
        if [[ $1 == $arch ]]; then
            # return True
            return 1
        fi
    done
    # return False
    return 0
}

# parse args
if [[ $# -eq 1 ]]; then
    is_arch_available $1
    if [[ $? -eq 1 ]]; then
        ARCH="$1"
    else
        echo -e "\033[101mArchitecture '$1' is not supported\033[0m"
        exit 1
    fi
fi

docker run --detach -it -v "${GOPATH}/src/netspot:/go/src/netspot" --name "${CONTAINER}" "${IMAGE}"
trap "docker rm -f ${CONTAINER}" EXIT

for arch in ${ARCH}; do
    echo -en "\n"
    printf "%${COLUMNS}s\n" | tr " " "="
    echo $arch
    printf "%${COLUMNS}s\n" | tr " " "="
    echo -en "\n"
    docker exec ${CONTAINER} /go/src/netspot/dev/build/builder.sh $arch
    echo -en "\n"
done

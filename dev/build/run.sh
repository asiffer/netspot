#!/bin/bash
#
#

CONTAINER=netspot-build
IMAGE=alpine-crossbuild-libpcap:latest

docker run --detach -it -v "/home/asr/go/src/netspot:/go/src/netspot" --name "${CONTAINER}" "${IMAGE}"
trap "docker rm -f ${CONTAINER}" EXIT

docker cp ./builder.sh ${CONTAINER}:/builder.sh
for arch in x86_64-linux arm-linux aarch64-linux; do
    echo -en "\n"
    printf "%${COLUMNS}s\n" | tr " " "="
    echo $arch
    printf "%${COLUMNS}s\n" | tr " " "="
    echo -en "\n"
    docker exec ${CONTAINER} /builder.sh $arch
    echo -en "\n"
done

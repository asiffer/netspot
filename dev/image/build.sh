#!/bin/bash
#
# This script build the docker image
# used to cross compile netspot
docker build -t alpine-crossbuild-libpcap:latest .

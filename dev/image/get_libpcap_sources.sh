#!/bin/bash
# This script install the sources
# of libpcap
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
# LIBPCAP_DIR
# LIBPCAP_VERSION
#

BASE_URL=https://www.tcpdump.org/release

mkdir -p $LIBPCAP_DIR
cd $LIBPCAP_DIR
wget ${BASE_URL}/libpcap-$LIBPCAP_VERSION.tar.gz
cd $LIBPCAP_DIR
tar -xvf libpcap-$LIBPCAP_VERSION.tar.gz

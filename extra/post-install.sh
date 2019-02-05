#!/bin/sh

# Create 'netspot' user
adduser --no-create-home --disabled-password --disabled-login netspot

# Add sniffing packets capabilities to the binary /usr/local/bin/netspot
# CAP_NET_RAW is used to capture packets
# CAP_NET_ADMIN is used to put interface into promiscuous mode (it 
# is possibly useless but it may depends on the how the IDS is used)
#
# Added capabilities:
#    CAP_NET_RAW
#           * use RAW and PACKET sockets;
#           * bind to any address for transparent proxying.
#    CAP_NET_ADMIN
#           Perform various network-related operations:
#           * interface configuration;
#           * administration of IP firewall, masquerading, and accounting;
#           * modify routing tables;
#           * bind to any address for transparent proxying;
#           * set type-of-service (TOS)
#           * clear driver statistics;
#           * set promiscuous mode;
#           * enabling multicasting;
setcap 'CAP_NET_RAW+eip CAP_NET_ADMIN+eip' /usr/local/bin/netspot
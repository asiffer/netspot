---
title: Architecture
weight: 20
summary: netspot is also simple by its design
---

The picture below details the internal structure of `netspot`. It aims to present how the IDS is designed and it is also likely to help both the user and the developer to better understand the tool.


<object data="/assets/archi.svg" type="image/svg+xml" style="width: 100%;"></object>


At the lowest level, `netspot` parse packets and increment some basic **counters**. This part is performed by the `miner` subpackage.
The source can either be an network interface or a .pcap file (network capture).

At a given frequency, counter values are retrieved so as to build **statistics**, this is the role of the `analyzer`. The statistics are the measures monitored by `netspot`.

Every statistic embeds an instance of the `SPOT` algorithm to monitor itself. This algorithm learns the *normal* behaviour of the statistic and constantly updates its knowledge. When an abnormal value occurs, `SPOT` triggers an alarm.

The analyzer forwards statistics values, SPOT thresholds and SPOT alarms to the `exporter`. This last component dispatch
these information modules that binds to different backends 
(console, file, socket or InfluxDB database).
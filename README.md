# netspot

<table>
<tr>
<img src="assets/netspot.png" alt="drawing" width="300" align="left" style="display: block"/>
</tr>
<tr>
<code>netspot</code> is a simple <it>anomaly-based</it> network IDS written in <code>Go</code> (based on <a href="https://github.com/google/gopacket"><code>GoPacket</code></a>).

The <code>netspot</code> core uses <a href="https://asiffer.github.io/libspot/"><code>SPOT</code></a>, a statistical learning algorithm so as to detect abnormal behaviour in network traffic. 

<code>netspot</code> works as a server and can be controlled trough an HTTP REST API (a <code>Go</code> RPC endpoint is also available).
The current package embeds a client: <code>netspotctl</code> but the latter could be in a different package in the future.
</tr>
</table>

<!-- <div style="width: 100%; display: block">-</div> -->

## Table of contents
- [The SPOT algorithm through a single picture](#the-spot-algorithm-through-a-single-picture)
- [Installation](#installation)
	- [From sources](#from-sources)
    - [Debian package](#debian-package)
    - [Docker container](#docker-container)
- [Get started](#get-started)
- [REST API](#rest-api)
- [Architecture overview](#architecture-overview)
    - [Miner](#miner)
    - [Analyzer](#analyzer)
    - [Alarms](#alarms)
- [Notes](#notes)


## The SPOT algorithm through a single picture

As *a good sketch is better than a long speech*, we illustrate what the [`SPOT`](https://asiffer.github.io/libspot/) algorithm does below.

<center><img src="assets/plot.png" alt="drawing" width="90%"/></center>

## Installation

### From sources

You naturally have to clone the git repository and build the executables. The building process requires the `Go` compiler (I use the version `1.10.4` on linux/amd64 and `1.12.9` on linux/arm)
and some dependencies that you can get with `make deps`.

```bash
# clone
git clone github.com/asiffer/netspot.git
cd netspot
# get dependencies
make deps
# build
make
# install (it may require root privileges)
make install
```
The installation step put the two executables in `$(DESTDIR)/usr/bin`, the configuration file in `$(DESTDIR)/etc/netspot` and the `systemd` service file in `$(DESTDIR)/lib/systemd/system`. By default `$(DESTDIR)` is empty.

### Debian package

A debian package is also available on the release section. Two architectures are available `amd64` and `armhf` (for a Raspberry Pi for instance).

### Docker container

A `docker` image (based on `alpine`) also exists. Some options can naturally be added to start a new container.

```bash
docker run --rm --name=netspot \
                --net=host \
                -p 11000:11000 \
                -p 11001:11001 \
                -v netspot.toml:/etc/netspot/netspot.toml \
                asiffer/netspot-amd64:1.3
```


## Get started




## REST API

The current implementation of `netspot` embeds a `Go` client. However, `netspot` can be managed by other clients since it exposes a REST API.

The description of the API respects the [OpenAPI](https://swagger.io/specification/) standard and can be found [here](api/openapi.yaml). 
The endpoints are detailed in the [api](api/) folder.

## Architecture overview

<center><img src="assets/archi.svg" alt="Architecture" width="90%" /></center>

<!-- ![miner](assets/archi.svg) -->

At the lowest level, `netspot` parse packets and increment some basic **counters**. This part is performed by the `miner` subpackage.
The source can either be an network interface or a .pcap file (network capture).

At a given frequency, counter values are retrieved so as to build **statistics**, this is the role of the `analyzer`. The statistics are the measures monitored by `netspot`.

Every statistic embeds an instance of the `SPOT` algorithm to monitor itself. This algorithm learns the *normal* behaviour of the statistic and constantly updates its knowledge. When an abnormal value occurs, `SPOT` triggers an alarm.

A logging system stores the stat values and the corresponding thresholds either to files or to an [influxdb](https://www.influxdata.com/) instance.

### Miner

The goal of the `miner` is threefold:
* parse incoming packets
* dispatch layers to the concerned counters
* send snapshots along time (at the desired frequency)

<!-- ![miner](assets/miner.svg) -->
<center><img src="assets/miner.svg" alt="Miner" width="90%" /></center>

The dispatching is done concurrently to increase performances. However when a snapshot has to be done, packets parsing is paused and we wait for all the counters to finish to process the last layers they receive. It seems like it's long, but actually it's quite fast.

Many counters are already implemented within `netspot` like
* number of SYN packets;
* number of ICMP packets;
* number of IP packets;
* number of unique source IP addresses...

but `netspot` is designed to be modular, so every user is free to developed its own counter insofar as it respects the basic counter layout.



### Analyzer

Above the counters we can build network statistics on fixed-time intervals \(t\). For instance with the counters #SYN and #IP, the ratio of SYN packets can be computed: R_SYN = #SYN/#IP.

Every statistic embeds a `SPOT` instance to monitor itself. Like the counters, you can define all the desired statistics in so far as the required counters are implemented.

### Alarms

When a `SPOT` instance finds an abnormal value, it merely logs it (currently to a file or to InfluxDB).


## Notes

### Version 1.3

The IDS is quite ready for a release!
* New counters and new stats
* New HTTP API with OpenAPI spec
* Cleaner code
* New distributions options (Debian package, Docker image, `armhf` binaries)


### Version 1.2

Bye, bye Python... Welcome Go! The IDS has been reimplemented in `Go` for performances and concurrency reasons.

A controller (CLI) is also provided so as to manage the NetSpot service. I don't know if I will put it in another package later.

More tests are always needed.

### Version 1.1

This version is cleaner than the previous one. Some object have been added so as to balance the tasks. The interactive console is also simpler.

Now, I am reflecting on improving performances. Python is not very efficient for this purpose so I will probably use another programming language for specific and highly parallelizable tasks.

Sorry Scapy, but you take too long time to parse and dispatch packets...


### Version 1.0

This first version is ugly: everything is a big class! No, not really but the size of the main object has increased greatly with the new incoming ideas. So the next version will try to split it into smaller classes.

Moreover, there are not any unit tests (see cfy for good arguments), but the next version will be more serious (I hope).

There are probably many bugs, don't be surprised.

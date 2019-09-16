# netspot

<!-- ![NetSpot_logo](assets/netspot.png) -->

<img src="assets/netspot.png" alt="drawing" width="300"/>

`netspot` is a simple *anomaly-based* network IDS written in `Go` (based on [`GoPacket`](https://github.com/google/gopacket)). 
The `netspot` core uses [`SPOT`](https://asiffer.github.io/libspot/), a statistical learning algorithm so as to detect abnormal behaviour in network traffic. As *a good sketch is better than a long speech*, we illustrate what `netspot` does below.

<img src="assets/plot.png" alt="drawing" width="80%" align="center"/>

<!-- ![NetSpot_logo](assets/plot.png) -->

`netspot` works as a server and can be controlled trough an HTTP REST API (a `Go` RPC endpoint is also available).
The current package embeds a client: `netspotctl` but the latter could be in a different package in the future.

## Installation

### From sources

You naturally have to clone the git repository and build the executables. The building process requires the `Go` compiler
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

```sh
docker run --rm --name=netspot \
                --net=host \
                -p 11000:11000 \
                -p 11001:11001 \
                asiffer/netspot-amd64:1.3
```

### Snap package

Finally, `netspot` can also be installed through a `snap` package.


## REST API



## Architecture overview

![miner](assets/archi.svg)

At the lowest level, `netspot` parse packets and increment some basic **counters**. This part is performed by the `miner` subpackage.
The source can either be an network interface or a .pcap file (network capture).

At a given frequency, counter values are retrieved so as to build **statistics**, this is the role of the `analyzer`. The statistics are the measures monitored by `netspot`.

Every statistic embeds an instance of the `SPOT` algorithm to monitor itself. This algorithm learns the *normal* behaviour of the statistic and constantly updates its knowledge. When an abnormal value occurs, `SPOT` triggers an alarm.

A logging system stores the stat values and the corresponding thresholds either to files or to an [influxdb](https://www.influxdata.com/) instance.

## Miner

The goal of the `miner` is threefold:
* parse incoming packets
* dispatch layers to the concerned counters
* send snapshots along time (at the desired frequency)

![miner](assets/miner.svg)

The dispatching is done concurrently to increase performances. However when a snapshot has to be done, packets parsing is paused and we wait for all the counters to finish to process the last layers they receive. It seems like it's long, but actually it's quite fast.

Many counters are already implemented within `netspot` like
* number of SYN packets;
* number of ICMP packets;
* number of IP packets;
* number of unique source IP addresses...

but `netspot` is designed to be modular, so every user is free to developed its own counter insofar as it respects the basic counter layout.



### Statistics
Above the counters we can build network statistics on fixed-time intervals \(t\). For instance with the counters #SYN and #IP, the ratio of SYN packets can be computed: R_SYN = #SYN/#IP.

Every statistic embeds a `SPOT` instance to monitor itself. Like the counters, you can define all the desired statistics in so far as the required counters are implemented.

### Alarms

When a `SPOT` instance finds an abnormal value, it merely logs it (currently to a file or to InfluxDB).

## Usage

This tool is available through a debian package. Next I will show how to implement Counters/Stats and how to use the controller.

## Notes

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

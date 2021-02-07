![Build](https://github.com/asiffer/netspot/workflows/Build/badge.svg)
![Test](https://github.com/asiffer/netspot/workflows/Test/badge.svg)

# netspot

![netspot](assets/netspot6.png)

A simple IDS with statistical learning

Visit the project page: **https://asiffer.github.io/netspot/**

## Overview

**netspot** is a simple anomaly-based network IDS written in `Go` (based on [GoPacket](https://github.com/google/gopacket))

The **netspot** core uses [**SPOT**](https://asiffer.github.io/libspot/), a statistical learning algorithm so as to detect abnormal behaviour in network traffic (see the
picture below).

**netspot** is provided as a single and statically-compiled binary ([musl](https://www.musl-libc.org/) + [libpcap](https://www.tcpdump.org/)).

<!-- ![SPOT algorithm](assets/netspot4.png) -->


## Installation

### Binaries

The latest compiled binaries can be found below.

[![amd64](https://img.shields.io/badge/v2.0a-amd64-5e81ac?logo=go)](https://github.com/asiffer/netspot/releases/download/v2.0a/netspot-2.0a-amd64-linux-static)
[![arm](https://img.shields.io/badge/v2.0a-arm-81a1c1?logo=go)](https://github.com/asiffer/netspot/releases/download/v2.0a/netspot-2.0a-arm-linux-static)
[![arm64](https://img.shields.io/badge/v2.0a-arm64-88c0d0?logo=go)](https://github.com/asiffer/netspot/releases/download/v2.0a/netspot-2.0a-arm64-linux-static)

### Building from sources

To build **netspot** from sources, you mainly need a `Go` compiler and `libpcap-dev`. Through the basic `make` command, **netspot** is dynamically
linked to `libpcap` and your `libc` (generally GNU `libc`).

```sh
git clone https://github.com/asiffer/netspot.git
cd netspot
make
```

To perform a static build, you rather have to use `musl`. The [dev](dev/) folder
presents how **netspot** is (cross-)built based on the `golang:alpine` docker image.

## Get started

Basically, you can run `netspot` on a network interface. In the example below,
`netspot` monitors the `PERF` statistics (packet processing rate) on the `eth0` interface. 
The computation period is `1s` and the values are printed to the console (`-v`).

```sh
netspot run -d eth0 -s PERF -p 1s -v
```

You can also analyze a capture file.
```sh
netspot run -d file.pcap -s PERF -s R_SYN -p 500ms -v
```

All these command-line options can be set in a config file:
```toml
# netspot.toml

[miner]
device = "~/file.pcap"

[analyzer]
period = "500ms"
stats = ["PERF", "R_SYN"]

[exporter.console]
data = true
```

```sh
netspot run --config netspot.toml
```


All the available statistics can be listed with the `netspot  ls` command.



To print the default config (in TOML format only), you can run the following command:
```sh
netspot defaults
```

### Netspot As A Service

Even if it is not the main way to use **netspot**, it can 
run as a service, exposing a minimal REST API.

```sh
netspot serve
```

By default it listens at `tcp://localhost:11000` but it can be changed with the `-e` flag. For instance, if you want to monitor the loopback
interface, the API trafic must not pollute what **netspot** monitors,
so you can consider a unix socket.

```sh
netspot serve -e unix:///tmp/netspot.sock
```

| Method | Path          | Description                               |
| ------ | ------------- | ----------------------------------------- |
| `GET`  | `/api/config` | Get the current config (JSON output)      |
| `POST` | `/api/config` | Change the config (JSON expected)         |
| `POST` | `/api/run`    | Manage the status of netspot (start/stop) |



## Architecture overview

![architecture](assets/netspot-archi.png)


At the lowest level, `netspot` parse packets and increment some basic **counters**. This part is performed by the `miner` subpackage.
The source can either be an network interface or a .pcap file (network capture).

At a given frequency, counter values are retrieved so as to build **statistics**, this is the role of the `analyzer`. The statistics are the measures monitored by `netspot`.

Every statistic embeds an instance of the `SPOT` algorithm to monitor itself. This algorithm learns the *normal* behaviour of the statistic and constantly updates its knowledge. When an abnormal value occurs, `SPOT` triggers an alarm.

The analyzer forwards statistics values, SPOT thresholds and SPOT alarms to the `exporter`. This last component dispatch
these information modules that binds to different backends 
(console, file, socket or InfluxDB database).


## Roadmap

- [ ] Rework unit tests
- [ ] Enrich documentation
- [ ] Frontend? TUI?
- [ ] Port `netspot` to arduino :)

## Notes

### Version 2.0a

This is the second big refactoring. Many things have changed, making the way to use **netspot** more *modern*.

- Single and statically-compiled binary. Forget about the server, just run the binary on what you want (a server mode still exists but it is rather minimal)
- Better performances! I think that **netspot** can process 
twice as fast: **1M pkt/s** on my affordable desktop and **100K pkt/s** on a Raspberry 3B+. 
- Developper process has been improved so as to "easily" add new counters, statistics and exporting modules.

### Version 1.3

The IDS is quite ready for a release!
* New counters and new stats
* New HTTP API with OpenAPI spec
* Cleaner code
* New distributions options (Debian package, Docker image, `armhf` and `aarch64` binaries)


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

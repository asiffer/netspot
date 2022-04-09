![Release](https://shields.io/github/v/release/asiffer/netspot?include_prereleases)
![Build](https://github.com/asiffer/netspot/workflows/Build/badge.svg)
![Test](https://github.com/asiffer/netspot/workflows/Test/badge.svg)
[![Coverage](https://codecov.io/gh/asiffer/netspot/branch/master/graph/badge.svg)](https://codecov.io/gh/asiffer/netspot)
[![GoReport](https://goreportcard.com/badge/github.com/asiffer/netspot)](https://goreportcard.com/report/github.com/asiffer/netspot)

![netspot](assets/netspot6.png)

A simple IDS with statistical learning

Visit the project page: **https://asiffer.github.io/netspot/**

Join the [**Discord**](https://discord.gg/82HW2fydvD) server!

- [Overview](#overview)
- [Installation](#installation)
  - [Binaries](#binaries)
  - [Building from sources](#building-from-sources)
  - [(NEW) Docker](#new-docker)
- [Get started](#get-started)
  - [Inline](#inline)
  - [Service](#service)
  - [Architecture overview](#architecture-overview)
  - [Roadmap](#roadmap)
- [Notes](#notes)
  - [Version 2.0a](#version-20a)
  - [Version 1.3](#version-13)
  - [Version 1.2](#version-12)
  - [Version 1.1](#version-11)
  - [Version 1.0](#version-10)

## Overview

**netspot** is a simple anomaly-based network IDS written in `Go` (based on [GoPacket](https://github.com/google/gopacket))

The **netspot** core uses [**SPOT**](https://asiffer.github.io/libspot/), a statistical learning algorithm so as to detect abnormal behaviour in network traffic (see the
picture below).

**netspot** is provided as a single and statically-compiled binary ([musl](https://www.musl-libc.org/) + [libpcap](https://www.tcpdump.org/)).

## Installation

### Binaries

The latest compiled binaries can be found below.

[![amd64](https://img.shields.io/badge/v2.1.1-amd64-5e81ac?logo=go)](https://github.com/asiffer/netspot/releases/download/v2.1.1/netspot-2.1.1-amd64-linux-static)
[![arm](https://img.shields.io/badge/v2.1.1-arm-81a1c1?logo=go)](https://github.com/asiffer/netspot/releases/download/v2.1.1/netspot-2.1.1-arm-linux-static)
[![arm64](https://img.shields.io/badge/v2.1.1-arm64-88c0d0?logo=go)](https://github.com/asiffer/netspot/releases/download/v2.1.1/netspot-2.1.1-arm64-linux-static)

### Building from sources

To build **netspot** from sources, you mainly need a `Go` compiler (`>=1.16`) and `libpcap-dev`. Through the basic `make` command, **netspot** is dynamically
linked to `libpcap` and your `libc` (generally GNU `libc`).

```sh
git clone https://github.com/asiffer/netspot.git
cd netspot
make
```

Then you install it with `make install` (as root). Since version 2.1 it also installs a systemd service file.

To perform a static build, you rather have to use `musl`. The [dev](dev/) folder
details how **netspot** is (cross-)built based on the `golang:alpine` docker image.

### Docker

`netspot` is now available through a docker image, hosted on Github. You can have a look to the [local registry](https://github.com/users/asiffer/packages/container/package/netspot) to pull the image.

Once you have pulled the image, you can run `netspot` interactively through:

```sh
docker run --rm -it --name netspot --cap-add NET_ADMIN --network host netspot:latest
```

### (NEW) Systemd

Since version `2.1`, `netspot` is also available through a [systemd portable service](https://systemd.io/PORTABLE_SERVICES/).

You can also grab the `netspot.service` file and install it at /usr/local/lib/systemd/system/.

## Get started

### Inline

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

All the available statistics can be listed with the `netspot ls` command.

To print the default config (in TOML format only), you can run the following command:

```sh
netspot defaults
```

### Service

Even if it is not the main way to use **netspot**, it can
run as a service, exposing a minimal REST API.

```sh
netspot serve
```

By default it listens at `tcp://localhost:11000`, and you can visit `http://localhost:11000` to look at the simple dashboard that displays
the current config of `netspot`.

![dashboard](assets/dashboard.png)

Naturally, depending on the interface(s) you monitor, you would like to change the API endpoint not to pollute what `netspot` is analyzing.
You can be changed it with the `-e` flag. For instance, you can consider a unix socket.

```sh
netspot serve -e unix:///tmp/netspot.sock
```

The server exposes few methods that allows to do roughly everything.

| Method | Path           | Description                               |
| ------ | -------------- | ----------------------------------------- |
| `GET`  | `/api/config`  | Get the current config (JSON output)      |
| `POST` | `/api/config`  | Change the config (JSON expected)         |
| `POST` | `/api/run`     | Manage the status of netspot (start/stop) |
| `GET`  | `/api/stats`   | Get the list of available statistics      |
| `GET`  | `/api/devices` | Get the list of available interfaces      |

In addition, a `Go` client is available in the `api/client` subpackage.

```sh
go get -u github.com/asiffer/netspot/api/client
```

## Developer

`netspot` is rather modular so as to let people enriching it. Developers can notably
create their own counters, statistics or exporting modules. You should have a look to
the [project website](https://asiffer.github.io/netspot/developer/).

### Architecture overview

![architecture](assets/netspot-archi.png)

At the lowest level, `netspot` parses packets and increment some basic **counters**. This part is performed by the `miner` subpackage.
The packet source can either be an network interface or a .pcap file (network capture).

At a given period (for instance every second), counter values are retrieved so as to build **statistics**. This is the role of the `analyzer`. The statistics are the measures monitored by `netspot`.

Every statistic embeds an instance of the `SPOT` algorithm to monitor itself. This algorithm learns the _normal_ behaviour of the statistic and constantly updates its knowledge. When an abnormal value occurs, `SPOT` triggers an alarm.

Finally, the `analyzer` forwards stat values, SPOT thresholds and SPOT alarms to the `exporter`. This last component dispatch
these information to some modules that binds to different backends
(console, file, socket or InfluxDB database currently).

### Roadmap

Here are some ideas to improve netspot:

- [ ] InfluxDB (v2+) exporting module
- [ ] ElasticSearch exporting module
- [ ] MySQL exporting module
- [ ] New counters/statistics
- [ ] Provide a bare metal snap
- [ ] Web-based GUI (or TUI) for the netspot service
- [ ] Port `netspot` to arduino or other small dev board

### Building documentation

The docs is built with [`mkdocs`](https://www.mkdocs.org/). You also need some extensions:

```shell
pip3 install mkdocs mkdocs-markdownextradata-plugin mkdocs-material pymdown-extensions
```

## Notes

### Version 2.0a

This is the second big refactoring. Many things have changed, making the way to use **netspot** more _modern_.

- Single and statically-compiled binary. Forget about the server, just run the binary on what you want (a server mode still exists but it is rather minimal)
- Better performances! I think that **netspot** can process
  twice as fast: **1M pkt/s** on my affordable desktop and **100K pkt/s** on a Raspberry 3B+.
- Developper process has been improved so as to "easily" add new counters, statistics and exporting modules.

### Version 1.3

The IDS is quite ready for a release!

- New counters and new stats
- New HTTP API with OpenAPI spec
- Cleaner code
- New distributions options (Debian package, Docker image, `armhf` and `aarch64` binaries)

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
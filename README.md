<p align="center">
    <img src="https://raw.githubusercontent.com/asiffer/netspot/master/assets/netspot.png" width="100%" alt="asiffer/netspot">
</p>

## Overview

`netspot` is an anomaly-based network intrusion detection system (A-NIDS) written in `Go` leveraging both [`XDP`](https://en.wikipedia.org/wiki/Express_Data_Path) and [`gopacket`](https://github.com/google/gopacket) to inspect packets.

`netspot` aggregates network metrics on time slots and uses the [SPOT algorithm](https://asiffer.github.io/libspot/) to detect abnormal events. Simple.

> [!NOTE]  
> A brand new `v3` is now available. It is quite different from previous versions while doing the same job (flagging network anomalies with the SPOT algorithm). This new version drops side features like dashboard, api, docker... focusing on its first goals.

## About

This project is an extension of research work previously published at TrustCom'20 conference. 
If `netspot` contributes to a project that leads to a publication, please acknowledge this fact 
by citing this work.

```bibtex
@inproceedings{siffer2020netspot,
  title={Netspot: A simple Intrusion Detection System with statistical learning},
  author={Siffer, Alban and Fouque, Pierre-Alain and Termier, Alexandre and Largouet, Christine},
  booktitle={2020 IEEE 19th international conference on trust, security and privacy in computing and communications (TrustCom)},
  pages={911--918},
  year={2020},
  organization={IEEE}
}
```




## Installation

Download statically-compiled binaries from the [latest release](https://github.com/asiffer/netspot/releases/latest). 
Binaries for `amd64`, `arm64` and `armv7` are available.

Otherwise you can build directly from source:

```shell
go install github.com/asiffer/netspot@latest
```

> [!WARNING]
> The output binary notably needs `libpcap.so.1` installed 

## Getting Started

`netspot` basically needs a **source** (NIC or `.pcap` file) and a **tick** (time slot size)

```shell
netspot --source file.pcap --tick 250ms --all-stats
```

You can then track the anomalies through stdout logs.

```shell
10:38:18.176 INFO   Source defined source:200704011400.dump
10:38:18.176 INFO   Stats monitored stats:["BPS","PPS","APS","DPE","RACK","SYNFIN"]
10:38:18.176 INFO   Loading collector collector:gopacket
10:38:18.176 INFO   Open 200704011400.dump
10:38:18.176 INFO   Starting
10:38:19.904 INFO   Stat fit source_time_ns:1175404101043862000 stat:BPS timesince_ns:500162607000
10:38:19.904 INFO   Stat fit source_time_ns:1175404101043862000 stat:PPS timesince_ns:500162607000
10:38:19.904 INFO   Stat fit source_time_ns:1175404101043862000 stat:APS timesince_ns:500162607000
10:38:19.904 INFO   Stat fit source_time_ns:1175404101043862000 stat:DPE timesince_ns:500162607000
10:38:19.904 INFO   Stat fit source_time_ns:1175404101043862000 stat:RACK timesince_ns:500162607000
10:38:19.904 INFO   Stat fit source_time_ns:1175404101043862000 stat:SYNFIN timesince_ns:500162607000
10:38:19.906 WARN   Anomaly detected probability:0.0003707745276524869 source_time_ns:1175404101543963000 stat:APS threshold:58.972433165294696 timesince_ns:500662708000 value:58.990582191780824
10:38:19.909 WARN   Anomaly detected probability:0 source_time_ns:1175404102544374000 stat:APS threshold:58.972433165294696 timesince_ns:501663119000 value:59.19199346405229
...
```

If you want to monitor a network interface, you can use either `gopacket` (userspace level) or `xdp` (kernel level) collector.

```shell
sudo netspot --source eth0 --collector xdp --tick 1s --all-stats
```

## Advanced usage

### JSONL

All the collected data can be stored in a `jsonl` file (JSON records) to forward them to other tools or just analyze them afterwards.

```shell
netspot --source file.pcap --tick 250ms --output out.jsonl
```

You can parse it with the following json schema.

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "Netspot Record",
  "type": "object",
  "additionalProperties": false,
  "properties": {
    "time": {
      "type": "string",
      "format": "date-time"
    },
    "stats": {
      "type": "object",
      "propertyNames": { "$ref": "#/$defs/StatName" },
      "additionalProperties": { "$ref": "#/$defs/StatValue" }
    },
    "alerts": {
      "type": "array",
      "items": { "$ref": "#/$defs/StatName" }
    }
  },
  "required": ["time", "stats"],
  "$defs": {
    "StatName": {
      "type": "string",
      "enum": ["APS", "BPS", "DPE", "PPS", "RACK", "SYNFIN"]
    },
    "StatValue": {
      "type": "object",
      "additionalProperties": false,
      "properties": {
        "value": { "$ref": "#/$defs/NonFiniteFloat" },
        "spot_result": { "type": "integer" },
        "excess_threshold": { "$ref": "#/$defs/NonFiniteFloat" },
        "anomaly_threshold": { "$ref": "#/$defs/NonFiniteFloat" },
        "alert": { "$ref": "#/$defs/Alert" }
      },
      "required": ["value", "spot_result", "excess_threshold", "anomaly_threshold"]
    },
    "Alert": {
      "type": "object",
      "additionalProperties": false,
      "properties": {
        "probability": { "$ref": "#/$defs/NonFiniteFloat" }
      },
      "required": ["probability"]
    },
    "NonFiniteFloat": {
      "oneOf": [
        { "type": "number" },
        { "type": "string", "enum": ["NaN", "Inf", "-Inf"] }
      ]
    }
  }
}
```

> [!WARNING]
> As JSON does not manage `NaN` and `Inf`, you probably need some manual tweaks to make it serve your purpose.

### Spot algorithm

`netspot` defines some default parameters to monitor network statistics. You can modify them (per stat) through CLI flags.

| Parameter    | Type    | Flag                       | Description                               | Default value |
| ------------ | ------- | -------------------------- | ----------------------------------------- | ------------- |
| `q`          | `float` | `--spot-<stat>-q`          | Abnormal event probability                | `5e-4`        |
| `level`      | `float` | `--spot-<stat>-level`      | Out of tail distribution probability      | `0.98`        |
| `max-excess` | `uint`  | `--spot-<stat>-max-excess` | Maximum number of tail data for fitting   | `1000`        |
| `low`        |         | `--spot-<stat>-low`        | Monitor low values instead of high values |               |


See [libspot](https://asiffer.github.io/libspot/parameters/) to get the full picture.

## Benchmark

Here are some `netspot` benchmarks run on packet captures from [MAWILab](http://www.fukuda-lab.org/mawilab/index.html) (on Intel(R) Core(TM) i7-10850H CPU).

| source                                                                               | tick  | mean_time |     pkts | avg_pkt_size | pps     | gbps    |
| :----------------------------------------------------------------------------------- | :---- | --------: | -------: | -----------: | ------- | ------- |
| [200704011400.dump](http://www.fukuda-lab.org/mawilab/v1.1/2007/04/01/20070401.html) | 100ms |   3.76222 |  9503556 |        633.4 | 2526048 | 12.8    |
| [200704011400.dump](http://www.fukuda-lab.org/mawilab/v1.1/2007/04/01/20070401.html) | 500ms |   3.13571 |  9503556 |        633.4 | 3030746 | 15.3574 |
| [201302281400.dump](http://www.fukuda-lab.org/mawilab/v1.1/2013/02/28/20130228.html) | 100ms |   8.97875 | 25581540 |        759.9 | 2849120 | 17.3204 |
| [201302281400.dump](http://www.fukuda-lab.org/mawilab/v1.1/2013/02/28/20130228.html) | 500ms |   8.21167 | 25581540 |        759.9 | 3115267 | 18.9383 |
| [202005091400.pcap](http://www.fukuda-lab.org/mawilab/v1.1/2020/05/09/20200509.html) | 100ms |    29.205 | 90422857 |        431.9 | 3096143 | 10.6978 |
| [202005091400.pcap](http://www.fukuda-lab.org/mawilab/v1.1/2020/05/09/20200509.html) | 500ms |    26.525 | 90422857 |        431.9 | 3408967 | 11.7787 |

We can notice that `netspot` can treats about **3 millions packets/s** (with the `gopacket` collector). 
Also this benchmark shows that the bitrate processing ability (`gbps`) is not a fundamental constant. 
This is consistent with the fact that netspot does not inspect payloads.

## Contributing

This project is open to contributions! Here is the classical workflow: open an issue then we could discuss (among humans) about the bug/feature and plan something (or close it).

## Notes

### Version 3.0

Third big refactoring. Removing all the side stuff (API, docker, dashboard, systemd etc.) so as to _do one thing and do it well_.
The goal was to be able to flag network anomalies (from NIC or .pcap) while keeping collected data (for further analysis), not much.

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
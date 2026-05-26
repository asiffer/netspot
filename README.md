<p align="center">
    <img src="https://raw.githubusercontent.com/asiffer/netspot/master/assets/netspot.png" width="100%" alt="asiffer/netspot">
</p>

## Overview

`netspot` is an anomaly-based network intrusion detection system (A-NIDS) written in `Go` leveraging both [`XDP`](https://en.wikipedia.org/wiki/Express_Data_Path) and [`gopacket`](https://github.com/google/gopacket) to inspect packets.

`netspot` aggregates network metrics on time slots and uses the [SPOT algorithm](https://asiffer.github.io/libspot/) to detect abnormal events. Simple.

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

> [WARNING]
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

> [WARNING]
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
---
title: Basic usage
summary: Run `netspot` for the first time!
weight: 10
---

## One-liner

Basically, you can run `netspot` on a network interface. In the example below,
`netspot` monitors the `PERF` statistics (packet processing rate) on the `eth0` interface. 
The computation period is `1s` and the values are printed to the console (`-v`).

```sh
netspot run -d eth0 -s PERF -p 1s -v
```

You can also analyze a capture file with several statistics.
```sh
netspot run -d file.pcap -s PERF -s R_SYN -s R_ARP -p 500ms -v
```

## Config file

All the command-line options can be set in a config file (see the [configuration](/get-started/configuration/) section for more options):
```toml
# netspot.toml

[miner]
device = "~/file.pcap"

[analyzer]
period = "500ms"
stats = ["PERF", "R_SYN", "R_ARP"]

[exporter.console]
data = true
```

```sh
netspot run --config netspot.toml
```



## Data output

`netspot` outputs two things:

- the network statistics (namely the *data*, at every `period`)
- the alarms (when a computed stat is abnormal)

A **stat record** is a simple map `STAT: value` with a timestamp. When `netspot` learns,
it also gathers the decision thresholds `STAT_UP: upper_threshold` and `STAT_DOWN: lower_threshold`.

An **alarm** is generated once a statistics is beyond a threshold. It contains the value of 
the statistics and its probability to occur (the lower, the more abnormal).

`netspot` can dispatch these two streams to different exporting modules. In the previous example, data (not the alarms) are sent to the console (`-v` flag), but actually you can 
also send it to a file (see below), a socket or an influxdb database.

```sh
# storing data to a file
netspot run -d file.pcap -s PERF -s R_SYN -p 500ms -f /tmp/data.json
```

There is not a short CLI flag for every module. In the general case, you
have to use the `module.submodule.option` scheme (like in the config file):
```sh
netspot run -d file.pcap -s PERF -s R_SYN -p 500ms --exporter.file.data /tmp/data.json
```

Or you can set it in the config file:
```toml
# netspot.toml
[exporter.file]
# Path to the file which will store the data.
# The value can contain a '%s' which will be replaced by 
# the series name. 
data = "/tmp/netspot_%s_data.json"
# Same as the data but for the alarms
#alarm = "/tmp/netspot_%s_alarm.json"
```
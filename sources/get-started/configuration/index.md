---
title: Configuration
summary: All the stuff you can tune
weight: 15
---

The configuration of netspot follows its architecture.
There are three main sections: `miner`, 
`analyzer` and `exporter`. In addition, there is the 
`spot` section to configure the Spot algorithm.


## Miner

The miner is responsible of packet parsing. So you give it
all the necessary options to sniff either a network interface or
a pcap file. 

The main parameter is `device` that defines the packets source .
By default, it is set to `"any"`, meaning that it sniffs all the 
network interfaces. 
In addition you will find all the classical options you may pass
to `libpcap`.

!!! danger
    You must take care of the `timeout` parameter. By default, it is set to `0s`, meaning that packets are directly sent to netspot. If this value is changed, you are likely to have a time lag in the statistics computation.


```toml
# the Miner module manages the packets parsing
# and the counters
[miner]
# name of the interface to listen or dump/pcap file path
#device = "any"
device = "eth0"
#device = "/tmp/file.pcap"
# interface only
promiscuous = true
snapshot_len = 65535
# instant mode
timeout = "0s"
```

## Analyzer

The analyzer computes the statistics. So we only need to
set a list of statistics to monitor (parameter `stats`) and
to define the computation `period`.
The `period` is relative to the `device` netspot sniffs. If the `device` is a pcap file, the source of time is the 
capture timestamps while it is the real time in the network interface case.

!!! info
    You should set the `period` according to your device so as to make the stats computation relevant. For example, it is useless to set a very low value like `1ms` if you have few packets a second. 
    In practice, you should tune 
    this parameter to ensure a rather **low variance** of the computed statistics (i.e. stable values).


```toml
# The Analyzer module manages the statistics
# and send data to the exporter
[analyzer]
# time between two statistics computations
period = "1s"
# stats to load at startup
#stats = ["AVG_PKT_SIZE"]
stats = ["PERF", "R_SYN", "R_ACK"]
#stats = [
#    "PERF", 
#    "R_ACK", 
#    "R_ARP", 
#    "R_DST_SRC",
#    "R_DST_SRC_PORT", 
#    "R_ICMP", 
#    "R_IP", 
#    "R_SYN", 
#    "TRAFFIC"
#]
```



## Exporter

The exporter dispatches statistics and alarms to
the desired backend. 
The exporter gathers several basic modules like
the `console`, the `file` or the `socket`. In addition, netspot has also a module to send data
to `influxdb`.

For all the modules, you may notice that there are always two streams: data and alarms. You can
activate them independently.

### Console 

The configuration of this module could not be easier.

```toml
# The exporter print or send data according
# to the loaded modules (and their options)
# [exporter]
[exporter.console]
# print data to the console
data = true
# print alarms to the console
alarm = true
```

### File

The `file` module has a basic template. You can
add `%s` in the output file: this will be replaced by the name of the running series.
Data (and alarms) are stored as `json` records.

```toml
[exporter.file]
# Path to the file which will store the data.
# The value can contain a '%s' which will be
# replaced by the series name. 
data = "/tmp/netspot_%s_data.json"
# Same as the data but for the alarms
#alarm = "/tmp/netspot_%s_alarm.json"
```

### Socket

The `socket` allows to send data in a "generic way", meaning without setting the protocol upon.

In comparison to the above modules, you can add a `tag` into the sent data and change their format. 
Currently three formats are supported: `csv`, `json` and `gob` (golang binary format).

```toml
[exporter.socket]
# Path to the socket which will receive the data
# The format is the following: <proto>://<address>
data = "unix:///tmp/netspot_data.socket"
# Path to the socket which will receive the alarms
# The format is the following: <proto>://<address>
alarm = "unix:///tmp/netspot_alarm.socket"
# Additional tag when data are sent
tag = "netspot"
# Format of the data (accept csv, json or gob)
format = "json"
```

### InfluxDB

Finally, you can send netspot data to an [InfluxDB](https://www.influxdata.com/) database (version `v1.x`). 
Classically, you need to define the endpoint, some credentials and the name of the database.
Furthermore, for performance reasons, you can tune the `batch_size` (number of records to cache before sending).
Like in the `socket` module, you can define an `agent_name` which is a kind of tag 
(it can be very convenient for InfluxDB).

```toml
[exporter.influxdb1]
#data = true
#alarm = true
#address = "http://127.0.0.1:8086"
#database = "netspot"
#username = "netspot"
#password = "netspot"
#batch_size = 5
#agent_name = "local"
```


## Spot

The Spot section is specific to the configuration of the detection algorithm.
To well understand the parameters, we advise to look at the [Spot details](/advanced/spot/).
The parameters given in the `[spot]` section
are the default parameter of the Spot instances monitoring the network statistics.

```toml
# spot section manages the default spot parameters
[spot]
# depth = 50
# q = 1e-4
# n_init = 1000
# level = 0.98
# up = true
# down = false
# alert = true
# bounded = true
# max_excess = 200
```

However, you can define another Spot configuration
for some statistics. You only need to create a 
section `[spot.STAT_NAME]` and overwrite some parameters.

```toml
# you can add a specific section for a statistic
# it overrides the default values
[spot.R_SYN]
q = 1e-5
```
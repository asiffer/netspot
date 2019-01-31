# NetSpot

![NetSpot_logo](assets/bat.png)

`netspot` is a basic *anomaly-based* network IDS written in `Go` (based on [`GoPacket`](https://github.com/google/gopacket)). 
The `netspot` core uses [`SPOT`](https://asiffer.github.io/libspot/), a statistical learning algorithm so as to detect abnormal behaviour.

`netspot` works as a server and can be controlled trough an HTTP API.
The current package embeds a client: `netspotctl` but the latter could be in a different package in the future.


## Details

### Counters
The idea is to increment network counters with `GoPacket`. Examples of counters are the following:
* number of SYN packets;
* number of ICMP packets;
* number of IP packets;
* number of unique source IP addresses...

Some are already implemented, but we can add more as desired.


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

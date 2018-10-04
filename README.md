# netspot

`netspot` is a basic network IDS, in `python3`, based on [`Scapy`](https://scapy.net/) which embeds [`SPOT`](https://asiffer.github.io/libspot/), a statistical learning algorithm. `netspot` is managed by a command line interface.

This is merely a proof of concept, not a well implemented tool (not yet) 

## Details

### Counters
The idea is to increment network counters with `Scapy`. Examples of counters are the following:
* number of SYN packets;
* number of ICMP packets;
* number of IP packets;
* number of unique source IP addresses...

Some are already implemented, but we can add more as desired.


### Statistics
Above the counters we can build network statistics on fixed-time intervals \(t\). For instance with the counters #SYN and #IP, the ratio of SYN packets can be computed: R_SYN = #SYN/#IP.

Every statistic embeds a `SPOT` instance to monitor itself. Like the counters, you can define all the desired statistics in so far as the required counters are implemented.

### Alarms

When a `SPOT` instance finds an abnormal value, it merely logs it (currently to a file or a socket).

## Usage

This tool is available through a python3 package. First I have to understand how to make releases on Gitlab and then I will show you.

## Notes
### Version 1.0

This first version is ugly: everything is a big class! No, not really but the size of the main object has increased greatly with the new incoming ideas. So the next version will try to split it into smaller classes.

Moreover, there are not any unit tests (see cfy for good arguments), but the next version will be more serious (I hope).

There are probably many bugs, don't be surprised.
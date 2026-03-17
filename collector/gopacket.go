package collector

import (
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

var DummyTime = time.Unix(-(1<<63 - 1), rand.Int64())

// GoPacketCollector is a collector that uses gopacket
// to read packets from a pcap file
type GoPacketCollector struct {
	CollectorHooks
	handle *pcap.Handle
	source string
	tick   time.Duration
}

// NewGoPacketCollector returns a new GoPacketCollector
// initialized with the given source and tick duration.
//
// Parameters:
// - source: The path to the pcap file to be read.
// - tick: The duration between data sends.
//
// Returns:
// - A new instance of GoPacketCollector.
func NewGoPacketCollector(source string, tick time.Duration) Collector {
	return &GoPacketCollector{
		CollectorHooks: NewCollectorHooks(),
		source:         source,
		tick:           tick,
	}
}

// Config returns the configuration of the collector.
//
// Returns:
// - A CollectorConfig struct containing the source and tick duration.
func (c *GoPacketCollector) Config() CollectorConfig {
	return CollectorConfig{
		Source: c.source,
		Tick:   c.tick,
	}
}

// Load opens the pcap file specified in the source.
//
// Returns:
// - An error if the pcap file cannot be opened.
func (c *GoPacketCollector) Load() error {
	var err error
	c.log(fmt.Sprintf("Open %s", c.source))
	c.handle, err = pcap.OpenOffline(c.source)
	return err
}

// Unload closes the pcap handle if it is open.
//
// Returns:
// - An error if no handle is found.
func (c *GoPacketCollector) Unload() error {
	if c.handle == nil {
		return fmt.Errorf("no handle found")
	}
	c.handle.Close()
	return nil
}

// Start begins reading packets from the pcap file and processes them.
// It sends data at intervals specified by the tick duration or when
// a stop signal is received.
//
// Parameters:
// - stop: A channel to signal the collector to stop.
//
// Returns:
// - An error if there is an issue reading packets.
func (c *GoPacketCollector) Start(stop chan bool) {
	// init a Data structure. All the counters are zero
	data := Data{}

	var eth layers.Ethernet
	var arp layers.ARP
	var icmp layers.ICMPv4
	var icmp6 layers.ICMPv6
	var ip4 layers.IPv4
	var ip6 layers.IPv6
	// var ip6Frag layers.IPv6Fragment
	var tcp layers.TCP
	var udp layers.UDP
	parser := gopacket.NewDecodingLayerParser(layers.LayerTypeEthernet,
		&eth, &arp, &icmp, &icmp6, &ip4, &ip6, &tcp, &udp)
	parser.IgnoreUnsupported = true
	decoded := []gopacket.LayerType{}

	// init reference time
	lastSendTimestamp := DummyTime

	for {
		select {
		case <-stop:
			c.log("receiving stop signal")
			c.log("sending last data")
			c.send(&data)
			c.end(nil)
			return
		default:
			payload, metadata, err := c.handle.ZeroCopyReadPacketData()
			if err != nil {
				c.send(&data)
				c.end(err)
				return
			}

			// check the clock
			if lastSendTimestamp.Equal(DummyTime) {
				// first packet
				lastSendTimestamp = metadata.Timestamp
			} else if metadata.Timestamp.Sub(lastSendTimestamp) > c.tick {
				// send data
				c.send(&data)
				// update the last send timestamp
				lastSendTimestamp = metadata.Timestamp
			}

			// update TIME, PKT and BYTES
			data.TIME = uint64(metadata.Timestamp.UnixNano())
			data.PKT += 1
			data.BYTES += uint64(len(payload))

			if err := parser.DecodeLayers(payload, &decoded); err != nil {
				continue
			}

			for _, layerType := range decoded {
				switch layerType {
				case layers.LayerTypeARP:
					data.ARP += 1
				case layers.LayerTypeICMPv4:
					data.ICMP += 1
				case layers.LayerTypeICMPv6:
					data.ICMP6 += 1
				case layers.LayerTypeIPv6:
					data.IP6 += 1
					// TODO: fragmented packet
				// case layers.LayerTypeIPv6Fragment:
				// 	if ip6Frag.MoreFragments || ip6Frag.FragmentOffset != 0 {
				// 		data.FRAG += 1
				// 	}
				case layers.LayerTypeIPv4:
					data.IP += 1
					if ip4.Flags&layers.IPv4MoreFragments != 0 || ip4.FragOffset != 0 {
						data.FRAG += 1
					}
				case layers.LayerTypeTCP:
					data.TCP += 1
					if tcp.ACK {
						data.ACK += 1
					}
					if tcp.SYN {
						data.SYN += 1
					}
					if tcp.FIN {
						data.FIN += 1
					}
					if tcp.RST {
						data.RST += 1
					}
					data.TCP_DST_PORTS[tcp.DstPort] += 1
				case layers.LayerTypeUDP:
					data.UDP += 1

				}
			}
		}
	}
}

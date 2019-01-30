// tap.go
package miner

import (
	"fmt"
	"netspot/miner/counters"
	"os"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/rs/zerolog/log"
)

var (
	parser *gopacket.DecodingLayerParser
)

var (
	AvailableDevices DeviceList    // list of available interfaces
	device           string        // name of the device (interface of pcap file)
	iface            bool          // tells if the packet source is an interface
	snapshotLen      int32         // the maximum size to read for each packet
	promiscuous      bool          // promiscuous mode of the interface
	timeout          time.Duration // time to wait if nothing happens
	nbParsedPkts     uint64        // the number of parsed packets
	sniffing         bool          // tells if the package is currently sniffing
)

var (
	SourceTime time.Time      // the clock given by the packet capture
	tickPeriod time.Duration  // time between two data sending (if stat computation)
	last       time.Time      // time of the last data sending
	ticker     chan time.Time // channel sending time (at a given frequency in practice)
	sendTicks  bool           // tells if ticks have to be sent
)

var (
	handle *pcap.Handle
	err    error
)

func (dl DeviceList) contains(str string) bool {
	for _, s := range dl {
		if s == str {
			return true
		}
	}
	return false
}

func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// GetAvailableDevices returns the current available interfaces
func GetAvailableDevices() DeviceList {
	dl, _ := pcap.FindAllDevs()
	devNames := make([]string, 0)
	for _, dev := range dl {
		devNames = append(devNames, dev.Name)
	}
	return devNames
}

// Tick returns a channel sending frequent tick (at each duration)
func Tick(d time.Duration) chan time.Time {
	last = SourceTime.Add(6 * time.Hour)
	tickPeriod = d
	sendTicks = true
	return ticker
}

func updateTime(t time.Time) {
	SourceTime = t

	if sendTicks {
		if SourceTime.Sub(last) < 0 {
			last = SourceTime
			return
		} else if SourceTime.Sub(last) > tickPeriod {
			last = SourceTime
			ticker <- SourceTime

		}
	}

}

// release changes the state of the package.
func release() {
	// log.Debug().Msg("Releasing ")
	sniffing = false
	sendTicks = false
	stopAllCounters()
	close(events)
}

// Sniff Open the device and start to sniff packets.
// These packets are sent to the the dispatcher
func Sniff() {
	// Open the handler according to the source (interface or file)
	if iface {
		handle, err = pcap.OpenLive(device, snapshotLen, promiscuous, pcap.BlockForever)
	} else {
		handle, err = pcap.OpenOffline(device)
	}
	if err != nil {
		log.Error().Msg(fmt.Sprintf("Error while opening device (%s)", err))
	}
	defer handle.Close()
	nbParsedPkts = 0
	// init the packet source from the handler
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	// Start all the counters (if they are not running)
	startAllCounters()
	sniffing = true
	// loop over the incoming packets
	for {
		select {
		// stop the loop
		// case <-stopSniff:
		case e := <-events:
			// event 0 means "stop sniffing"
			if e == 0 {
				log.Info().Msg("Stop sniffing")
				release()
				return
			}
		// parse a packet
		case packet, ok := <-packetSource.Packets():
			// check whether it is the last packet or not
			if ok {
				updateTime(packet.Metadata().Timestamp)
				go dispatch(packet)
				nbParsedPkts++
			} else {
				// if there is no packet anymore, we stop it
				log.Info().Msgf("No packets to parse anymore (%d parsed packets)", GetNbParsedPkts())
				log.Info().Msg("Stop sniffing")
				release()
				return
			}
		}

	}
}

// dispatch Sends the incoming packet to the loaded
// counters
func dispatch(pkt gopacket.Packet) {
	var ip *layers.IPv4
	var tcp *layers.TCP
	var icmp *layers.ICMPv4

	ipLayer := pkt.Layer(layers.LayerTypeIPv4)
	if ipLayer != nil {
		ip, _ = ipLayer.(*layers.IPv4)
		tcpLayer := pkt.Layer(layers.LayerTypeTCP)
		if tcpLayer != nil {
			tcp, _ = tcpLayer.(*layers.TCP)
		} else {
			icmpLayer := pkt.Layer(layers.LayerTypeICMPv4)
			if icmpLayer != nil {
				icmp, _ = icmpLayer.(*layers.ICMPv4)
			}
		}
	}

	for _, ctr := range counterMap {
		switch ctr.(type) {
		case counters.IpCtrInterface:
			if ip != nil {
				ipctr, _ := ctr.(counters.IpCtrInterface)
				ipctr.LayPipe() <- ip
				// ipctr, ok := ctr.(counters.IpCtrInterface)
				// if ok {
				// 	ipctr.LayPipe() <- ip
				// }
			}
		case counters.TcpCtrInterface:
			if tcp != nil {
				tcpctr, _ := ctr.(counters.TcpCtrInterface)
				tcpctr.LayPipe() <- tcp
				// tcpctr, ok := ctr.(counters.TcpCtrInterface)
				// if ok {
				// 	tcpctr.LayPipe() <- tcp
				// }
			}
		case counters.IcmpCtrInterface:
			if icmp != nil {
				icmpctr, _ := ctr.(counters.IcmpCtrInterface)
				icmpctr.LayPipe() <- icmp
				// icmpctr, ok := ctr.(counters.IcmpCtrInterface)
				// if ok {
				// 	icmpctr.LayPipe() <- icmp
				// }
			}
		}
	}
}

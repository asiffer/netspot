// tap.go

package miner

import (
	"fmt"
	"netspot/miner/counters"
	"os"
	"sync"
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
	// AvailableDevices is the list of available interfaces
	AvailableDevices []string
	device           string        // name of the device (interface of pcap file)
	iface            bool          // tells if the packet source is an interface
	snapshotLen      int32         // the maximum size to read for each packet
	promiscuous      bool          // promiscuous mode of the interface
	timeout          time.Duration // time to wait if nothing happens
	nbParsedPkts     uint64        // the number of parsed packets
	sniffing         bool          // tells if the package is currently sniffing
)

var (
	// SourceTime is the clock given by the packet capture
	SourceTime time.Time
	tickPeriod time.Duration  // time between two data sending (if stat computation)
	last       time.Time      // time of the last data sending
	ticker     chan time.Time // channel sending time (at a given frequency in practice)
	sendTicks  bool           // tells if ticks have to be sent
)

var (
	handle *pcap.Handle
	err    error
)

func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// GetAvailableDevices returns the current available interfaces
func GetAvailableDevices() []string {
	dl, err := pcap.FindAllDevs()
	devNames := make([]string, 0)
	if err == nil {
		for _, dev := range dl {
			devNames = append(devNames, dev.Name)
		}
	} else {
		fmt.Println(err.Error())
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
func release(goroutinePool *sync.WaitGroup) {
	goroutinePool.Wait()
	sniffing = false
	stopAllCounters()
	if sendTicks {
		// a last tick is send to trigger the analyzer tick event
		ticker <- time.Now()
		sendTicks = false
	}
	close(events)
}

// func release() {
// 	sniffing = false
// 	sendTicks = false
// 	stopAllCounters()
// 	close(events)
// }

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

	// reinit the number of parse packets
	nbParsedPkts = 0
	// init the packet source from the handler
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	// Start all the counters (if they are not running)
	startAllCounters()
	// now we are sniffing!
	sniffing = true
	// pool of goroutines (for every packets)
	var goroutinePool sync.WaitGroup

	// loop over the incoming packets
	for {
		select {
		case e := <-events:
			// event 0 means "stop sniffing"
			if e == 0 {
				log.Info().Msg("Stop sniffing")
				release(&goroutinePool)
				return
			}
		// parse a packet
		case packet, ok := <-packetSource.Packets():
			// check whether it is the last packet or not
			if ok {
				updateTime(packet.Metadata().Timestamp)
				// add the packet parsing to the goroutine pool
				goroutinePool.Add(1)
				go dispatch(&goroutinePool, packet)
				nbParsedPkts++
			} else {
				// if there is no packet anymore, we stop it
				log.Info().Msgf("No packets to parse anymore (%d parsed packets)", GetNbParsedPkts())
				log.Info().Msg("Stop sniffing")
				release(&goroutinePool)
				return
			}
		}

	}
}

// dispatch Sends the incoming packet to the loaded
// counters
func dispatch(goroutinePool *sync.WaitGroup, pkt gopacket.Packet) {
	defer goroutinePool.Done()

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
		case counters.IPCtrInterface:
			if ip != nil {
				ipctr, _ := ctr.(counters.IPCtrInterface)
				ipctr.LayPipe() <- ip
			}
		case counters.TCPCtrInterface:
			if tcp != nil {
				tcpctr, _ := ctr.(counters.TCPCtrInterface)
				tcpctr.LayPipe() <- tcp
			}
		case counters.ICMPCtrInterface:
			if icmp != nil {
				icmpctr, _ := ctr.(counters.ICMPCtrInterface)
				icmpctr.LayPipe() <- icmp
			}
		}
	}
}

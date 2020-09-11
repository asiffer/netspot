// miner.go

// Package miner aims to read either network interfaces or
// network captures to increment basic counters. This is
// the lowest layer of netspot.
package miner

import (
	"errors"
	"fmt"
	"netspot/miner/counters"
	"os"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/pfring"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

//----------------------------------------------------------------------------//
//----------------------------- PACKAGE VARIABLES ----------------------------//
//----------------------------------------------------------------------------//

// EventChannel defines a channel to send events
type EventChannel chan uint8

// DataChannel defines a channel to send counters data
type DataChannel chan map[string]uint64

// TimeChannel defines a channel to send time ticks
type TimeChannel chan time.Time

// Logger
var (
	minerLogger zerolog.Logger
)

// Storing/Accessing the counters
var (
	mux, valmux          sync.RWMutex         // Locker for the counter map access
	internalEventChannel = make(EventChannel) // to receive events
)

// Status
var (
	availableDevices = GetAvailableDevices()
	device           string        // name of the device (interface of pcap file)
	iface            bool          // tells if the packet source is an interface
	snapshotLen      int32         // the maximum size to read for each packet
	promiscuous      bool          // promiscuous mode of the interface
	timeout          time.Duration // time to wait if nothing happens
	sniffing         bool          // tells if the package is currently sniffing
)

// Time
var (
	// SourceTime is the clock given by the packet capture
	SourceTime = time.Now()
	last       = time.Now()
)

// Packet sniffing/parsing
var (
	handle     *pcap.Handle
	ringHandle *pfring.Ring
	parser     *gopacket.DecodingLayerParser
	err        error
)

// Dispatcher
var dispatcher = NewDispatcher()

// Events
const (
	// STOP stops a counter
	STOP uint8 = 0
	// GET trigger a snapshot of the counters
	GET uint8 = 1
	// FLUSH trigger a snapshot of the counters and reset them
	FLUSH uint8 = 2
)

// Custom error ============================================================= //
// ========================================================================== //
// ========================================================================== //

func init() {}

// InitLogger initializes a specific logger for the miner package
func InitLogger() {
	minerLogger = log.With().Str("module", "MINER").Logger()
}

// Zero aims to zero the internal state of the miner. So it removes all
// the loaded counters, reset some variables
func Zero() error {
	if IsSniffing() {
		minerLogger.Error().Msg("Cannot reload, sniffing in progress")
		return errors.New("Cannot reload, sniffing in progress")
	}

	// reset the dispatcher
	dispatcher = NewDispatcher()

	// everything is ok
	minerLogger.Info().Msg("Miner package has been reset")
	return nil

}

// Information function ===================================================== //
// ========================================================================== //
// ========================================================================== //

// GetAvailableDevices returns the current available interfaces
func GetAvailableDevices() []string {
	dl, err := pcap.FindAllDevs()
	if err != nil {
		minerLogger.Error().Msgf("Error while listing network interfaces: %v", err)
		return nil
	}
	devices := make([]string, 0)
	for _, dev := range dl {
		devices = append(devices, dev.Name)
	}
	return devices
}

// GetAvailableCounters returns the list of the implemented
// counters (within the 'counters' subpackage)
func GetAvailableCounters() []string {
	names := make([]string, 0, len(counters.AvailableCounters))
	for n := range counters.AvailableCounters {
		names = append(names, n)
	}
	return names
}

// i/o function ============================================================= //
// ========================================================================== //
// ========================================================================== //

// Load loads a counter given its name
func Load(name string) error {
	return dispatcher.load(name)
}

// Unload unloads a counter given its name
func Unload(name string) error {
	return dispatcher.unload(name)
}

// UnloadAll removes all the counter from the miner
func UnloadAll() {
	// we just create a new dispatcher
	dispatcher = NewDispatcher()
}

// GetLoadedCounters returns the list of counters
// loaded by the miner
func GetLoadedCounters() []string {
	return dispatcher.loadedCounters()
}

// Sniffing function ======================================================== //
// ========================================================================== //
// ========================================================================== //

func closeDevice() {
	if ringHandle != nil {
		ringHandle.Close()
	}
	if handle != nil {
		handle.Close()
	}
}

func openDevice() (*gopacket.PacketSource, error) {
	var err error
	if iface {
		// in case of network interface
		// use PF_RING
		flag := pfring.FlagTimestamp
		if promiscuous {
			flag = flag | pfring.FlagPromisc
		}
		ringHandle, err = pfring.NewRing(device, uint32(snapshotLen), flag)
		if err != nil {
			return nil, err
		}
		// PF_RING has a ton of optimizations and tweaks to
		// make sure you get just the packets you want. For
		// example, if you're only using pfring to read packets,
		// consider running:
		ringHandle.SetSocketMode(pfring.ReadOnly)

		// If you only care about packets received on your
		// interface (not those transmitted by the interface),
		// you can run:
		ringHandle.SetDirection(pfring.ReceiveOnly)

		if err = ringHandle.Enable(); err != nil { // Must do this!, or you get no packets!
			return nil, err
		}
		return gopacket.NewPacketSource(ringHandle, layers.LinkTypeEthernet), nil
	}

	// Otherwise use libpcap
	handle, err = pcap.OpenOffline(device)

	if err != nil {
		minerLogger.Error().Msg(fmt.Sprintf("Error while opening device (%s)", err))
	}

	// init the packet source from the handler
	return gopacket.NewPacketSource(handle, handle.LinkType()), nil
}

// sniffAndYield opens a device and starts to sniff.
// It sends counters snapshot at given period
func sniffAndYield(period time.Duration, event EventChannel, data DataChannel) error {
	defer close(data)

	packetSource, err := openDevice()
	minerLogger.Debug().Msgf("Device %s open (%v)", device, err)
	if err != nil {
		return err
	}
	minerLogger.Debug().Msgf("Device %s open", device)
	defer closeDevice()
	minerLogger.Debug().Msgf("Device %s open", device)

	packetsChan := packetSource.Packets()
	// init timestamp
	firstPacket := <-packetsChan
	lastTimestamp := firstPacket.Metadata().Timestamp
	// Start all the counters (if they are not running)
	dispatcher.init()
	// now we are sniffing!
	sniffing = true

	// loop over the incoming packets
	for {
		select {
		// manage events
		case e, _ := <-event:
			switch e {
			case STOP:
				// the counters are stopped
				minerLogger.Debug().Msg("Receiving STOP")
				dispatcher.terminate()
				sniffing = false
				return nil
			case GET:
				// counter values are retrieved and sent
				// to the channel.
				minerLogger.Debug().Msg("Receiving GET")
				data <- dispatcher.getAll()
			case FLUSH:
				// counter values are retrieved and sent
				// to the channel. The counters are also
				// reset.
				minerLogger.Debug().Msg("Receiving FLUSH")
				data <- dispatcher.flushAll()
			default:
				minerLogger.Debug().Msgf("Receiving unknown event (%v)", e)
			}
		// parse packet
		case packet, ok := <-packetsChan:
			// check whether it is the last packet or not
			if ok {
				SourceTime = packet.Metadata().Timestamp
				// update time and check if data must be sent through the channel
				if lastTimestamp.Add(period).Before(SourceTime) {
					lastTimestamp = SourceTime
					dispatcher.terminate()
					data <- dispatcher.flushAll()
				}
				dispatcher.pool.Add(1)
				// in all cases dispatch the packet to the counters
				go dispatcher.dissect(packet)
			} else {
				// if there is no packet anymore, we stop it
				minerLogger.Info().Msgf("No packets to parse anymore (%d parsed packets).",
					dispatcher.receivedPackets)
				dispatcher.terminate()
				sniffing = false
				return nil
			}
		}

	}

}

// sniff open the device and start to sniff packets.
// These packets are sent to the the dispatcher which
// increment counters. It does not flush automatically
func sniff(event EventChannel, data DataChannel) error {
	defer close(data)

	packetSource, err := openDevice()
	if err != nil {
		return err
	}
	defer closeDevice()

	packetsChan := packetSource.Packets()
	// Start all the counters (if they are not running)
	dispatcher.init()
	// now we are sniffing!
	sniffing = true

	// loop over the incoming packets
	for {
		select {
		// manage events
		case e, _ := <-event:
			switch e {
			case STOP:
				// the counters are stopped
				minerLogger.Debug().Msg("Receiving STOP")
				dispatcher.terminate()
				sniffing = false
				return nil
			case GET:
				// counter values are retrieved and sent
				// to the channel.
				minerLogger.Debug().Msg("Receiving GET")
				data <- dispatcher.getAll()
			case FLUSH:
				// counter values are retrieved and sent
				// to the channel. The counters are also
				// reset.
				minerLogger.Debug().Msg("Receiving FLUSH")
				data <- dispatcher.flushAll()
			default:
				minerLogger.Debug().Msgf("Receiving unknown event (%v)", e)
			}
		// parse packet
		case packet, ok := <-packetsChan:
			// check whether it is the last packet or not
			if ok {
				dispatcher.pool.Add(1)
				// dispatch packet to the counters
				go dispatcher.dissect(packet)
			} else {
				// if there is no packet anymore, we stop it
				minerLogger.Info().Msgf("No packets to parse anymore (%d parsed packets).",
					dispatcher.receivedPackets)
				dispatcher.terminate()
				sniffing = false
				return nil
			}
		}

	}

}

// Side functions =========================================================== //
// ========================================================================== //
// ========================================================================== //

func contains(list []string, str string) bool {
	for _, s := range list {
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

// Main ===================================================================== //
// ========================================================================== //
// ========================================================================== //

func main() {}

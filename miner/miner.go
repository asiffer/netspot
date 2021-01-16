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
	"path"
	"strings"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
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
	mux, valmux          sync.RWMutex            // Locker for the counter map access
	internalEventChannel = make(EventChannel, 1) // to receive events
	// externalDataChannel  = make(DataChannel, 1)  // to send data to the analyzer
)

// Status
var (
	availableDevices = GetAvailableDevices()
	device           string        // name of the device (interface of pcap file)
	iface            bool          // tells if the packet source is an interface
	snapshotLen      int32         // the maximum size to read for each packet
	promiscuous      bool          // promiscuous mode of the interface
	timeout          time.Duration // time to wait if nothing happens
)

// Packet sniffing/parsing
var (
	handle *pcap.Handle
	// ringHandle *pfring.Ring
	parser *gopacket.DecodingLayerParser
	err    error
)

// Dispatcher
var dispatcher = NewDispatcher()

// Events
const (
	ERR uint8 = 255
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

func init() {
	reset()
}

// reset the variables of the package
func reset() {
	// create data channel
	// they must be buffered to avoid blocking steps.
	// I had this kind of probleme when a STOP msg is sent
	// to the analyzer during the flush of the miner: the miner
	// wants to send flushed data to the analyzer, but the
	// analyzer waits for the miner to end.
	internalEventChannel = make(EventChannel, 1)
	// externalDataChannel = make(DataChannel, 1)

	// reset the dispatcher
	dispatcher = NewDispatcher()

	// update device list
	availableDevices = GetAvailableDevices()

	// sniffing
	sniffing = NewSniffingStatus()

	// source time
	sourceTime = NewSourceTime()
}

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

	reset()
	// everything is ok
	minerLogger.Info().Msg("Miner package (re)loaded")
	return nil

}

// GetSeriesName returns the name of the series according
// to the device it sniffs
func GetSeriesName() string {
	if IsDeviceInterface() {
		t := time.Now()
		f := t.Format(time.StampMilli)
		f = strings.Replace(f, " ", "-", -1)
		return fmt.Sprintf("%s-%s", device, f)
	}
	p := path.Base(device)
	ext := path.Ext(p)
	return strings.Replace(p, ext, "", -1)
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

// IsSniffing returns the sniffing status
func IsSniffing() bool {
	return sniffing.Status()
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
	dispatcher.counters = nil
	dispatcher.list = nil
	dispatcher = nil
	// we just create a new dispatcher
	dispatcher = NewDispatcher()
}

// GetLoadedCounters returns the list of counters
// loaded by the miner
func GetLoadedCounters() []string {
	return dispatcher.loadedCounters()
}

// GetSourceTime returns the time given by the current packet source
func GetSourceTime() int64 {
	return sourceTime.GetNano()
}

// Sniffing function ======================================================== //
// ========================================================================== //
// ========================================================================== //

func release() {
	// close(externalDataChannel)
	// externalDataChannel = make(DataChannel, 1)
	sniffing.End()
}

// openDevice returns a handle on the packet source (interface or file)
func openDevice() (*pcap.Handle, error) {

	// Offline mode ----------------------------------------------------------
	// -----------------------------------------------------------------------
	if !IsDeviceInterface() {
		return pcap.OpenOffline(device)
	}

	// Online mode -----------------------------------------------------------
	// -----------------------------------------------------------------------
	inactive, err := pcap.NewInactiveHandle(device)
	if err != nil {
		return nil, err
	}
	defer inactive.CleanUp()

	// config ----------------------------------------------
	if err := inactive.SetSnapLen(int(snapshotLen)); err != nil {
		return nil, err
	}

	if timeout == 0 {
		if err := inactive.SetImmediateMode(true); err != nil {
			return nil, err
		}
	} else {
		if err := inactive.SetTimeout(timeout); err != nil {
			return nil, err
		}
	}

	if err := inactive.SetPromisc(promiscuous); err != nil {
		return nil, err
	}

	// Finally, create the actual handle by calling Activate:
	return inactive.Activate() // after this, inactive is no longer valid
}

// sniff open the device and call either the offline sniffer or the
// online one
func sniff(period time.Duration, data DataChannel) {
	// data channel should be closed to send a 'nil' object
	// to the analyzer. This is the way the analyzer understands
	// that the miner has ended.
	// close the channel when ends
	defer close(data)

	// Open the device
	handle, err := openDevice()
	if err != nil {
		minerLogger.Error().Msgf("Fail to open the device '%s': %v", device, err)
		internalEventChannel <- ERR
		return
	}
	defer handle.Close()
	minerLogger.Debug().Msgf("Device %s is open", device)

	// TODO: BPF hook
	// if err := handle.SetBPFFilter(bpf); err != nil {
	// 	minerLogger.Error().Msgf("Fail to open the device '%s': %v", device, err)
	// 	event <- ERR
	// 	return err
	// }

	// Create the packet source
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	// set decoding options
	packetSource.DecodeOptions.Lazy = true
	packetSource.DecodeOptions.NoCopy = true
	packetSource.Lazy = true
	packetSource.NoCopy = true
	// packet channel
	packetChan := packetSource.Packets()
	// Start all the counters (if they are not running)
	dispatcher.init()
	// run
	if IsDeviceInterface() {
		err = sniffOnline(packetChan, period, data)
	} else {
		err = sniffOffline(packetChan, period, data)
	}

	if err != nil {
		minerLogger.Error().Msgf("Error while sniffing: %v", err)
		internalEventChannel <- ERR
	}
}

// Start starts the miner and demands it to send
// counter values at given period. It returns the
// channel where counters are sent
func Start(period time.Duration) (DataChannel, error) {
	if IsSniffing() {
		return nil, fmt.Errorf("Already sniffing")
	}
	if len(dispatcher.loadedCounters()) == 0 {
		return nil, fmt.Errorf("No counters loaded")
	}

	minerLogger.Info().Msgf("Start sniffing %s", device)
	minerLogger.Debug().Msgf("Loaded counters: %v", dispatcher.loadedCounters())

	// Create data channel. It is closed at the end of the
	// sniff function
	data := make(DataChannel, 1)
	// sniff
	go sniff(period, data)

	// wait for sniffing
	for !IsSniffing() {
		select {
		case <-internalEventChannel: // error case
			return nil, fmt.Errorf("Something bad happened")
		default:
			// pass
		}
	}
	return data, nil
}

// Stop stops to sniff the device
// It waits until the miner is stopped
// (returns always nil)
func Stop() error {
	if !IsSniffing() {
		minerLogger.Warn().Msg("The miner is already stopped")
		return nil
	}
	minerLogger.Info().Msgf("Stopping counter...")
	// send STOP msg to the sniff function
	internalEventChannel <- STOP

	minerLogger.Info().Msgf("Wait for stop sniffing...")
	for IsSniffing() {
		// wait for stop sniffing
	}
	return nil
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

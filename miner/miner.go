// miner.go

// Package miner aims to read either network interfaces or
// network captures to increment basic counters.
package miner

import (
	"errors"
	"fmt"
	"netspot/miner/counters"
	"os"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

//----------------------------------------------------------------------------//
//----------------------------- PACKAGE VARIABLES ----------------------------//
//----------------------------------------------------------------------------//

// EventChannel defines a channel to send events
type EventChannel chan uint8

// DataChannel defines a channel to send counters data
type DataChannel chan map[int]uint64

// TimeChannel defines a channel to send time ticks
type TimeChannel chan time.Time

// Logger
var (
	minerLogger zerolog.Logger
)

// Storing/Accessing the counters
var (
	counterMap          map[int]counters.BaseCtrInterface // Map id->counter
	counterID           int                               // Id to store counters in counterMap
	counterValues       map[int]uint64                    // temp container to store counter values
	mux, valmux         sync.RWMutex                      // Locker for the counter map access
	defaultEventChannel EventChannel                      // to receive events
	defaultDataChannel  DataChannel                       // internal channel to send snapshots
	// builtinCounterMap   map[int]string
)

// var (
// 	// Miner built-in counters
// 	builtinCounters []string
// )

// Status
var (
	// AvailableDevices is the list of available interfaces
	AvailableDevices []string
	device           string        // name of the device (interface of pcap file)
	iface            bool          // tells if the packet source is an interface
	snapshotLen      int32         // the maximum size to read for each packet
	promiscuous      bool          // promiscuous mode of the interface
	timeout          time.Duration // time to wait if nothing happens
	// nbParsedPkts     uint64        // the number of parsed packets
	sniffing bool // tells if the package is currently sniffing
)

// Time
var (
	// SourceTime is the clock given by the packet capture
	SourceTime         time.Time
	tickPeriod         time.Duration // time between two data sending (if stat computation)
	last               time.Time     // time of the last data sending
	defaultTimeChannel TimeChannel   // channel sending time (at a given frequency in practice)
	sendTicks          bool          // tells if ticks have to be sent
	remoteTimeChannel  TimeChannel
)

// Packet sniffing/parsing
var (
	handle *pcap.Handle
	parser *gopacket.DecodingLayerParser
	err    error
)

// Events
const (
	// STOP stops a counter
	STOP uint8 = 0
	// GET trigger a snapshot of the counters
	GET uint8 = 1
	// FLUSH trigger a snapshot of the counters and reset them
	FLUSH uint8 = 3
)

// Specific IDs
// const (
// 	// SourceTimeID is the default id of the SOURCE_TIME built-in counter
// 	SourceTimeID int = 1000
// 	// RealTimeID is the default id of the REAL_TIME built-in counter
// 	RealTimeID int = 1001
// 	// PacketsID is the default id of the PACKETS built-in counter
// 	PacketsID int = 1002
// )

func init() {
	// Default configuration
	viper.SetDefault("miner.device", "any")
	viper.SetDefault("miner.snapshot_len", int32(65535))
	viper.SetDefault("miner.promiscuous", true)
	viper.SetDefault("miner.timeout", 30*time.Second)
	// Default configuration
	zerolog.SetGlobalLevel(zerolog.Disabled)
	// Reset all variables
	Zero()
}

// initChannels creates the default event channel and the
// default data channel
func initChannels() {
	defaultDataChannel = make(DataChannel)
	defaultEventChannel = make(EventChannel)
}

// InitLogger initializes a specific logger for the miner package
func InitLogger() {
	minerLogger = log.With().Str("module", "MINER").Logger()
}

// func loadBuiltinCounters() {
// 	for _, ctr := range builtinCounters {
// 		id := LoadFromName(ctr)
// 		// keep the id in the memory
// 		builtinCounterMap[ctr] = id
// 	}
// }

// registerAsBuiltin registers a miner built-in counter to the
// counter subpackage. It adds it to the built-in counter list
// func registerAsBuiltin(name string, sc counters.CounterConstructor) {
// 	counters.Register(name, sc)
// 	builtinCounters = append(builtinCounters, name)
// }

// Zero aims to zero the internal state of the miner. So it removes all
// the loaded counters, reset some variables
// [and read the config file](NOT ANYMORE).
func Zero() error {
	if !IsSniffing() {
		AvailableDevices = GetAvailableDevices()

		initChannels()

		// sniff variables
		sniffing = false

		// time variables
		SourceTime = time.Now()
		// defaultTicker = make(chan time.Time)
		sendTicks = false
		last = SourceTime

		// counter loader
		// builtinCounterMap = make(map[int]string)
		counterMap = make(map[int]counters.BaseCtrInterface)
		counterID = 0 // 0 is never used

		// everything is ok
		minerLogger.Info().Msg("Miner package (re)loaded")
		return nil
	}
	minerLogger.Error().Msg("Cannot reload, sniffing in progress")
	return errors.New("Cannot reload, sniffing in progress")
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

// GetAvailableCounters returns the list of the implemented
// counters (within the 'counters' subpackage)
func GetAvailableCounters() []string {
	names := make([]string, 0, len(counters.AvailableCounters))
	for n := range counters.AvailableCounters {
		names = append(names, n)
	}
	return names
}

//------------------------------------------------------------------------------
// SIDES FUNCTIONS (UNEXPORTED)
//------------------------------------------------------------------------------
func idFromName(ctrname string) int {
	for id, ctr := range counterMap {
		if ctr.Name() == ctrname {
			return id
		}
	}
	return -1
}

func isAlreadyLoaded(ctrname string) bool {
	for _, ctr := range counterMap {
		if ctrname == ctr.Name() {
			return true
		}
	}
	return false
}

// func isBuiltin(ctrname string) bool {
// 	for _, ctr := range builtinCounters {
// 		if ctrname == ctr {
// 			return true
// 		}
// 	}
// 	return false
// }

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

//------------------------------------------------------------------------------
// MAIN
//------------------------------------------------------------------------------

func main() {

}

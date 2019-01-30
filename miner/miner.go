// miner.go

// Package miner aims to read either network interfaces or
// network captures to increment basic counters.
package miner

import "C"

import (
	"errors"
	"fmt"
	"netspot/miner/counters"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type DeviceList []string

var (
	counterMap map[int]counters.BaseCtrInterface // Map id->counter
	counterId  int                               // Id to store counters in counterMap
	mux        sync.RWMutex                      // Locker for the counter map access
	events     chan int                          // to receive events

)

func init() {
	// devices
	AvailableDevices = GetAvailableDevices()

	// Default configuration
	viper.SetDefault("miner.device", GetAvailableDevices()[0])
	viper.SetDefault("miner.snapshot_len", int32(65535))
	viper.SetDefault("miner.promiscuous", true)
	viper.SetDefault("miner.timeout", 30*time.Second)
}

func init() {
	// sniff variables
	sniffing = false
	// stopSniff = make(chan int)
	nbParsedPkts = 0

	// time variables
	SourceTime = time.Now()
	ticker = make(chan time.Time)
	sendTicks = false
	last = SourceTime

	// counter loader
	counterMap = make(map[int]counters.BaseCtrInterface)
	counterId = 0 // 0 is never used

	// events
	events = make(chan int)

	// everything is ok
	// log.Info().Msg("Miner package initialized")
}

// InitConfig initializes the miner package from the config file
func InitConfig() {
	SetDevice(viper.GetString("miner.device"))
	SetSnapshotLen(viper.GetInt32("miner.snapshot_len"))
	SetPromiscuous(viper.GetBool("miner.promiscuous"))
	SetTimeout(viper.GetDuration("miner.timeout"))

	log.Debug().Msg(fmt.Sprint("Available counters: ", counters.GetAvailableCounters()))
	log.Info().Msg("Miner package configured")
}

// RawStatus returns the current status of the miner through a
// basic map. It is designed to a future print.
func RawStatus() map[string]string {
	m := make(map[string]string)
	// m["parsed packets"] = fmt.Sprintf("%d", nbParsedPkts)
	m["promiscuous"] = fmt.Sprintf("%v", promiscuous)
	m["timeout"] = fmt.Sprint(timeout)
	m["device"] = device
	m["snapshot length"] = fmt.Sprintf("%d", snapshotLen)
	return m
}

// Zero aims to zero the internal state of the miner. So it removes all
// the loaded counters, reset some variables and read the config file.
func Zero() error {
	if !IsSniffing() {
		AvailableDevices = GetAvailableDevices()

		SetDevice(viper.GetString("miner.device"))
		SetSnapshotLen(viper.GetInt32("miner.snapshot_len"))
		SetPromiscuous(viper.GetBool("miner.promiscuous"))
		SetTimeout(viper.GetDuration("miner.timeout"))

		// sniff variables
		sniffing = false
		// stopSniff = make(chan int)
		nbParsedPkts = 0

		// time variables
		SourceTime = time.Now()
		ticker = make(chan time.Time)
		sendTicks = false
		last = SourceTime

		// counter loader
		counterMap = make(map[int]counters.BaseCtrInterface)
		counterId = 0 // 0 is never used

		// events
		events = make(chan int)

		// everything is ok
		log.Info().Msg("Miner package reloaded")
		return nil
	} else {
		log.Error().Msg("Cannot reload, sniffing in progress")
		return errors.New("Cannot reload, sniffing in progress")
	}
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

//------------------------------------------------------------------------------
// EXPORTED FUNCTIONS (BASIC TYPES)
//------------------------------------------------------------------------------

//export DisableLogging
func DisableLogging() {
	log.Warn().Msg("Disabling logging")
	zerolog.SetGlobalLevel(zerolog.Disabled)
}

//export SetLogging
func SetLogging(level int) {
	l := zerolog.Level(level)
	zerolog.SetGlobalLevel(l)
	log.Warn().Msgf("Enabling logging (level %s)", l.String())
}

//export IsSniffing
func IsSniffing() bool {
	return sniffing
}

//export GetSourceTime
func GetSourceTime() int64 {
	return SourceTime.UnixNano()
}

//export GetNbParsedPkts
func GetNbParsedPkts() uint64 {
	return nbParsedPkts
}

//export GetNumberOfDevices
func GetNumberOfDevices() int {
	return len(GetAvailableDevices())
}

//export IsDeviceInterface
func IsDeviceInterface() bool {
	return iface
}

//export IsPromiscuous
func IsPromiscuous() bool {
	return promiscuous
}

//export SetPromiscuous
func SetPromiscuous(b bool) int {
	promiscuous = b
	log.Debug().Msgf("Promiscuous set to %v", b)
	return 0
}

//export SetSnapshotLen
func SetSnapshotLen(sl int32) int {
	snapshotLen = sl
	log.Debug().Msgf("Snapshot length set to %d", sl)
	return 0
}

//export StartSniffing
func StartSniffing() {
	if !sniffing {
		log.Info().Msgf("Start sniffing %s", device)
		log.Debug().Msg(fmt.Sprint("Loaded counters: ", GetLoadedCounters()))
		events = make(chan int)
		go Sniff()
		for !sniffing {
			// wait for sniffing
		}
	}

}

//export StartSniffingAndWait
func StartSniffingAndWait() {
	if !sniffing {
		log.Info().Msgf("Start sniffing %s", device)
		log.Debug().Msg(fmt.Sprint("Loaded counters: ", GetLoadedCounters()))
		events = make(chan int)
		Sniff()
	}
	log.Info().Msg("Sniffing stopped")
}

//export StopSniffing
func StopSniffing() {
	if sniffing {
		// send the event 0
		events <- 0
	}
}

//export GetNumberOfLoadedCounters
func GetNumberOfLoadedCounters() int {
	return len(counterMap)
}

//export UnloadAll
func UnloadAll() {
	for k, _ := range counterMap {
		Unload(k)
	}
	// reset the counter
	counterId = 0
}

//export GetCounterValue
func GetCounterValue(id int) uint64 {
	mux.Lock()

	if counterMap[id].IsRunning() {
		// send the signal
		counterMap[id].SigPipe() <- uint8(1)
		// return the value
		defer mux.Unlock()
		return <-counterMap[id].ValPipe()
	} else {
		defer mux.Unlock()
		return counterMap[id].Value()
	}
}

//export Unload
func Unload(id int) {
	log.Debug().Msgf("Unloading counter %s", counterMap[id].Name())
	delete(counterMap, id)
}

//export Reset
func Reset(id int) int {
	ctr, exists := counterMap[id]
	if exists {
		ctr.Reset()
		return 0
	} else {
		return -1
	}
}

//export ResetAll
func ResetAll() {
	mux.Lock()
	for _, ctr := range counterMap {
		ctr.Reset()
	}
	mux.Unlock()
}

//------------------------------------------------------------------------------
// EXPORTED FUNCTIONS (GOLANG ONLY)
//------------------------------------------------------------------------------

func Events() chan int {
	return events
}

// SetTimeout
func SetTimeout(d time.Duration) {
	timeout = d
	log.Debug().Msgf("Timeout set to %s", d)
}

// SetDevice sets the device to listen. It can be either an interface or
// a capture file (ex: .pcap)
func SetDevice(dev string) int {
	if AvailableDevices.contains(dev) {
		device = dev
		iface = true
	} else if fileExists(dev) {
		device = dev
		iface = false
	} else {
		log.Debug().Msgf("Unknown device (%s)", dev)
		return 1
	}
	log.Info().Msgf(`Set device to "%s"`, dev)
	return 0
}

// GetDevice returns the current device (interface name or capture file)
func GetDevice() string {
	return device
}

// LoadFromName loads a new counter and returns its id
func LoadFromName(ctrname string) int {
	ctr := counterFromName(ctrname)
	id, _ := load(ctr)
	if sniffing && id >= 0 {
		mux.Lock()
		startCounter(id)
		mux.Unlock()
	}
	return id
}

// UnloadFromName unloads a counter and return 0 if the operation
// has correctly been done. Returns -1 if the counter does not exist
func UnloadFromName(ctrname string) int {
	id := idFromName(ctrname)
	if id == -1 {
		return -1
	} else {
		Unload(id)
		log.Debug().Msgf("Unloading %s", ctrname)
		return 0
	}
}

// GetLoadedCounters returns a slice of the names of
// the loaded counters
func GetLoadedCounters() []string {
	nbCounters := len(counterMap)
	names := make([]string, 0, nbCounters)
	for _, ctr := range counterMap {
		names = append(names, ctr.Name())
	}
	return names
}

//------------------------------------------------------------------------------
// UNEXPORTED FUNCTIONS
//------------------------------------------------------------------------------

func startAllCounters() error {
	// updateCounterTypes()
	log.Info().Msg("Starting all counters")
	for _, ctr := range counterMap {
		if !ctr.IsRunning() {
			go counters.Run(ctr)
		}
	}
	return nil
}

func startCounter(id int) error {
	ctr, ok := counterMap[id]
	if ok {
		if !ctr.IsRunning() {
			go counters.Run(ctr)
		} else {
			return errors.New(fmt.Sprintf("The counter %s is already running", ctr.Name()))
		}
	} else {
		return errors.New(fmt.Sprintf("The counter %s is not loaded", ctr.Name()))
	}
	return nil
}

func stopAllCounters() error {
	mux.Lock()
	for _, ctr := range counterMap {
		if ctr.IsRunning() {
			ctr.SigPipe() <- 0
		}
	}
	mux.Unlock()
	log.Info().Msg("Stopping all counters")
	return nil
}

// counterFromName returns the BaseCtrInterface related to the
// given name. It returns an error when the desired counter does
// not exist.
func counterFromName(name string) counters.BaseCtrInterface {
	cc, exists := counters.AVAILABLE_COUNTERS[name]
	if exists {
		return cc()
	} else {
		log.Error().Msg("Unknown counter")
		return nil
	}
}

// func counterFromName(name string) counters.BaseCtrInterface {
// 	switch name {
// 	case "IP":
// 		return &counters.IP{
// 			IpCtr:   counters.NewIpCtr(),
// 			Counter: 0}
// 	case "SYN":
// 		return &counters.SYN{
// 			TcpCtr:  counters.NewTcpCtr(),
// 			Counter: 0}
// 	case "ACK":
// 		return &counters.ACK{
// 			TcpCtr:  counters.NewTcpCtr(),
// 			Counter: 0}
// 	case "IP_BYTES":
// 		return &counters.IP_BYTES{
// 			IpCtr:   counters.NewIpCtr(),
// 			Counter: 0}
// 	case "NB_UNIQ_SRC_ADDR":
// 		return &counters.NB_UNIQ_SRC_ADDR{
// 			IpCtr: counters.NewIpCtr(),
// 			Addr:  make(map[string]bool)}
// 	case "NB_UNIQ_DST_ADDR":
// 		return &counters.NB_UNIQ_DST_ADDR{
// 			IpCtr: counters.NewIpCtr(),
// 			Addr:  make(map[string]bool)}
// 	case "ICMP":
// 		return &counters.ICMP{
// 			IcmpCtr: counters.NewIcmpCtr(),
// 			Counter: 0}
// 	default:
// 		log.Error().Msgf("Unknown counter (%s)", name)
// 		return nil
// 	}
// }

func load(ctr counters.BaseCtrInterface) (int, error) {
	if ctr != nil {
		if isAlreadyLoaded(ctr.Name()) {
			msg := fmt.Sprintf("Counter %s already loaded", ctr.Name())
			log.Debug().Msgf("Counter %s already loaded", ctr.Name())
			return -2, errors.New(msg)
		} else {
			counterId = counterId + 1
			counterMap[counterId] = ctr
			log.Debug().Msgf("Loading counter %s", ctr.Name())
			return counterId, nil
		}
	}
	log.Error().Msg("Cannot load null counter")
	return -1, errors.New("Cannot load null counter")
}

//------------------------------------------------------------------------------
// MAIN
//------------------------------------------------------------------------------

func main() {}

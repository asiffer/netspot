// analyzer.go

// Package analyzer is the core of netspot as it controls the miner, the
// stats computations and the logs.
package analyzer

import (
	"errors"
	"fmt"
	"math"
	"netspot/config"
	"netspot/exporter"
	"netspot/miner"
	"netspot/stats"
	"sync"
	"time"

	"github.com/asiffer/gospot"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

//------------------------------------------------------------------------------
// GLOBAL VARIABLES
//------------------------------------------------------------------------------

// package variables to manage stats
var (
	// counterID     = make(map[string]int)                 // map CounterName -> counterID (within miner CounterMap)
	statMap = make(map[string]stats.StatInterface) // map StatId -> Stat
	// statID        = 0                                    // id of the next loaded stat
	statValues = make(map[string]float64) // the last computed values of the statistics
	// counterValues = make(map[string]uint64)              // temp container of the counter values
	period = 0 * time.Second // time between two stat updates (= window size)
)

// mutex
var (
	mux  sync.RWMutex // Locker for the counter map access
	smux sync.RWMutex // Locker for the stat map access
)

// i/o channels
var (
	defaultEventChannel = make(chan int)                // channel to send/receive events
	defaultDataChannel  = make(chan map[string]float64) // channel to sen/receive data
	running             = false                         // to check if the stat are computed
)

var (
	stopChannel = make(chan int, 1)
)

// Events to get data
const (
	// STOP aims to stop the analyzer (which stops the miner)
	STOP int = 0
	// STAT aims to get the stats values
	STAT int = 1
	// PERF aims to get the current miner performances
	PERF int = 2
)

var (
	analyzerLogger zerolog.Logger
)

var err error

// AlreadyLoadedError is raised when a stat is
// loaded twice
type AlreadyLoadedError struct {
	Query string
	Type  string
}

func (a *AlreadyLoadedError) Error() string {
	return fmt.Sprintf("%s %s already loaded", a.Type, a.Query)
}

//------------------------------------------------------------------------------
// INITIALIZATION
//------------------------------------------------------------------------------

// init sets the default configuration
func init() {
	reset()
}

// init or reset all the stats variables
func reset() {
	// counterID = make(map[string]int)
	statMap = make(map[string]stats.StatInterface)
	// statID = 0
	statValues = make(map[string]float64)
	// counterValues = make(map[string]uint64)
	defaultEventChannel = make(chan int)
	running = false
}

// InitConfig load the stats according to
// the config file
func InitConfig() error {
	p, err := config.GetDuration("analyzer.period")
	if err != nil {
		return err
	}
	SetPeriod(p)

	if config.HasKey("analyzer.stats") {
		toLoad, err := config.GetStringList("analyzer.stats")
		if err != nil {
			return err
		}
		for _, s := range toLoad {
			if err := LoadFromName(s); err != nil {
				return err
			}
		}
	}

	analyzerLogger.Debug().Msgf("Available stats: %s", GetAvailableStats())
	analyzerLogger.Info().Msg("Analyzer package configured")
	return nil
}

// InitLogger initialize the sublogger for ANALYZER
func InitLogger() {
	analyzerLogger = log.With().Str("module", "ANALYZER").Logger()
}

//------------------------------------------------------------------------------
// UNEXPORTED FUNCTIONS
//------------------------------------------------------------------------------

// GENERIC ---------------------------------------------------------------------

func find(sl []string, str string) int {
	for i, s := range sl {
		if s == str {
			return i
		}
	}
	return -1
}

// SPECIFIC --------------------------------------------------------------------

func release() {
	running = false
	// stopChannel <- 1
	// Stop()
}

func isLoaded(statname string) bool {
	_, exists := statMap[statname]
	return exists
}

func getcounterValues(ctrvalues map[string]uint64, ctrnames []string) []uint64 {
	values := make([]uint64, len(ctrnames))
	for i, name := range ctrnames {
		values[i] = ctrvalues[name]
	}
	return values
}

func load(stat stats.StatInterface) error {
	if stat == nil {
		return fmt.Errorf("Cannot load null stat")
	}

	if isLoaded(stat.Name()) {
		return &AlreadyLoadedError{Type: "Stat", Query: stat.Name()}
	}
	// load the counters
	for _, ctrname := range stat.Requirement() {
		if err := miner.Load(ctrname); err != nil {
			switch err.(type) {
			case *AlreadyLoadedError:
				analyzerLogger.Debug().Msgf(err.Error())
			default:
				return fmt.Errorf("Error while loading counters of stat %s: %v", stat.Name(), err)
			}
		}
	}
	statMap[stat.Name()] = stat
	analyzerLogger.Debug().Msgf("Loading stat %s", stat.Name())
	return nil

}

// unload
func unload(name string) error {
	var index int
	// check the potential counters to remove
	stat, exists := statMap[name]
	if !exists {
		return fmt.Errorf("Unknown Stat %s", name)
	}
	counters2remove := stat.Requirement()
	// look if these counters are requested by other stats
	for n, stat := range statMap {
		if n != name {
			for _, req := range stat.Requirement() {
				index = find(counters2remove, req)
				if index >= 0 {
					// if a stat still need this counter, we remove it from the list
					counters2remove = append(counters2remove[:index], counters2remove[(index+1):]...)
				}
			}
		}
	}

	// We remove all the useless counters
	for _, ctr := range counters2remove {
		if err := miner.Unload(ctr); err != nil {
			return fmt.Errorf("Error while unloading %s: %v", ctr, err)
		}
	}
	// we remove the stat
	delete(statMap, name)
	return nil
}

//------------------------------------------------------------------------------
// EXPORTED FUNCTIONS
//------------------------------------------------------------------------------

// StatStatus returns the status of the dspot instance monitoring that stat
func StatStatus(s string) (gospot.DSpotStatus, error) {
	if !isLoaded(s) {
		return statMap[s].Status(), nil
	}
	return gospot.DSpotStatus{}, fmt.Errorf("Stat %s is not loaded", s)
}

// RawStatus returns the current status of the analyzer through a
// basic map. It is designed to a future print.
func RawStatus() map[string]string {
	m := make(map[string]string)
	m["period"] = fmt.Sprint(period)
	m["statistics"] = fmt.Sprint(GetLoadedStats())
	return m
}

// GenericStatus returns the current status of the analyzer through a
// basic map. It is designed to JSON marshalling.
func GenericStatus() map[string]interface{} {
	return map[string]interface{}{
		"period":     period,
		"statistics": GetLoadedStats(),
	}
}

// Zero aims to zero the internal state of the analyzer. So it removes all
// the loaded stats, initialize some variables [and read the config file](NOT ANYMORE).
func Zero() error {
	if IsRunning() {
		analyzerLogger.Error().Msg("Cannot reload, monitoring in progress")
		return errors.New("Cannot reload, monitoring in progress")
	}

	// Reset
	reset()

	analyzerLogger.Info().Msg("Analyzer package (re)loaded")
	return nil
}

// SetPeriod sets the duration between two stat computations
func SetPeriod(d time.Duration) {
	if !running {
		period = d
		analyzerLogger.Debug().Msgf("Period set to %s", d)
	} else {
		analyzerLogger.Error().Msgf("Cannot change period while sniffing")
	}
}

// GetPeriod returns the current duration between two stat computations
func GetPeriod() time.Duration {
	return period
}

// GetLoadedStats returns the slice of the names of the loaded statistics
func GetLoadedStats() []string {
	list := make([]string, 0)
	for _, s := range statMap {
		list = append(list, s.Name())
	}
	return list
}

// GetNumberOfLoadedStats returns the number of loaded statistics
func GetNumberOfLoadedStats() int {
	return len(statMap)
}

// GetAvailableStats returns the slice of the names of the available
// statistics
func GetAvailableStats() []string {
	list := make([]string, 0)
	for name := range stats.AvailableStats {
		list = append(list, name)
	}
	return list
}

// LoadFromName loads the statistics corresponding to the given name
// and returns the id where it is internally stored. An error is returned
// when the statistics is unknown.
func LoadFromName(statname string) error {
	stat, err := stats.StatFromName(statname)
	if err != nil {
		return fmt.Errorf("Error while getting statistics %s: %v", statname, err)
	}
	return load(stat)
}

// UnloadFromName removes the statistics, so it will not be monitored.
// It returns 0, nil if the unload is ok, or -1, error otherwise.
func UnloadFromName(statname string) error {
	if isLoaded(statname) {
		return unload(statname)
	}
	return fmt.Errorf("Stat %s is not loaded", statname)
}

// UnloadAll removes all the previously loaded statistics
func UnloadAll() {
	for i := range statMap {
		delete(statMap, i)
	}
	miner.UnloadAll()
}

// IsRunning checks whether the statistics are currently computed
func IsRunning() bool {
	return running
}

// Stop stops the analysis and check that it is well stopped
func Stop() error {
	if running {
		analyzerLogger.Debug().Msg("Sending STOP message")
		// send the STOP msg
		defaultEventChannel <- STOP
		// check that it did stop
		timeout2 := time.After(2 * time.Second)
		timeout5 := time.After(5 * time.Second)
		for {
			select {
			case <-stopChannel:
				// good
				return nil
			case <-timeout2:
				// warning
				analyzerLogger.Warn().Msg("Analyzer is still running")
			case <-timeout5:
				// timeout
				analyzerLogger.Warn().Msg("Timeout reached")
				return errors.New("The analyzer is not well stopped")
			}
		}
	}
	return errors.New("The analyzer is not running")
}

// StatValues return a current snapshot of the stat values (and their thresholds)
func StatValues() map[string]float64 {
	if running {
		defaultEventChannel <- STAT
		return <-defaultDataChannel
	}
	return nil
}

// Start starts the analysis
// func Start() error {
func Start() error {
	if len(GetLoadedStats()) == 0 {
		return errors.New("No stats loaded")
	}
	analyzerLogger.Info().Msg("Starting stats computation")
	analyzerLogger.Debug().Msg(fmt.Sprint("Loaded stats: ", GetLoadedStats()))
	defaultEventChannel, defaultDataChannel = GoRun()
	return nil
}

// StartAndWait starts the analysis. It will stop only when no packets
// have to be processed (ex: pcap file)
func StartAndWait() error {
	if len(GetLoadedStats()) == 0 {
		return errors.New("No stats loaded")
	}
	analyzerLogger.Info().Msg("Starting stats computation")
	analyzerLogger.Debug().Msg(fmt.Sprint("Loaded stats: ", GetLoadedStats()))
	return Run()
}

//------------------------------------------------------------------------------
// PARAMOUNT BUT UNEXPORTED FUNCTIONS
//------------------------------------------------------------------------------

func checkSpotOutput(stat stats.StatInterface, val float64, res int, t time.Time) {
	var sa exporter.SpotAlert
	if res == 1 {
		sa = exporter.SpotAlert{
			Status:      "UP_ALERT",
			Stat:        stat.Name(),
			Value:       val,
			Code:        res,
			Probability: stat.UpProbability(val),
		}
	} else if res == -1 {
		sa = exporter.SpotAlert{
			Status:      "DOWN_ALERT",
			Stat:        stat.Name(),
			Value:       val,
			Code:        res,
			Probability: stat.DownProbability(val),
		}
	} else {
		// do nothing
		return
	}

	if err := exporter.Warn(t, &sa); err != nil {
		analyzerLogger.Error().Msgf("Error while sending alarms: %v", err)
	}
}

func analyze(m map[string]uint64) {
	curtime := miner.SourceTime

	// the locker is needed in case of a snapshot
	// smux.Lock()
	for _, stat := range statMap {
		name := stat.Name()

		downTh, upTh := stat.GetThresholds()

		// if upTh is NaN, it means that up data are not monitored or
		// the calibration has not finished
		if !math.IsNaN(upTh) {
			statValues[name+"_UP"] = upTh
		}

		// if downTh is NaN, it means that down data are not monitored or
		// the calibration has not finished
		if !math.IsNaN(downTh) {
			statValues[name+"_DOWN"] = downTh
		}

		// compute the statistics
		ctrValues := getcounterValues(m, stat.Requirement())
		statValue := stat.Compute(ctrValues)

		// check if the computed statistics is a number
		if !math.IsNaN(statValue) {
			// feed DSpot
			res := stat.Update(statValue)
			// check alert
			checkSpotOutput(stat, statValue, res, curtime)
		}
		// store stats data
		statValues[name] = statValue

	}
	// smux.Unlock()
	// send data to the exporter
	if err := exporter.Write(curtime, statValues); err != nil {
		analyzerLogger.Error().Msgf("Error while exporting values: %v", err)
	}
}

func run(eventChannel chan int, dataChannel chan map[string]float64) error {
	// start the exporter
	if err := exporter.Start("test"); err != nil {
		return err
	}
	// defer close
	defer exporter.Close()
	// display basic information
	analyzerLogger.Info().Msg("Start running")
	// set the running state
	running = true
	// sniff
	minerData, err := miner.StartAndYield(period)
	if err != nil {
		running = false
		return fmt.Errorf("Error while starting the miner: %v", err)
	}
	// loop
	for {
		select {
		case e := <-eventChannel:
			switch e {
			case STOP: // stop order
				analyzerLogger.Debug().Msg("Receiving STOP message")
				// stop the miner
				if err := miner.Stop(); err != nil {
					analyzerLogger.Error().Msgf("Error while stopping miner: %v", err)
				}
				// minerEvent <- miner.STOP
				// release (put running to false)
				running = false
				stopChannel <- 1
				analyzerLogger.Info().Msg("Stopping stats computation (controller)")
				return nil
			case STAT: // send data
				smux.Lock()
				snapshot := make(map[string]float64)
				for s, v := range statValues {
					snapshot[s] = v
				}
				smux.Unlock()
				dataChannel <- snapshot
			}
		case m := <-minerData:
			if m == nil {
				// release
				running = false
				analyzerLogger.Info().Msg("Stopping stats computation (miner)")
				return nil
			}

			// analyze the stats values (feed dspot, log data/thresholds)
			analyze(m)

		}
	}
}

// Run open the device to listen
func Run() error {
	return run(defaultEventChannel, defaultDataChannel)
}

// GoRun starts the analyzer and return two communication channels
// event and data
func GoRun() (chan int, chan map[string]float64) {
	eventChannel := make(chan int)
	dataChannel := make(chan map[string]float64)
	go run(eventChannel, dataChannel)
	return eventChannel, dataChannel
}

//------------------------------------------------------------------------------
// MAIN
//------------------------------------------------------------------------------

func main() {
	// nothing
}

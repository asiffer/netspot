// analyzer.go

// Package analyzer is the core of netspot as it controls the miner, the
// stats computations and the logs.
package analyzer

import (
	"errors"
	"fmt"
	"math"
	"netspot/influxdb"
	"netspot/miner"
	"netspot/stats"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/asiffer/gospot"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

//------------------------------------------------------------------------------
// GLOBAL VARIABLES
//------------------------------------------------------------------------------

var (
	counterID           map[string]int              // map CounterName -> counterID (within miner CounterMap)
	statMap             map[int]stats.StatInterface // map StatId -> Stat
	statID              int                         // id of the next loaded stat
	statValues          map[string]float64          // the last computed values of the statistics
	counterValues       map[string]uint64           // temp container of the counter values
	period              time.Duration               // time between two stat updates (= window size)
	mux                 sync.RWMutex                // Locker for the counter map access
	smux                sync.RWMutex                // Locker for the stat map access
	defaultEventChannel chan int                    // channel to send/receive events
	defaultDataChannel  chan map[string]float64     // channel to sen/receive data
	running             bool                        // to check if the stat are computed
)

var (
	logDataToFile       bool           // if data/thresholds are logged to file
	logDataToInfluxDB   bool           // if data/thresholds are logged to influxdb
	rawDataLogger       zerolog.Logger // log raw statistics
	thresholdLogger     zerolog.Logger // log raw statistics thresholds
	anomalyLogger       zerolog.Logger // log anomalies
	rawDataOutputFile   string         // the path of the file containing data
	thresholdOutputFile string         // the path of the file containing thresholds
	anomalyOutputFile   string         // the path of the file containing the anomalies
	seriesName          string         // the name of the influxdb series
	outputDir           string         // directory where raw data are stored (2 files)
)

const (
	rawDataFileNameFormat   = "netspot_raw_%s.json"       // format to name the file where data are logged
	thresholdFileNameFormat = "netspot_threshold_%s.json" // format to name the file where thresholds are logged
	anomalyFileNameFormat   = "netspot_anomaly_%s.json"   // format to name the file where anomalies are logged
)

var err error

//------------------------------------------------------------------------------
// INITIALIZATION
//------------------------------------------------------------------------------

// init sets the default configuration
func init() {
	// default config
	viper.SetDefault("analyzer.period", 2*time.Second)
	viper.SetDefault("analyzer.datalog.influxdb", false)
	viper.SetDefault("analyzer.datalog.file", true)
	viper.SetDefault("analyzer.datalog.output_dir", "/tmp")
	viper.SetDefault("analyzer.stats", []string{})
}

// init initializes package global variables
func init() {
	counterID = make(map[string]int)
	statMap = make(map[int]stats.StatInterface)
	statID = 0
	statValues = make(map[string]float64)
	counterValues = make(map[string]uint64)
	defaultEventChannel = make(chan int)
	defaultDataChannel = make(chan map[string]float64)
	running = false
}

// InitConfig initialize the loggers and load the stats according to
// the config file
func InitConfig() {
	// settings
	SetPeriod(viper.GetDuration("analyzer.period"))
	SetOutputDir(viper.GetString("analyzer.datalog.output_dir"))
	logDataToFile = viper.GetBool("analyzer.datalog.file")
	logDataToInfluxDB = viper.GetBool("analyzer.datalog.influxdb")

	if logDataToInfluxDB {
		influxdb.InitConfig()
	}

	for _, s := range viper.GetStringSlice("analyzer.stats") {
		LoadFromName(s)
	}

	log.Info().Msgf("Available stats: %s", GetAvailableStats())
	log.Info().Msg("Analyzer package configured")
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

// autoSetSeriesName automaticcaly sets the name of the series for the
// incoming sniff
func autoSetSeriesName() {
	// Interface: device_time
	if miner.IsDeviceInterface() {
		seriesName = miner.GetDevice() + "_" + seriesNameFromCurrentTime()
	} else {
		// PCAP: file
		seriesName = path.Base(miner.GetDevice())
	}
}

func seriesNameFromCurrentTime() string {
	t := strings.Replace(time.Now().Format(time.Stamp), "_", "0", -1)
	t = strings.Replace(t, " ", "_", -1)
	return t
}

func release() {
	running = false
	// close(defaultEventChannel)
}

// func updateAndReset() {
// 	// m, _ := miner.SnapAndReset()
// 	m := miner.Snapshot(true, nil, nil)
// 	// if err != nil {
// 	for name, val := range m {
// 		counterValues[name] = val
// 		log.Debug().Msg(fmt.Sprint(counterValues))
// 	}

// }
// log.Error().Msg(err.Error())
// mux.Lock()
// var val uint64
// var err error
// for ctrname, ctrid := range counterID {
// 	val, err = miner.GetCounterValue(ctrid)
// 	if err != nil {
// 		log.Error().Msgf("The ID of %s (%d) is wrong for the miner", ctrname, ctrid)
// 	} else {
// 		counterValues[ctrname] = val
// 		miner.Reset(ctrid)
// 	}
// }
// log.Debug().Msg(fmt.Sprint(counterValues))
// mux.Unlock()
// }

func isLoaded(statname string) int {
	for i, stat := range statMap {
		if statname == stat.Name() {
			return i
		}
	}
	return -1
}

func getcounterValues(ctrnames []string) []uint64 {
	values := make([]uint64, len(ctrnames))
	for i, name := range ctrnames {
		values[i] = counterValues[name]
	}
	return values
}

func compute(id int) float64 {
	counters := statMap[id].Requirement()
	values := getcounterValues(counters)
	return statMap[id].Compute(values)
}

func load(stat stats.StatInterface) (int, error) {
	var id int
	if stat != nil {
		if isLoaded(stat.Name()) > 0 {
			log.Debug().Msgf("Stat %s already loaded", stat.Name())
			msg := fmt.Sprintf("Stat %s already loaded", stat.Name())
			return -2, errors.New(msg)
		}
		// load the counters
		for _, ctrname := range stat.Requirement() {
			id = miner.LoadFromName(ctrname)
			if id > 0 {
				counterID[ctrname] = id
			}
		}
		// increment the stat container
		statID = statID + 1
		statMap[statID] = stat
		log.Debug().Msgf("Loading stat %s", stat.Name())
		return statID, nil
	}
	return -1, errors.New("Cannot load null stat")
}

// unload
func unload(id int) (int, error) {
	var index int
	// check the potential counters to remove
	stat, exists := statMap[id]
	if !exists {
		msg := fmt.Sprintf("Unknown Stat id %d", id)
		return -1, errors.New(msg)
	}
	counters2remove := stat.Requirement()
	// look if these counters are requested by other stats
	for i, stat := range statMap {
		if i != id {
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
		delete(counterID, ctr)
		delete(counterValues, ctr)
		miner.UnloadFromName(ctr)
	}
	// we remove the stat
	delete(statMap, id)
	return 0, nil
}

//------------------------------------------------------------------------------
// EXPORTED FUNCTIONS
//------------------------------------------------------------------------------

// StatStatus returns the status of the dspot instance monitoring that stat
func StatStatus(s string) (gospot.DSpotStatus, error) {
	i := isLoaded(s)
	if i >= 0 {
		return statMap[i].DSpot().Status(), nil
	}
	msg := fmt.Sprintf("%s statistics is not loaded", s)
	return gospot.DSpotStatus{}, errors.New(msg)
}

// RawStatus returns the current status of the analyzer through a
// basic map. It is designed to a future print.
func RawStatus() map[string]string {
	m := make(map[string]string)
	m["period"] = fmt.Sprint(period)
	if logDataToInfluxDB {
		m["influxdb logging"] = fmt.Sprintf("%v (%s)", logDataToInfluxDB, influxdb.GetAddress())
	} else {
		m["influxdb logging"] = "no"
	}

	if logDataToFile {
		m["file logging"] = fmt.Sprintf("%v (%s)", logDataToFile, outputDir)
	} else {
		m["file logging"] = "no"
	}
	m["statistics"] = fmt.Sprint(GetLoadedStats())
	return m
}

// InitDataLogging creates new loggers (file/influxdb). Normally, it is called when
// start to run.
func InitDataLogging() {
	// create the name of the new incoming series
	autoSetSeriesName()

	if logDataToFile {
		// Data logger
		p := filepath.Join(outputDir, fmt.Sprintf(rawDataFileNameFormat, seriesName))
		f, err := os.Create(p)
		if err != nil {
			log.Fatal().Msgf("Error while creating raw data log file (%s)", p)
		}
		rawDataLogger = zerolog.New(f).With().Logger()
		rawDataOutputFile = f.Name()

		// Thresholds logger
		p = filepath.Join(outputDir, fmt.Sprintf(thresholdFileNameFormat, seriesName))
		f, err = os.Create(p)
		if err != nil {
			log.Fatal().Msgf("Error while creating threshold log file (%s)", p)
		}
		thresholdLogger = zerolog.New(f).With().Logger()
		thresholdOutputFile = f.Name()

		// Anomalies logger
		p = filepath.Join(outputDir, fmt.Sprintf(anomalyFileNameFormat, seriesName))
		f, err = os.Create(p)
		if err != nil {
			log.Fatal().Msgf("Error while creating anomaly log file (%s)", p)
		}
		anomalyLogger = zerolog.New(f).With().Logger()
		anomalyOutputFile = f.Name()
	} else {
		rawDataLogger = zerolog.New(nil).With().Logger()
		thresholdLogger = zerolog.New(nil).With().Logger()
		anomalyLogger = zerolog.New(nil).With().Logger()
	}
}

// Zero aims to zero the internal state of the miner. So it removes all
// the loaded stats, initialize some variables and read the config file.
// In particular it creates a new series name, so the data logs will be placed in
// new files/db.series
func Zero() error {
	if !IsRunning() {
		// SetPeriod(viper.GetDuration("analyzer.period"))
		// stats := viper.GetStringSlice("an")
		// // period = viper.GetDuration("analyzer.period")
		// logDataToFile = viper.GetBool("analyzer.datalog.file")
		// logDataToInfluxDB = viper.GetBool("analyzer.datalog.influxdb")

		// package variables
		counterID = make(map[string]int)
		statMap = make(map[int]stats.StatInterface)
		statID = 0
		statValues = make(map[string]float64)
		counterValues = make(map[string]uint64)
		defaultEventChannel = make(chan int)
		running = false

		InitConfig()
		log.Info().Msg("Analyzer package reloaded")
		return nil
	}
	log.Error().Msg("Cannot reload, monitoring in progress")
	return errors.New("Cannot reload, monitoring in progress")
}

// Config return the configuration of the analyzer/miner
func Config() map[string]interface{} {
	return viper.AllSettings()
}

// DisableLogging disable the log output. Warning! It disables the log
// for all the modules using zerolog
func DisableLogging() {
	log.Info().Msg("Disabling logging")
	zerolog.SetGlobalLevel(zerolog.Disabled)
}

// SetLogging set the minimum level of the output logs.
// - panic (zerolog.PanicLevel, 5)
// - fatal (zerolog.FatalLevel, 4)
// - error (zerolog.ErrorLevel, 3)
// - warn (zerolog.WarnLevel, 2)
// - info (zerolog.InfoLevel, 1)
// - debug (zerolog.DebugLevel, 0)
func SetLogging(level int) {
	l := zerolog.Level(level)
	zerolog.SetGlobalLevel(l)
	log.Info().Msgf("Enabling logging (level %s)", l.String())
}

// SetOutputDir change the directory where the raw stats are saved (and thresholds)
func SetOutputDir(dir string) {
	if !running {
		absPath, err := filepath.Abs(dir)
		file, err := os.Open(absPath)
		f, err := file.Stat()
		switch {
		case err != nil:
			// handle the error and return
			log.Error().Msgf("Error while changing data log directory (%s)", err)
			return
		case f.IsDir():
			// it's a directory
			outputDir = absPath
		default:
			// it's not a directory (but a file)
			outputDir = filepath.Dir(absPath)
		}
	} else {
		log.Error().Msgf("Cannot change output directory while sniffing")
	}
}

// SetPeriod sets the duration between two stat computations
func SetPeriod(d time.Duration) {
	if !running {
		period = d
		log.Debug().Msgf("Period set to %s", d)
	} else {
		log.Error().Msgf("Cannot change period while sniffing")
	}
}

// SetSeriesName sets the name of the series within InfluxDB
func SetSeriesName(s string) {
	if !running {
		seriesName = s
		log.Debug().Msgf(`Series name set to "%s"`, s)
	} else {
		log.Error().Msgf("Cannot change series name while sniffing")
	}
}

// GetPeriod returns the current duration between two stat computations
func GetPeriod() time.Duration {
	return period
}

// GetThresholdOutputFile returns the file where the computed thresholds will be logged
func GetThresholdOutputFile() string {
	// return thresholdOutputFile
	return filepath.Join(outputDir, fmt.Sprintf(thresholdFileNameFormat, seriesName))
}

// GetRawDataOutputFile returns the file where the raw statistics will be logged
func GetRawDataOutputFile() string {
	// return rawDataOutputFile
	return filepath.Join(outputDir, fmt.Sprintf(rawDataFileNameFormat, seriesName))
}

// GetAnomalyOutputFile returns the file where the anomalies will be logged
func GetAnomalyOutputFile() string {
	return filepath.Join(outputDir, fmt.Sprintf(anomalyFileNameFormat, seriesName))
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
func LoadFromName(statname string) (int, error) {
	stat, err := stats.StatFromName(statname)
	if err != nil {
		return -1, err
	}
	id, err := load(stat)
	return id, err
}

// LoadFromNameWithCustomConfig loads the statistics corresponding to the given name
// and returns the id where it is internally stored. An error is returned
// when the statistics is unknown. An additional Map parameter is given so as to
// change the DSpot attributes.
func LoadFromNameWithCustomConfig(statname string, config map[string]interface{}) (int, error) {
	stat, err := stats.StatFromNameWithCustomConfig(statname, config)
	if err != nil {
		return -1, err
	}
	id, err := load(stat)
	return id, err
}

// UnloadFromName removes the statistics, so it will not be monitored.
// It returns 0, nil if the unload is ok, or -1, error otherwise.
func UnloadFromName(statname string) (int, error) {
	id := isLoaded(statname)
	if id > 0 {
		ret, err := unload(id)
		return ret, err
	}
	msg := fmt.Sprintf("%s statistics is not loaded", statname)
	return -1, errors.New(msg)
}

// UnloadAll removes all the previously loaded statistics
func UnloadAll() {
	for i := range statMap {
		delete(statMap, i)
	}
	for i := range counterValues {
		delete(counterValues, i)
	}
	for i := range counterID {
		delete(counterID, i)
	}
	miner.UnloadAll()
	statID = 0
}

// IsRunning checks whether the statistics are currently computed
func IsRunning() bool {
	return running
}

// StopStats stops the analysis
func StopStats() {
	if running {
		defaultEventChannel <- 0
		log.Info().Msg("Stopping stats computation")
	}
}

// StatValues return a current snapshot of the stat values (and their thresholds)
func StatValues() map[string]float64 {
	if running {
		defaultEventChannel <- 1
		return <-defaultDataChannel
	}
	return nil
}

// StartStats starts the analysis
// func StartStats() error {
func StartStats() error {
	log.Info().Msg("Starting stats computation")
	log.Debug().Msg(fmt.Sprint("Loaded stats: ", GetLoadedStats()))
	defaultEventChannel, defaultDataChannel = GoRun()
	return nil
}

// StartStatsAndWait starts the analysis. It will stop only when no packets
// have to be processed (ex: pcap file)
func StartStatsAndWait() error {
	log.Info().Msg("Starting stats computation")
	log.Debug().Msg(fmt.Sprint("Loaded stats: ", GetLoadedStats()))
	Run()
	return nil
}

//------------------------------------------------------------------------------
// PARAMOUNT BUT UNEXPORTED FUNCTIONS
//------------------------------------------------------------------------------

// checkSpotOutput analyze the output provided by (D)SPOT.
// id is the stat identifier in statMap
// val is the stat value
// res is the (D)Spot output
// t is the current time given by the miner (miner.SourceTime)
func checkSpotOutput(id int, val float64, res int, t time.Time) {
	if res == 1 {
		alog := anomalyLogger.Warn().Time("time", t)
		alog.Str("Status", "UP_ALERT").Str("Stat", statMap[id].Name()).
			Float64("Value", val).Int("Spot", res).
			Float64("Probability", statMap[id].DSpot().UpProbability(val)).Msg("Alarm!")
		// log.Warn().Str("Status", "UP_ALERT").
		// 	Str("Stat", statMap[id].Name()).
		// 	Float64("Value", val).Int("Spot", res).
		// 	Float64("Probability", statMap[id].DSpot().UpProbability(val)).Msg("Alarm!")
	} else if res == -1 {
		alog := anomalyLogger.Warn().Time("time", t)
		alog.Str("Status", "DOWN_ALERT").
			Str("Stat", statMap[id].Name()).
			Float64("Value", val).Int("Spot", res).
			Float64("Probability", statMap[id].DSpot().DownProbability(val)).Msg("Alarm!")
		// log.Warn().Str("Status", "DOWN_ALERT").
		// 	Str("Stat", statMap[id].Name()).
		// 	Float64("Value", val).Int("Spot", res).
		// 	Float64("Probability", statMap[id].DSpot().DownProbability(val)).Msg("Alarm!")
	}
}

// analyze loops over the computed stats and send it to DSpot instances
// The stat values and the computed thresholds are logged (file and/or influxdb)
func analyze() {
	var val float64
	var upTh, downTh float64
	var res int
	var name string
	curtime := miner.SourceTime

	dlog := rawDataLogger.Log().Time("time", curtime)
	tlog := thresholdLogger.Log().Time("time", curtime)

	// the locker is needed in case of a snapshot
	smux.Lock()

	for id, stat := range statMap {
		name = stat.Name()

		// log thresholds
		upTh = stat.DSpot().GetUpperThreshold()
		downTh = stat.DSpot().GetLowerThreshold()

		fmt.Println(upTh)
		// if upTh is NaN, it means that up data are not monitored or
		// the calibration has not finished
		if !math.IsNaN(upTh) {
			tlog.Float64(name+"_UP", upTh)
			statValues[name+"_UP"] = upTh
		}

		// if downTh is NaN, it means that down data are not monitored or
		// the calibration has not finished
		if !math.IsNaN(downTh) {
			tlog.Float64(name+"_DOWN", downTh)
			statValues[name+"_DOWN"] = downTh
		}

		// compute the statistics
		val = stat.Compute(getcounterValues(stat.Requirement()))

		// check if the computed statistics is a number
		if !math.IsNaN(val) {
			// feed DSpot
			res = stat.Update(val)
			// check alert
			checkSpotOutput(id, val, res, curtime)
		}

		// log data
		dlog.Float64(name, val)
		statValues[name] = val

	}
	smux.Unlock()
	log.Debug().Msg(fmt.Sprint(statValues))
	dlog.Msg("")
	tlog.Msg("")
	// if data have to be sent to InfluxDB
	if logDataToInfluxDB {
		influxdb.PushRecord(statValues, seriesName, curtime)
	}
}

// Run open the device to listen
func Run() {
	// initialize files/influxdb to log data and thresholds
	InitDataLogging()
	log.Info().Msg("Start running")
	log.Info().Msgf("Raw data are logged to %s", rawDataOutputFile)
	log.Info().Msgf("Thresholds are logged to %s", thresholdOutputFile)
	log.Info().Msgf("Anomalies are logged to %s", anomalyOutputFile)

	// number of records
	records := 0
	// set the running state
	running = true
	// sniff
	minerEvent, minerData := miner.GoSniffAndYieldChannel(period)
	// loop
	for {
		log.Debug().
			Str("Function", "Sniff").
			Int("minerEvent", len(minerEvent)).
			Int("minerData", len(minerData)).
			Msg("")

		select {
		case e := <-defaultEventChannel:
			if e == 0 { // stop order
				release()
				// stop the miner
				minerEvent <- miner.STOP
				log.Info().Msg("Stopping stats computation (controller)")
				return
			}
		case m := <-minerData:
			if m != nil {
				// retrieve the counter values
				for name, id := range counterID {
					counterValues[name] = m[id]
					log.Debug().Msg(fmt.Sprint(counterValues))
				}
				records++
				// analyze the stats values (feed dspot, log data/thresholds)
				analyze()
			} else {
				release()
				log.Warn().Msg("Stopping stats computation (miner)")
				log.Info().Msgf("Number of records: %d", records)
				return
			}
		}
	}
}

// GoRun open the device to listen
func GoRun() (chan int, chan map[string]float64) {
	eventChannel := make(chan int)
	dataChannel := make(chan map[string]float64)

	go func() {
		// initialize files/influxdb to log data and thresholds
		InitDataLogging()
		log.Info().Msg("Start running")
		log.Info().Msgf("Raw data are logged to %s", rawDataOutputFile)
		log.Info().Msgf("Thresholds are logged to %s", thresholdOutputFile)
		log.Info().Msgf("Anomalies are logged to %s", anomalyOutputFile)
		// set the running state
		running = true
		// sniff
		minerEvent, minerData := miner.GoSniffAndYieldChannel(period)
		// loop
		for {
			log.Debug().
				Str("Function", "Sniff").
				Int("minerEvent", len(minerEvent)).
				Int("minerData", len(minerData)).
				Msg("")

			select {
			case e := <-eventChannel:
				switch e {
				case 0: // stop order
					log.Info().Msg("Receiving STOP message")
					// stop the miner
					minerEvent <- miner.STOP
					// release all
					// release
					running = false
					// close(eventChannel)
					//
					log.Info().Msg("Stopping stats computation (controller)")
					return
				case 1: // send data
					smux.Lock()
					snapshot := make(map[string]float64)
					for s, v := range statValues {
						snapshot[s] = v
					}
					smux.Unlock()
					fmt.Println(snapshot)
					dataChannel <- snapshot
				}
			case m := <-minerData:
				if m != nil {
					// retrieve the counter values
					for name, id := range counterID {
						counterValues[name] = m[id]
						log.Debug().Msg(fmt.Sprint(counterValues))
					}
					// analyze the stats values (feed dspot, log data/thresholds)
					analyze()
				} else {
					// release
					running = false
					// close(eventChannel)
					//
					log.Warn().Msg("Stopping stats computation (miner)")
					return
				}
			}
		}
	}()
	return eventChannel, dataChannel
}

// // GoRun open the device to listen
// func GoRun() (chan int, chan map[string]float64) {
// 	eventChannel := make(chan int)
// 	dataChannel := make(chan map[string]float64)

// 	go func() {
// 		// initialize files/influxdb to log data and thresholds
// 		InitDataLogging()
// 		log.Info().Msg("Start running")
// 		log.Info().Msgf("Raw data are logged to %s", rawDataOutputFile)
// 		log.Info().Msgf("Thresholds are logged to %s", thresholdOutputFile)
// 		log.Info().Msgf("Anomalies are logged to %s", anomalyOutputFile)

// 		// set the tick period to the miner
// 		// miner.SetTickPeriod(period)
// 		// set the running state
// 		running = true
// 		// sniff
// 		minerEvent, minerData, minerTime := miner.GoSniff()
// 		// loop
// 		for {
// 			log.Debug().
// 				Str("Function", "GoSniff").
// 				Int("eventChannel", len(eventChannel)).
// 				Int("dataChannel", len(dataChannel)).
// 				Int("minerTime", len(minerTime)).
// 				Msg("")
// 			select {
// 			case e := <-eventChannel:
// 				switch e {
// 				case 0: // stop order
// 					log.Info().Msg("Receiving STOP message")
// 					// stop the miner
// 					minerEvent <- miner.STOP
// 					// release all
// 					release()
// 					log.Info().Msg("Stopping stats computation (controller)")
// 					return
// 				case 1: // send data
// 					smux.Lock()
// 					snapshot := make(map[string]float64)
// 					for s, v := range statValues {
// 						snapshot[s] = v
// 					}
// 					smux.Unlock()
// 					fmt.Println(snapshot)
// 					dataChannel <- snapshot
// 				}
// 			case <-minerTime:
// 				if !miner.IsSniffing() {
// 					release()
// 					log.Warn().Msg("Stopping stats computation (miner)")
// 					return
// 				}
// 				// get the stats values (call miner counters etc.)
// 				// counters are reset
// 				m := miner.Snapshot(true, minerEvent, minerData)
// 				for name, val := range m {
// 					counterValues[name] = val
// 					log.Debug().Msg(fmt.Sprint(counterValues))
// 				}
// 				// analyze the stats values (feed dspot, log data/thresholds/anomalies)
// 				analyze()
// 			}
// 		}
// 	}()
// 	return eventChannel, dataChannel
// }

// // run open the device to listen
// func run() {
// 	var ticker <-chan time.Time
// 	var e int

// 	// initialize files/influxdb to log data and thresholds
// 	InitDataLogging()

// 	// Define the clock
// 	if miner.IsDeviceInterface() {
// 		//seriesName = miner.GetDevice() + "_" + seriesNameFromCurrentTime()
// 		// live ticker
// 		ticker = time.Tick(period)
// 	} else {
// 		// the series name is then the name of the file
// 		//seriesName = path.Base(miner.GetDevice())
// 		// the timestamps of the capture define the clock
// 		ticker = miner.Tick(period)
// 	}

// 	// set the running state
// 	running = true

// 	// loop
// 	for {
// 		select {
// 		case e = <-events:
// 			if e == 0 { // stop order
// 				release()
// 				log.Info().Msg("Stopping stats computation (controller)")
// 				return
// 			}
// 		case <-ticker:
// 			if !miner.IsSniffing() {
// 				release()
// 				log.Warn().Msg("Stopping stats computation (miner)")
// 				return
// 			}
// 			// get the stats values (call miner counters etc.)
// 			updateAndReset()
// 			// analyze the stats values (feed dspot, log data/thresholds)
// 			analyze()
// 		}
// 	}
// }

//------------------------------------------------------------------------------
// MAIN
//------------------------------------------------------------------------------

func main() {
	// nothing
}

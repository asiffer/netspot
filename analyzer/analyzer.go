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

// Formats of the log files
const (
	rawDataFileNameFormat   = "netspot_raw_%s.json"       // format to name the file where data are logged
	thresholdFileNameFormat = "netspot_threshold_%s.json" // format to name the file where thresholds are logged
	anomalyFileNameFormat   = "netspot_anomaly_%s.json"   // format to name the file where anomalies are logged
)

// Events to get data
const (
	// STAT aims to get the stats values
	STAT int = 1
	// PERF aims to get the current miner performances
	PERF int = 2
)

var (
	analyzerLogger zerolog.Logger // Logger of the module
)

var err error

//------------------------------------------------------------------------------
// INITIALIZATION
//------------------------------------------------------------------------------

// init sets the default configuration
func init() {
	// default config
	viper.SetDefault("analyzer.period", 2*time.Second)
	viper.SetDefault("data.influxdb", false)
	viper.SetDefault("data.file", true)
	viper.SetDefault("data.output_dir", "/tmp")
	viper.SetDefault("analyzer.stats", []string{})
	// Reset all variables
	Zero()

}

// InitConfig initialize the loggers and load the stats according to
// the config file
func InitConfig() {
	// settings
	SetPeriod(viper.GetDuration("analyzer.period"))
	SetOutputDir(viper.GetString("data.output_dir"))
	logDataToFile = viper.GetBool("data.file")
	logDataToInfluxDB = viper.GetBool("data.influxdb")

	if logDataToInfluxDB {
		influxdb.InitConfig()
	}

	for _, s := range viper.GetStringSlice("analyzer.stats") {
		LoadFromName(s)
	}

	analyzerLogger.Debug().Msgf("Available stats: %s", GetAvailableStats())
	analyzerLogger.Info().Msg("Analyzer package configured")
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

// autoSetSeriesName automatically sets the name of the series for the
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
			analyzerLogger.Debug().Msgf("Stat %s already loaded", stat.Name())
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
		analyzerLogger.Debug().Msgf("Loading stat %s", stat.Name())
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
	// fmt.Println(counters2remove)
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
	m["influxdb"] = fmt.Sprintf("%t", logDataToInfluxDB)
	m["file"] = fmt.Sprintf("%t", logDataToFile)
	m["output"] = outputDir
	m["statistics"] = fmt.Sprint(GetLoadedStats())
	return m
}

// GenericStatus returns the current status of the analyzer through a
// basic map. It is designed to JSON marshalling.
func GenericStatus() map[string]interface{} {
	return map[string]interface{}{
		"period":     period,
		"influxdb":   logDataToInfluxDB,
		"file":       logDataToFile,
		"output":     outputDir,
		"statistics": GetLoadedStats(),
	}
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
			analyzerLogger.Fatal().Msgf("Error while creating raw data log file (%s)", p)
		}
		rawDataLogger = zerolog.New(f).With().Logger()
		rawDataOutputFile = f.Name()

		// Thresholds logger
		p = filepath.Join(outputDir, fmt.Sprintf(thresholdFileNameFormat, seriesName))
		f, err = os.Create(p)
		if err != nil {
			analyzerLogger.Fatal().Msgf("Error while creating threshold log file (%s)", p)
		}
		thresholdLogger = zerolog.New(f).With().Logger()
		thresholdOutputFile = f.Name()

		// Anomalies logger
		p = filepath.Join(outputDir, fmt.Sprintf(anomalyFileNameFormat, seriesName))
		f, err = os.Create(p)
		if err != nil {
			analyzerLogger.Fatal().Msgf("Error while creating anomaly log file (%s)", p)
		}
		anomalyLogger = zerolog.New(f).With().Logger()
		anomalyOutputFile = f.Name()
	} else {
		rawDataLogger = zerolog.New(nil).With().Logger()
		thresholdLogger = zerolog.New(nil).With().Logger()
		anomalyLogger = zerolog.New(nil).With().Logger()
	}
}

// Zero aims to zero the internal state of the analyzer. So it removes all
// the loaded stats, initialize some variables [and read the config file](NOT ANYMORE).
func Zero() error {
	if IsRunning() {
		analyzerLogger.Error().Msg("Cannot reload, monitoring in progress")
		return errors.New("Cannot reload, monitoring in progress")
	}

	// Reset the miner
	miner.Zero()

	// package variables
	counterID = make(map[string]int)
	statMap = make(map[int]stats.StatInterface)
	statID = 0
	statValues = make(map[string]float64)
	counterValues = make(map[string]uint64)
	defaultEventChannel = make(chan int)
	running = false

	// InitConfig()
	analyzerLogger.Info().Msg("Analyzer package (re)loaded")
	return nil

}

// Config return the configuration of the analyzer/miner
func Config() map[string]interface{} {
	return viper.AllSettings()
}

// DisableLogging disable the log output. Warning! It disables the log
// for all the modules using zerolog
func DisableLogging() {
	analyzerLogger.Info().Msg("Disabling logging")
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
	analyzerLogger.Info().Msgf("Enabling logging (level %s)", l.String())
}

// SetFileLogging (des)activate the data/thresholds/anomaly logging
// into files (files are saved in the outputDir directory)
func SetFileLogging(flog bool) error {
	if running {
		analyzerLogger.Error().Msgf("Cannot change file logging while sniffing")
		return errors.New("Sniffing in progress")
	}
	logDataToFile = flog
	analyzerLogger.Debug().Msgf("File logging set to %t", flog)
	return nil
}

// SetInfluxDBLogging (des)activate the data/thresholds logging
// into influxdb (! anomalies are not logged into influxdb !)
func SetInfluxDBLogging(ilog bool) error {
	if running {
		analyzerLogger.Error().Msgf("Cannot change influxdb logging while sniffing")
		return errors.New("Sniffing in progress")
	}

	// both are true
	if logDataToInfluxDB && ilog {
		return errors.New("InfluxDB logging is already activated")
	}

	// Here ilog != logDataToInfluxDB
	if ilog {
		logDataToInfluxDB = true
		// we have to init (see InitConfig function)
		influxdb.InitConfig()
	} else {
		logDataToInfluxDB = false
		// close the connection
		err := influxdb.Close()
		if err != nil {
			analyzerLogger.Error().Msgf("Error while closing InfluxDB connection (%s)", err)
			return err
		}
	}
	analyzerLogger.Debug().Msgf("InfluxDB logging set to %t", ilog)
	return nil
}

// SetOutputDir change the directory where the raw stats are saved (and thresholds)
func SetOutputDir(dir string) error {
	if running {
		analyzerLogger.Error().Msgf("Cannot change output directory while sniffing")
		return errors.New("Sniffing in progress")
	}

	absPath, err := filepath.Abs(dir)
	if err != nil {
		analyzerLogger.Error().Msgf("Error while changing data log directory (%s)", err)
		return err
	}

	file, err := os.Open(absPath)
	if err != nil {
		analyzerLogger.Error().Msgf("Error while changing data log directory (%s)", err)
		return err
	}

	f, err := file.Stat()
	if err != nil {
		analyzerLogger.Error().Msgf("Error while changing data log directory (%s)", err)
		return err
	}

	// setting the directory
	if f.IsDir() {
		// it's a directory
		outputDir = absPath
	} else {
		// it's not a directory (but a file)
		outputDir = filepath.Dir(absPath)
	}

	analyzerLogger.Debug().Msgf("Output directory set to %s", outputDir)
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

// SetSeriesName sets the name of the series within InfluxDB
func SetSeriesName(s string) {
	if !running {
		seriesName = s
		analyzerLogger.Debug().Msgf(`Series name set to "%s"`, s)
	} else {
		analyzerLogger.Error().Msgf("Cannot change series name while sniffing")
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
	return load(stat)
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
		analyzerLogger.Info().Msg("Stopping stats computation")
	}
}

// StatValues return a current snapshot of the stat values (and their thresholds)
func StatValues() map[string]float64 {
	if running {
		defaultEventChannel <- STAT
		return <-defaultDataChannel
	}
	return nil
}

// Perfs return the current miner performances
// func Perfs() map[string]float64 {
// 	if running {
// 		defaultEventChannel <- PERF
// 		return <-defaultDataChannel
// 	}
// 	return nil
// }

// StartStats starts the analysis
// func StartStats() error {
func StartStats() error {
	if len(GetLoadedStats()) == 0 {
		return errors.New("No stats loaded")
	}
	analyzerLogger.Info().Msg("Starting stats computation")
	analyzerLogger.Debug().Msg(fmt.Sprint("Loaded stats: ", GetLoadedStats()))
	defaultEventChannel, defaultDataChannel = GoRun()
	return nil
}

// StartStatsAndWait starts the analysis. It will stop only when no packets
// have to be processed (ex: pcap file)
func StartStatsAndWait() error {
	if len(GetLoadedStats()) == 0 {
		return errors.New("No stats loaded")
	}
	analyzerLogger.Info().Msg("Starting stats computation")
	analyzerLogger.Debug().Msg(fmt.Sprint("Loaded stats: ", GetLoadedStats()))
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
	} else if res == -1 {
		alog := anomalyLogger.Warn().Time("time", t)
		alog.Str("Status", "DOWN_ALERT").
			Str("Stat", statMap[id].Name()).
			Float64("Value", val).Int("Spot", res).
			Float64("Probability", statMap[id].DSpot().DownProbability(val)).Msg("Alarm!")
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
		if dspot := stat.DSpot(); dspot != nil {
			upTh = dspot.GetUpperThreshold()
			downTh = dspot.GetLowerThreshold()

			// fmt.Println(upTh)
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

	// Debug stat values
	debug := analyzerLogger.Debug()
	for name, value := range statValues {
		debug.Float64(name, value)
	}
	debug.Msg("Stat values")
	// analyzerLogger.Debug().Msgf("%v", statValues)

	// flush
	dlog.Msg("")
	tlog.Msg("")
	// if data have to be sent to InfluxDB
	if logDataToInfluxDB {
		influxdb.PushRecord(statValues, seriesName, curtime)
	}
}

func startRunningInfo() {
	analyzerLogger.Info().Msg("Start running")
	analyzerLogger.Info().Msgf("Raw data are logged to %s", rawDataOutputFile)
	analyzerLogger.Info().Msgf("Thresholds are logged to %s", thresholdOutputFile)
	analyzerLogger.Info().Msgf("Anomalies are logged to %s", anomalyOutputFile)
}

// Run open the device to listen
func Run() {
	// initialize files/influxdb to log data and thresholds
	InitDataLogging()
	// display basic information
	startRunningInfo()
	// number of records
	records := 0
	// set the running state
	running = true
	// sniff
	minerEvent, minerData := miner.GoSniffAndYieldChannel(period)
	// loop
	for {
		// Aimed to debug the channels
		// analyzerLogger.Debug().
		// 	Str("Function", "Run").
		// 	Int("minerEvent", len(minerEvent)).
		// 	Int("minerData", len(minerData)).
		// 	Msg("")

		select {
		case e := <-defaultEventChannel:
			if e == 0 { // stop order
				release()
				// stop the miner
				minerEvent <- miner.STOP
				analyzerLogger.Info().Msg("Stopping stats computation (controller)")
				return
			}
		case m := <-minerData:
			if m != nil {
				// retrieve the counter values
				for name, id := range counterID {
					counterValues[name] = m[id]
					// analyzerLogger.Debug().Msg(fmt.Sprint(counterValues))
				}
				records++
				// analyze the stats values (feed dspot, log data/thresholds)
				analyze()
			} else {
				release()
				analyzerLogger.Warn().Msg("Stopping stats computation (miner)")
				analyzerLogger.Info().Msgf("Number of records: %d", records)
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
		// display basic information
		startRunningInfo()
		// set the running state
		running = true
		// sniff
		minerEvent, minerData := miner.GoSniffAndYieldChannel(period)
		// loop
		for {
			// Aimed to debug the channels
			// analyzerLogger.Debug().
			// 	Str("Function", "GoRun").
			// 	Int("minerEvent", len(minerEvent)).
			// 	Int("minerData", len(minerData)).
			// 	Msg("")

			select {
			case e := <-eventChannel:
				switch e {
				case 0: // stop order
					analyzerLogger.Info().Msg("Receiving STOP message")
					// stop the miner
					minerEvent <- miner.STOP
					// release all
					// release
					running = false
					// close(eventChannel)
					//
					analyzerLogger.Info().Msg("Stopping stats computation (controller)")
					return
				case STAT: // send data
					smux.Lock()
					snapshot := make(map[string]float64)
					for s, v := range statValues {
						snapshot[s] = v
					}
					smux.Unlock()
					dataChannel <- snapshot
					// case PERF: // send PERF signal (perfs are sent to minerData)
					// 	minerEvent <- miner.PERF
				}
			case m := <-minerData:
				if m != nil {
					// retrieve the counter values
					for name, id := range counterID {
						counterValues[name] = m[id]
						// analyzerLogger.Debug().Msg(fmt.Sprint(counterValues))
					}
					// analyze the stats values (feed dspot, log data/thresholds)
					analyze()

				} else {
					// release
					running = false
					analyzerLogger.Warn().Msg("Stopping stats computation (miner)")
					return
				}
			}
		}
	}()
	return eventChannel, dataChannel
}

//------------------------------------------------------------------------------
// MAIN
//------------------------------------------------------------------------------

func main() {
	// nothing
}
